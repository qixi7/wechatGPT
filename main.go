package main

import (
	"fmt"
	"github.com/qixi7/xlog"
	"net/http"
	"time"
	"wechatGPT/config"
	"wechatGPT/server"
)

func main() {
	xlog.InfoF("hello world!")
	http.DefaultClient.Timeout = 2 * time.Minute
	//http.DefaultTransport.(*http.Transport).TLSClientConfig =
	//	&tls.Config{InsecureSkipVerify: true}

	httpPort := config.Get().HttpPort
	httpSvr := server.NewHttpServer(fmt.Sprintf(":%d", httpPort))
	httpSvr.StartServer()
}
