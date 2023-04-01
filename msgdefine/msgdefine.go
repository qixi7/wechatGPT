package msgdefine

import "encoding/xml"

// 微信发来的消息
type TextMsg struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	MsgId        int64  `xml:"MsgId"`
	MsgDataId    int64  `xml:"MsgDataId"`
	Idx          int64  `xml:"Idx"`
	Event        string `xml:"Event"`
}

func ToTextMsg(body []byte) *TextMsg {
	var msg TextMsg
	err := xml.Unmarshal(body, &msg)
	if err != nil {
		panic(err)
	}
	return &msg
}

func (msg *TextMsg) ToXml() []byte {
	body, err := xml.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}

// 回复给微信的消息
type TextRes struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

func ToTextRes(body []byte) *TextRes {
	var msg TextRes
	err := xml.Unmarshal(body, &msg)
	if err != nil {
		panic(err)
	}
	return &msg
}

func (msg *TextRes) ToXml() []byte {
	body, err := xml.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}