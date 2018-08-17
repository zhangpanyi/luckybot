package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"github.com/boltdb/bolt"
)

var (
	// 领完了
	ErrNothingLeft = errors.New("nothing left")
	// 重复领取
	ErrRepeatReceive = errors.New("repeat receive")
	// 没有激活
	ErrNotActivated = errors.New("not activated")
	// 已经激活
	ErrAlreadyActivated = errors.New("already activated")
	// 没有权限
	ErrPermissionDenied = errors.New("permission denied")
	// 红包已过期
	ErrLuckyMoneydExpired = errors.New("lucky money expired")
)

// ********************** 结构图 **********************
// {
// 	"luckymoney": {
// 		"sid": {
// 			"seq": 0,					// 红包领取序列
// 			"worst": 0,					// 手气最烂序列
// 			"best": 0,					// 手气最佳序列
// 			"base": types.LuckyMoney	// 红包基本信息
//			"users": {					// 红包已领用户
//				"user_id": ""
//			}
// 			"record": {					// 红包领取记录
// 				"seq": types.LuckyMoneyRecord
// 			}
//			"expired": true				// 红包是否过期
// 		},
//		"mapping": {					// 红包编号映射
//			<sn>: <sid>
//		},
//		"sequeue": 0,					// 红包ID生成序列
//		"last_expired": 0,				// 上次过期红包ID
// 	}
// }
// ***************************************************

// 红包存储
type LuckyMoneyStorage struct {
}

// 创建新红包
func (handler *LuckyMoneyStorage) NewLuckyMoney(data *LuckyMoney, luckyMoneyArr []int) (*LuckyMoney, error) {
	err := blotDB.Update(func(tx *bolt.Tx) error {
		// 生成红包ID
		rootBucket, err := ensureBucketExists(tx, "luckymoney")
		if err != nil {
			return err
		}
		if rootBucket.Sequence() < DefaultLuckyMoneyID {
			if err = rootBucket.SetSequence(DefaultLuckyMoneyID); err != nil {
				return err
			}
		}
		data.ID, err = rootBucket.NextSequence()
		if err != nil {
			return err
		}

		// 生成序列号
		sn, err := handler.generateSN(tx, data.ID)
		if err != nil {
			return err
		}
		data.SN = sn

		// 序列化数据
		data.Received = 0
		data.Active = false
		jsb, err := json.Marshal(data)
		if err != nil {
			return err
		}

		// 插入基本信息
		sid := strconv.FormatUint(data.ID, 10)
		bucket, err := ensureBucketExists(tx, "luckymoney", sid)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte("base"), jsb)
		if err != nil {
			return err
		}

		// 插入领取用户
		_, err = ensureBucketExists(tx, "luckymoney", sid, "users")
		if err != nil {
			return err
		}

		// 插入领取记录
		worstSeq, bestSeq, err := handler.insertRecord(tx, sid, luckyMoneyArr)
		if err != nil {
			return err
		}

		// 插入已领取序列
		err = bucket.Put([]byte("seq"), []byte("0"))
		if err != nil {
			return err
		}

		// 插入手气最佳序列
		err = bucket.Put([]byte("best"), []byte(strconv.Itoa(bestSeq)))
		if err != nil {
			return err
		}

		// 插入手气最烂序列
		err = bucket.Put([]byte("worst"), []byte(strconv.Itoa(worstSeq)))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return data, nil
}

// 是否过期
func (handler *LuckyMoneyStorage) IsExpired(id uint64) bool {
	var expired bool
	sid := strconv.FormatUint(id, 10)
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", sid)
		if err != nil {
			return err
		}
		expired = bucket.Get([]byte("expired")) != nil
		return nil
	})

	if err != nil {
		return false
	}
	return expired
}

// 设置过期
func (handler *LuckyMoneyStorage) SetExpired(id uint64) error {
	sid := strconv.FormatUint(id, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", sid)
		if err != nil {
			return err
		}
		return bucket.Put([]byte("expired"), []byte("true"))
	})
}

// 是否已领取
func (handler *LuckyMoneyStorage) IsReceived(id uint64, userID int64) (bool, error) {
	received := false
	sid := strconv.FormatUint(id, 10)
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", sid, "users")
		if err != nil {
			return err
		}
		received = bucket.Get([]byte(strconv.FormatInt(userID, 10))) != nil
		return nil
	})
	if err != nil {
		return received, err
	}
	return received, nil
}

