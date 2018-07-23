package generate

import (
	"errors"
	"math/rand"
	"time"
)

var (
	// 金额不足
	ErrTooLittleMoney = errors.New("each person is at least 0.01")
	// 小数点错误
	ErrDecimalPlaces = errors.New("accurate to two decimal places")
	// 红包数量太少
	ErrorTooLittleNumber = errors.New("number must be more than 0")
)

// 生成算法
func Generate(amount uint32, number uint32) ([]int, error) {
	// 检查红包数量
	if number < 1 {
		return nil, ErrorTooLittleNumber
	}

	// 最小金额为0.01
	if number > amount {
		return nil, ErrTooLittleMoney
	}
	return generateResultSet(int(amount), int(number))
}

// 随机器
var randx = rand.New(rand.NewSource(time.Now().UnixNano()))

// 打乱数组
func randomShuffle(array []int) []int {
	for i := range array {
		j := randx.Intn(i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
}

// 生成结果集
func generateResultSet(amount int, number int) ([]int, error) {
	result := make([]int, 0, number)
	for i := 1; i < number; i++ {
		value := 1
		safeAmount := int(float64(amount-(number-i)) / float64((number - i)))
		if safeAmount > 1 {
			value = 1 + randx.Intn(safeAmount-1)
		}
		amount = amount - value
		result = append(result, value)
	}
	result = append(result, amount)
	return randomShuffle(result), nil
}
