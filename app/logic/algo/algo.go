package algo

import (
	"errors"
	"math/big"
	"math/rand"
	"time"
)

var (
	// 金额不足
	ErrTooLittleMoney = errors.New("each person is at least 0.01")
	// 红包数量太少
	ErrorTooLittleNumber = errors.New("number must be more than 0")
)

// 生成算法
func Generate(amount *big.Float, precision, number int) ([]*big.Float, error) {
	if precision < 0 {
		precision = 0
	}

	base := big.NewInt(10)
	base.Exp(base, big.NewInt(int64(precision)), nil)
	wei, _ := big.NewFloat(0).SetString(base.String())

	product := big.NewFloat(0).Mul(wei, amount)
	newAmount, _ := product.Int(big.NewInt(0))
	arr, err := generate(newAmount, number)
	if err != nil {
		return nil, err
	}

	result := make([]*big.Float, 0, number)
	for i := 0; i < len(arr); i++ {
		tmp := big.NewFloat(0).SetInt(arr[i])
		result = append(result, tmp.Quo(tmp, wei))
	}
	return result, nil
}

// 零值
var ZERO = big.NewInt(0)

// 随机器
var randx = rand.New(rand.NewSource(time.Now().UnixNano()))

// 打乱数组
func randomShuffle(array []*big.Int) []*big.Int {
	for i := range array {
		j := randx.Intn(i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
}

// 大整数相减
func subBigInt(x *big.Int, y *big.Int) *big.Int {
	return big.NewInt(0).Sub(x, y)
}

// 生成算法
func generate(amount *big.Int, number int) ([]*big.Int, error) {
	if amount.Cmp(ZERO) == -1 {
		return nil, ErrTooLittleMoney
	}
	if number < 1 {
		return nil, ErrorTooLittleNumber
	}
	return generateResultSet(amount, number)
}

// 生成结果集
func generateResultSet(amount *big.Int, number int) ([]*big.Int, error) {
	one := big.NewInt(1)
	result := make([]*big.Int, 0, number)
	for i := 1; i < number; i++ {
		value := big.NewInt(1)
		x, ok := big.NewFloat(0).SetString(
			subBigInt(amount, big.NewInt(int64(number-1))).String())
		if !ok {
			return nil, errors.New("invalid number")
		}
		y := big.NewFloat(float64(number - i))
		safeAmount, _ := big.NewFloat(0).Quo(x, y).Int(big.NewInt(0))
		if safeAmount.Cmp(one) == 1 {
			value.Add(one, big.NewInt(0).Rand(randx, safeAmount.Sub(safeAmount, one)))
		}
		amount.Sub(amount, value)
		result = append(result, value)
	}
	result = append(result, amount)
	return randomShuffle(result), nil
}