// 获取上次过期红包
func (handler *LuckyMoneyStorage) GetLastExpired() (uint64, error) {
	var id uint64
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney")
		if err != nil {
			return err
		}

		sid := bucket.Get([]byte("last_expired"))
		if sid == nil {
			return nil
		}

		id, err = strconv.ParseUint(string(sid), 10, 64)
		if err != nil {
			return nil
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return id, nil
}

// 设置上次过期红包
func (handler *LuckyMoneyStorage) SetLastExpired(id uint64) error {
	sid := strconv.FormatUint(id, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney")
		if err != nil {
			return err
		}
		return bucket.Put([]byte("last_expired"), []byte(sid))
	})
}

// 获取红包信息
func (handler *LuckyMoneyStorage) GetLuckyMoney(id uint64) (*LuckyMoney, uint32, error) {
	var received uint32
	var base LuckyMoney
	sid := strconv.FormatUint(id, 10)
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", sid)
		if err != nil {
			return err
		}

		// 获取红包信息
		jsb := bucket.Get([]byte("base"))
		if err = json.Unmarshal(jsb, &base); err != nil {
			return err
		}

		// 已领取数量
		seq := bucket.Get([]byte("seq"))
		numReceived, err := strconv.Atoi(string(seq))
		if err != nil {
			return err
		}

		// 剩余红包数量
		received = uint32(numReceived)
		return nil
	})

	if err != nil {
		return nil, 0, err
	}
	return &base, received, nil
}

// 根据SN获取红包ID
func (handler *LuckyMoneyStorage) GetLuckyMoneyIDBySN(sn string) (uint64, error) {
	var id uint64
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", "mapping")
		if err != nil {
			return err
		}

		value := bucket.Get([]byte(sn))
		if value == nil {
			return errors.New("not found")
		}

		id, err = strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return id, nil
}

// 激活红包
func (handler *LuckyMoneyStorage) ActiveLuckyMoney(id uint64, userID, chatID int64, messageID int32) error {
	sid := strconv.FormatUint(id, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", sid)
		if err != nil {
			return err
		}

		// 检查状态
		if bucket.Get([]byte("expired")) != nil {
			return ErrLuckyMoneydExpired
		}

		// 获取红包基本信息
		var base LuckyMoney
		jsb := bucket.Get([]byte("base"))
		if err = json.Unmarshal(jsb, &base); err != nil {
			return err
		}

		if base.Active {
			return ErrAlreadyActivated
		}

		if base.SenderID != userID {
			return ErrPermissionDenied
		}

		// 更新红包基本信息
		base.Active = true
		base.GroupID = chatID
		base.MessageID = messageID
		if jsb, err = json.Marshal(&base); err != nil {
			return err
		}
		return bucket.Put([]byte("base"), jsb)
	})
}

// 领取红包
func (handler *LuckyMoneyStorage) ReceiveLuckyMoney(id uint64, userID int64, firstName string) (int, int, error) {

	received, err := handler.IsReceived(id, userID)
	if err != nil {
		return 0, 0, err
	}

	if received {
		return 0, 0, ErrRepeatReceive
	}

	value := 0
	count := 0
	sid := strconv.FormatUint(id, 10)
	err = blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", sid)
		if err != nil {
			return err
		}

		// 检查状态
		if bucket.Get([]byte("expired")) != nil {
			return ErrLuckyMoneydExpired
		}

		// 已领取数量
		seq := bucket.Get([]byte("seq"))
		numReceived, err := strconv.Atoi(string(seq))
		if err != nil {
			return err
		}

		// 红包是否充足
		var base LuckyMoney
		jsb := bucket.Get([]byte("base"))
		if err = json.Unmarshal(jsb, &base); err != nil {
			return err
		}
		if uint32(numReceived) >= base.Number {
			return ErrNothingLeft
		}

		// 红包是否激活
		if !base.Active {
			return ErrNotActivated
		}

		// 是否重复领取
		usersBucket, err := getBucketIfExists(tx, "luckymoney", sid, "users")
		if err != nil {
			return err
		}
		key := []byte(strconv.FormatInt(userID, 10))
		if usersBucket.Get(key) != nil {
			return ErrRepeatReceive
		}

		// 执行领取红包
		newSeq := numReceived + 1
		value, err = handler.receiveLuckyMoney(tx, sid, newSeq, &LuckyMoneyUser{
			UserID:    userID,
			FirstName: firstName,
		})
		if err != nil {
			return err
		}
		base.Received += uint32(value)

		// 更新红包信息
		if jsb, err = json.Marshal(&base); err != nil {
			return err
		}
		if err = bucket.Put([]byte("base"), jsb); err != nil {
			return err
		}
		if err = usersBucket.Put(key, []byte("")); err != nil {
			return err
		}
		if err = bucket.Put([]byte("seq"), []byte(strconv.Itoa(newSeq))); err != nil {
			return err
		}

		count = int(base.Number - uint32(newSeq))
		return nil
	})

	if err != nil {
		return 0, 0, err
	}
	return value, count, nil
}

