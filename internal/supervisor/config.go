package supervisor

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/processName"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"syscall"
	"strconv"
	"errors"
)

const (
	AUTORESTART_ALWAYS = "Always"
	AUTORESTART_NEVER = "Never"
	AUTORESTART_LIMITED_TIMES = "Limited"
	SIGNAL_SIGTERM = "SIGTERM"
	SIGNAL_SIGINT = "SIGINT"
	SIGNAL_SIGQUIT = "SIGQUIT"
)

type UnitListConfig []*UnitConfig

func (u UnitListConfig) GetMaxStopTime() uint {
	var maxStopTime uint
	for _, conf := range u {
		if conf.Stoptime > maxStopTime {
			maxStopTime = conf.Stoptime
		}
	}
	return maxStopTime
}

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
	Umask           string   `conf:"Umask"`          // ???
	Stdout          *string	 `conf:"Stdout"`         // Файл для перенаправления вместо стандартного потока вывода. null если не нужно. Это поле нужно парсить
	Stderr          *string	 `conf:"Stderr"`         // Файл для перенаправления вместо стандартного потока ошибок. null если не нужно. Это поле нужно парсить
	Workingdir      *string	 `conf:"Workingdir"`     // установка каталога для процесса (относится к chroot)

	processName     string         `conf:"-"`
	processArgs     []string       `conf:"-"`
	binPath         string         `conf:"-"`
	signal          syscall.Signal `conf:"-"`
	autorestart     bool           `conf:"-"`
	restartTimes    *uint          `conf:"-"`
}

func (u UnitConfig) validate() error {
	if u.AutoRestart != AUTORESTART_ALWAYS && u.AutoRestart != AUTORESTART_NEVER && u.AutoRestart != AUTORESTART_LIMITED_TIMES {
		return errors.New("В конфигурации в пункте AutoRestart присвоено недопустимое значение (" + u.AutoRestart + "). Допустимые значения: " + 
			AUTORESTART_ALWAYS + ", " + AUTORESTART_NEVER + ", " + AUTORESTART_LIMITED_TIMES)
	}
	if u.Signal != SIGNAL_SIGTERM && u.Signal != SIGNAL_SIGINT && u.Signal != SIGNAL_SIGQUIT {
		return errors.New("В конфигурации в пункте Signal присвоено недопустимое значение (" + u.Signal + "). Допустимые значения: " + 
			SIGNAL_SIGTERM  + ", " + SIGNAL_SIGINT + ", " + SIGNAL_SIGQUIT)
	}
	if u.AutoRestart == AUTORESTART_LIMITED_TIMES && u.Restartretries == 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Restartretries присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Restartretries=0).")
	}
	if u.AutoRestart == AUTORESTART_NEVER && u.Restartretries != 0 {
		return errors.New("В конфигурации в пунктах AutoRestart, Restartretries присвоено недопустимое сочетание значений (" + u.AutoRestart + " + Restartretries=" +
			strconv.FormatUint(uint64(u.Restartretries), 10) + "). Restartretries должен быть нулем если вы не хотите авторестарта")
	}
	if u.Replicas == 0 {
		return errors.New("В конфигурации в пункте Replicas присвоено недопустимое значение (0). Минимум 1")
	}
	if u.Replicas > 100 {
		return errors.New("В конфигурации в пункте Replicas присвоено недопустимое значение (" + strconv.FormatUint(uint64(u.Replicas), 10) + "). Максимум 100. Совсем совесть потерял?")
	}
	if u.Stdout != nil && *u.Stdout == "" {
		return errors.New("В конфигурации в пункте Stdout присвоено недопустимое значение (пустая строка)")
	}
	if u.Stderr != nil && *u.Stderr == "" {
		return errors.New("В конфигурации в пункте Stderr присвоено недопустимое значение (пустая строка)")
	}
	return nil
}

func (u *UnitConfig) parse() error {
	if err := u.parseName(); err != nil {
		return err
	}
	u.parseAutorestart()
	u.parseSignal()
	return nil
}

func (u *UnitConfig) parseName() error {
	name, path, args, err := processName.Parse(u.Cmd)
	if err != nil {
		return err
	}
	u.processName = name
	u.binPath = path
	u.processArgs = append([]string{}, args...)
	u.processArgs = append(u.processArgs, u.Args...)
	return nil
}

func (u *UnitConfig) parseAutorestart() {
	switch u.AutoRestart {
	case AUTORESTART_ALWAYS:
		u.autorestart = true
		u.restartTimes = nil
	case AUTORESTART_LIMITED_TIMES:
		u.autorestart = true
		u.restartTimes = &u.Restartretries
	case AUTORESTART_NEVER:
		u.autorestart = false
		u.restartTimes = nil
	}
}

func (u *UnitConfig) parseSignal() {
	switch u.Signal {
	case SIGNAL_SIGTERM:
		u.signal = syscall.SIGTERM
	case SIGNAL_SIGINT:
		u.signal = syscall.SIGINT
	case SIGNAL_SIGQUIT:
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
		Stdout: u.Stdout,
		Stderr: u.Stderr,
		Workingdir: u.Workingdir,
		processName: u.processName,
		processArgs: u.processArgs,
		binPath: u.binPath,
		signal: u.signal,
		autorestart: u.autorestart,
		restartTimes: u.restartTimes,
	}
}

func (u UnitConfig) getProcessMeta() dto.ProcessMeta {
	return dto.ProcessMeta{
		Name: u.processName,
		BinPath: u.binPath,
		Args: u.processArgs,
		Env: u.Env,
		ProcessPath: u.Workingdir,
		Autostart: u.Autostart,
		Autorestart: u.autorestart,
		RestartTimes: u.restartTimes,
		StopSignal: u.signal,
		Exitcodes: u.Exitcodes,
		Starttime: u.Starttime,
		Stoptime: u.Stoptime,
	}
}
