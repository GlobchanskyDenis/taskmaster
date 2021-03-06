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

func server() {
	conn := socket.NewServer(SOCKET_TEST_PATH)

	if err := conn.DialSocket(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	
	defer func(conn socket.Server) {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}(conn)

	for {
		line, err := conn.Listen()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Client said '%s'\n", string(line))
		if err := conn.Answer([]byte("hi, i'm server, this is my answer")); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	}
}

/*	Тут тестирую многофункциональное соединение с сокетом  */
func connectionAsync() {
	conn := socket.New(SOCKET_CONN_TEST_PATH_2, SOCKET_CONN_TEST_PATH_1)

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
		if err := conn.Write([]byte("hi from other socket connection")); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		time.Sleep(time.Second * 3)
	}
}

func main() {
	connectionAsync()
}