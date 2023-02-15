package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
)

// MD5 hashes using md5 algorithm
func MD5(text string) string {
	data := []byte(text)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// SHA256 Hashes the passed bytes to a SHA256 string
func SHA256(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
