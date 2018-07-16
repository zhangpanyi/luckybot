package storage

import (
	"strconv"

	"github.com/boltdb/bolt"
)

// 订户存储
type SubscriberStorage struct {
}

// 获取订阅者
func (*SubscriberStorage) GetSubscribers(botID int64) ([]int64, error) {
	var subscribers []int64
	err := blotDB.View(func(tx *bolt.Tx) error {
		key := strconv.FormatInt(botID, 10)
		bucket, err := getBucketIfExists(tx, "subscriber", key)
		if err != nil && err != ErrNoBucket {
			return err
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
func (*SubscriberStorage) AddSubscriber(botID, userID int64) error {
	return blotDB.Batch(func(tx *bolt.Tx) error {
		key := strconv.FormatInt(botID, 10)
		bucket, err := ensureBucketExists(tx, "subscriber", key)
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
func (*SubscriberStorage) GetSubscriberCount(botID int64) (int, error) {
	var count int
	err := blotDB.View(func(tx *bolt.Tx) error {
		key := strconv.FormatInt(botID, 10)
		bucket, err := ensureBucketExists(tx, "subscriber", key)
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
