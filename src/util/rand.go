package util

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math/rand"
	"time"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

//RandString init a rand string
func RandString(n int) string {
	len := len(letterBytes)
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len)]
	}
	return string(b)
}

// Sha1Hash sha1 hash
func Sha1Hash(s string) string {
	t := sha1.New()
	io.WriteString(t, s)
	return fmt.Sprintf("%x", t.Sum(nil))
}

// RandomNumStr random num to string
func RandomNumStr() string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := rand.Intn(99999999)
	numstr := fmt.Sprintf("%08v", num)
	return numstr
}

//RandMaskToken init a rand mask_token
func RandMaskToken(s string, n int) string {
	l := len(s)
	b := make([]byte, n)
	for i := range b {
		b[i] = s[rand.Intn(l)]
	}
	return string(b)
}

// Rand4Num return 4 random number
func Rand4Num() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vCode := fmt.Sprintf("%04v", rnd.Int31n(99999999))
	return vCode[0:4]
}
