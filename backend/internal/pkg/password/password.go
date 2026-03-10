package password

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost 默认加密成本
	DefaultCost = 12
)

// Hash 加密密码
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify 验证密码
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
