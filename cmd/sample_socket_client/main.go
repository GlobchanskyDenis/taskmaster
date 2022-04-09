package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/socket"
	"fmt"
	"os"
	"time"
	"context"
)

const SOCKET_TEST_PATH = "/var/run/unixsock_test.sock"
const SOCKET_CONN_TEST_PATH_1 = "/var/run/unixsock_test_1.sock"
const SOCKET_CONN_TEST_PATH_2 = "/var/run/unixsock_test_2.sock"

/*	Реализация интерфейса */
type Printer struct {}

func (printer Printer) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func client() {
	conn := socket.NewClient(SOCKET_TEST_PATH)

	if err := conn.DialSocket(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	defer func(conn socket.Client) {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}(conn)

	for {
		for {
			if err := conn.DialServer(); err != nil {
				fmt.Printf("Ожидаю создания сервера. %s\n", err.Error())
				time.Sleep(time.Second * 3)
			} else {
				break
			}
		}
		line, err := conn.TransmitReceive([]byte("hi, im client, this is my request to server"))
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("server responsed '%s'\n", string(line))
		time.Sleep(time.Second * 3)
	}
}

/*	Тут тестирую многофункциональное соединение с сокетом  */
func connectionAsync() {
	println("I will listen from " + SOCKET_CONN_TEST_PATH_1 + " and write in " + SOCKET_CONN_TEST_PATH_2)
	conn := socket.New(SOCKET_CONN_TEST_PATH_1, SOCKET_CONN_TEST_PATH_2)

	ctx, _ := context.WithCancel(context.Background())

	if err := conn.ReaderStartAsync(ctx, Printer{}); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	defer func(conn socket.Conn) {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	}(conn)

	for {
		fmt.Printf("sending hi\n")
		if err := conn.Write([]byte("hi from other socket connection\n")); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		time.Sleep(time.Second * 3)
	}
}

func connection() {
	conn := socket.New(SOCKET_CONN_TEST_PATH_1, SOCKET_CONN_TEST_PATH_2)

	defer func(conn socket.Conn) {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	}(conn)

	if err := conn.ReaderStartSync(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	for {
		if message, err := conn.Read(); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("received: %s\n", string(message))
		}
		time.Sleep(time.Millisecond * 1500)

		fmt.Printf("sending hi\n")
		if err := conn.Write([]byte("hi from other socket connection")); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		time.Sleep(time.Millisecond * 1500)

		
	}
}

func main() {
	connection()
}