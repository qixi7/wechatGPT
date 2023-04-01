package server

import (
	"github.com/qixi7/xlog"
	"net/http"
)

type HttpServer struct {
	httpSvr    http.Server
	handlerMap map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewHttpServer(addr string) *HttpServer {
	s := &HttpServer{
		httpSvr:    http.Server{Addr: addr, Handler: http.NewServeMux()},
		handlerMap: make(map[string]func(w http.ResponseWriter, r *http.Request)),
	}
	s.registerHttpHandler()
	return s
}

func (s *HttpServer) StartServer() {
	svrMtx := s.httpSvr.Handler.(*http.ServeMux)
	for p, h := range s.handlerMap {
		svrMtx.HandleFunc(p, h)
	}

	xlog.InfoF("httpSvr listen: %s", s.httpSvr.Addr)
	err := s.httpSvr.ListenAndServe()
	if err != nil {
		xlog.Errorf("httpSvr ListenAndServe err=%v", err)
	}
}

func (s *HttpServer) handleFunc(p string, h func(w http.ResponseWriter, r *http.Request)) {
	s.handlerMap[p] = h
}
