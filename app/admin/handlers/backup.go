package handlers

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/zhangpanyi/luckymoney/app/storage"
)

// 备份数据库
func Backup(w http.ResponseWriter, r *http.Request) {
	// 验证权限
	if !authentication(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 备份数据库
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	size, err := storage.Backup(buf)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	// 返回数据库
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="master.db"`)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
