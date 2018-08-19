package models

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/zhangpanyi/luckymoney/app/storage"
)

// 账户信息
type Account struct {
	Symbol string `json:"symbol"` // 货币符号
	Amount uint32 `json:"amount"` // 资产金额
	Locked uint32 `json:"locked"` // 锁定金额
}

var (
	// 资产不足
	ErrInsufficientAmount = errors.New("insufficient amount")
	// 没有此类型账户
	ErrNoSuchTypeAccount = errors.New("no such type of account")
)

// ********************** 结构图 **********************
// {
//	"accounts": {
// 		<user_id>: {
// 			<symbol>: {			// 账户信息
// 				"amount": 0,	// 资产金额
// 				"locked": 0		// 锁定金额
//			}
// 		}
//	}
// ***************************************************

// 账户模型
type AccountModel struct {
}

// 获取账户列表
func (model *AccountModel) GetAccounts(userID int64) ([]*Account, error) {
	accounts := make([]*Account, 0)
	key := strconv.FormatInt(userID, 10)
	err := storage.DB.View(func(tx *bolt.Tx) error {
		bucket, err := storage.GetBucketIfExists(tx, "accounts", key)
		if err != nil {
			return err
		}

		err = bucket.ForEach(func(k, v []byte) error {
			var account Account
			if err = json.Unmarshal(v, &account); err != nil {
				return err
			}
			accounts = append(accounts, &account)
			return nil
		})

		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// 获取账户信息
func (model *AccountModel) GetAccount(userID int64, symbol string) (*Account, error) {
	var acount Account
	key := strconv.FormatInt(userID, 10)
	err := storage.DB.View(func(tx *bolt.Tx) error {
		bucket, err := storage.GetBucketIfExists(tx, "accounts", key)
		if err != nil {
			return err
		}

		jsb := bucket.Get([]byte(symbol))
		if jsb == nil {
			return ErrNoSuchTypeAccount
		}

		if err = json.Unmarshal(jsb, &acount); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &acount, nil
}

// 存款
func (model *AccountModel) Deposit(userID int64, symbol string, amount uint32) error {
	key := strconv.FormatInt(userID, 10)
	return storage.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "accounts", key)
		if err != nil {
			return err
		}

		var acount Account
		jsb := bucket.Get([]byte(symbol))
		if jsb == nil {
			acount.Symbol = symbol
			acount.Amount = amount
		} else {
			if err = json.Unmarshal(jsb, &acount); err != nil {
				return err
			}
			acount.Amount += amount
		}

		jsb, err = json.Marshal(&acount)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(symbol), jsb); err != nil {
			return err
		}
		return nil
	})
}

// 取款
func (model *AccountModel) Withdraw(userID int64, symbol string, amount uint32) error {
	key := strconv.FormatInt(userID, 10)
	return storage.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "accounts", key)
		if err != nil {
			return err
		}

		var acount Account
		jsb := bucket.Get([]byte(symbol))
		if jsb == nil {
			return ErrNoSuchTypeAccount
		}

		if err = json.Unmarshal(jsb, &acount); err != nil {
			return err
		}

		if amount > acount.Amount {
			return ErrInsufficientAmount
		}
		acount.Amount -= amount

		jsb, err = json.Marshal(&acount)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(symbol), jsb); err != nil {
			return err
		}
		return nil
	})
}

// 锁定账户
func (model *AccountModel) LockAccount(userID int64, symbol string, amount uint32) (*Account, error) {
	var acount Account
	key := strconv.FormatInt(userID, 10)
	err := storage.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "accounts", key)
		if err != nil {
			return err
		}

		jsb := bucket.Get([]byte(symbol))
		if jsb == nil {
			return ErrNoSuchTypeAccount
		}

		if err = json.Unmarshal(jsb, &acount); err != nil {
			return err
		}

		if amount > acount.Amount {
			return ErrInsufficientAmount
		}
		acount.Amount -= amount
		acount.Locked += amount

		jsb, err = json.Marshal(&acount)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(symbol), jsb); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &acount, nil
}

// 解锁账户
func (model *AccountModel) UnlockAccount(userID int64, symbol string, amount uint32) error {
	key := strconv.FormatInt(userID, 10)
	return storage.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "accounts", key)
		if err != nil {
			return err
		}

		var account Account
		jsb := bucket.Get([]byte(symbol))
		if jsb == nil {
			return ErrNoSuchTypeAccount
		}

		if err = json.Unmarshal(jsb, &account); err != nil {
			return err
		}

		if amount > account.Locked {
			return ErrInsufficientAmount
		}
		account.Amount += amount
		account.Locked -= amount

		jsb, err = json.Marshal(&account)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(symbol), jsb); err != nil {
			return err
		}
		return nil
	})
}

// 从锁定账户转账
func (model *AccountModel) TransferFromLockAccount(from, to int64, symbol string,
	amount uint32) (*Account, *Account, error) {

	var toAccount Account
	var fromAccount Account
	toKey := strconv.FormatInt(to, 10)
	fromKey := strconv.FormatInt(from, 10)
	err := storage.DB.Update(func(tx *bolt.Tx) error {
		// 获取桶
		toBucket, err := storage.EnsureBucketExists(tx, "accounts", toKey)
		if err != nil {
			return err
		}

		fromBucket, err := storage.EnsureBucketExists(tx, "accounts", fromKey)
		if err != nil {
			return err
		}

		// 获取账户信息
		jsb := fromBucket.Get([]byte(symbol))
		if jsb == nil {
			return ErrNoSuchTypeAccount
		}
		if err = json.Unmarshal(jsb, &fromAccount); err != nil {
			return err
		}

		if amount > fromAccount.Locked {
			return ErrInsufficientAmount
		}
		fromAccount.Locked -= amount

		jsb, err = json.Marshal(&fromAccount)
		if err != nil {
			return err
		}

		if err = fromBucket.Put([]byte(symbol), jsb); err != nil {
			return err
		}

		// 转移锁定资产
		jsb = toBucket.Get([]byte(symbol))
		if jsb == nil {
			toAccount.Symbol = symbol
			toAccount.Amount = amount
			toAccount.Locked = 0
		} else {
			if err = json.Unmarshal(jsb, &toAccount); err != nil {
				return err
			}
			toAccount.Amount += amount
		}

		jsb, err = json.Marshal(&toAccount)
		if err != nil {
			return err
		}

		if err = toBucket.Put([]byte(symbol), jsb); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return &fromAccount, &toAccount, nil
}
