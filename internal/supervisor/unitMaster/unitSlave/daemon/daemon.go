package daemon

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	// "github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/osProcess"
	// "syscall"
	"context"
	"time"
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

type processAsync struct {
	pid            int
	statusCode     uint
	status         string
	lastError      error
	exitCode       int
	lastChangeTime time.Time
	
	process    *os.Process
	mu         *sync.Mutex
}

func new(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, meta dto.ProcessMeta) *daemon {
	return &daemon{
		ctx: ctx,
		receiver: receiver,
		sender: sender,
		ProcessMeta: meta,
		processAsync: processAsync{
			statusCode: constants.STATUS_DEAD,
			status: "Процесс пока не был запущен",
			mu: &sync.Mutex{},
		},
	}
}

func RunAsync(ctx context.Context, receiver <-chan dto.Command, sender chan<- dto.CommandResult, meta dto.ProcessMeta) {
	d := new(ctx, receiver, sender, meta)
	if err := d.createProcess(); err != nil {
		println("ERROR!!!")
		println(err.Error())
		d.handleError(err)
	} else {
		/*	Блокирующая команда. Заканчивается только по сигналу останова (gracefull shutdown через контекст)  */
		d.listen()
	}
}

func (d *daemon) createProcess() error {
	return d.startProcess()
	// if d.Name != "" {
	// 	process, err := osProcess.New(d.Name, d.BinPath, d.Args, d.Env, syscall.SIGINT) // SIGINT - сигнал который получит процесс в случае остановки родительского процесса
	// 	if err != nil {
	// 		return err
	// 	}
	// 	d.mu.Lock()
	// 	d.process = process
	// 	d.pid = process.Pid
	// 	d.statusCode = constants.STATUS_ACTIVE
	// 	d.lastChangeTime = time.Now()
	// 	d.mu.Unlock()
	// } else if d.pid != 0 {
	// 	process, err := osProcess.GetByPid(d.pid)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	d.mu.Lock()
	// 	d.process = process
	// 	d.pid = process.Pid
	// 	d.statusCode = constants.STATUS_ACTIVE
	// 	d.lastChangeTime = time.Now()
	// 	d.mu.Unlock()
	// }
	// return nil
}

func (d *daemon) listen() {
	// fmt.Printf("\tДемон %d %s стартовал\n", d.pid, d.Name)
	go d.handleProcessInterrupt()
	// defer fmt.Printf("\tДемон %d %s завершил работу функции listen\n", d.pid, d.Name)
	for {
		select {
		case command := <- d.receiver:
			// fmt.Printf("\tДемон %d %s получил команду %#v\n", d.pid, d.Name, command)
			d.commandFactory(command)
			// fmt.Printf("\tДемон %d %s отослал команду\n", d.pid, d.Name)
		case <- d.ctx.Done():
			// fmt.Printf("\tДемон %d %s получил сигнал завершения программы\n", d.pid, d.Name)
			if err := d.killProcess(); err != nil {
				d.handleError(err)
			}
			// fmt.Printf("\tДемон %d %s завершил работу\n", d.pid, d.Name)
			return
		}
	}
}

/*	Все данные подготовлены ДО данного обращения  */
func (d *daemon) sendStatusResult() {
	// println("\tstatus command handling")
	// defer println("\tstatus command was sent")
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
	}
}

func (d *daemon) handleError(err error) {
	// fmt.Printf("\tДемон %d %s получил ошибку %s\n", d.pid, d.Name, err)
	d.mu.Lock()
	d.statusCode = constants.STATUS_ERROR
	d.lastError = err
	d.status = "Произошла ошибка"
	d.lastChangeTime = time.Now()
	d.mu.Unlock()
}

func (d *daemon) handleProcessInterrupt() {
	/*	Процесс неактивен, команда ожидания блокирует процесс до сигнала останова, поэтому в случае ошибки в моем коде тут может зависать  */
	exitState, err := d.process.Wait()
	if err != nil {
		d.handleError(err)
	} else {
		fmt.Printf("\tPid %d \tExited? %#v \tExitCode %d \tString %s\n", exitState.Pid(), exitState.Exited(), exitState.ExitCode(), exitState.String())
		d.mu.Lock()
		d.statusCode = constants.STATUS_DEAD
		d.status = "Процесс убит извне " + exitState.String()
		d.lastChangeTime = time.Now()
		d.exitCode = exitState.ExitCode()
		d.mu.Unlock()
	}
}

func (d *daemon) isActive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.process != nil && d.statusCode == constants.STATUS_ACTIVE {
		return true
	}
	return false
}

func (d *daemon) isDead() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.process == nil && d.statusCode == constants.STATUS_DEAD {
		return true
	}
	return false
}

func (d *daemon) isStopped() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.process != nil && d.statusCode == constants.STATUS_STOPPED {
		return true
	}
	return false
}
