package supervisor

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/daemon"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"context"
)

type Supervisor interface {
	NewUnit(rawName string, env []string)
	GetStatusAllUnitsCli()
}

type supervisor struct {
	ctx      context.Context
	unitList []*unitMeta
}

func New(ctx context.Context) Supervisor {
	return &supervisor{
		ctx: ctx,
	}
}

func (s *supervisor) NewUnit(rawName string, env []string) {
	newSender := make(chan dto.Command)
	newReceiver := make(chan dto.CommandResult)
	ctx, _ := context.WithCancel(s.ctx)

	unit := newUnit(newSender, newReceiver)

	s.unitList = append(s.unitList, unit)

	/*	Запускаю горутину процесса и сразу опрашиваю ее статус чтобы узнать ее pid  */
	go daemon.RunAsync(ctx, newSender, newReceiver, rawName, env)
	unit.getStatus()
}

func (s *supervisor) GetStatusAllUnitsCli() {
	for _, unit := range s.unitList {
		unit.getStatus()
		unit.printShortStatus()
	}
}
