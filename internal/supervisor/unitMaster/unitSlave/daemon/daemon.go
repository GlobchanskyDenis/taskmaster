package daemon

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/process"
	"bufio"
	"context"
	"time"
	"sync"
	"fmt"
	"os/exec"
	"syscall"
	"io"
	"errors"
)

type daemon struct {
	dto.ProcessMeta
	processAsync
	ctx      context.Context
	receiver <-chan dto.Command
	sender   chan<- dto.CommandResult
	logger   dto.ILogger
}

type processAsync struct {
	pid            int
	statusCode     uint
	status         string
	lastError      error
	exitCode       int
	lastChangeTime time.Time
	startTime      time.Time
	stdout         io.ReadCloser
	stderr         io.ReadCloser
	logs           []string

	restartedTimes uint
	
	cmd *exec.Cmd
	mu  *sync.Mutex
}

func new(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, meta dto.ProcessMeta, logger dto.ILogger) *daemon {
	return &daemon{
		ctx: ctx,
		receiver: receiver,
		sender: sender,
		logger: logger,
		ProcessMeta: meta,
		processAsync: processAsync{
			statusCode: constants.STATUS_NOT_STARTED,
			status: "Процесс пока не был запущен",
			mu: &sync.Mutex{},
		},
	}
}

func RunAsync(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, meta dto.ProcessMeta, logger dto.ILogger) {
	d := new(ctx, receiver, sender, meta, logger)
	if meta.Autostart == true {
		if err := d.newProcess(); err != nil {
			println("ERROR!!!")
			println(err.Error())
			d.handleError(err)
		}
	}

	go d.listen()
}

func (d *daemon) newProcess() error {
	/*	processPath -- либо текущая папка либо (если задано) из конфигурационника  */
	var processPath string = "./"
	if d.ProcessPath != nil {
		processPath = *d.ProcessPath
	}

	/*	Логика с Umask  */
	syscall.Umask(d.Umask)

	/*	Создаем процесс  */
	cmd, stdout, stderr, err := process.New(d.BinPath + d.Name, processPath, d.Args, d.Env)
	if err != nil {
		return err
	}

	/*	Сохраняем горутино безопасно  */
	d.mu.Lock()
	d.cmd = cmd
	d.pid = cmd.Process.Pid
	d.statusCode = constants.STATUS_ACTIVE
	d.status = "Процесс успешно стартовал"
	d.lastChangeTime = time.Now()
	d.startTime = time.Now()
	d.stdout = stdout
	d.stderr = stderr
	d.logs = nil
	d.mu.Unlock()

	go d.listenStdout()
	go d.listenStderr()
	go d.handleAutorestart()

	return nil
}

func (d *daemon) logError(err error, message string) {
	d.logger.LogError(map[string]interface{}{
		"entity": "process",
		"cmd": d.Name,
	}, err, message)
}

func (d *daemon) logWarning(err error, message string) {
	d.logger.LogWarning(map[string]interface{}{
		"entity": "process",
		"cmd": d.Name,
	}, err, message)
}

func (d *daemon) logInfo(message string) {
	d.logger.LogInfo(map[string]interface{}{
		"entity": "process",
		"cmd": d.Name,
	}, message)
}

/*	Блокирующая функция - делает столько авторестартов, сколько задано конфигурационником и только тогда функция завершается  */
func (d *daemon) handleAutorestart() {
	for {
		/*	Функция прервется когда процесс завершится (по любой причине)  */
		d.handleProcessInterrupt()

		/*	Проверяю разрешен ли рестарт  */
		if d.isRestartPermitted() == false {
			return
		}

		/*	Принудительно стартую. Если ошибка - пофиг, значит плюс еще одна итерация к авторестарту  */
		_ = d.newProcess()
	}
}

/*	Функция отвечает на вопрос - делать ли рестарт или нет  */
func (d *daemon) isRestartPermitted() bool {
	/*	В случае если процесс был остановлен вручную - авторестарт делать не нужно  */
	if d.isStopped() == true {
		return false
	}
	/*	В случае если авторестарт выключен в конфигурационнике -- аналогично  */
	if d.Autorestart == false {
		return false
	}
	if d.IsRestartByRestartTimes() == true {
		/*	Если превысил максимальное количество рестартов  */
		if d.RestartTimes != nil && *d.RestartTimes <= d.restartedTimes {
			return false
		} else {
			d.restartedTimes++
		}
	}
	if d.IsRestartByExitCodeCheck() == true {
		/*	Если код завершения не занесен в конфигурационник как разрешенный для авторестарта - авторестарт не состоится  */
		if d.isExitCodePermittedForRestart() == false {
			return false
		}
	}
	return true
}

/*	Авторестарт разрешен только если код завершения процесса (хрунится внутри) разрешен в конфигурационнике  */
func (d *daemon) isExitCodePermittedForRestart() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, exitCode := range d.Exitcodes {
		if d.exitCode == exitCode {
			return true
		}
	}
	return false
}

