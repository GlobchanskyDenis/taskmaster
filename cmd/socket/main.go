package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor"
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
	if err := parseFlags(); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	unitListConfig, err := initializeConfigs(configPath)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	newSupervisor := supervisor.New(ctx)
	if err := newSupervisor.StartByConfig(unitListConfig); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	time.Sleep(time.Second * 1)

	newSupervisor.StatusAll(Printer{})

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("stop")
	if err := newSupervisor.Stop("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 3)

	println("start")
	if err := newSupervisor.Start("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	newSupervisor.StatusAll(Printer{})

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

	/*	Чтобы не усложнять бизнес логику - простой способ подождать завершения всех
	**	воркеров. Время стоит расчитывать как сумму из:
	**	-- времени таймаута БД (в строке dsn к базе данных jwtgost конфигуратора)
	**	-- времени ожидания внешних сервисов (в настройках конфигуратора)
	**	+ 100 миллисекунд на одного воркера (про запас)  */
	time.Sleep(time.Duration(waitTime) * time.Millisecond * 1000 + 500)
}
