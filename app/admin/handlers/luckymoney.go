package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 获取红包请求
type GetLuckymoneyRequest struct {
	UserID int64 `json:"user_id"` // 用户ID
	Offset uint  `json:"offset"`  // 偏移量
	Limit  uint  `json:"limit"`   // 返回数量
	Tonce  int64 `json:"tonce"`   // 时间戳
}

// 获取红包响应
type GetLuckymoneyRespone struct {
	Sum    int                  `json:"sum"`    // 动作总量
	Count  int                  `json:"count"`  // 返回数量
	Result []*models.LuckyMoney `json:"result"` // 红包列表
}

// 获取红包信息
func GetLuckymoney(w http.ResponseWriter, r *http.Request) {
	// 跨域访问
	allowAccessControl(w)

	// 验证权限
	sessionID, data, ok := authentication(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(makeErrorRespone("", ""))
		return
	}

	// 解析请求参数
	var request GetLuckymoneyRequest
	if err := json.Unmarshal(data, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 获取红包列表
	model := models.LuckyMoneyModel{}
	ids, sum, err := model.Collection(request.UserID, true, uint(request.Offset), uint(request.Limit), true)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	if request.Offset >= sum {
		request.Offset = request.Offset - sum
	} else if len(ids) < int(request.Limit) {
		request.Offset = 0
		request.Limit = uint(int(request.Limit) - len(ids))
	} else {
		request.Limit = 0
		request.Offset = 0
	}

	historyIds, historySum, err := model.Collection(request.UserID, false, uint(request.Offset), uint(request.Limit), true)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 获取红包数据
	idset := make([]uint64, 0, len(ids)+len(historyIds))
	copy(idset, ids)
	copy(idset[len(ids):], historyIds)
	result := make([]*models.LuckyMoney, 0, len(ids)+len(historyIds))
	for i := 0; i < len(idset); i++ {
		data, _, err := model.GetLuckyMoney(idset[i])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(makeErrorRespone(sessionID, err.Error()))
			return
		}
		result = append(result, data)
	}

	// 序列化结果
	respone := GetLuckymoneyRespone{Sum: int(sum + historySum), Count: len(result), Result: nil}
	jsb, err := json.Marshal(respone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 返回余额信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(makeRespone(sessionID, jsb))
}
