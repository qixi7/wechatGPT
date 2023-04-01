package main

import (
	"fmt"
	"github.com/qixi7/xlog"
	"net/http"
	"net/url"
	"time"
	"wechatGPT/config"
	"wechatGPT/server"
)

func main() {
	xlog.InfoF("hello world!")
	http.DefaultClient.Timeout = 2 * time.Minute
	proxyStr := config.Get().ProxyUrl
	if proxyStr != "" {
		proxyUrl, err := url.Parse(proxyStr)
		if err != nil {
			xlog.Errorf("proxyUrl err=%v", err)
			return
		}
		http.DefaultClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		xlog.InfoF("Set ProxyUrl=%s", proxyStr)
	}

	//http.DefaultTransport.(*http.Transport).TLSClientConfig =
	//	&tls.Config{InsecureSkipVerify: true}

	httpPort := config.Get().HttpPort
	httpSvr := server.NewHttpServer(fmt.Sprintf(":%d", httpPort))
	httpSvr.StartServer()
}
