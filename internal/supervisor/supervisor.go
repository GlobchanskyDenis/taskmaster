package supervisor

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor/unitMaster"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"context"
	"errors"
	"sync"
)

type Supervisor interface {
	StartByConfig(confList dto.UnitListConfig) error
	StatusAll(dto.IPrinter)
	Status(string, dto.IPrinter, uint) error
	Stop(string, dto.IPrinter) error
	Start(string, dto.IPrinter) error
	Restart(string, dto.IPrinter) error
}

type supervisor struct {
	ctx      context.Context
	wg       *sync.WaitGroup
	unitList []*unitMaster.Unit
}

func New(ctx context.Context) Supervisor {
	return &supervisor{
		ctx: ctx,
		wg: &sync.WaitGroup{},
	}
}

func (s *supervisor) StartByConfig(confList dto.UnitListConfig) error {
	/*	Валидируем и парсим все конфиги  */
	for _, conf := range confList {
		if err := conf.Validate(); err != nil {
			return err
		}
		if err := conf.Parse(); err != nil {
			return err
		}
	}

	/*	Запускаем процессы  */
	for _, conf := range confList {
		master := unitMaster.New(s.ctx, conf)
		s.unitList = append(s.unitList, master)
	}

	/*	Получаем статусы процессов чтобы сразу знать их pid  */
	for _, master := range s.unitList {
		s.wg.Add(1)
		go master.GetStatusAsync(s.wg, 0)
	}
	s.wg.Wait()
	return nil
}

func (s *supervisor) StatusAll(printer dto.IPrinter) {
	for _, master := range s.unitList {
		s.wg.Add(1)
		go master.GetStatusAsync(s.wg, 0)
	}
	s.wg.Wait()

	for _, master := range s.unitList {
		master.PrintShortStatus(printer)
	}
}

func (s *supervisor) Status(processName string, printer dto.IPrinter, amountLogs uint) error {
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go master.GetStatusAsync(s.wg, amountLogs)
	s.wg.Wait()

	master.PrintFullStatus(printer)
	return nil
}

func (s *supervisor) Stop(processName string, printer dto.IPrinter) error {
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go master.StopAsync(s.wg)
	s.wg.Wait()
	return nil
}

func (s *supervisor) Start(processName string, printer dto.IPrinter) error {
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go master.StartAsync(s.wg)
	s.wg.Wait()
	return nil
}

func (s *supervisor) Restart(processName string, printer dto.IPrinter) error {
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go master.RestartAsync(s.wg)
	s.wg.Wait()
	return nil
}

// TODO -- в конфигурационнике после парсинга перепроверять чтобы не было одинаковых процессов
func (s *supervisor) findMasterByProcessName(processName string) (*unitMaster.Unit, error) {
	for _, master := range s.unitList {
		if master.GetName() == processName {
			return master, nil
		}
	}
	return nil, errors.New("Процесс " + processName + " не найден")
}
