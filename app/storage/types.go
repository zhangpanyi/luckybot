package storage

// 资产信息
type Asset struct {
	Asset  string `json:"asset"`  // 资产名称
	Amount uint32 `json:"amount"` // 资产总额
	Freeze uint32 `json:"freeze"` // 冻结资产
}

// 红包信息
type LuckyMoney struct {
	ID         uint64 `json:"id"`          // 红包ID
	GroupID    int64  `json:"group_id"`    // 群组ID
	MessageID  int32  `json:"message_id"`  // 消息ID
	SenderID   int64  `json:"sneder_id"`   // 发送者
	SenderName string `json:"sneder_name"` // 发送者名字
	Asset      string `json:"asset"`       // 资产类型
	Amount     uint32 `json:"amount"`      // 红包总额
	Received   uint32 `json:"received"`    // 领取金额
	Number     uint32 `json:"number"`      // 红包个数
	Lucky      bool   `json:"lucky"`       // 是否随机
	Value      uint32 `json:"value"`       // 单个价值
	Active     bool   `json:"active"`      // 是否激活
	Memo       string `json:"memo"`        // 红包留言
	Timestamp  int64  `json:"timestamp"`   // 时间戳
}

// 红包用户
type LuckyMoneyUser struct {
	UserID    int64  `json:"user_id"`    // 用户ID
	FirstName string `json:"first_name"` // 用户名
}

// 红包记录
type LuckyMoneyRecord struct {
	Value int             `json:"value"`          // 红包金额
	User  *LuckyMoneyUser `json:"user,omitempty"` // 用户信息
}
