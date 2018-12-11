package models

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/zhangpanyi/luckybot/app/fmath"
	"github.com/zhangpanyi/luckybot/app/storage"
)

// 账户数据
type Account struct {
	Symbol  string     `json:"symbol"`  // 货币符号
	Amount  *big.Float `json:"amount"`  // 资产金额
	Locked  *big.Float `json:"locked"`  // 锁定金额
	Disable bool       `json:"disable"` // 禁用账户
}

// 标准化
func (account *Account) Normalization() {
	if account.Amount != nil {
		account.Amount.SetPrec(fmath.Prec())
	}
	if account.Locked != nil {
		account.Locked.SetPrec(fmath.Prec())
	}
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
			account.Normalization()
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
	var account Account
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

		if err = json.Unmarshal(jsb, &account); err != nil {
			return err
		}
		account.Normalization()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &account, nil
}

// 账户存款操作
func (model *AccountModel) Deposit(userID int64, symbol string, amount *big.Float) (*Account, error) {
	var account Account
	key := strconv.FormatInt(userID, 10)
	err := storage.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "accounts", key)
		if err != nil {
			return err
		}

		jsb := bucket.Get([]byte(symbol))
		if jsb == nil {
			account.Symbol = symbol
			account.Amount = amount
			account.Locked = big.NewFloat(0)
		} else {
			if err = json.Unmarshal(jsb, &account); err != nil {
				return err
			}
			account.Normalization()
			account.Amount.Add(account.Amount, amount)
		}

		jsb, err = json.Marshal(&account)
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
	return &account, nil
}

// 账户取款操作
func (model *AccountModel) Withdraw(userID int64, symbol string, amount *big.Float) (*Account, error) {
	var account Account
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

		if err = json.Unmarshal(jsb, &account); err != nil {
			return err
		}
		account.Normalization()

		if account.Locked.Cmp(amount) == 1 {
			return ErrInsufficientAmount
		}
		account.Locked.Sub(account.Locked, amount)

		jsb, err = json.Marshal(&account)
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
	return &account, nil
}

// 锁定账户资金
func (model *AccountModel) LockAccount(userID int64, symbol string, amount *big.Float) (*Account, error) {
	var account Account
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

		if err = json.Unmarshal(jsb, &account); err != nil {
			return err
		}
		account.Normalization()

		if amount.Cmp(account.Amount) == 1 {
			return ErrInsufficientAmount
		}
		account.Locked.Add(account.Locked, amount)
		account.Amount.Sub(account.Amount, amount)

		jsb, err = json.Marshal(&account)
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
	return &account, nil
}

// 解锁账户资金
func (model *AccountModel) UnlockAccount(userID int64, symbol string, amount *big.Float) (*Account, error) {
	var account Account
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

		if err = json.Unmarshal(jsb, &account); err != nil {
			return err
		}
		account.Normalization()

		if amount.Cmp(account.Locked) == 1 {
			return ErrInsufficientAmount
		}
		account.Locked.Sub(account.Locked, amount)
		account.Amount.Add(account.Amount, amount)

		jsb, err = json.Marshal(&account)
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
	return &account, nil
}

// 从锁定账户转账
func (model *AccountModel) TransferFromLockAccount(from, to int64, symbol string,
	amount *big.Float) (*Account, *Account, error) {

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
		fromAccount.Normalization()

		if amount.Cmp(fromAccount.Locked) == 1 {
			return ErrInsufficientAmount
		}
		fromAccount.Locked.Sub(fromAccount.Locked, amount)

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
			toAccount.Locked = big.NewFloat(0)
		} else {
			if err = json.Unmarshal(jsb, &toAccount); err != nil {
				return err
			}
			toAccount.Normalization()
			toAccount.Amount.Add(toAccount.Amount, amount)
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
