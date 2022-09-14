package helpers

import (
	"math/rand"
	"strings"
	"time"
)

func RandomString(length int, salt string) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890" + strings.Replace(salt, "-", "", -1))
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandomNumber(length int) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("1234567890")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
