package random

import (
	"fmt"
	"math/rand"
	"time"
)

func GetRandomString(count int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < count; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// pseudo-random number in [min,max)
func RandInt(min int, max int) int64 {
	if min > max {
		panic(fmt.Sprintf("min: %d can not greater then max: %d", min, max))
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return int64(min + r.Intn(max-min))
}
