package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/black40x/golog"
	"github.com/black40x/tunl-core/commands"
	"github.com/black40x/tunl-core/tunl"
	"github.com/black40x/tunl-server/cmd/tui"
	"github.com/black40x/tunl-server/cmd/utils"
	"github.com/black40x/tunl-server/ui"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/net/netutil"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

const BrowserWarningCookieName = "tunl-online-skip-warning"
const BrowserWarningHeaderName = "Tunl-Online-Skip-Warning"

type TunlHttp struct {
	tunl        *TunlServer
	conf        *Config
	httpSrv     *http.Server
	log         *golog.Logger
	ctx         context.Context
	appTemplate *template.Template
}

type AppMessage struct {
	Title        string
	Data         string
	NoScriptText string
}

func NewTunlHttp(conf *Config, log *golog.Logger, ctx context.Context) *TunlHttp {
	srv := &TunlHttp{
		conf: conf,
		ctx:  ctx,
		log:  log,
	}
	srv.prepareTemplate()

	return srv
}

func (s *TunlHttp) prepareTemplate() {
	appUI, _ := fs.Sub(ui.AppUI, "app/build")
	s.appTemplate, _ = template.ParseFS(appUI, "index.html")
}

func (s *TunlHttp) handle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	connId := params["subdomain"]

	if connId == "" || s.tunl.pool.Get(connId) == nil {
		s.browserError(connId, ErrorUndefinedClient, w)
		return
	} else {
		if s.conf.Tunl.BrowserWarning && r.Header.Get(BrowserWarningHeaderName) == "" {
			if utils.IsBrowserRequest(r) && !utils.HasCookie(r, BrowserWarningCookieName) {
				s.browserWarning(s.tunl.pool.Get(connId), w)
				return
			}
		}

		req := commands.HttpRequest{
			Uuid:          uuid.New().String(),
			Method:        r.Method,
			Proto:         r.Proto,
			Uri:           r.URL.String(),
			ContentLength: r.ContentLength,
			RemoteAddr:    r.RemoteAddr,
		}

		for _, v := range r.Cookies() {
			req.Cookies = append(req.Cookies, &commands.Cookie{
				Name:     v.Name,
				Value:    v.Value,
				Path:     v.Path,
				Domain:   v.Domain,
				Expires:  v.Expires.UnixMicro(),
				HttpOnly: v.HttpOnly,
				Secure:   v.Secure,
			})
		}

		for k, v := range r.Header {
			req.Header = append(req.Header, &commands.Header{Key: k, Value: v})
		}

		chResp := s.tunl.pool.MakeResponseChan(req.Uuid)
		chBody := s.tunl.pool.MakeBodyChunkChan(req.Uuid)
		defer s.tunl.pool.CloseChannels(req.Uuid)

		_, err := s.tunl.pool.Get(connId).Conn().Send(&req)
		if err != nil {
			s.browserError(connId, ErrorConnectClient, w)
			return
		} else {
			if r.ContentLength > 0 {
				re := bufio.NewReader(r.Body)
				buf := make([]byte, 0, tunl.ReaderSize)
				for {
					n, err := re.Read(buf[:cap(buf)])
					buf = buf[:n]
					if n == 0 {
						if err == io.EOF {
							break
						}
					} else {
						s.tunl.pool.Get(connId).Conn().Send(&commands.BodyChunk{
							Uuid: req.Uuid,
							Body: buf,
							Eof:  false,
						})
					}
				}
			}

			var bodySize int64 = 0
			select {
			case res := <-chResp:
				if int(res.ErrorCode) == int(tunl.ErrorClientResponse) {
					s.browserError(connId, ErrorCode(tunl.ErrorClientResponse), w)
					break
				}

				for _, h := range res.Header {
					for _, v := range h.GetValue() {
						w.Header().Add(h.GetKey(), v)
					}
				}
				w.WriteHeader(int(res.Status))

				if res.ContentLength != 0 {
				L:
					for {
						select {
						case body := <-chBody:
							bodySize += int64(len(body.Body))
							w.Write(body.Body)
							if (bodySize >= res.ContentLength && res.ContentLength != -1) || body.Eof {
								break L
							}
						case <-time.After(time.Second * 60):
							break L
						}
					}
				}
			case <-time.After(time.Second * 30):
				s.browserError(connId, ErrorReceiveData, w)
			}
		}
	}
}

func (s *TunlHttp) Shutdown() {
	if s.log != nil {
		s.log.Info("Shutdown server")
	}

	if s.httpSrv != nil {
		s.httpSrv.Shutdown(s.ctx)
	}
}

func (s *TunlHttp) startTunl() {
	s.tunl = NewTunlServer(s.conf.Tunl, s.log)

	if s.log != nil {
		s.log.Info("(TCP) Starting tunl server...")
	}

	go func() {
		err := s.tunl.Run(s.conf)

		if err != nil {
			tui.PrintError(err)

			if s.log != nil {
				s.log.Error("(TCP) " + err.Error())
			}

			os.Exit(1)
		}
	}()
}

func (s *TunlHttp) Start() {
	s.startTunl()

	appAssets, _ := fs.Sub(ui.AppUI, "app/build/static")

	r := mux.NewRouter()
	r.Host("cdn." + s.conf.Tunl.Domain).
		PathPrefix("/app/static/").
		Handler(http.StripPrefix("/app/static/", utils.NoFileListing(
			http.FileServer(http.FS(appAssets)),
		)))
	r.Host("{subdomain:[0-9a-z/-]+}." + s.conf.Tunl.Domain).HandlerFunc(s.handle)

	addr := s.conf.Server.HTTPAddr + ":" + s.conf.Server.HTTPPort
	concurrency := runtime.NumCPU() * 2
	listener, _ := net.Listen("tcp", addr)
	listener = netutil.LimitListener(listener, concurrency*s.conf.Server.ThreadCount)

	s.httpSrv = &http.Server{
		Addr:        addr,
		Handler:     r,
		ReadTimeout: time.Duration(s.conf.Server.ReadTimeout) * time.Second,
		IdleTimeout: time.Duration(s.conf.Server.IdleTimeout) * time.Second,
	}
	s.httpSrv.SetKeepAlivesEnabled(s.conf.Server.KeepAlive)

	if s.log != nil {
		s.log.Info(fmt.Sprintf("(HTTP) Starting server at %s, SSL mod: %v", addr, s.conf.Server.SSL))
	}

	go func() {
		if s.conf.Server.SSL {
			if err := s.httpSrv.ServeTLS(listener, s.conf.Server.CertFile, s.conf.Server.KeyFile); err != nil {
				if err != http.ErrServerClosed {
					if s.log != nil {
						s.log.Error("(HTTP) " + err.Error())
					}

					tui.PrintError(err)
					os.Exit(1)
				}
			}
		} else {
			if err := s.httpSrv.Serve(listener); err != nil {
				if err != http.ErrServerClosed {
					if s.log != nil {
						s.log.Error("(HTTP) " + err.Error())
					}

					tui.PrintError(err)
					os.Exit(1)
				}
			}
		}
	}()
}
