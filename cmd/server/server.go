package server

import (
	"fmt"
	"github.com/black40x/golog"
	"github.com/black40x/tunl-core/commands"
	"github.com/black40x/tunl-core/tunl"
	"net"
	"time"
)

type TunlServer struct {
	ln   net.Listener
	pool *ConnectionPool
	conf *Tunl
	log  *golog.Logger
}

func NewTunlServer(conf *Tunl, log *golog.Logger) *TunlServer {
	return &TunlServer{
		pool: NewConnectionPool(conf.UriPrefixSize),
		conf: conf,
		log:  log,
	}
}

func (s *TunlServer) GetPool() *ConnectionPool {
	return s.pool
}

func (s *TunlServer) processCommand(cmd *commands.Transfer, conn *tunl.TunlConn) {
	switch cmd.GetCommand().(type) {
	case *commands.Transfer_ClientConnect:
		if s.log != nil {
			s.log.Info(fmt.Sprintf("(TCP) %s try to connect", conn.Conn.RemoteAddr().String()))
		}

		if !s.conf.ServerPrivate || (s.conf.ServerPrivate && s.conf.ServerPassword == cmd.GetClientConnect().Password) {
			scheme := "http"
			if s.conf.SchemeHttps {
				scheme = "https"
			}
			pubUrl := fmt.Sprintf("%s://%s.%s", scheme, conn.ID, s.conf.Domain)

			conn.Send(&commands.ServerConnect{
				Prefix:    conn.ID,
				PublicUrl: pubUrl,
				Expire:    int64(s.conf.ClientExpireAt),
			})

			s.pool.Get(conn.ID).SetAllowed(true)
			s.pool.Get(conn.ID).SetHost(pubUrl)

			if s.log != nil {
				s.log.Info(fmt.Sprintf("(TCP) %s allow with URL %s", conn.Conn.RemoteAddr().String(), pubUrl))
			}
		}
		if s.conf.ServerPrivate && s.conf.ServerPassword != cmd.GetClientConnect().Password {
			conn.Send(&commands.Error{
				Code:    tunl.ErrorUnauthorized,
				Message: "invalid server password",
			})

			if s.log != nil {
				s.log.Warning(fmt.Sprintf("(TCP) %s invalid server password", conn.Conn.RemoteAddr().String()))
			}
		}
	case *commands.Transfer_HttpResponse:
		if s.pool.Get(conn.ID).IsAllowed() {
			ch := s.pool.GetResponseChan(cmd.GetHttpResponse().Uuid)
			if ch != nil {
				ch <- cmd.GetHttpResponse()
			}
		}
	case *commands.Transfer_BodyChunk:
		if s.pool.Get(conn.ID).IsAllowed() {
			ch := s.pool.GetBodyChunkChan(cmd.GetBodyChunk().Uuid)
			if ch != nil {
				ch <- cmd.GetBodyChunk()
			}
		}
	}
}

func (s *TunlServer) Start(conf Config) error {
	var err error
	s.ln, err = net.Listen("tcp", conf.Tunl.Addr+":"+conf.Tunl.Port)
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

			if s.log != nil {
				s.log.Warning("(TCP) client queue is full")
			}

			continue
		}

		c.SetOnCommand(func(cmd *commands.Transfer) {
			s.processCommand(cmd, c)
		})
		c.SetOnDisconnected(func() {
			if s.log != nil {
				s.log.Info(fmt.Sprintf("(TCP) %s disconnect", c.Conn.RemoteAddr()))
			}
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
