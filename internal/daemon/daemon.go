package daemon

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/osProcess"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/processName"
	"context"
	"errors"
	"sync"
	"fmt"
	"os"
)

type daemon struct {
	processMeta
	processAsync
	ctx                  context.Context
	receiver             <-chan dto.Command
	sender               chan<- dto.CommandResult
}

type processMeta struct {
	name       string
	path       string
	args       []string
	env        []string
	rawName    string
}

type processAsync struct {
	pid        int
	statusCode uint
	status     string
	lastError  error
	
	process    *os.Process
	wasStopped bool
	mu         *sync.Mutex
}

func new(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, rawName string, env []string) *daemon {
	return &daemon{
		ctx: ctx,
		receiver: receiver,
		sender: sender,
		processMeta: processMeta{
			rawName: rawName,
			env: env,
		},
		processAsync: processAsync{
			statusCode: constants.STATUS_DEAD,
			wasStopped: true,
			status: "Процесс пока не был запущен",
			mu: &sync.Mutex{},
		},
	}
}

func RunAsync(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, rawName string, env []string) {
	d := new(ctx, receiver, sender, rawName, env)
	if err := d.parseProcessName(); err != nil {
		d.handleError(err)
	}
	if err := d.startProcess(); err != nil {
		d.handleError(err)
	}

	/*	Блокирующая команда. Заканчивается только по сигналу останова (gracefull shutdown через контекст)  */
	d.listen()
}

func (d *daemon) parseProcessName() error {
	name, path, args, err := processName.Parse(d.rawName)
	if err != nil {
		return err
	}
	d.name = name
	d.path = path
	d.args = args
	return nil
}

func (d *daemon) startProcess() error {
	if d.name != "" {
		process, err := osProcess.New(d.name, d.path, d.args, d.env)
		if err != nil {
			return err
		}
		d.process = process
	} else if d.pid != 0 {
		process, err := osProcess.GetByPid(d.pid)
		if err != nil {
			return err
		}
		d.process = process
	}
	return nil
}

func (d *daemon) listen() {
	fmt.Printf("\tДемон %d %s стартовал\n", d.pid, d.name)
	go d.handleProcessStop()
	select {
	case command := <- d.receiver:
		fmt.Printf("\tДемон %d %s получил команду %#v\n", d.pid, d.name, command)
		d.commandFactory(command)
		fmt.Printf("\tДемон %d %s отослал команду\n", d.pid, d.name)
	case <- d.ctx.Done():
		fmt.Printf("\tДемон %d %s получил сигнал завершения программы\n", d.pid, d.name)
		if err := d.killProcess(); err != nil {
			d.handleError(err)
		}
		fmt.Printf("\tДемон %d %s завершил работу\n", d.pid, d.name)
		return
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
		Status: d.status,
		StatusCode: d.statusCode,
		Error: d.lastError,
	}
}

func (d *daemon) killProcess() error {
	if d.process != nil {
		if err := d.process.Kill(); err != nil {
			return err
		}
		d.process = nil
	}
	return nil
}

func (d *daemon) handleError(err error) {
	fmt.Printf("\tДемон %d %s получил ошибку %s\n", d.pid, d.name, err)
	d.sender <- dto.CommandResult{
		Error: err,
		StatusCode: constants.STATUS_ERROR,
		Status: "Произошла ошибка",
	}
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