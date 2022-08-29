package server

import (
	"fmt"
	"github.com/black40x/tunl-core/commands"
	"github.com/black40x/tunl-core/tunl"
	"net"
	"time"
)

type TunlServer struct {
	ln   net.Listener
	pool *ConnectionPool
	conf *TunlConfig
}

func NewTunlServer(conf *TunlConfig) *TunlServer {
	return &TunlServer{
		pool: NewConnectionPool(conf.UriPrefixSize),
		conf: conf,
	}
}

func (s *TunlServer) GetPool() *ConnectionPool {
	return s.pool
}

func (s *TunlServer) processCommand(cmd *commands.Transfer) {
	switch cmd.GetCommand().(type) {
	case *commands.Transfer_HttpResponse:
		ch := s.pool.GetResponseChan(cmd.GetHttpResponse().Uuid)
		if ch != nil {
			ch <- cmd.GetHttpResponse()
		}
	case *commands.Transfer_BodyChunk:
		ch := s.pool.GetBodyChunkChan(cmd.GetBodyChunk().Uuid)
		if ch != nil {
			ch <- cmd.GetBodyChunk()
		}
	}
}

func (s *TunlServer) Start(conf Config) error {
	var err error
	s.ln, err = net.Listen("tcp", conf.Tunl.TunlAddr+":"+conf.Tunl.TunlPort)
	if err != nil {
		return err
	}

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			continue
		}

		c := tunl.NewTunlConn(conn)

		if s.pool.Count()+1 > s.conf.MaxClients {
			c.Send(&commands.Error{
				Code:    tunl.ErrorServerFull,
				Message: "server full",
			})
			c.Close()
			continue
		}

		c.SetOnCommand(func(cmd *commands.Transfer) {
			s.processCommand(cmd)
		})
		c.SetOnDisconnected(func() {
			s.pool.Drop(c.ID)
		})
		c.SetOnError(func(err error) {})
		c.SetExpireAt(time.Now().Add(time.Minute * time.Duration(conf.Tunl.ClientExpireAt)))
		s.pool.Push(c)

		go c.HandleConnection()
		go c.HandleExpire()

		// ToDo - Add password protect
		c.Send(&commands.ServerConnect{
			Prefix:    c.ID,
			PublicUrl: fmt.Sprintf(conf.Tunl.ClientPublicAddr, c.ID),
			Expire:    int64(conf.Tunl.ClientExpireAt),
			Version:   Version,
		})
	}
}

func (s *TunlServer) Run(conf *Config) error {
	err := s.Start(*conf)
	if err != nil {
		return err
	}
	return nil
}
