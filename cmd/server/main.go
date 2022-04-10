package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/socket"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/cli_parser"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	// "github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/file_logger"
	"context"
	"syscall"
	"os"
	"time"
	"os/signal"
	"fmt"
)

const SOCKET_CONN_PATH_LISTEN = "/var/run/unixsock_2.sock"
const SOCKET_CONN_PATH_WRITE = "/var/run/unixsock_1.sock"

/*	Реализация интерфейса */
type Printer struct {
	conn socket.Conn
}

func (printer Printer) Printf(format string, args ...interface{}) {
	if err := printer.conn.Write([]byte(fmt.Sprintf(format, args...))); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

/*	Тут тестирую многофункциональное соединение с сокетом  */
func socketLoop(svisor supervisor.Supervisor) {
	conn := socket.New(SOCKET_CONN_PATH_LISTEN, SOCKET_CONN_PATH_WRITE)

	if err := conn.ReaderStartSync(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	defer func(conn socket.Conn) {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
	}(conn)

	printer := Printer{conn: conn}

	for {
		request, err := conn.Read()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		parsedCommand, err := cli_parser.ParseCliCommand(string(request))
		if err != nil {
			if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR)); err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			continue
		}

		switch parsedCommand.CommandType {
		case constants.COMMAND_STATUS:
			if len(parsedCommand.Args) == 0 {
				if err := svisor.Status(parsedCommand.UnitName, printer, 10); err != nil {
					if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR)); err != nil {
						fmt.Printf("Error: %s\n", err)
						return
					}
				}
			} else {
				switch parsedCommand.Args[0].Number {
				case constants.ARGUMENT_STATUS_ALL:
					svisor.StatusAll(printer)
				case constants.ARGUMENT_STATUS_NUMBER:
					if err := svisor.Status(parsedCommand.UnitName, printer, parsedCommand.Args[0].Value); err != nil {
						if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR)); err != nil {
							fmt.Printf("Error: %s\n", err)
							return
						}
					}
				}
			}
		case constants.COMMAND_STOP:
			if err := svisor.Stop(parsedCommand.UnitName, printer); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR)); err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}
			}
		case constants.COMMAND_START:
			if err := svisor.Start(parsedCommand.UnitName, printer); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR)); err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}
			}
		case constants.COMMAND_RESTART:
			if err := svisor.Restart(parsedCommand.UnitName, printer); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR)); err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}
			}
		case constants.COMMAND_KILL:
		default:
			println("unknown command")
		}		
	}
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

	socketLoop(newSupervisor)

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
