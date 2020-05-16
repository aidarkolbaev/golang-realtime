package utils

import (
	"math/rand"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

var (
	src        = rand.NewSource(time.Now().UnixNano())
	emailRegex = regexp.MustCompile("(?i)^[a-z0-9_.+-]+@[a-z0-9-]+\\.[a-z0-9-.]+$")
	nameRegex  = regexp.MustCompile("(?i)^[a-zа-яА-Я0-9]+[a-zа-яА-Я0-9 :_-]*[a-zа-яА-Я0-9]+$")
	urlRegex   = regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
	colors     = []string{
		"white", "black", "red", "maroon", "yellow", "lime", "green", "aqua", "teal", "blue", "navy", "fuchsia", "purple",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Returns a random string of the specified length
func RandString(length int) string {
	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// ParseInt converts val to int by min max conditions, on error returns default value
func ParseInt(val string, def, min, max int) int {
	v, _ := strconv.Atoi(val)
	if v < min || v > max {
		v = def
	}
	return v
}

func InArray(arr []string, val string) bool {
	for _, s := range arr {
		if s == val {
			return true
		}
	}
	return false
}

func IsLengthValid(str string, minLen, maxLen int) bool {
	length := utf8.RuneCountInString(str)
	return length >= minLen && length <= maxLen
}

func IsEmailValid(email string) bool {
	return IsLengthValid(email, 2, 50) && emailRegex.MatchString(email)
}

func IsNameValid(name string) bool {
	return IsLengthValid(name, 2, 100) && nameRegex.MatchString(name)
}

func IsUrlValid(url string) bool {
	return urlRegex.MatchString(url)
}

func GetRandomColor() string {
	return colors[rand.Intn(len(colors)-1)]
}
