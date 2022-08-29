package utils

import "math/rand"

func RandomString(size int) string {
	const str = "0123456789abcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	for i, b := range bytes {
		bytes[i] = str[b%byte(len(str))]
	}

	return string(bytes)
}
