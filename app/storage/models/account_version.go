package models

import (
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/zhangpanyi/luckybot/app/storage"
)

// 触发原因
type Reason int

const (
	_                     Reason = iota
	ReasonGive                   // 发红包
	ReasonSystem                 // 系统发放
	ReasonReceive                // 领取红包
	ReasonGiveBack               // 退还红包
	ReasonDeposit                // 充值
	ReasonWithdraw               // 提现
	ReasonWithdrawSuccess        // 提现成功
	ReasonWithdrawFailure        // 提现失败
)

// 版本信息
type Version struct {
	ID              uint64     `json:"id"`                           // 版本ID
	Symbol          string     `json:"symbol"`                       // 代币符号
	Balance         *big.Float `json:"balance"`                      // 余额变化
	Locked          *big.Float `json:"locked"`                       // 锁定变化
	Fee             *big.Float `json:"fee"`                          // 手续费
	Amount          *big.Float `json:"amount"`                       // 剩余金额
	Timestamp       int64      `json:"Timestamp"`                    // 时间戳
	Reason          Reason     `json:"reason"`                       // 触发原因
	RefLuckyMoneyID *uint64    `json:"ref_lucky_money_id,omitempty"` // 关联红包ID
	RefBlockHeight  *uint64    `json:"ref_block_height,omitempty"`   // 关联区块高度
	RefTxID         *string    `json:"ref_tx_id,omitempty"`          // 关联交易ID
	RefUserID       *int64     `json:"ref_user_id,omitempty"`        // 关联用户ID
	RefUserName     *string    `json:"ref_user_name,omitempty"`      // 关联用户名
	RefAddress      *string    `json:"ref_address,omitempty"`        // 关联地址
	RefMemo         *string    `json:"ref_memo,omitempty"`           // 关联备注信息
}

// ********************** 结构图 **********************
// {
//	"account_versions": {
// 		<user_id>: {
// 			<seq>: Version	// 版本信息
// 		}
//	}
// ***************************************************

// 账户版本模型
type AccountVersionModel struct {
}

// 插入版本
func (model *AccountVersionModel) InsertVersion(userID int64, version *Version) (*Version, error) {
	key := strconv.FormatInt(userID, 10)
	version.Timestamp = time.Now().UTC().Unix()
	err := storage.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "account_versions", key)
		if err != nil {
			return err
		}

		seq, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		version.ID = seq
		jsb, err := json.Marshal(version)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(strconv.FormatUint(seq, 10)), jsb)
	})

	if err != nil {
		return nil, err
	}

	return version, nil
}

// 获取版本
func (model *AccountVersionModel) GetVersions(userID int64, offset, limit uint, reverse bool) ([]*Version, int, error) {
	sum := 0
	jsonarray := make([][]byte, 0)
	key := strconv.FormatInt(userID, 10)
	err := storage.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := storage.GetBucketIfExists(tx, "account_versions", key)
		if err != nil {
			if err != storage.ErrNoBucket {
				return err
			}
			return nil
		}

		var idx uint
		filter := func(idx uint, k, v []byte) bool {
			if v != nil {
				if idx >= offset {
					jsonarray = append(jsonarray, v)
					if len(jsonarray) >= int(limit) {
						return false
					}
				}
				idx++
			}
			return true
		}

		if reverse {
			for i := bucket.Sequence(); i >= uint64(1); i-- {
				k := []byte(strconv.FormatUint(i, 10))
				if !filter(idx, k, bucket.Get(k)) {
					break
				}
				idx++
			}
		} else {
			for i := uint64(1); i <= bucket.Sequence(); i++ {
				k := []byte(strconv.FormatUint(i, 10))
				if !filter(idx, k, bucket.Get(k)) {
					break
				}
				idx++
			}
		}
		sum = bucket.Stats().KeyN
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	versions := make([]*Version, 0)
	for i := 0; i < len(jsonarray); i++ {
		jsb := jsonarray[i]
		var version Version
		if err = json.Unmarshal(jsb, &version); err != nil {
			return nil, 0, err
		}
		versions = append(versions, &version)
	}
	return versions, sum, nil
}
