package daemon

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/osProcess"
	// "github.com/GlobchanskyDenis/taskmaster.git/pkg/processName"
	"syscall"
	"context"
	"errors"
	"sync"
	"fmt"
	"os"
)

type daemon struct {
	dto.ProcessMeta
	processAsync
	ctx                  context.Context
	receiver             <-chan dto.Command
	sender               chan<- dto.CommandResult
}

// type processMeta struct {
// 	name       string
// 	path       string
// 	args       []string
// 	env        []string
// 	rawName    string
// }

type processAsync struct {
	pid        int
	statusCode uint
	status     string
	lastError  error
	
	process    *os.Process
	wasStopped bool
	mu         *sync.Mutex
}

func new(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, meta dto.ProcessMeta) *daemon { // rawName string, env []string
	return &daemon{
		ctx: ctx,
		receiver: receiver,
		sender: sender,
		ProcessMeta: meta, /*processMeta{
			rawName: rawName,
			env: env,
		},*/
		processAsync: processAsync{
			statusCode: constants.STATUS_DEAD,
			wasStopped: true,
			status: "Процесс пока не был запущен",
			mu: &sync.Mutex{},
		},
	}
}

func RunAsync(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, meta dto.ProcessMeta) { // rawName string, env []string
	d := new(ctx, receiver, sender, meta) // rawName, env
	// if err := d.parseProcessName(); err != nil {
	// 	d.handleError(err)
	// }
	if err := d.startProcess(); err != nil {
		d.handleError(err)
	}

	/*	Блокирующая команда. Заканчивается только по сигналу останова (gracefull shutdown через контекст)  */
	d.listen()
}

// func (d *daemon) parseProcessName() error {
// 	name, path, args, err := processName.Parse(d.rawName)
// 	if err != nil {
// 		return err
// 	}
// 	d.name = name
// 	d.path = path
// 	d.args = args
// 	return nil
// }

func (d *daemon) startProcess() error {
	if d.Name != "" {
		process, err := osProcess.New(d.Name, d.BinPath, d.Args, d.Env, syscall.SIGINT) // SIGINT - сигнал который получит процесс в случае остановки родительского процесса
		if err != nil {
			return err
		}
		d.mu.Lock()
		d.process = process
		d.pid = process.Pid
		d.statusCode = constants.STATUS_ACTIVE
		d.mu.Unlock()
	} else if d.pid != 0 {
		process, err := osProcess.GetByPid(d.pid)
		if err != nil {
			return err
		}
		d.mu.Lock()
		d.process = process
		d.pid = process.Pid
		d.statusCode = constants.STATUS_ACTIVE
		d.mu.Unlock()
	}
	return nil
}

func (d *daemon) listen() {
	fmt.Printf("\tДемон %d %s стартовал\n", d.pid, d.Name)
	go d.handleProcessStop()
	defer fmt.Printf("\tДемон %d %s завершил работу функции listen\n", d.pid, d.Name)
	for {
		select {
		case command := <- d.receiver:
			fmt.Printf("\tДемон %d %s получил команду %#v\n", d.pid, d.Name, command)
			d.commandFactory(command)
			fmt.Printf("\tДемон %d %s отослал команду\n", d.pid, d.Name)
		case <- d.ctx.Done():
			fmt.Printf("\tДемон %d %s получил сигнал завершения программы\n", d.pid, d.Name)
			if err := d.killProcess(); err != nil {
				d.handleError(err)
			}
			fmt.Printf("\tДемон %d %s завершил работу\n", d.pid, d.Name)
			return
		}
	}
}

func (d *daemon) commandFactory(command dto.Command) {
	switch command.Type {
	case constants.COMMAND_STATUS:
		d.sendStatusResult(command)
	case constants.COMMAND_STOP:
		d.handleError(errors.New("Команда останова демона не готова"))
	case constants.COMMAND_START:
		d.handleError(errors.New("Команда старта демона не готова"))
	case constants.COMMAND_RESTART:
		d.handleError(errors.New("Команда рестарта демона не готова"))
	case constants.COMMAND_KILL:
		if err := d.killProcess(); err != nil {
			d.handleError(err)
		}
	default:
		d.handleError(errors.New(fmt.Sprintf("Команда не найдена %d", int(constants.COMMAND_STATUS))))
	}
}

/*	Все данные подготовлены ДО данного обращения  */
func (d *daemon) sendStatusResult(command dto.Command) {
	println("\tstatus command handling")
	defer println("\tstatus command was sent")
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sender <- dto.CommandResult{
		Pid: d.pid,
		Name: d.Name,
		Status: d.status,
		StatusCode: d.statusCode,
		Error: d.lastError,
	}
}

func (d *daemon) killProcess() error {
	if d.process != nil && d.wasStopped == false {
		if err := d.process.Kill(); err != nil {
			return err
		}
		d.process = nil
	}
	return nil
}

func (d *daemon) handleError(err error) {
	fmt.Printf("\tДемон %d %s получил ошибку %s\n", d.pid, d.Name, err)
	d.mu.Lock()
	d.statusCode = constants.STATUS_ERROR
	d.lastError = err
	d.status = "Произошла ошибка"
	d.mu.Unlock()
}

func (d *daemon) handleProcessStop() {
	/*	Процесс неактивен, команда ожидания блокирует процесс до сигнала останова, поэтому в случае ошибки в моем коде тут может зависать  */
	exitState, err := d.process.Wait()
	if err != nil {
		d.handleError(err)
	} else {
		d.pid = exitState.Pid()
		fmt.Printf("\tPid %d \tExited? %#v \tExitCode %d \tString %s\n", exitState.Pid(), exitState.Exited(), exitState.ExitCode(), exitState.String())
	}
	
	// d.sender <- dto.CommandResult{
	// 	Pid: d.pid,
	// 	Status: exitState.String(),
	// 	StatusCode: constants.STATUS_STOPPED,
	// }
}