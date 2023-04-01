package util

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"sort"
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

// StatusCtxKey is a context key to record a future HTTP response status code.
var StatusCtxKey = &contextKey{"Status"}

// 检查微信验证签名
func CheckSignature(signature, timestamp, nonce, token string) bool {
	// 1）将token、timestamp、nonce三个参数进行字典序排序
	sl := []string{token, timestamp, nonce}
	sort.Strings(sl)
	// 2）将三个参数字符串拼接成一个字符串进行sha1加密
	sum := sha1.Sum([]byte(sl[0] + sl[1] + sl[2]))
	// 3）开发者获得加密后的字符串可与 signature 对比，标识该请求来源于微信
	return signature == hex.EncodeToString(sum[:])
}

// PlainText writes a string to the response, setting the Content-Type as
// text/plain.
func PlainText(w http.ResponseWriter, r *http.Request, v string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write([]byte(v))
}