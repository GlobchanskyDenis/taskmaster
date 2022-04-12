package supervisor

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor/unitMaster"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"context"
	"errors"
	"sync"
)

var _ Supervisor = (*supervisor)(nil)

type Supervisor interface {
	StartByConfig(dto.UnitListConfig) error
	StatusAll(dto.IPrinter)
	Status(string, dto.IPrinter, uint) error
	Stop(string, dto.IPrinter) error
	Start(string, dto.IPrinter) error
	Restart(string, dto.IPrinter) error
}

type supervisor struct {
	ctx      context.Context
	wg       *sync.WaitGroup
	logger   dto.ILogger
	unitList []*unitMaster.Unit
}

func New(ctx context.Context, logger dto.ILogger) Supervisor {
	return &supervisor{
		ctx: ctx,
		wg: &sync.WaitGroup{},
		logger: logger,
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
		master := unitMaster.New(s.ctx, conf, s.logger)
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
	s.logInfo("Получена команда status-all")
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
	s.logInfo("Получена команда status процесса " + processName)
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		s.logWarning(err, "")
		return err
	}

	s.wg.Add(1)
	go master.GetStatusAsync(s.wg, amountLogs)
	s.wg.Wait()

	master.PrintFullStatus(printer)
	return nil
}

func (s *supervisor) Stop(processName string, printer dto.IPrinter) error {
	s.logInfo("Получена команда stop процесса " + processName)
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		s.logWarning(err, "")
		return err
	}

	s.wg.Add(1)
	go master.StopAsync(s.wg)
	s.wg.Wait()
	return nil
}

func (s *supervisor) Start(processName string, printer dto.IPrinter) error {
	s.logInfo("Получена команда start процесса " + processName)
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		s.logWarning(err, "")
		return err
	}

	s.wg.Add(1)
	go master.StartAsync(s.wg)
	s.wg.Wait()
	return nil
}

func (s *supervisor) Restart(processName string, printer dto.IPrinter) error {
	s.logInfo("Получена команда restart процесса " + processName)
	master, err := s.findMasterByProcessName(processName)
	if err != nil {
		s.logWarning(err, "")
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

func (s *supervisor) logInfo(message string) {
	s.logger.LogInfo(map[string]interface{}{
		"entity": "supervisor",
	}, message)
}

func (s *supervisor) logWarning(err error, message string) {
	s.logger.LogWarning(map[string]interface{}{
		"entity": "supervisor",
	}, err, message)
}
