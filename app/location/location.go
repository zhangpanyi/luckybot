package location

import (
	"time"
)

var loc *time.Location

const (
	RFC3339LITE = "2006-01-02 15:04:05"
)

func init() {
	var err error
	loc, err = time.LoadLocation("Hongkong")
	if err != nil {
		panic(err)
	}
}

// 格式化时间
func Format(timestamp int64) string {
	utctime := time.Unix(timestamp, 0)
	return utctime.Format(RFC3339LITE)
}