func (d *daemon) listen() {
	for {
		select {
		case command := <- d.receiver:
			d.commandFactory(command)
		case <- d.ctx.Done():
			if err := d.killProcess(); err != nil {
				d.handleError(err)
			}
			return
		}
	}
}

func (d *daemon) listenStdout() {
	scanner := bufio.NewScanner(d.stdout)
	/*	Выход из цикла только при получении EOF (которое получаем при завершении процесса)  */
	for scanner.Scan() {
		/*	Тут мы делаем все что нужно для обработки потока вывода процесса (в данной реализации это логгирование в файл и
		**	сохранение логов в самой горутине для команды status)  */
		newLogLine := scanner.Bytes()
		/*	Добавляю лог в слайс для быстрой отдачи при команде status */
		d.addLog(string(newLogLine))
		/*	Логгирую в файл  */
		d.logInfo(string(newLogLine))
	}
	if err := scanner.Err(); err != nil {
		d.handleError(err)
	}
}

func (d *daemon) listenStderr() {
	scanner := bufio.NewScanner(d.stderr)
	/*	Выход из цикла только при получении EOF (которое получаем при завершении процесса)  */
	for scanner.Scan() {
		/*	Тут мы делаем все что нужно для обработки потока вывода процесса (в данной реализации это логгирование в файл и
		**	сохранение логов в самой горутине для команды status)  */
		newLogLine := scanner.Bytes()
		/*	Добавляю лог в слайс для быстрой отдачи при команде status */
		d.addLog(string(newLogLine))
		/*	Логгирую в файл  */
		d.logWarning(nil, string(newLogLine))
	}
	if err := scanner.Err(); err != nil {
		d.handleError(err)
	}
}

func (d *daemon) addLog(line string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logs = append(d.logs, time.Now().Format("2006-01-02 15:04:05") + " " + line)
}

/*	Все данные подготовлены ДО данного обращения  */
func (d *daemon) sendStatusResult(amountLogs uint) {
	logs := d.getLogsMu(amountLogs)
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sender <- dto.CommandResult{
		Pid: d.pid,
		Name: d.Name,
		Status: d.status,
		StatusCode: d.statusCode,
		Error: d.lastError,
		ExitCode: d.exitCode,
		ChangeTime: d.lastChangeTime,
		Logs: logs,
	}
}

func (d *daemon) getLogsMu(amountLogs uint) []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.logs) == 0 {
		return []string{}
	}
	if len(d.logs) <= int(amountLogs) {
		return d.logs
	}
	return d.logs[len(d.logs) - int(amountLogs):]
}

func (d *daemon) handleError(err error) {
	d.mu.Lock()
	d.statusCode = constants.STATUS_ERROR
	d.lastError = err
	d.status = "Произошла ошибка"
	d.lastChangeTime = time.Now()
	d.mu.Unlock()
	/*	Логгирую в файл  */
	d.logError(err, "")
}

func (d *daemon) handleProcessInterrupt() {
	/*	Процесс неактивен, команда ожидания блокирует процесс до сигнала останова, поэтому в случае ошибки в моем коде тут может зависать  */
	if d.cmd != nil {
		exitState, err := d.cmd.Process.Wait()
		if err != nil {
			d.handleError(err)
		} else {
			fmt.Printf("\tPid %d \tExited? %#v \tExitCode %d \tString %s\n", exitState.Pid(), exitState.Exited(), exitState.ExitCode(), exitState.String())
			/*	Проверяем логику из конфигурационника -- отработал ли процесс минимально положенное время  */
			if uint(time.Now().Sub(d.startTime).Seconds()) < d.Starttime {
				d.handleError(errors.New("Процесс завершился раньше ожидаемого времени"))
			} else {
				d.mu.Lock()
				if d.statusCode != constants.STATUS_STOPPED {
					d.statusCode = constants.STATUS_DEAD
					d.status = "Процесс убит извне " + exitState.String()
					d.lastChangeTime = time.Now()
				}
				d.exitCode = exitState.ExitCode()
				d.mu.Unlock()
				/*	Логгирую в файл  */
				d.logWarning(nil, "Процесс остановился " + exitState.String())
			}
		}
	} else {
		fmt.Printf("Не могу отслеживать прерывание процесса так как он не создан (равен nil)")
	}
}

func (d *daemon) isActive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cmd != nil && d.statusCode == constants.STATUS_ACTIVE {
		return true
	}
	return false
}

func (d *daemon) isDead() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cmd == nil && d.statusCode == constants.STATUS_DEAD {
		return true
	}
	return false
}

func (d *daemon) isStopped() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cmd != nil && d.statusCode == constants.STATUS_STOPPED {
		return true
	}
	return false
}

func (d *daemon) isNotStarted() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.statusCode == constants.STATUS_NOT_STARTED {
		return true
	}
	return false
}
