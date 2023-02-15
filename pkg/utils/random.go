package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func CreateRandomIntRange(min int, max int) int {
	return rand.Intn(max-min) + min
}

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
