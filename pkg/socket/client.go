package socket

import (
	"encoding/binary"
	"bytes"
	"io"
	"net"
)

type Client interface {
	Close() error
	DialSocket() error
	DialServer() error
	TransmitReceive([]byte) ([]byte, error)
	Transmit([]byte) error
}

type client struct {
	pathSocket string
	uaddr      *net.UnixAddr
	uconn      *net.UnixConn
}

func NewClient(unixSocketPath string) Client {
	return &client{
		pathSocket: unixSocketPath,
	}
}

func (entity *client) Close() error {
	if entity.uconn != nil {
		if err := entity.uconn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (entity *client) DialSocket() error {
	// Get unix socket address based on file path
	uaddr, err := net.ResolveUnixAddr("unix", entity.pathSocket)
	if err != nil {
		println("dial socket error")
		return err
	}
	entity.uaddr = uaddr

	return nil
}

func (entity *client) DialServer() error {
	// Connect server with unix socket
	uconn, err := net.DialUnix("unix", nil, entity.uaddr)
	if err != nil {
		return err
	}
	entity.uconn = uconn
	return nil
}

func (entity *client) TransmitReceive(data []byte) ([]byte, error) {

	/*	preparing data  */
	buf := new(bytes.Buffer)
	msglen := uint32(len(data))
	binary.Write(buf, binary.BigEndian, &msglen)
	data = append(buf.Bytes(), data...)

	/*	write request data  */
	if _, err := entity.uconn.Write(data); err != nil {
		return nil, err
	}

	/*	receiving response data  */
	var reqLen uint32
	lenBytes := make([]byte, 4)
	if _, err := io.ReadFull(entity.uconn, lenBytes); err != nil {
		return nil, err
	}
	lenBuf := bytes.NewBuffer(lenBytes)
	if err := binary.Read(lenBuf, binary.BigEndian, &reqLen); err != nil {
		return nil, err
	}
	reqBytes := make([]byte, reqLen)
	if _, err := io.ReadFull(entity.uconn, reqBytes); err != nil {
		return nil, err
	}

	return reqBytes, nil
}

func (entity *client) Transmit(data []byte) error {
	/*	preparing data  */
	buf := new(bytes.Buffer)
	msglen := uint32(len(data))
	binary.Write(buf, binary.BigEndian, &msglen)
	data = append(buf.Bytes(), data...)

	/*	write request data  */
	if _, err := entity.uconn.Write(data); err != nil {
		return err
	}
	return nil
}