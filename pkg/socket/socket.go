package socket

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"context"
	"time"
)

type Conn interface{
	Close() error
	ReaderStartAsync(context.Context, dto.IPrinter) error
	ReaderStartSync() error
	Read() ([]byte, error)
	Write([]byte) error
}

type connection struct {
	clientSocketFilePath string
	server               Server
}

func New(readSocketFilePath, writeSocketFilePath string) Conn {
	return &connection{
		clientSocketFilePath: writeSocketFilePath,
		server: NewServer(readSocketFilePath),
	}
}

/*	Контекст для gracefull shutdown, IPrinter - для инверсии зависимостей
**	Фактически этот метод запускает чтение в асинхроне и сам отписывает в интерфейс принтера  */
func (conn *connection) ReaderStartAsync(ctx context.Context, printer dto.IPrinter) error {
	/*	Устанавливаю соединение с сокетом для сервера  */
	if err := conn.server.DialSocket(); err != nil {
		return err
	}

	/*	Далее сервер будет работать в асинхроне и сам писать в терминал  */
	var serverReadChan = make(chan []byte)
	go conn.readServerLoop(serverReadChan)
	go conn.listen(ctx, printer, serverReadChan)

	return nil
}

func (conn *connection) ReaderStartSync() error {
	/*	Устанавливаю соединение с сокетом для сервера  */
	if err := conn.server.DialSocket(); err != nil {
		return err
	}
	return nil
}

func (conn *connection) readServerLoop(resultChan chan<- []byte) {
	for {
		result, err := conn.server.ListenWithoutAnswer()
		if err != nil {
			println("Чтение сокета сервером завершено с ошибкой " + err.Error())
			close(resultChan)
			return
		}
		resultChan <- result
	}
}

func (conn *connection) listen(ctx context.Context, printer dto.IPrinter, resultChan <-chan []byte) {
	for {
		select {
		case <- ctx.Done():
			if err := conn.Close(); err != nil {
				println(err.Error())
			}
			return
		case result, ok := <- resultChan:
			if ok {
				printer.Printf("%s", result)
			} else {
				println("channel was closed")
			}
		}
	}
}

func (conn *connection) Read() ([]byte, error) {
	result, err := conn.server.ListenWithoutAnswer()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (conn *connection) Close() error {
	if err := conn.server.Close(); err != nil {
		return err
	}
	return nil
}

func (conn *connection) Write(payload []byte) error {
	client := NewClient(conn.clientSocketFilePath)

	defer func(client Client) {
		if err := client.Close(); err != nil {
			println("client close error " + err.Error())
		}
	}(client)

	if err := client.DialSocket(); err != nil {
		return err
	}

	/*	Пока не соединимся - не закончим  */
	for {
		if err := client.DialServer(); err != nil {
			// fmt.Printf("Ожидаю создания сервера. %s\n", err.Error())
			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}

	/*	Обратно от сервера мы не ожидаем ничего  */
	if err := client.Transmit(payload); err != nil {
		return err
	}

	return nil
}
