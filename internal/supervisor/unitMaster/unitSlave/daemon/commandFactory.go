package daemon

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"errors"
	"time"
	"fmt"
)

func (d *daemon) commandFactory(command dto.Command) {
	switch command.Type {
	case constants.COMMAND_STATUS:
		d.sendStatusResult(command.AmountLogs)
	case constants.COMMAND_STOP:
		if err := d.stopProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult(0)
	case constants.COMMAND_START:
		if err := d.startProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult(0)
	case constants.COMMAND_RESTART:
		if err := d.restartProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult(0)
	case constants.COMMAND_KILL:
		if err := d.killProcess(); err != nil {
			d.handleError(err)
		}
	default:
		d.handleError(errors.New(fmt.Sprintf("Команда не найдена %d", int(constants.COMMAND_STATUS))))
		d.sendStatusResult(0)
	}
}

func (d *daemon) killProcess() error {
	if d.isActive() == true {
		if err := d.cmd.Process.Kill(); err != nil {
			return err
		}
		d.mu.Lock()
		d.cmd = nil
		d.statusCode = constants.STATUS_DEAD
		d.status = "Процесс убит командой пользователя"
		d.lastChangeTime = time.Now()
		d.mu.Unlock()
	} else if d.isDead() == false {
		/*	*Добиваю* процесс (не обрабатываю ошибку)  */
		d.cmd.Process.Kill()
		d.mu.Lock()
		d.cmd = nil
		d.statusCode = constants.STATUS_DEAD
		d.status = "Процесс убит командой пользователя"
		d.lastChangeTime = time.Now()
		d.mu.Unlock()
	}
	return nil
}

func (d *daemon) stopProcess() error {
	if d.isActive() == true {
		if err := d.cmd.Process.Signal(d.ProcessMeta.StopSignal); err != nil {
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
	if d.isDead() == false {
		/*	Сначала надо добить старый процесс  */
		if err := d.killProcess(); err != nil {
			return err
		}
		if err := d.newProcess(); err != nil {
			return err
		}
	}
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