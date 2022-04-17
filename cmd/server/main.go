package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/socket"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/cli_parser"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
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

func helpCommandOutput() string {
	return `	status [-n value] unit_name	Статус юнита
	status -all			Статус всех юнитов (сокращенный)
	stop unit_name			Остановка юнита
	start unit_name			Запуск юнита
	restart unit_name		Перезапуск юнита
	kill unit_name			Полное удаление юнита (до перезапуска)`
}

/*	Тут тестирую многофункциональное соединение с сокетом  */
func socketLoop(svisor supervisor.Supervisor, exit chan os.Signal) {
	conn := socket.New(SOCKET_CONN_PATH_LISTEN, SOCKET_CONN_PATH_WRITE)

	if err := conn.ReaderStartSync(); err != nil {
		fmt.Printf("Error: %s\n", err)
		exit <- syscall.SIGQUIT
	}

	defer func(conn socket.Conn) {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
			exit <- syscall.SIGQUIT
		}
	}(conn)

	printer := Printer{conn: conn}

	for {
		request, err := conn.Read()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			exit <- syscall.SIGQUIT
		}

		parsedCommand, err := cli_parser.ParseCliCommand(string(request))
		if err != nil {
			if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
				fmt.Printf("Error: %s\n", err)
				exit <- syscall.SIGQUIT
			}
			continue
		}

		switch parsedCommand.CommandType {
		case constants.COMMAND_STATUS:
			if len(parsedCommand.Args) == 0 {
				if err := svisor.Status(parsedCommand.UnitName, printer, 10); err != nil {
					if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
						fmt.Printf("Error: %s\n", err)
						exit <- syscall.SIGQUIT
					}
				}
			} else {
				switch parsedCommand.Args[0].Number {
				case constants.ARGUMENT_STATUS_ALL:
					svisor.StatusAll(printer)
				case constants.ARGUMENT_STATUS_NUMBER:
					if err := svisor.Status(parsedCommand.UnitName, printer, parsedCommand.Args[0].Value); err != nil {
						if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
							fmt.Printf("Error: %s\n", err)
							exit <- syscall.SIGQUIT
						}
					}
				}
			}
		case constants.COMMAND_STOP:
			if err := svisor.Stop(parsedCommand.UnitName, printer); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
					fmt.Printf("Error: %s\n", err)
					exit <- syscall.SIGQUIT
				}
			}
		case constants.COMMAND_START:
			if err := svisor.Start(parsedCommand.UnitName, printer); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
					fmt.Printf("Error: %s\n", err)
					exit <- syscall.SIGQUIT
				}
			}
		case constants.COMMAND_RESTART:
			if err := svisor.Restart(parsedCommand.UnitName, printer); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
					fmt.Printf("Error: %s\n", err)
					exit <- syscall.SIGQUIT
				}
			}
		case constants.COMMAND_KILL:
			if err := svisor.Kill(parsedCommand.UnitName, printer); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
					fmt.Printf("Error: %s\n", err)
					exit <- syscall.SIGQUIT
				}
			}
		case constants.COMMAND_HELP:
			if err := conn.Write([]byte(constants.YELLOW + helpCommandOutput() + constants.NO_COLOR + "\n")); err != nil {
				fmt.Printf("Error: %s\n", err)
				exit <- syscall.SIGQUIT
			}
		case constants.COMMAND_EXIT:
			println("Получил от клиента сигнал остановки программы")
			exit <- syscall.SIGQUIT
		case constants.COMMAND_RECONFIG:
			if unitListConfig, err := initializeConfigs(configPath); err != nil {
				if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
					fmt.Printf("Error: %s\n", err)
					exit <- syscall.SIGQUIT
				}
			} else {
				if err := svisor.UpdateByConfig(unitListConfig); err != nil {
					if err := conn.Write([]byte(constants.RED + err.Error() + constants.NO_COLOR + "\n")); err != nil {
						fmt.Printf("Error: %s\n", err)
						exit <- syscall.SIGQUIT
					}
				}
			}
		default:
			continue
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

	exit := make(chan os.Signal, 1)

	go socketLoop(newSupervisor, exit)

	/*	Механизм gracefull shutdown реализуется тут  */
	waitForGracefullShutdown(cancel, unitListConfig.GetMaxStopTime(), exit)
}

func waitForGracefullShutdown(cancel context.CancelFunc, waitTime uint, exit chan os.Signal) {
	/*	Отлавливаю системный вызов останова программы. Это блокирующая операция  */
	signal.Notify(exit,
		syscall.SIGTERM, /*  Согласно всякой документации именно он должен останавливать прогу, но на деле его мы не находим. Оставил его просто на всякий случай  */
		syscall.SIGINT,  /*  Останавливает прогу когда она запущена из терминала и останавливается через CTRL+C  */
		syscall.SIGQUIT, /*  Останавливает демона systemd  */
	)
	<-exit

	/*	Посылаю каждому воркеру сигнал останова  */
	cancel()

	/*	Ожидаем завершения всех команд. Время ожидания берется из конфигурационника  */
	time.Sleep(time.Duration(waitTime) * time.Millisecond * 1000 + 500)
}
