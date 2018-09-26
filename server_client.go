package rtcp

import (
	"net"
	"sync/atomic"
	"time"
)

func newServerClient(server *Server, conn net.Conn) *ServerClient {
	client := new(ServerClient)
	client.Client = newClient(conn)
	client.s = server
	client.logger = server.logger
	return client
}

type ServerClient struct {
	*Client
	s *Server
}

func (c *ServerClient) keepalive() {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Printf("recovered from panic: %v", err)
		}
		c.logger.Printf("close client with ip %s.", c.IP())
		c.s.removeClient(c.IP())
	}()
	for {
		<-time.After(time.Second * 10)
		c.logger.Printf("try send ping for client with ip %s.", c.IP())
		_, _, err := c.send(CMDPing, nil)
		if err != nil {
			if err == ErrClientBusy {
				c.logger.Printf("client is busy.")
				continue
			}
			c.logger.Printf("fail to send request: %s.", err.Error())
			return
		}
		c.logger.Printf("received keepalive pong from client.")
		if c.s.OnPong != nil {
			c.s.OnPong(c)
		}
	}
}

func (c *ServerClient) Send(d []byte) (header Header, data []byte, err error) {
	return c.send(CMDData, d)
}

func (c *ServerClient) send(CMD string, d []byte) (header Header, data []byte, err error) {
	c.logger.Printf("send command %s %s.", CMD, d)
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		err = ErrClientBusy
		return
	}
	defer atomic.CompareAndSwapInt32(&c.state, 1, 0)
	err = c.conn.SetDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return
	}
	err = Write(c.conn, []byte(CMD), d)
	if err != nil {
		return
	}
	return Read(c.conn)
}
