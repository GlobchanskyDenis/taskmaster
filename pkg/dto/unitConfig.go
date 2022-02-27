package dto

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/parser"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"syscall"
	"strconv"
	"errors"
	"fmt"
)

type UnitConfig struct {
	Cmd				string   `conf:"Cmd"`            // Команда, которую мы запускаем. Отсюда же мы берем и имя процесса
	Args			[]string `conf:"Args"`		     // Дополнительные аргументы для запуска. Можно указывать аргументы и напрямую в поле команды
	Env				[]string `conf:"Env"`            // Переменные окружения для данного процесса
	Replicas        uint     `conf:"Replicas"`       // Количество процессов которые необходимо запустить
	Autostart       bool     `conf:"Autostart"`      // Указывается необходимость автоматического старта при запуске программы
	AutoRestart     string   `conf:"Autorestart"`    // Это поле нужно парсить. Варианты Always, Never, ...
	Starttime       uint     `conf:"Starttime"`      // Время для запуска процесса. После этого программа решает - успешно ли запущен процесс
	Stoptime        uint     `conf:"Stoptime"`       // Время после отправки сигнала quitsit до отправки сигнала SIGKILL. Также эти значения участвуют в реализации механизма Gracefull Shutdown
	Restartretries  uint     `conf:"Restartretries"` // Количество попыток запуска программы
	Signal          string   `conf:"Signal"`         // Каким сигналом останавливать процесс. Это поле нужно парсить. Есть три варианта сигнала останова SIGTERM, SIGINT, SIGQUIT
	Exitcodes       []int    `conf:"Exitcodes"`      // Список кодов завершения программы, после которых можно делать рестарт программы
	Umask           int      `conf:"Umask"`          // Маска прав доступа 0 - нет доп ограничений 7 - запретить все права
	Workingdir      *string	 `conf:"Workingdir"`     // установка каталога для процесса (относится к chroot)

	ProcessName     string         `conf:"-"`
	ProcessArgs     []string       `conf:"-"`
	BinPath         string         `conf:"-"`
	signal          syscall.Signal `conf:"-"`
	autorestart     bool           `conf:"-"`
	RestartTimes    *uint          `conf:"-"`
}

func (u UnitConfig) Validate() error {
	if u.AutoRestart != constants.AUTORESTART_ALWAYS && u.AutoRestart != constants.AUTORESTART_NEVER && u.AutoRestart != constants.AUTORESTART_LIMITED_TIMES && u.AutoRestart != constants.AUTORESTART_UNEXPECTED_EXITS {
		return errors.New("В конфигурации в пункте AutoRestart присвоено недопустимое значение (" + u.AutoRestart + "). Допустимые значения: " + 
		constants.AUTORESTART_ALWAYS + ", " + constants.AUTORESTART_NEVER + ", " + constants.AUTORESTART_LIMITED_TIMES + ", " + constants.AUTORESTART_UNEXPECTED_EXITS)
	}
	if u.Signal != constants.SIGNAL_SIGTERM && u.Signal != constants.SIGNAL_SIGINT && u.Signal != constants.SIGNAL_SIGQUIT {
		return errors.New("В конфигурации в пункте Signal присвоено недопустимое значение (" + u.Signal + "). Допустимые значения: " + 
		constants.SIGNAL_SIGTERM  + ", " + constants.SIGNAL_SIGINT + ", " + constants.SIGNAL_SIGQUIT)
	}

	if u.AutoRestart == constants.AUTORESTART_ALWAYS && u.Restartretries != 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Restartretries присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Restartretries=" +
			strconv.FormatUint(uint64(u.Restartretries), 10) + "). Restartretries должен быть нулем если вы хотите перманентный авторестарт")
	}
	if u.AutoRestart == constants.AUTORESTART_ALWAYS && len(u.Exitcodes) != 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Exitcodes присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Exitcodes=" +
		fmt.Sprintf("%#v", u.Exitcodes) + "). Exitcodes не должны указываться, т.к авторестарт будет в любом случае")
	}
	if u.AutoRestart == constants.AUTORESTART_NEVER && u.Restartretries != 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Restartretries присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Restartretries=" +
			strconv.FormatUint(uint64(u.Restartretries), 10) + "). Restartretries должен быть нулем если вы не хотите авторестарта")
	}
	if u.AutoRestart == constants.AUTORESTART_NEVER && len(u.Exitcodes) != 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Exitcodes присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Exitcodes=" +
		fmt.Sprintf("%#v", u.Exitcodes) + "). Exitcodes не должны указываться, т.к авторестарта не будет")
	}
	if u.AutoRestart == constants.AUTORESTART_LIMITED_TIMES && u.Restartretries == 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Restartretries присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Restartretries=0).")
	}
	if u.AutoRestart == constants.AUTORESTART_LIMITED_TIMES && len(u.Exitcodes) != 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Exitcodes присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Exitcodes=" +
		fmt.Sprintf("%#v", u.Exitcodes) + "). Exitcodes не должны указываться, т.к данный вариант авторестарта предполагает рестарт вне зависимости от кодов завершения процесса")
	}
	if u.AutoRestart == constants.AUTORESTART_UNEXPECTED_EXITS && u.Restartretries != 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Restartretries присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Restartretries=" +
			strconv.FormatUint(uint64(u.Restartretries), 10) + "). Restartretries должен быть нулем если вы не хотите авторестарта")
	}
	if u.AutoRestart == constants.AUTORESTART_UNEXPECTED_EXITS && len(u.Exitcodes) == 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Exitcodes присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Exitcodes=[]). Exitcodes должны указываться")
	}
	if u.Replicas == 0 {
		return errors.New("В конфигурации в пункте Replicas присвоено недопустимое значение (0). Минимум 1")
	}
	if u.Replicas > 20 {
		return errors.New("В конфигурации в пункте Replicas присвоено недопустимое значение (" + strconv.FormatUint(uint64(u.Replicas), 10) + "). Максимум 20. Совсем совесть потерял?")
	}
	return nil
}

