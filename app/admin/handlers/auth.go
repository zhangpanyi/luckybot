package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/zhangpanyi/luckybot/app/admin/crypto"
	"github.com/zhangpanyi/luckybot/app/config"
)

var once sync.Once
var authenticator *Authenticator

// 会话
type Session struct {
	ID        string
	IP        string
	Key       string
	ExpiredAt int64
}

// 密文消息
type Ciphertext struct {
	ACK      bool   `json:"ack,omitempty"`
	Session  string `json:"session,omitempty"`
	Data     string `json:"data"`
	Checksum uint32 `json:"checksum"`
}

// 解密消息
func (msg *Ciphertext) Decode(key [16]byte) ([]byte, bool) {
	src, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		return nil, false
	}
	if crc32.ChecksumIEEE(src) != msg.Checksum {
		return nil, false
	}
	data := crypto.DecryptAES(src, key)
	return data, true
}

// 加密消息
func (msg *Ciphertext) Encode(src []byte, key [16]byte) {
	data := crypto.DncryptAES(src, key)
	msg.Checksum = crc32.ChecksumIEEE(data)
	msg.Data = base64.StdEncoding.EncodeToString(data)
}

// 验证器
type Authenticator struct {
	mutex    sync.Mutex
	sessions map[string]*Session
}

// 创建验证器
func NewAuthenticatorOnce() {
	once.Do(func() {
		authenticator = &Authenticator{
			sessions: make(map[string]*Session),
		}
		go authenticator.checkExpiredSessions()
	})
}

// 验证身份
func (a *Authenticator) Auth(ip string, code string) (string, bool) {
	serveCfg := config.GetServe()
	ok := totp.Validate(code, serveCfg.SecretKey)
	if !ok {
		return "", false
	}

	var buf [32]byte
	rand.Read(buf[:])
	session := Session{
		ID:        hex.EncodeToString(buf[:]),
		IP:        ip,
		Key:       code,
		ExpiredAt: time.Now().Add(time.Minute * 5).Unix(),
	}

	a.mutex.Lock()
	a.sessions[session.ID] = &session
	defer a.mutex.Unlock()
	return session.ID, true
}

// 保持活跃
func (a *Authenticator) KeepAlive(sessionID string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	session, ok := a.sessions[sessionID]
	if !ok {
		return
	}
	session.ExpiredAt += 60 * 5
}

// 解码消息
func (a *Authenticator) Decode(data []byte) (string, []byte, bool) {
	var msg Ciphertext
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", nil, false
	}

	a.mutex.Lock()
	session, ok := a.sessions[msg.Session]
	if !ok {
		a.mutex.Unlock()
		return "", nil, false
	}
	a.mutex.Unlock()

	var key [16]byte
	for i := 0; i < len(key); i++ {
		if 1 >= len(session.Key) {
			break
		}
		key[i] = byte(session.Key[i])
	}

	jsb, ok := msg.Decode(key)
	if !ok {
		return "", nil, false
	}
	return session.ID, jsb, true
}

// 编码消息
func (a *Authenticator) Encode(sessionID string, data []byte) []byte {
	a.mutex.Lock()
	session, ok := a.sessions[sessionID]
	if !ok {
		a.mutex.Unlock()
		msg := Ciphertext{ACK: false}
		jsb, _ := json.Marshal(&msg)
		return jsb
	}
	a.mutex.Unlock()

	var key [16]byte
	for i := 0; i < len(key); i++ {
		if 1 >= len(session.Key) {
			break
		}
		key[i] = byte(session.Key[i])
	}

	msg := Ciphertext{ACK: true}
	msg.Encode(data, key)
	jsb, _ := json.Marshal(&msg)
	return jsb
}

// 检查Sessions
func (a *Authenticator) checkExpiredSessions() {
	ticker := time.NewTicker(time.Second * 1)
	for range ticker.C {
		a.mutex.Lock()
		now := time.Now().Unix()
		for k, v := range a.sessions {
			if v.ExpiredAt <= now {
				delete(a.sessions, k)
			}
		}
		a.mutex.Unlock()
	}
}

// 身份认证请求
type AuthRequest struct {
	Code  string `json:"code"`  // 验证码
	Tonce int64  `json:"tonce"` // 时间戳
}

// 身份认证响应
type AuthRespone struct {
	SessionID string `json:"session_id"` // 会话ID
}

// 身份认证
func Authentication(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone("", ""))
		return
	}

	var request AuthRequest
	if err := json.Unmarshal(data, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone("", ""))
		return
	}

	sessionID, ok := authenticator.Auth("192.168.121", request.Code)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(makeErrorRespone("", ""))
		return
	}

	respone := AuthRespone{SessionID: sessionID}
	jsb, err := json.Marshal(&respone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(makeRespone(sessionID, jsb))
}
