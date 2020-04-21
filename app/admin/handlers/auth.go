package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
	"luckybot/app/admin/crypto"
	"luckybot/app/config"
)

var once sync.Once
var authenticator *Authenticator

// 会话
type Session struct {
	ID        string
	IP        string
	Key       string
	CreatedAt int64
	ExpiredAt int64
}

// 密文消息
type Ciphertext struct {
	ACK      bool   `json:"ack"`
	Session  string `json:"session,omitempty"`
	Data     string `json:"data"`
	Checksum uint32 `json:"checksum"`
}

// 解密消息
func (msg *Ciphertext) Decode(key [16]byte) ([]byte, bool) {
	if len(msg.Data) == 0 {
		return nil, false
	}

	src, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		return nil, false
	}

	checksum := crc32.ChecksumIEEE(src)
	if checksum != msg.Checksum {
		return nil, false
	}

	data, err := crypto.AesDecrypt(src, key)
	if err != nil {
		return nil, false
	}
	return data, true
}

// 加密消息
func (msg *Ciphertext) Encode(src []byte, key [16]byte) error {
	data, err := crypto.AesEncrypt(src, key)
	if err != nil {
		return err
	}
	msg.Checksum = crc32.ChecksumIEEE(data)
	msg.Data = base64.StdEncoding.EncodeToString(data)
	return nil
}

// 验证器
type Authenticator struct {
	mutex    sync.Mutex
	codes    sync.Map
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

// 获取密钥
func (a *Authenticator) getKey(code string) [16]byte {
	var key [16]byte
	for i := 0; i < len(key); i++ {
		if i >= len(code) {
			key[i] = byte('0')
			continue
		}
		key[i] = byte(code[i])
	}
	return key
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

// 验证身份
func (a *Authenticator) Auth(ip string, data []byte) (string, bool) {
	var msg Ciphertext
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", false
	}

	serveCfg := config.GetServe()
	code, err := totp.GenerateCode(serveCfg.SecretKey, time.Now())
	if err != nil {
		return "", false
	}

	key := a.getKey(code)
	src, ok := msg.Decode(key)
	if !ok {
		return "", false
	}

	var request AuthRequest
	if err := json.Unmarshal(src, &request); err != nil {
		return "", false
	}

	var buf [16]byte
	rand.Read(buf[:])
	session := Session{
		ID:        hex.EncodeToString(buf[:]),
		IP:        ip,
		Key:       code,
		CreatedAt: time.Now().Unix(),
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

	jsb, ok := msg.Decode(a.getKey(session.Key))
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
		msg := Ciphertext{}
		msg.ACK = false
		jsb, _ := json.Marshal(&msg)
		return jsb
	}
	a.mutex.Unlock()

	msg := Ciphertext{}
	msg.ACK = true
	msg.Encode(data, a.getKey(session.Key))
	jsb, _ := json.Marshal(&msg)
	return jsb
}

// 身份认证请求
type AuthRequest struct {
	Tonce int64 `json:"tonce"` // 时间戳
}

// 身份认证响应
type AuthRespone struct {
	SessionID string `json:"session_id"` // 会话ID
}

// 身份认证
func Authentication(w http.ResponseWriter, r *http.Request) {
	// 跨域访问
	allowAccessControl(w)

	// 解析请求参数
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(makeErrorRespone("", ""))
		return
	}

	// 验证数据合法性
	address := strings.Split(r.RemoteAddr, ":")[0]
	sessionID, ok := authenticator.Auth(address, data)
	if !ok {
		w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
	w.Write(makeRespone(sessionID, jsb))
}
