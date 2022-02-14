package supervisor

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/daemon"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"context"
	"sync"
	// "time"
)

type Supervisor interface {
	// NewUnit(rawName string, env []string) // сделать неэкспортируемым. Оно должно быть
	GetStatusAllUnitsCli()
	SetUnitsByConfig(confList UnitListConfig) error
}

type supervisor struct {
	ctx      context.Context
	wg       *sync.WaitGroup
	unitList []*unitMeta
}

func New(ctx context.Context) Supervisor {
	return &supervisor{
		ctx: ctx,
		wg: &sync.WaitGroup{},
	}
}

func (s *supervisor) newUnit(conf *UnitConfig) { //, stopSignal syscall.Signal //rawName string, env []string
	newSender := make(chan dto.Command)
	newReceiver := make(chan dto.CommandResult)
	ctx, _ := context.WithCancel(s.ctx)

	unit := newUnit(newSender, newReceiver)
	s.unitList = append(s.unitList, unit)

	/*	Запускаю горутину процесса и сразу опрашиваю ее статус чтобы узнать ее pid  */
	go daemon.RunAsync(ctx, newSender, newReceiver, conf.getProcessMeta()) // rawName, env
	unit.getStatus()
}

func (s *supervisor) GetStatusAllUnitsCli() {
	for _, unit := range s.unitList {
		s.wg.Add(1)
		go unit.getStatusAsync(s.wg)
	}
	s.wg.Wait()

	for _, unit := range s.unitList {
		unit.printShortStatus()
	}
}

func (s *supervisor) SetUnitsByConfig(confList UnitListConfig) error {
	for _, conf := range confList {
		if err := conf.validate(); err != nil {
			return err
		}
		if err := conf.parse(); err != nil {
			return err
		}
	}
	for _, conf := range confList {
		s.newUnit(conf)
	}
	return nil
}
