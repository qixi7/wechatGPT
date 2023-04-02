package server

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"wechatGPT/chatgpt/officalchatgpt"
	"wechatGPT/config"
	"wechatGPT/msgdefine"
	"wechatGPT/util"

	"github.com/qixi7/xlog"
)

var drawPicRe *regexp.Regexp
var msgID2Ret sync.Map     // key->msgID, value->chatGTP结果
var msgID2ReEnter sync.Map // key->msgID, value->是否重试消息

const maxWXMsgSize = 2048

func init() {
	drawPicRe = regexp.MustCompile(`^@pic(.*)`)
}

func testChat(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		verifiyUrl(w, r)
	case "POST":
		replyMsg(w, r)
	}
	r.Body.Close()
}

func sendWXMsg(w http.ResponseWriter, from, to, msg string) {
	// 分段发送
	for len(msg) > maxWXMsgSize {
		sendWXOne(w, from, to, msg[:maxWXMsgSize])
		msg = msg[maxWXMsgSize:]
	}
	if len(msg) > 0 {
		sendWXOne(w, from, to, msg)
	}
}

func sendWXOne(w http.ResponseWriter, from, to, msg string) {
	textRes := &msgdefine.TextRes{
		ToUserName:   to,
		FromUserName: from,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      msg,
	}
	_, err := w.Write(textRes.ToXml())
	if err != nil {
		xlog.Errorf("write back to wx err=%v", err)
		return
	}
	xlog.InfoF("send msg success!")
}

func replyMsg(w http.ResponseWriter, r *http.Request) {
	xlog.InfoF("replyMsg called!")
	// 解析消息
	body, _ := ioutil.ReadAll(r.Body)
	xmlMsg := msgdefine.ToTextMsg(body)

	xlog.InfoF("[消息接收] Type: %s, Event=%s, MsgID=%d, From: %s, To: %s, Content: %s",
		xmlMsg.MsgType, xmlMsg.Event, xmlMsg.MsgId, xmlMsg.FromUserName, xmlMsg.ToUserName, xmlMsg.Content)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// 回复消息
	replyStr := "Unknown"
	// 查找该msgID是否已经在发起请求了
	chatRet, exist := msgID2Ret.LoadOrStore(xmlMsg.MsgId, "")
	if exist {
		// 重试的微信消息
		msgID2ReEnter.Store(xmlMsg.MsgId, true)
		if chatRet.(string) == "" {
			// 正在请求中
			time.Sleep(6 * time.Second)
			return
		}
		// 已经有了结果, 但是还没通知
		replyStr = chatRet.(string)
		msgID2Ret.Delete(xmlMsg.MsgId)
		msgID2ReEnter.Delete(xmlMsg.MsgId)
	} else {
		// 不存在已有, 第一次需要请求chatGPT
		switch xmlMsg.MsgType {
		case "event":
			if xmlMsg.Event == "subscribe" {
				replyStr = "靓女, 恭喜你发现了新大陆, 现在来问我问题吧~"
			}
		case "text":
			// 【收到不支持的消息类型，暂无法显示】
			if strings.Contains(xmlMsg.Content, "【收到不支持的消息类型，暂无法显示】") {
				xlog.InfoF("recv not support msg.")
				return
			}
			// 去chatgpt请求, 最多等待 15s, 超时返回空值
			chatHandler := officalchatgpt.NewOfficialHandler(
				"gpt-3.5-turbo",
				config.Get().ApiKey)
			replyStr, _ = chatHandler.ReqChatGPT(xmlMsg)
			msgID2Ret.Store(xmlMsg.MsgId, replyStr)
		}
	}

	// 如果有微信重试消息过来, 那么这次send会失败
	if _, ok := msgID2ReEnter.Load(xmlMsg.MsgId); ok {
		time.Sleep(6 * time.Second)
		return
	}

	sendWXMsg(w, xmlMsg.ToUserName, xmlMsg.FromUserName, replyStr)
}

func verifiyUrl(w http.ResponseWriter, r *http.Request) {
	xlog.InfoF("verifiyUrl called!")
	sign := getUrlArg(r, "signature")
	timestamp := getUrlArg(r, "timestamp")
	nonce := getUrlArg(r, "nonce")
	echoStr := getUrlArg(r, "echostr")
	xlog.InfoF("sign=%s", sign)
	xlog.InfoF("timestamp=%s", timestamp)
	xlog.InfoF("nonce=%s", nonce)
	xlog.InfoF("echoStr=%s", echoStr)

	wxToken := config.Get().EncryptToken
	// 校验
	if util.CheckSignature(sign, timestamp, nonce, wxToken) {
		util.PlainText(w, r, echoStr)
		xlog.InfoF("verifyUrl success")
		return
	}
	xlog.InfoF("verifyUrl failed!")
}

//获取URL的GET参数
func getUrlArg(r *http.Request, name string) string {
	var arg string
	values := r.URL.Query()
	arg = values.Get(name)
	return arg
}

func (s *HttpServer) registerHttpHandler() {
	s.handleFunc("/testchat", testChat)
}
