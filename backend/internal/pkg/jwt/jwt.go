package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims JWT 声明
type Claims struct {
	UserID                   string `json:"user_id"`
	CollaborationWorkspaceID string `json:"collaboration_workspace_id"`
	Email                    string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 Token
func GenerateToken(secret string, userID, tenantID, email string, expiresInMinutes int) (string, error) {
	expiresAt := time.Now().Add(time.Duration(expiresInMinutes) * time.Minute)

	claims := &Claims{
		UserID:                   userID,
		CollaborationWorkspaceID: tenantID,
		Email:                    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析 Token
func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		// jwt/v5 使用 errors.Is 来检查错误类型
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken 刷新 Token
func RefreshToken(tokenString, secret string, expiresInMinutes int) (string, error) {
	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		return "", err
	}

	// 生成新的 Token
	return GenerateToken(secret, claims.UserID, claims.CollaborationWorkspaceID, claims.Email, expiresInMinutes)
}