func (u *UnitConfig) Parse() error {
	if err := u.parseName(); err != nil {
		return err
	}
	u.parseAutorestart()
	u.parseSignal()
	return nil
}

func (u *UnitConfig) parseName() error {
	name, path, args, err := parser.ParseProcessName(u.Cmd)
	if err != nil {
		return err
	}
	u.ProcessName = name
	u.BinPath = path
	u.ProcessArgs = append([]string{}, args...)
	u.ProcessArgs = append(u.ProcessArgs, u.Args...)
	return nil
}

func (u *UnitConfig) parseAutorestart() {
	switch u.AutoRestart {
	case constants.AUTORESTART_ALWAYS:
		u.autorestart = true
		u.RestartTimes = nil
	case constants.AUTORESTART_LIMITED_TIMES:
		u.autorestart = true
		u.RestartTimes = &u.Restartretries
	case constants.AUTORESTART_NEVER:
		u.autorestart = false
		u.RestartTimes = nil
	case constants.AUTORESTART_UNEXPECTED_EXITS:
		u.autorestart = true
		u.RestartTimes = nil
	}
}

func (u *UnitConfig) parseSignal() {
	switch u.Signal {
	case constants.SIGNAL_SIGTERM:
		u.signal = syscall.SIGTERM
	case constants.SIGNAL_SIGINT:
		u.signal = syscall.SIGINT
	case constants.SIGNAL_SIGQUIT:
		u.signal = syscall.SIGQUIT
	}
}

func (u UnitConfig) copy() UnitConfig {
	return UnitConfig{
		Cmd: u.Cmd,
		Args: u.Args,
		Env: u.Env,
		Replicas: u.Replicas,
		Autostart: u.Autostart,
		AutoRestart: u.AutoRestart,
		Starttime: u.Starttime,
		Stoptime: u.Stoptime,
		Restartretries: u.Restartretries,
		Signal: u.Signal,
		Exitcodes: u.Exitcodes,
		Umask: u.Umask,
		Workingdir: u.Workingdir,
		ProcessName: u.ProcessName,
		ProcessArgs: u.ProcessArgs,
		BinPath: u.BinPath,
		signal: u.signal,
		autorestart: u.autorestart,
		RestartTimes: u.RestartTimes,
	}
}

func (u UnitConfig) GetProcessMeta() ProcessMeta {
	return ProcessMeta{
		Name: u.ProcessName,
		BinPath: u.BinPath,
		Args: u.ProcessArgs,
		Env: u.Env,
		ProcessPath: u.Workingdir,
		Autostart: u.Autostart,
		Autorestart: u.autorestart,
		RestartTimes: u.RestartTimes,
		StopSignal: u.signal,
		Exitcodes: u.Exitcodes,
		Starttime: u.Starttime,
		Stoptime: u.Stoptime,
		Umask: u.Umask,
	}
}
