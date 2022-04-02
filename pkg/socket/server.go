package socket

import (
	"encoding/binary"
	"bytes"
	"io"
	"os"
	"net"
)

type Server interface {
	Close() error
	DialSocket() error
	Listen() (string, error)
	ListenWithoutAnswer() (string, error)
	Answer(response string) error
}

type server struct {
	pathSocket        string
	listener          *net.UnixListener
	messageConnection *net.UnixConn
}

func NewServer(unixSocketPath string) Server {
	return &server{
		pathSocket: unixSocketPath,
	}
}

func (entity *server) Close() error {
	if entity.listener != nil {
		if err := entity.listener.Close(); err != nil {
			return err
		}
		entity.listener = nil
	}
	return nil
}

func (entity *server) DialSocket() error {
	// Remove socket file
	os.Remove(entity.pathSocket)

	// Get unix socket address based on file path
	uaddr, err := net.ResolveUnixAddr("unix", entity.pathSocket)
	if err != nil {
		return err
	}
 
	// Listen on the address
	unixListener, err := net.ListenUnix("unix", uaddr)
	if err != nil {
		return err
	}

	entity.listener = unixListener

	return nil
}

func (entity *server) Listen() (string, error) {
	uconn, err := entity.listener.AcceptUnix()
	if err != nil {
		return "", err
	}
	entity.messageConnection = uconn

	requestMessageByte, err := entity.parseRequest()
	if err != nil {
		return "", err
	}

	return string(requestMessageByte), nil
}

func (entity *server) ListenWithoutAnswer() (string, error) {
	uconn, err := entity.listener.AcceptUnix()
	if err != nil {
		return "", err
	}
	entity.messageConnection = uconn

	defer func(entity *server) {
		if entity.messageConnection != nil {
			if err := entity.messageConnection.Close(); err != nil {
				println("during close error " + err.Error())
			}
		}
		entity.messageConnection = nil
	}(entity)

	requestMessageByte, err := entity.parseRequest()
	if err != nil {
		return "", err
	}

	return string(requestMessageByte), nil
}

func (entity *server) parseRequest() ([]byte, error) {
	var reqLen uint32
	lenBytes := make([]byte, 4)
	if _, err := io.ReadFull(entity.messageConnection, lenBytes); err != nil {
		return nil, err
	}
 
	lenBuf := bytes.NewBuffer(lenBytes)
	if err := binary.Read(lenBuf, binary.BigEndian, &reqLen); err != nil {
		return nil, err
	}
 
	reqBytes := make([]byte, reqLen)
	_, err := io.ReadFull(entity.messageConnection, reqBytes)
 
	if err != nil {
		return nil, err
	}
 
	return reqBytes, nil
}

func (entity *server) Answer(response string) error {
	var data = []byte(response)
	defer func(entity *server) {
		if entity.messageConnection != nil {
			if err := entity.messageConnection.Close(); err != nil {
				println("during close error " + err.Error())
			}
		}
		entity.messageConnection = nil
	}(entity)

	buf := new(bytes.Buffer)
	msglen := uint32(len(data))
 
	if err := binary.Write(buf, binary.BigEndian, &msglen); err != nil {
		return err
	}
	data = append(buf.Bytes(), data...)

	if _, err := entity.messageConnection.Write(data); err != nil {
		return err
	}

	return nil
}
