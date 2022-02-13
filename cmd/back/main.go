package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor"
	"context"
	"syscall"
	"os"
	"time"
	"os/signal"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	newSupervisor := supervisor.New(ctx)

	newSupervisor.NewUnit("vim", []string{})
	// newSupervisor.NewUnit("vim", []string{})
	// newSupervisor.NewUnit("vim", []string{})
	// newSupervisor.NewUnit("vim", []string{})

	time.Sleep(time.Second * 2)

	newSupervisor.GetStatusAllUnitsCli()

	waitForGracefullShutdown(cancel)
}

func waitForGracefullShutdown(cancel context.CancelFunc) {
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
	time.Sleep(3000 * time.Millisecond)
}
