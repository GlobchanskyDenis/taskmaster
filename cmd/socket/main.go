package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/file_logger"
	"context"
	"syscall"
	"os"
	"time"
	"os/signal"
	"fmt"
)

/*	Реализация интерфейса */
type Printer struct {}

func (printer Printer) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func main() {
	/*	Обрабатываем флаги (если они есть в наличии)  */
	if err := parseFlags(); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	/*	Инициализируем конфиги всех пакетов  */
	unitListConfig, err := initializeConfigs(configPath)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	/*	Инициализируем логгера. Сама сущность хранится в самом пакете  */
	if err := file_logger.NewLogger(); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	/*	Создаем и инициализируем наш супервизор  */
	ctx, cancel := context.WithCancel(context.Background())
	newSupervisor := supervisor.New(ctx, file_logger.GLogger)
	if err := newSupervisor.StartByConfig(unitListConfig); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	/*	Это просто эмуляция команд. Только на время отладки  */
	debugSimple(newSupervisor)

	/*	Механизм gracefull shutdown реализуется тут  */
	waitForGracefullShutdown(cancel, unitListConfig.GetMaxStopTime())
}

func waitForGracefullShutdown(cancel context.CancelFunc, waitTime uint) {
	/*	Отлавливаю системный вызов останова программы. Это блокирующая операция  */
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGTERM, /*  Согласно всякой документации именно он должен останавливать прогу, но на деле его мы не находим. Оставил его просто на всякий случай  */
		syscall.SIGINT,  /*  Останавливает прогу когда она запущена из терминала и останавливается через CTRL+C  */
		syscall.SIGQUIT, /*  Останавливает демона systemd  */
	)
	<-quit

	/*	Посылаю каждому воркеру сигнал останова  */
	cancel()

	/*	Ожидаем завершения всех команд. Время ожидания берется из конфигурационника  */
	time.Sleep(time.Duration(waitTime) * time.Millisecond * 1000 + 500)
}
