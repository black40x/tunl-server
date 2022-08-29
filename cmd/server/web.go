package server

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/black40x/tunl-core/commands"
	"github.com/black40x/tunl-core/tunl"
	"github.com/black40x/tunl-server/cmd/tui"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/net/netutil"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

type TunlHttp struct {
	tunl    *TunlServer
	conf    *Config
	httpSrv *http.Server
	ctx     context.Context
}

type JsonData map[string]interface{}

func NewTunlHttp(conf *Config, ctx context.Context) *TunlHttp {
	return &TunlHttp{
		conf: conf,
		ctx:  ctx,
	}
}

func (s *TunlHttp) responseJSON(w http.ResponseWriter, status int, data JsonData, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (s *TunlHttp) handle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if params["subdomain"] == "" {
		s.responseJSON(w, http.StatusForbidden, JsonData{"error": "Empty client ID"}, nil)
	} else {
		if s.tunl.pool.Get(params["subdomain"]) == nil {
			s.responseJSON(w, http.StatusForbidden, JsonData{"error": "Undefined client ID"}, nil)
		} else {
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

			_, err := s.tunl.pool.Get(params["subdomain"]).Send(&req)
			if err != nil {
				s.responseJSON(w, http.StatusBadRequest, JsonData{"error": "Can't send request to the client"}, nil)
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
							s.tunl.pool.Get(params["subdomain"]).Send(&commands.BodyChunk{
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
							case <-time.After(time.Second * 60): // ToDo - increase wait time + add chanel client disconnect in pool!!!
								break L
							}
						}
					}
				case <-time.After(time.Second * 30):
					s.responseJSON(w, http.StatusBadRequest, JsonData{"error": "Can't receive request from client"}, nil)
				}
			}
		}
	}
}

func (s *TunlHttp) Shutdown() {
	if s.httpSrv != nil {
		s.httpSrv.Shutdown(s.ctx)
	}
}

func (s *TunlHttp) startTunl() {
	s.tunl = NewTunlServer(s.conf.Tunl)
	go func() {
		err := s.tunl.Run(s.conf)

		if err != nil {
			tui.PrintError(err)
			os.Exit(1)
		}
	}()
}

func (s *TunlHttp) Start() {
	s.startTunl()

	r := mux.NewRouter()
	r.Host(s.conf.Tunl.ClientSubDomain).HandlerFunc(s.handle)

	addr := s.conf.Base.HTTPAddr + ":" + s.conf.Base.HTTPPort
	concurrency := runtime.NumCPU() * 2
	listener, _ := net.Listen("tcp", addr)
	listener = netutil.LimitListener(listener, concurrency*s.conf.Base.ThreadCount)

	s.httpSrv = &http.Server{
		Addr:        addr,
		Handler:     r,
		ReadTimeout: time.Duration(s.conf.Base.ReadTimeout) * time.Second,
		IdleTimeout: time.Duration(s.conf.Base.IdleTimeout) * time.Second,
	}
	s.httpSrv.SetKeepAlivesEnabled(s.conf.Base.KeepAlive)

	go func() {
		if s.conf.Base.SSL {
			if err := s.httpSrv.ServeTLS(listener, s.conf.Base.CertFile, s.conf.Base.KeyFile); err != nil {
				if err != http.ErrServerClosed {
					tui.PrintError(err)
					os.Exit(1)
				}
			}
		} else {
			if err := s.httpSrv.Serve(listener); err != nil {
				if err != http.ErrServerClosed {
					tui.PrintError(err)
					os.Exit(1)
				}
			}
		}
	}()
}
