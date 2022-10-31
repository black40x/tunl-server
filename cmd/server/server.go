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
	conf *Tunl
}

func NewTunlServer(conf *Tunl) *TunlServer {
	return &TunlServer{
		pool: NewConnectionPool(conf.UriPrefixSize),
		conf: conf,
	}
}

func (s *TunlServer) GetPool() *ConnectionPool {
	return s.pool
}

func (s *TunlServer) processCommand(cmd *commands.Transfer, conn *tunl.TunlConn) {
	switch cmd.GetCommand().(type) {
	case *commands.Transfer_ClientConnect:
		if !s.conf.ServerPrivate || (s.conf.ServerPrivate && s.conf.ServerPassword == cmd.GetClientConnect().Password) {
			conn.Send(&commands.ServerConnect{
				Prefix:    conn.ID,
				PublicUrl: fmt.Sprintf(s.conf.ClientPublicAddr, conn.ID),
				Expire:    int64(s.conf.ClientExpireAt),
			})
			s.pool.SetAllowed(conn.ID)
		}
		if s.conf.ServerPrivate && s.conf.ServerPassword != cmd.GetClientConnect().Password {
			conn.Send(&commands.Error{
				Code:    tunl.ErrorUnauthorized,
				Message: "invalid server password",
			})
		}
	case *commands.Transfer_HttpResponse:
		if s.pool.IsAllowed(conn.ID) {
			ch := s.pool.GetResponseChan(cmd.GetHttpResponse().Uuid)
			if ch != nil {
				ch <- cmd.GetHttpResponse()
			}
		}
	case *commands.Transfer_BodyChunk:
		if s.pool.IsAllowed(conn.ID) {
			ch := s.pool.GetBodyChunkChan(cmd.GetBodyChunk().Uuid)
			if ch != nil {
				ch <- cmd.GetBodyChunk()
			}
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
				Message: "server client queue full",
			})
			c.Close()
			continue
		}

		c.SetOnCommand(func(cmd *commands.Transfer) {
			s.processCommand(cmd, c)
		})
		c.SetOnDisconnected(func() {
			s.pool.Drop(c.ID)
		})
		c.SetOnError(func(err error) {})
		c.SetExpireAt(time.Now().Add(time.Minute * time.Duration(conf.Tunl.ClientExpireAt)))
		s.pool.Push(c)

		go c.HandleConnection()
		go c.HandleExpire()

		c.Send(&commands.ServerHeader{
			Version: Version,
			Private: s.conf.ServerPrivate,
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
