package config

import (
	"encoding/json"
	"github.com/qixi7/xlog"
	"os"
)

type ChatConfig struct {
	ProxyUrl      string `json:"proxy_url"`       // 代理url
	ApiKey        string `json:"api_key"`         // gtp apikey
	HttpPort      int32  `json:"http_port"`       // 监听的http端口
	EncryptToken  string `json:"encrypt_token"`   // 加密Token
	EncryptAESKey string `json:"encrypt_aes_key"` // AES Key
}

var config *ChatConfig

func Get() *ChatConfig {
	return config
}

func init() {
	// 默认值
	config = &ChatConfig{
		HttpPort: 80,
	}
	f, err := os.Open("config.json")
	if err != nil {
		xlog.Errorf("open config err: %v", err)
		return
	}
	defer f.Close()
	encoder := json.NewDecoder(f)
	err = encoder.Decode(config)
	if err != nil {
		xlog.Warnf("decode config err: %v", err)
		return
	}
	xlog.InfoF("config is : %v", config)
}
