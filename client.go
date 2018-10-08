package rtcp

import (
	"net"
	"time"
)

const (
	CMDPing = "ping"
	CMDPong = "pong"
	CMDData = "data"
)

func Dial(addr string, timeout time.Duration, logger Logger) (client *Client, err error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return
	}
	client = newClient(conn, logger)
	client.timeout = timeout
	client.logger = logger
	return
}

func newClient(conn net.Conn, logger Logger) *Client {
	client := new(Client)
	client.internalClient = newInternalClient(conn, logger)
	return client
}

type Client struct {
	OnData func(data []byte) ([]byte, error)
	timeout time.Duration
	*internalClient
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
		err := c.conn.SetReadDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return err
		}
		header, data, err := Read(c.conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				c.logger.Printf("read timeout and heart beat should received before timeout, so close the connection.")
				return err
			}
			return err
		}
		err = c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
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