// 获取最佳红包
func (handler *LuckyMoneyStorage) GetBestAndWorst(id uint64) (*LuckyMoneyRecord, *LuckyMoneyRecord, error) {
	var minRecord LuckyMoneyRecord
	var maxRecord LuckyMoneyRecord
	sid := strconv.FormatUint(id, 10)
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "luckymoney", sid)
		if err != nil {
			return err
		}

		// 获取序列号
		bestSeq := bucket.Get([]byte("best"))
		worstSeq := bucket.Get([]byte("worst"))
		if worstSeq == nil || bestSeq == nil {
			return errors.New("nou found")
		}

		// 获取红包信息
		recordBucket, err := getBucketIfExists(tx, "luckymoney", sid, "record")
		if err != nil {
			return err
		}

		minValueData := recordBucket.Get(worstSeq)
		maxValueData := recordBucket.Get(bestSeq)
		if minValueData == nil || maxValueData == nil {
			return errors.New("nou found")
		}

		if err = json.Unmarshal(minValueData, &minRecord); err != nil {
			return err
		}
		if err = json.Unmarshal(maxValueData, &maxRecord); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return &minRecord, &maxRecord, nil
}

// 遍历红包列表
func (handler *LuckyMoneyStorage) ForeachLuckyMoney(startID uint64, callback func(*LuckyMoney)) error {
	var base LuckyMoney
	return blotDB.View(func(tx *bolt.Tx) error {
		rootBucket, err := getBucketIfExists(tx, "luckymoney")
		if err != nil {
			return err
		}

		cursor := rootBucket.Cursor()
		seek := []byte(strconv.FormatUint(startID, 10))
		for k, v := cursor.Seek(seek); k != nil && v == nil; k, v = cursor.Next() {
			if bucket := rootBucket.Bucket(k); bucket != nil {
				jsb := bucket.Get([]byte("base"))
				if err = json.Unmarshal(jsb, &base); err != nil {
					continue
				}
				if callback != nil {
					callback(&base)
				}
			}
		}
		return nil
	})
}

// 生成序列号
func (handler *LuckyMoneyStorage) generateSN(tx *bolt.Tx, id uint64) (string, error) {
	bucket, err := ensureBucketExists(tx, "luckymoney", "mapping")
	if err != nil {
		return "", err
	}

	sn := ""
	token := make([]byte, 8)
	for {
		_, err = rand.Read(token)
		if err != nil {
			return "", err
		}
		sn = hex.EncodeToString(token)
		if bucket.Get([]byte(token)) == nil {
			sid := strconv.FormatUint(id, 10)
			return sn, bucket.Put([]byte(sn), []byte(sid))
		}
	}
}

// 创建领取记录
func (handler *LuckyMoneyStorage) insertRecord(tx *bolt.Tx, sid string, luckyMoneyArr []int) (int, int, error) {

	worstSeq, bestSeq := 0, 0
	minValue, maxValue := math.MaxInt32, 0
	recordBucket, err := ensureBucketExists(tx, "luckymoney", sid, "record")
	if err != nil {
		return 0, 0, err
	}
	for i := range luckyMoneyArr {
		seq, err := recordBucket.NextSequence()
		if err != nil {
			return 0, 0, err
		}

		val := LuckyMoneyRecord{Value: luckyMoneyArr[i]}
		jsb, err := json.Marshal(&val)
		if err != nil {
			return 0, 0, err
		}

		sseq := strconv.FormatUint(seq, 10)
		err = recordBucket.Put([]byte(sseq), jsb)
		if err != nil {
			return 0, 0, err
		}

		if luckyMoneyArr[i] < minValue {
			minValue = luckyMoneyArr[i]
			worstSeq = int(seq)
		}

		if luckyMoneyArr[i] > maxValue {
			maxValue = luckyMoneyArr[i]
			bestSeq = int(seq)
		}
	}
	return worstSeq, bestSeq, nil
}

// 领取红包
func (handler *LuckyMoneyStorage) receiveLuckyMoney(tx *bolt.Tx, sid string, seq int, user *LuckyMoneyUser) (int, error) {

	recordBucket, err := getBucketIfExists(tx, "luckymoney", sid, "record")
	if err != nil {
		return 0, err
	}

	var record LuckyMoneyRecord
	key := []byte(strconv.Itoa(seq))
	jsb := recordBucket.Get(key)
	if err = json.Unmarshal(jsb, &record); err != nil {
		return 0, err
	}
	record.User = user

	jsb, err = json.Marshal(&record)
	if err != nil {
		return 0, err
	}

	if err = recordBucket.Put(key, jsb); err != nil {
		return 0, err
	}
	return record.Value, nil
}
