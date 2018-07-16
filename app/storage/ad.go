package storage

import (
	"github.com/boltdb/bolt"
	"strconv"
)

// ********************** 结构图 **********************
// {
// 	"ad": {
// 		"bot_id": {
//			seq: val
// 		}
// 	}
// }
// ***************************************************

// 广告信息
type Ad struct {
	ID   uint32 `json:"id"`   // 广告ID
	Text string `json:"text"` // 广告文本
}

// 广告存储
type AdStorage struct {
}

// 获取广告
func (storage *AdStorage) GetAds(botID int64) ([]*Ad, error) {
	ads := make([]*Ad, 0)
	key := strconv.FormatInt(botID, 10)
	err := blotDB.View(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "ad", key)
		if err != nil {
			return err
		}
		return bucket.ForEach(func(k, v []byte) error {
			if v != nil {
				id, err := strconv.ParseUint(string(k), 10, 32)
				if err == nil {
					ads = append(ads, &Ad{ID: uint32(id), Text: string(v)})
				}
			}
			return nil
		})
	})

	if err != nil && err != ErrNoBucket {
		return nil, err
	}
	return ads, nil
}

// 添加广告
func (storage *AdStorage) AddAd(botID int64, ad string) (uint32, error) {
	var avID uint32
	key := strconv.FormatInt(botID, 10)
	err := blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := ensureBucketExists(tx, "ad", key)
		if err != nil {
			return err
		}

		seq, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		avID = uint32(seq)
		return bucket.Put([]byte(strconv.FormatUint(seq, 10)), []byte(ad))
	})

	if err != nil {
		return 0, err
	}
	return avID, nil
}

// 删除广告
func (storage *AdStorage) DelAd(botID int64, id uint32) error {
	sid := strconv.Itoa(int(id))
	key := strconv.FormatInt(botID, 10)
	return blotDB.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucketIfExists(tx, "ad", key)
		if err != nil {
			return err
		}
		return bucket.Delete([]byte(sid))
	})
}
