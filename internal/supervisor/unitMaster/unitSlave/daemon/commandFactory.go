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
		// d.logInfo("Получена команда status")
		d.sendStatusResult(command.AmountLogs)
	case constants.COMMAND_STOP:
		// d.logInfo("Получена команда stop")
		if err := d.stopProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult(0)
	case constants.COMMAND_START:
		// d.logInfo("Получена команда start")
		if err := d.startProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult(0)
	case constants.COMMAND_RESTART:
		// d.logInfo("Получена команда restart")
		if err := d.restartProcess(); err != nil {
			d.handleError(err)
		}
		d.sendStatusResult(0)
	case constants.COMMAND_KILL:
		// d.logInfo("Получена команда kill")
		if err := d.killProcess(); err != nil {
			d.handleError(err)
		}
	default:
		message := fmt.Sprintf("Команда не найдена %d", int(constants.COMMAND_STATUS))
		// d.logWarning(nil, message)
		d.handleError(errors.New(message))
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
	if d.isDead() == false && d.isNotStarted() == false {
		/*	Сначала надо добить старый процесс  */
		if err := d.killProcess(); err != nil {
			return err
		}
		if err := d.newProcess(); err != nil {
			return err
		}
		return nil
	}
	if d.isNotStarted() == true {
		if err := d.newProcess(); err != nil {
			return err
		}
		return nil
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