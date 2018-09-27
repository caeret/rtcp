package rtcp

import (
	"net"
	"strings"
	"sync/atomic"
)

func newInternalClient(conn net.Conn, logger Logger) *internalClient {
	client := new(internalClient)
	client.conn = conn
	client.logger = logger
	return client
}

type internalClient struct {
	conn   net.Conn
	logger Logger
	state  int32
}

func (c *internalClient) acquire() bool {
	return atomic.CompareAndSwapInt32(&c.state, 0, 1)
}

func (c *internalClient) release() bool {
	return atomic.CompareAndSwapInt32(&c.state, 1, 0)
}

func (c *internalClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *internalClient) IP() string {
	if c.conn != nil {
		strs := strings.Split(c.conn.RemoteAddr().String(), ":")
		if len(strs) > 0 {
			return strs[0]
		}
	}
	return ""
}