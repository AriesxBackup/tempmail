package tempmail

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

const (
	lowercaseCharset = "abcdefghijklmnopqrstuvwxyz"
	digitCharset     = "0123456789"
	uppercaseCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func GenerateRandomString(length int, charset string) string {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			result[i] = charset[i%len(charset)]
			continue
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

func GenerateUsername(length int) string {
	charset := lowercaseCharset + digitCharset
	return GenerateRandomString(length, charset)
}

func GeneratePassword(length int) string {
	charset := lowercaseCharset + uppercaseCharset + digitCharset
	return GenerateRandomString(length, charset)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func boolString(b bool) string {
	return strconv.FormatBool(b)
}
