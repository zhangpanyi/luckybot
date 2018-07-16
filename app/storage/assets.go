package storage

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/boltdb/bolt"
)

var (
	// 资产不足
	ErrLackOfAssets = errors.New("lack of assets")
	// 用户不存在
	ErrUserNotExist = errors.New("user does not exist")
	// 没有此类型资产
	ErrNoSuchTypeAsset = errors.New("no such type of assets")
)

// ********************** 结构图 **********************
// {
// 	"assets": {
// 		"user_id": {
// 			"asset": {			// 资产信息
// 				"amount": 0,	// 资产总额
// 				"freeze": 0		// 冻结资产
// 			}
// 		}
// 	}
// }
// ***************************************************

// 资产存储
type AssetStorage struct {
}

// 获取资产列表
func (storage *AssetStorage) GetAssets(userID int64) ([]*Asset, error) {
	assets := make([]*Asset, 0)
	key := strconv.FormatInt(userID, 10)
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "assets", key)
		if err != nil {
			return err
		}

		err = bucket.ForEach(func(k, v []byte) error {
			var asset Asset
			if err = json.Unmarshal(v, &asset); err != nil {
				return err
			}
			assets = append(assets, &asset)
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
	return assets, nil
}

// 获取资产信息
func (storage *AssetStorage) GetAsset(userID int64, asset string) (*Asset, error) {
	var result Asset
	key := strconv.FormatInt(userID, 10)
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "assets", key)
		if err != nil {
			return err
		}

		jsb := bucket.Get([]byte(asset))
		if jsb == nil {
			return ErrNoSuchTypeAsset
		}

		if err = json.Unmarshal(jsb, &result); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// 存款
func (storage *AssetStorage) Deposit(userID int64, asset string, amount uint32) error {
	key := strconv.FormatInt(userID, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := ensureBucketExists(tx, "assets", key)
		if err != nil {
			return err
		}

		var assetInfo Asset
		jsb := bucket.Get([]byte(asset))
		if jsb == nil {
			assetInfo.Asset = asset
			assetInfo.Amount = amount
		} else {
			if err = json.Unmarshal(jsb, &assetInfo); err != nil {
				return err
			}
			assetInfo.Amount += amount
		}

		jsb, err = json.Marshal(&assetInfo)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(asset), jsb); err != nil {
			return err
		}
		return nil
	})
}

// 取款
func (storage *AssetStorage) Withdraw(userID int64, asset string, amount uint32) error {
	key := strconv.FormatInt(userID, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := ensureBucketExists(tx, "assets", key)
		if err != nil {
			return err
		}

		var assetInfo Asset
		jsb := bucket.Get([]byte(asset))
		if jsb == nil {
			return ErrNoSuchTypeAsset
		}

		if err = json.Unmarshal(jsb, &assetInfo); err != nil {
			return err
		}

		if amount > assetInfo.Amount {
			return ErrLackOfAssets
		}
		assetInfo.Amount -= amount

		jsb, err = json.Marshal(&assetInfo)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(asset), jsb); err != nil {
			return err
		}
		return nil
	})
}

// 冻结资产
func (storage *AssetStorage) FrozenAsset(userID int64, asset string, amount uint32) error {
	key := strconv.FormatInt(userID, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := ensureBucketExists(tx, "assets", key)
		if err != nil {
			return err
		}

		var assetInfo Asset
		jsb := bucket.Get([]byte(asset))
		if jsb == nil {
			return ErrNoSuchTypeAsset
		}

		if err = json.Unmarshal(jsb, &assetInfo); err != nil {
			return err
		}

		if amount > assetInfo.Amount {
			return ErrLackOfAssets
		}
		assetInfo.Amount -= amount
		assetInfo.Freeze += amount

		jsb, err = json.Marshal(&assetInfo)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(asset), jsb); err != nil {
			return err
		}
		return nil
	})
}

// 解冻资产
func (storage *AssetStorage) UnfreezeAsset(userID int64, asset string, amount uint32) error {
	key := strconv.FormatInt(userID, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := ensureBucketExists(tx, "assets", key)
		if err != nil {
			return err
		}

		var assetInfo Asset
		jsb := bucket.Get([]byte(asset))
		if jsb == nil {
			return ErrNoSuchTypeAsset
		}

		if err = json.Unmarshal(jsb, &assetInfo); err != nil {
			return err
		}

		if amount > assetInfo.Freeze {
			return ErrLackOfAssets
		}
		assetInfo.Amount += amount
		assetInfo.Freeze -= amount

		jsb, err = json.Marshal(&assetInfo)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(asset), jsb); err != nil {
			return err
		}
		return nil
	})
}

// 转移冻结资金
func (storage *AssetStorage) TransferFrozenAsset(from, to int64, asset string, amount uint32) error {
	toKey := strconv.FormatInt(to, 10)
	fromKey := strconv.FormatInt(from, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		toBucket, err := ensureBucketExists(tx, "assets", toKey)
		if err != nil {
			return err
		}

		fromBucket, err := ensureBucketExists(tx, "assets", fromKey)
		if err != nil {
			return err
		}

		// 获取资产数据
		var assetInfo Asset
		jsb := fromBucket.Get([]byte(asset))
		if jsb == nil {
			return ErrNoSuchTypeAsset
		}
		if err = json.Unmarshal(jsb, &assetInfo); err != nil {
			return err
		}

		if amount > assetInfo.Freeze {
			return ErrLackOfAssets
		}
		assetInfo.Freeze -= amount

		jsb, err = json.Marshal(&assetInfo)
		if err != nil {
			return err
		}

		if err = fromBucket.Put([]byte(asset), jsb); err != nil {
			return err
		}

		// 转移冻结资产
		jsb = toBucket.Get([]byte(asset))
		if jsb == nil {
			assetInfo.Asset = asset
			assetInfo.Amount = amount
			assetInfo.Freeze = 0
		} else {
			if err = json.Unmarshal(jsb, &assetInfo); err != nil {
				return err
			}
			assetInfo.Amount += amount
		}

		jsb, err = json.Marshal(&assetInfo)
		if err != nil {
			return err
		}

		if err = toBucket.Put([]byte(asset), jsb); err != nil {
			return err
		}
		return nil
	})
}
