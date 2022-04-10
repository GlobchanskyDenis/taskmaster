package unitMaster

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor/unitMaster/unitSlave"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"context"
	"sync"
)

type Unit struct {
	name     string
	replicas []*unitSlave.Unit
	logger   dto.ILogger
}

func New(parentCtx context.Context, conf *dto.UnitConfig, logger dto.ILogger) *Unit {
	var master = &Unit{
		name: conf.ProcessName,
		logger: logger,
	}
	for i:=uint(0); i<conf.Replicas; i++ {
		master.replicas = append(master.replicas, unitSlave.New(parentCtx, conf, logger))
	}
	return master
}

func (master *Unit) GetStatusAsync(wg *sync.WaitGroup, amountLogs uint) {
	for _, slave := range master.replicas {
		wg.Add(1)
		go slave.GetStatusAsync(wg, amountLogs)
	}
	wg.Done()
}

func (master *Unit) StopAsync(wg *sync.WaitGroup) {
	for _, slave := range master.replicas {
		wg.Add(1)
		go slave.StopAsync(wg)
	}
	wg.Done()
}

func (master *Unit) StartAsync(wg *sync.WaitGroup) {
	for _, slave := range master.replicas {
		wg.Add(1)
		go slave.StartAsync(wg)
	}
	wg.Done()
}

func (master *Unit) RestartAsync(wg *sync.WaitGroup) {
	for _, slave := range master.replicas {
		wg.Add(1)
		go slave.RestartAsync(wg)
	}
	wg.Done()
}

func (master *Unit) PrintShortStatus(printer dto.IPrinter) {
	if len(master.replicas) > 1 {
		var sign string
		if master.isAllSlavesActive() == true {
			sign = "+"
		} else if master.isAllSlavesInactive() == true {
			sign = "-"
		} else {
			sign = "?"
		}

		printer.Printf("%s[%s] %s%s\n", constants.GREEN, sign, master.name, constants.NO_COLOR)
		for _, slave := range master.replicas {
			printer.Printf("    ")
			slave.PrintShortStatus(printer)
		}
	} else if len(master.replicas) == 1 {
		master.replicas[0].PrintShortStatus(printer)
	}
}

func (master *Unit) PrintFullStatus(printer dto.IPrinter) {
	if len(master.replicas) > 1 {
		printer.Printf("%s%s%s\n", constants.GREEN, master.name, constants.NO_COLOR)
		for num, slave := range master.replicas {
			slave.PrintFullStatus("   ", printer)
	
			/*	Если реплика не последняя -- добавляю еще один перевод строки  */
			if num < len(master.replicas) {
				printer.Printf("\n")
			}
		}
	} else {
		for _, slave := range master.replicas {
			slave.PrintFullStatus("", printer)
		}
	}
}

func (master *Unit) isAllSlavesActive() bool {
	if len(master.replicas) == 0 {
		return false
	}
	for _, slave := range master.replicas {
		if slave.GetStatusCode() != constants.STATUS_ACTIVE {
			return false
		}
	}
	return true
}

func (master *Unit) isAllSlavesInactive() bool {
	if len(master.replicas) == 0 {
		return false
	}
	for _, slave := range master.replicas {
		if slave.GetStatusCode() == constants.STATUS_ACTIVE {
			return false
		}
	}
	return true
}

func (master *Unit) GetName() string {
	return master.name
}
