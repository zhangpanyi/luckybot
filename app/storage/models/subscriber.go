package models

import (
	"strconv"

	"github.com/boltdb/bolt"
	"luckybot/app/storage"
)

// 订户模型
type SubscriberModel struct {
}

// 获取订阅者
func (*SubscriberModel) GetSubscribers() ([]int64, error) {
	var subscribers []int64
	err := storage.DB.View(func(tx *bolt.Tx) error {
		bucket, err := storage.GetBucketIfExists(tx, "subscribers")
		if err != nil {
			if err != storage.ErrNoBucket {
				return err
			}
			return nil
		}

		subscribers = make([]int64, 0, bucket.Stats().KeyN)
		bucket.ForEach(func(k, v []byte) error {
			userID, err := strconv.ParseInt(string(k), 10, 64)
			if err == nil {
				subscribers = append(subscribers, userID)
			}
			return nil
		})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return subscribers, nil
}

// 添加订阅者
func (*SubscriberModel) AddSubscriber(userID int64) error {
	return storage.DB.Batch(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "subscribers")
		if err != nil {
			return err
		}

		subscriber := strconv.FormatInt(userID, 10)
		if bucket.Get([]byte(subscriber)) != nil {
			return nil
		}
		return bucket.Put([]byte(subscriber), []byte(""))
	})
}

// 获取订阅者数量
func (*SubscriberModel) GetSubscriberCount() (int, error) {
	var count int
	err := storage.DB.View(func(tx *bolt.Tx) error {
		bucket, err := storage.EnsureBucketExists(tx, "subscribers")
		if err != nil {
			return err
		}
		count = bucket.Stats().KeyN
		return nil
	})

	if err != nil {
		return 0, err
	}
	return count, nil
}
