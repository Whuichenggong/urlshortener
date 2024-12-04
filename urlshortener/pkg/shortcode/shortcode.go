package shortcode

import "math/rand"

// GenerateIdShortCode() string

type ShortCode struct {
	length int
}

func NewShortCode(length int) *ShortCode {
	return &ShortCode{length: length}
}

const chars = "qwertyuiopasdfghjklzxcvbnm123456789"

func (s *ShortCode) GenerateShortCode() string {
	length := len(chars)
	result := make([]byte, s.length)

	for i := 0; i < s.length; i++ {
		result[i] = chars[rand.Intn(length)]
	}
	return string(result)
}
