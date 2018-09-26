package rtcp

import (
	"net"
	"strings"
	"time"
)

const (
	CMDPing = "ping"
	CMDPong = "pong"
	CMDData = "data"
)

func Dial(addr string, logger Logger) (client *Client, err error) {
	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		return
	}
	client = newClient(conn)
	client.logger = logger
	return
}

func newClient(conn net.Conn) *Client {
	client := new(Client)
	client.conn = conn
	return client
}

type Client struct {
	conn   net.Conn
	OnData func(data []byte) ([]byte, error)
	logger Logger
	state  int32
}

func (c *Client) Serve() error {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Printf("recovered from panic: %v", err)
		}
		c.logger.Printf("close client with ip %s.", c.IP())
		err := c.Close()
		if err != nil {
			c.logger.Printf("fail to close connection: %s", err.Error())
		}
	}()
	for {
		header, data, err := Read(c.conn)
		if err != nil {
			return err
		}
		switch header.CMDStr() {
		case CMDPing:
			err = Write(c.conn, []byte(CMDPong), nil)
			if err != nil {
				return err
			}
		case CMDData:
			if c.OnData != nil {
				b, err := c.OnData(data)
				if err != nil {
					return err
				}
				err = Write(c.conn, []byte(CMDData), b)
				if err != nil {
					return err
				}
			}
		}
	}
}

func (c *Client) IP() string {
	strs := strings.Split(c.conn.RemoteAddr().String(), ":")
	if len(strs) > 0 {
		return strs[0]
	}
	return ""
}

func (c *Client) Close() error {
	return c.conn.Close()
}
