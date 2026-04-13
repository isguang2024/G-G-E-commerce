package social

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSocialTokenSignAndParse(t *testing.T) {
	s := &service{secret: "unit-test-secret"}
	token, err := s.signSocialToken(socialTokenClaims{
		Intent:      socialIntentRegister,
		ProviderKey: "github",
		ProviderUID: "123",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	})
	if err != nil {
		t.Fatalf("signSocialToken error: %v", err)
	}

	claims, err := s.parseSocialToken(token)
	if err != nil {
		t.Fatalf("parseSocialToken error: %v", err)
	}
	if claims.Intent != socialIntentRegister {
		t.Fatalf("unexpected intent: %s", claims.Intent)
	}
	if claims.ProviderKey != "github" || claims.ProviderUID != "123" {
		t.Fatalf("unexpected provider claims: %+v", claims)
	}
}

func TestSocialTokenParseRejectExpired(t *testing.T) {
	s := &service{secret: "unit-test-secret"}
	token, err := s.signSocialToken(socialTokenClaims{
		Intent:      socialIntentLogin,
		ProviderKey: "github",
		ProviderUID: "123",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	})
	if err != nil {
		t.Fatalf("signSocialToken error: %v", err)
	}

	if _, err := s.parseSocialToken(token); err == nil {
		t.Fatal("expected parseSocialToken to fail for expired token")
	}
}
