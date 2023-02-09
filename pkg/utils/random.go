package utils

import "math/rand"

func CreateRandomIntRange(min int, max int) int {
	return rand.Intn(max-min) + min
}
