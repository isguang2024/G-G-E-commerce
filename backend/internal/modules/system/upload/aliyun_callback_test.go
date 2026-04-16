package upload

import (
	"context"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/url"
	"testing"
)

func TestOSSCallbackVerifierVerifySuccess(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	requestURI := "/callback/%E4%B8%AD%E6%96%87?id=1&index=2"
	body := []byte("bucket=test&object=demo.png")
	stringToSign, err := buildOSSCallbackStringToSign(requestURI, body)
	if err != nil {
		t.Fatalf("buildOSSCallbackStringToSign() error = %v", err)
	}
	digest := md5.Sum([]byte(stringToSign))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.MD5, digest[:])
	if err != nil {
		t.Fatalf("rsa.SignPKCS1v15() error = %v", err)
	}

	verifier := NewOSSCallbackVerifier()
	verifier.allowPublicKeyHost = func(*url.URL) bool { return true }
	verifier.fetchPublicKey = func(context.Context, string) ([]byte, error) {
		return publicKeyPEM, nil
	}
	err = verifier.Verify(context.Background(), OSSCallbackRequest{
		PublicKeyURLHeader: base64.StdEncoding.EncodeToString([]byte("https://gosspublic.alicdn.com/callback_pub_key_v1.pem")),
		Authorization:      base64.StdEncoding.EncodeToString(signature),
		RequestURI:         requestURI,
		Body:               body,
	})
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestOSSCallbackVerifierRejectsUnexpectedHostAndBadSignature(t *testing.T) {
	verifier := NewOSSCallbackVerifier()
	if err := verifier.Verify(context.Background(), OSSCallbackRequest{
		PublicKeyURLHeader: base64.StdEncoding.EncodeToString([]byte("https://example.com/callback.pem")),
		Authorization:      base64.StdEncoding.EncodeToString([]byte("bad-signature")),
		RequestURI:         "/callback",
		Body:               []byte("bucket=test"),
	}); !errors.Is(err, ErrOSSCallbackInvalidPublicKeyURL) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrOSSCallbackInvalidPublicKeyURL)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	verifier.allowPublicKeyHost = func(*url.URL) bool { return true }
	verifier.fetchPublicKey = func(context.Context, string) ([]byte, error) {
		return publicKeyPEM, nil
	}
	err = verifier.Verify(context.Background(), OSSCallbackRequest{
		PublicKeyURLHeader: base64.StdEncoding.EncodeToString([]byte("https://gosspublic.alicdn.com/callback_pub_key_v1.pem")),
		Authorization:      base64.StdEncoding.EncodeToString([]byte("not-a-real-signature")),
		RequestURI:         "/callback",
		Body:               []byte("bucket=test"),
	})
	if !errors.Is(err, ErrOSSCallbackInvalidSignature) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrOSSCallbackInvalidSignature)
	}
}

func TestBuildOSSCallbackStringToSignAndAllowedHost(t *testing.T) {
	got, err := buildOSSCallbackStringToSign("/index.php?id=1&index=2", []byte("bucket=test"))
	if err != nil {
		t.Fatalf("buildOSSCallbackStringToSign() error = %v", err)
	}
	want := "/index.php?id=1&index=2\nbucket=test"
	if got != want {
		t.Fatalf("buildOSSCallbackStringToSign() = %q, want %q", got, want)
	}
	if !isAllowedOSSCallbackPublicKeyHost(&url.URL{Scheme: "https", Host: "gosspublic.alicdn.com"}) {
		t.Fatalf("gosspublic.alicdn.com should be allowed")
	}
	if !isAllowedOSSCallbackPublicKeyHost(&url.URL{Scheme: "https", Host: "oss-cn-hangzhou.aliyuncs.com"}) {
		t.Fatalf(".aliyuncs.com host should be allowed")
	}
}
