package daemon

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/osProcess"
	"syscall"
	"errors"
	"time"
	"fmt"
)

func (d *daemon) commandFactory(command dto.Command) {
	switch command.Type {
	case constants.COMMAND_STATUS:
		d.sendStatusResult()
	case constants.COMMAND_STOP:
		if err := d.stopProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult()
	case constants.COMMAND_START:
		if err := d.startProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult()
	case constants.COMMAND_RESTART:
		if err := d.restartProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult()
	case constants.COMMAND_KILL:
		if err := d.killProcess(); err != nil {
			d.handleError(err)
		}
	default:
		d.handleError(errors.New(fmt.Sprintf("Команда не найдена %d", int(constants.COMMAND_STATUS))))
		d.sendStatusResult()
	}
}

func (d *daemon) killProcess() error {
	if d.isActive() == true {
		if err := d.process.Kill(); err != nil {
			return err
		}
		d.mu.Lock()
		d.process = nil
		d.statusCode = constants.STATUS_DEAD
		d.status = "Процесс убит командой пользователя"
		d.lastChangeTime = time.Now()
		d.mu.Unlock()
	} else if d.isDead() == false {
		/*	*Добиваю* процесс (не обрабатываю ошибку)  */
		d.process.Kill()
		d.mu.Lock()
		d.process = nil
		d.statusCode = constants.STATUS_DEAD
		d.status = "Процесс убит командой пользователя"
		d.lastChangeTime = time.Now()
		d.mu.Unlock()
	}
	return nil
}

func (d *daemon) stopProcess() error {
	if d.isActive() == true {
		if err := d.process.Signal(d.ProcessMeta.StopSignal); err != nil {
			return err
		}
		d.mu.Lock()
		d.statusCode = constants.STATUS_STOPPED
		d.status = "Процесс приостановлен командой пользователя"
		d.lastChangeTime = time.Now()
		d.mu.Unlock()
	}
	return nil
}

func (d *daemon) startProcess() error {
	/*	Если процесс уже активен -- ничего делать не надо  */
	if d.isActive() == true {
		return nil
	}

	process, err := osProcess.New(d.Name, d.BinPath, d.Args, d.Env, syscall.SIGINT) // SIGINT - сигнал который получит процесс в случае остановки родительского процесса
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.process = process
	d.pid = process.Pid
	d.statusCode = constants.STATUS_ACTIVE
	d.lastChangeTime = time.Now()
	d.exitCode = 0
	d.mu.Unlock()
	return nil
}

func (d *daemon) restartProcess() error {
	if err := d.stopProcess(); err != nil {
		return err
	}
	if err := d.startProcess(); err != nil {
		return err
	}
	return nil
}