package unitSlave

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor/unitMaster/unitSlave/daemon"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"context"
	"strconv"
	"time"
	"sync"
)

type Unit struct {
	pid      int
	name     string
	binPath  string
	procPath *string
	sender   chan<- dto.Command
	receiver <-chan dto.CommandResult
	logger   dto.ILogger

	statusCode     uint
	status         string
	lastError      error
	exitCode       int
	logs           []string
	lastChangeTime time.Time
}

func New(parentCtx context.Context, conf *dto.UnitConfig, logger dto.ILogger) *Unit {
	newSender := make(chan dto.Command)
	newReceiver := make(chan dto.CommandResult)
	var slave = &Unit{
		name: conf.ProcessName,
		binPath: conf.BinPath,
		procPath: conf.Workingdir,
		sender: newSender,
		receiver: newReceiver,
		logger: logger,
		statusCode: constants.STATUS_ACTIVE,
		status: "Процесс отправлен на запуск",
		lastError: nil,
	}

	ctx, _ := context.WithCancel(parentCtx)
	/*	Запускаю горутину процесса  */
	go daemon.RunAsync(ctx, newSender, newReceiver, conf.GetProcessMeta(), logger)
	return slave
}

func (slave *Unit) GetStatusAsync(wg *sync.WaitGroup, amountLogs uint) {
	slave.sender <- dto.Command{
		Type: constants.COMMAND_STATUS,
		AmountLogs: amountLogs,
	}
	result := <- slave.receiver
	slave.handleResponse(result)
	wg.Done()
}

func (slave *Unit) StopAsync(wg *sync.WaitGroup) {
	slave.sender <- dto.Command{
		Type: constants.COMMAND_STOP,
	}
	result := <- slave.receiver
	slave.handleResponse(result)
	wg.Done()
}

func (slave *Unit) StartAsync(wg *sync.WaitGroup) {
	slave.sender <- dto.Command{
		Type: constants.COMMAND_START,
	}
	result := <- slave.receiver
	slave.handleResponse(result)
	wg.Done()
}

func (slave *Unit) RestartAsync(wg *sync.WaitGroup) {
	slave.sender <- dto.Command{
		Type: constants.COMMAND_RESTART,
	}
	result := <- slave.receiver
	slave.handleResponse(result)
	wg.Done()
}

func (slave *Unit) PrintShortStatus(printer dto.IPrinter) {
	if slave.statusCode == constants.STATUS_ACTIVE {
		printer.Printf("%s[+] %5d %s%s\n", constants.GREEN, slave.pid, slave.name, constants.NO_COLOR)
	} else {
		printer.Printf("%s[-] %5d %s%s\n", constants.GREEN, slave.pid, slave.name, constants.NO_COLOR)
	}
}

func (slave *Unit) PrintFullStatus(prefix string, printer dto.IPrinter) {
	printer.Printf("%s%sprocess: %s%s\n", constants.GREEN, prefix, slave.name, constants.NO_COLOR)
	printer.Printf("%s%s     Path: binary (%s)", constants.GREEN, prefix, slave.binPath)
	if slave.procPath != nil {
		printer.Printf(" process (%s)%s\n", *slave.procPath, constants.NO_COLOR)
	} else {
		printer.Printf("%s\n", constants.NO_COLOR)
	}
	printer.Printf("%s%s   Active: %s%s%s since %s%s\n", constants.GREEN, prefix, constants.NO_COLOR, slave.stringStatusCode(), constants.GREEN, slave.stringChangeTime(), constants.NO_COLOR)
	printer.Printf("%s%s      Pid: %d%s\n", constants.GREEN, prefix, int(slave.pid), constants.NO_COLOR)
	printer.Printf("%s%s   Status: %s%s\n", constants.GREEN, prefix, slave.status, constants.NO_COLOR)
	if slave.lastError != nil {
		printer.Printf("%s%s    Error: %s%s%s%s\n", constants.GREEN, prefix, constants.RED, slave.lastError.Error(), constants.NO_COLOR, constants.NO_COLOR)
	}
	printer.Printf("\n")
	for _, log := range slave.logs {
		printer.Printf("%s%s%s\n", constants.GREEN, log, constants.NO_COLOR)
	}
}

func (slave *Unit) stringStatusCode() string {
	switch slave.statusCode {
	case constants.STATUS_ACTIVE:
		return constants.GREEN_BG + "active (running)" + constants.NO_COLOR
	case constants.STATUS_STOPPED:
		return "inactive (stopped)"
	case constants.STATUS_DEAD:
		return "inactive (dead)"
	case constants.STATUS_ERROR:
		return constants.RED + "error status" + constants.NO_COLOR + " exit code " + strconv.Itoa(slave.exitCode)
	default:
		return constants.RED_BG + "unknown status" + constants.NO_COLOR
	}
}

func (slave *Unit) stringChangeTime() string {
	return slave.lastChangeTime.Format(time.RFC3339)
}

func (slave *Unit) GetStatusCode() uint {
	return slave.statusCode
}

func (slave *Unit) handleResponse(result dto.CommandResult) {
	slave.pid = result.Pid
	slave.name = result.Name
	slave.statusCode = result.StatusCode // Моя внутренняя константа
	slave.status = result.Status
	slave.lastError = result.Error
	slave.exitCode = result.ExitCode // Системный код завершения программы
	slave.lastChangeTime = result.ChangeTime
	slave.logs = result.Logs
}
