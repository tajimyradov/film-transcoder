package utils

import (
	"math/rand"
)

func GenerateRandomCode(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
