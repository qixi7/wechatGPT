package officalchatgpt

/*
	【官方API请求ChatGPT的版本】
*/

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
	"wechatGPT/msgdefine"

	"github.com/qixi7/xlog"
)

const baseUrl = "https://api.openai.com/v1/chat/completions"
const maxHistory = 0

// 聊天历史记录
var chatHistory = map[string][]MsgItem{}
var chatMtx sync.Mutex

// ChatGPTResponseBody 请求体
type ChatGPTResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChoiceItem           `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChoiceItem struct {
	Index        int     `json:"index"`
	Msg          MsgItem `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type MsgItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatGPTRequestBody 响应体
type ChatGPTRequestBody struct {
	Model string    `json:"model"`
	Msgs  []MsgItem `json:"messages"`
}

type officialHandler struct {
	model  string
	apiKey string
}

func NewOfficialHandler(model, apiKey string) *officialHandler {
	return &officialHandler{
		model:  model,
		apiKey: apiKey,
	}
}

// 官方API速度快, 可以直接很快返回
func (h *officialHandler) ReqChatGPT(wrapMsg *msgdefine.TextMsg) (string, error) {
	// 失败重试
	var err error
	var chatMsg string
	tryCount := 2
	tmpCount := tryCount
	for tmpCount > 0 {
		chatMsg, err = h.completions(wrapMsg.FromUserName, wrapMsg.Content)
		if err != nil {
			tmpCount--
			xlog.Errorf("ReqChatGPT nowTryCount=%d, err=%v", tmpCount, err)
			time.Sleep(time.Second * 5)
			continue
		}
		// 请求成功, 存一下历史记录
		chatMsg = strings.TrimSpace(chatMsg)
		dealChatHistory(wrapMsg.FromUserName, wrapMsg.Content, chatMsg)
		return chatMsg, nil
	}

	// 失败了也要发消息通知用户
	errMsg := fmt.Sprintf("重试次数超过%d次, 这条消息不会回复了Sorry. o(╥﹏╥)o", tryCount)
	errMsg += fmt.Sprintf("\nerr=%v", err)
	errMsg += "\n======================我是分割线======================\n\n"
	errMsg += wrapMsg.Content
	return errMsg, nil
}

// gtp文本模型回复
//curl https://api.openai.com/v1/chat/completions \
//-H 'Content-Type: application/json' \
//-H 'Authorization: Bearer YOUR_API_KEY' \
//-d '{
//"model": "gpt-3.5-turbo",
//"messages": [{"role": "user", "content": "Hello!"}]
//}'
func (h *officialHandler) completions(recver, msg string) (string, error) {
	requestBody := ChatGPTRequestBody{Model: h.model}
	requestBody.Msgs = fillChatHistory(recver, msg)
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	xlog.InfoF("request gtp json string : %v", string(requestData))
	req, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := h.apiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return "", errors.New(fmt.Sprintf("请求GTP出错了，gtp api status code not equals 200,code is %d ,details:  %v ", response.StatusCode, string(body)))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	xlog.InfoF(fmt.Sprintf("response gtp json string : %v", string(body)))

	gptResponseBody := &ChatGPTResponseBody{}
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Msg.Content
	} else {
		reply = "chatGPT 啥都没返回, 好奇怪0^0"
	}
	return reply, nil
}

func dealChatHistory(recver, send, recv string) {
	if maxHistory == 0 {
		return
	}
	oneAsk := MsgItem{
		Role:    "user",
		Content: send,
	}
	oneRecv := MsgItem{
		Role:    "assistant",
		Content: recv,
	}
	chatMtx.Lock()
	// 只保存前n条用户的提问, 作为每次提问的上下文
	if info, ok := chatHistory[recver]; ok && len(info) >= maxHistory {
		chatHistory[recver] = append(chatHistory[recver][:0], chatHistory[recver][2:]...)
	}
	chatHistory[recver] = append(chatHistory[recver], oneAsk)
	chatHistory[recver] = append(chatHistory[recver], oneRecv)
	chatMtx.Unlock()
}

func fillChatHistory(recver, msg string) []MsgItem {
	var msgArr []MsgItem
	chatMtx.Lock()
	// 只保存前n条用户的提问, 作为每次提问的上下文
	for _, ask := range chatHistory[recver] {
		msgArr = append(msgArr, ask)
	}
	chatMtx.Unlock()
	msgArr = append(msgArr, MsgItem{
		Role:    "user",
		Content: msg,
	})
	return msgArr
}
