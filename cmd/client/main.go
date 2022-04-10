package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/socket"
	"fmt"
	"os"
	"context"
	"bufio"
)

const SOCKET_CONN_PATH_LISTEN = "/var/run/unixsock_1.sock"
const SOCKET_CONN_PATH_WRITE = "/var/run/unixsock_2.sock"

/*	Реализация интерфейса */
type Printer struct {}

func (printer Printer) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

/*	Тут тестирую многофункциональное соединение с сокетом  */
func connectionAsync() {
	conn := socket.New(SOCKET_CONN_PATH_LISTEN, SOCKET_CONN_PATH_WRITE)

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

	var line string

	for {
		line = scan()

		if err := conn.Write([]byte(line)); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	}
}

func scan() string {
	in := bufio.NewScanner(os.Stdin)
	in.Scan()
	if err := in.Err(); err != nil {
	  fmt.Fprintln(os.Stderr, "Ошибка ввода:", err)
	}
	return in.Text()
}

func main() {
	connectionAsync()
}