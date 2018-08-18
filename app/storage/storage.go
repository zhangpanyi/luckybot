package storage

import (
	"errors"

	"io"

	"github.com/boltdb/bolt"
)

// 数据库实例
var DB *bolt.DB

var (
	// 没有桶
	ErrNoBucket = errors.New("no bucket")
)

// 连接到数据库
func Connect(path string) error {
	var err error
	DB, err = bolt.Open(path, 0600, nil)
	return err
}

// 关闭连接
func Close() error {
	return DB.Close()
}

// 备份数据库
func Backup(writer io.Writer) (int64, error) {
	var size int64
	err := DB.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(writer)
		if err != nil {
			return err
		}
		size = tx.Size()
		return nil
	})
	return size, err
}

// 确保桶存在
func EnsureBucketExists(tx *bolt.Tx, args ...string) (*bolt.Bucket, error) {
	if len(args) == 0 {
		return nil, ErrNoBucket
	}

	bucket, err := tx.CreateBucketIfNotExists([]byte(args[0]))
	if err != nil {
		return nil, err
	}

	if len(args) > 1 {
		for _, name := range args[1:] {
			bucket, err = bucket.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return nil, err
			}
		}
	}
	return bucket, nil
}

// 获取桶若存在
func GetBucketIfExists(tx *bolt.Tx, args ...string) (*bolt.Bucket, error) {
	if len(args) == 0 {
		return nil, ErrNoBucket
	}

	bucket := tx.Bucket([]byte(args[0]))
	if bucket == nil {
		return nil, ErrNoBucket
	}

	if len(args) > 1 {
		for _, name := range args[1:] {
			bucket = bucket.Bucket([]byte(name))
			if bucket == nil {
				return nil, ErrNoBucket
			}
		}
	}
	return bucket, nil
}
