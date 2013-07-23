package main

import (
	"math/rand"
)

const (
	chars = `abcdefghijklmnopqrstuvwxyz1234567890`
)

func GenerateAlnumString(length int) string {
	key := make([]byte, length)
	idx := rand.Perm(len(chars))
	for i := 0; i < length; i++ {
		key[i] = chars[idx[i]]
	}
	return string(key)
}
