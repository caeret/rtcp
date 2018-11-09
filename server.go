package rtcp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func NewServer(addr string, timeout time.Duration, logger Logger) *Server {
	s := new(Server)
	s.addr = addr
	s.logger = logger
	s.clients = make(map[string]*ServerClient)
	s.timeout = timeout
	return s
}

type Server struct {
	addr      string
	listener  net.Listener
	clients   map[string]*ServerClient
	logger    Logger
	OnConnect func(*ServerClient) error
	OnPong    func(*ServerClient)
	timeout   time.Duration
	sync.RWMutex
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.listener = l

	s.logger.Printf("listen on addr %s", s.addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(time.Second)
				continue
			}
			return err
		}

		client := newServerClient(s, conn)

		if s.OnConnect != nil {
			err = s.OnConnect(client)
			if err != nil {
				s.logger.Printf("skip client: %s.", err)
				continue
			}
		}

		err = s.addClient(client)
		if err != nil {
			s.logger.Printf("fail to add client %s, so it will be closed: %s", client.IP(), err.Error())
			err = client.Close()
			if err != nil {
				s.logger.Printf("fail to close client: %s.", client.IP())
			}
			continue
		}

		s.logger.Printf("add client successfully.")

		go client.keepalive()
	}
}

func (s *Server) Client(IP string) *ServerClient {
	s.RLock()
	defer s.RUnlock()
	return s.clients[IP]
}

func (s *Server) Clients() (ret []*ServerClient) {
	s.RLock()
	defer s.RUnlock()
	ret = make([]*ServerClient, 0, len(s.clients))
	for _, client := range s.clients {
		ret = append(ret, client)
	}
	return
}

func (s *Server) addClient(client *ServerClient) error {
	s.logger.Printf("add new client with ip: %s.", client.IP())
	s.Lock()
	defer s.Unlock()
	ip := client.IP()
	if _, ok := s.clients[ip]; ok {
		return fmt.Errorf("client with ip %s already exists", ip)
	}
	s.clients[ip] = client
	return nil
}

func (s *Server) removeClient(ip string) {
	s.Lock()
	if old, ok := s.clients[ip]; ok {
		delete(s.clients, ip)
		err := old.Close()
		if err != nil {
			s.logger.Printf("fail to close client: %s", ip)
		}
	}
	s.Unlock()
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
