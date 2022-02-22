package daemon

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/process"
	"context"
	"time"
	"sync"
	"fmt"
	"os/exec"
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
	
	cmd *exec.Cmd
	mu  *sync.Mutex
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
	if err := d.newProcess(); err != nil {
		println("ERROR!!!")
		println(err.Error())
		d.handleError(err)
	} else {
		/*	Блокирующая команда. Заканчивается только по сигналу останова (gracefull shutdown через контекст)  */
		d.listen()
	}
}

func (d *daemon) newProcess() error {
	cmd, err := process.New(d.BinPath + d.Name, d.BinPath, d.Args, d.Env)
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.cmd = cmd
	d.pid = cmd.Process.Pid
	d.statusCode = constants.STATUS_ACTIVE
	d.status = "Процесс успешно стартовал"
	d.lastChangeTime = time.Now()
	d.mu.Unlock()
	go d.handleProcessInterrupt()
	return nil
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

/*	Все данные подготовлены ДО данного обращения  */
func (d *daemon) sendStatusResult() {
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
	d.mu.Lock()
	d.statusCode = constants.STATUS_ERROR
	d.lastError = err
	d.status = "Произошла ошибка"
	d.lastChangeTime = time.Now()
	d.mu.Unlock()
}

func (d *daemon) handleProcessInterrupt() {
	/*	Процесс неактивен, команда ожидания блокирует процесс до сигнала останова, поэтому в случае ошибки в моем коде тут может зависать  */
	if d.cmd != nil {
		exitState, err := d.cmd.Process.Wait()
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
