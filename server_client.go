package rtcp

import (
	"net"
	"time"
)

func newServerClient(server *Server, conn net.Conn) *ServerClient {
	client := new(ServerClient)
	client.internalClient = newInternalClient(conn, server.logger)
	client.s = server
	client.timeout = server.timeout
	client.lastHB = time.Now()
	return client
}

type ServerClient struct {
	*internalClient
	s       *Server
	timeout time.Duration
	lastHB  time.Time
}

func (c *ServerClient) LastHB() time.Time {
	return c.lastHB
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
		c.lastHB = time.Now()
		if c.s.OnPong != nil {
			c.s.OnPong(c)
		}
	}
}

func (c *ServerClient) Send(d []byte) (data []byte, err error) {
	_, data, err = c.send(CMDData, d)
	if err != nil && err != ErrClientBusy {
		c.s.removeClient(c.IP())
	}
	return
}

func (c *ServerClient) send(CMD string, d []byte) (header Header, data []byte, err error) {
	c.logger.Printf("send command %s %s.", CMD, d)
	if !c.acquire() {
		err = ErrClientBusy
		return
	}
	defer c.release()
	if c.timeout > 0 {
		err = c.conn.SetDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return
		}
	}
	err = Write(c.conn, []byte(CMD), d)
	if err != nil {
		return
	}
	return Read(c.conn)
}
