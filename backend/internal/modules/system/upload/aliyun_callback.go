package upload

import (
	"context"
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrOSSCallbackMissingPublicKeyURL = errors.New("oss callback missing x-oss-pub-key-url")
	ErrOSSCallbackMissingSignature    = errors.New("oss callback missing authorization")
	ErrOSSCallbackInvalidPublicKeyURL = errors.New("oss callback invalid public key url")
	ErrOSSCallbackInvalidSignature    = errors.New("oss callback invalid signature")
)

type OSSCallbackVerifier struct {
	client             *http.Client
	fetchPublicKey     func(context.Context, string) ([]byte, error)
	allowPublicKeyHost func(*url.URL) bool
}

type OSSCallbackRequest struct {
	PublicKeyURLHeader string
	Authorization      string
	RequestURI         string
	Body               []byte
}

func NewOSSCallbackVerifier() *OSSCallbackVerifier {
	client := &http.Client{Timeout: 3 * time.Second}
	return &OSSCallbackVerifier{
		client: client,
		fetchPublicKey: func(ctx context.Context, targetURL string) ([]byte, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
			if err != nil {
				return nil, err
			}
			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("fetch public key status: %d", resp.StatusCode)
			}
			return io.ReadAll(resp.Body)
		},
		allowPublicKeyHost: isAllowedOSSCallbackPublicKeyHost,
	}
}

func (v *OSSCallbackVerifier) Verify(ctx context.Context, req OSSCallbackRequest) error {
	if v == nil {
		v = NewOSSCallbackVerifier()
	}
	publicKeyURL, err := v.decodePublicKeyURL(req.PublicKeyURLHeader)
	if err != nil {
		return err
	}
	signature, err := decodeOSSCallbackSignature(req.Authorization)
	if err != nil {
		return err
	}
	publicKeyPEM, err := v.fetchPublicKey(ctx, publicKeyURL.String())
	if err != nil {
		return fmt.Errorf("fetch oss callback public key: %w", err)
	}
	publicKey, err := parseOSSCallbackPublicKey(publicKeyPEM)
	if err != nil {
		return fmt.Errorf("parse oss callback public key: %w", err)
	}
	stringToVerify, err := buildOSSCallbackStringToSign(req.RequestURI, req.Body)
	if err != nil {
		return err
	}
	digest := md5.Sum([]byte(stringToVerify))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.MD5, digest[:], signature); err != nil {
		return ErrOSSCallbackInvalidSignature
	}
	return nil
}

func (v *OSSCallbackVerifier) decodePublicKeyURL(headerValue string) (*url.URL, error) {
	if strings.TrimSpace(headerValue) == "" {
		return nil, ErrOSSCallbackMissingPublicKeyURL
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(headerValue))
	if err != nil {
		return nil, ErrOSSCallbackInvalidPublicKeyURL
	}
	targetURL, err := url.Parse(strings.TrimSpace(string(decoded)))
	if err != nil {
		return nil, ErrOSSCallbackInvalidPublicKeyURL
	}
	if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
		return nil, ErrOSSCallbackInvalidPublicKeyURL
	}
	if v.allowPublicKeyHost != nil && !v.allowPublicKeyHost(targetURL) {
		return nil, ErrOSSCallbackInvalidPublicKeyURL
	}
	return targetURL, nil
}

func isAllowedOSSCallbackPublicKeyHost(targetURL *url.URL) bool {
	if targetURL == nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(targetURL.Hostname()))
	if host == "gosspublic.alicdn.com" {
		return true
	}
	return strings.HasSuffix(host, ".aliyuncs.com")
}

func decodeOSSCallbackSignature(headerValue string) ([]byte, error) {
	if strings.TrimSpace(headerValue) == "" {
		return nil, ErrOSSCallbackMissingSignature
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(headerValue))
	if err != nil {
		return nil, ErrOSSCallbackInvalidSignature
	}
	return signature, nil
}

func buildOSSCallbackStringToSign(requestURI string, body []byte) (string, error) {
	if strings.TrimSpace(requestURI) == "" {
		return "", ErrOSSCallbackInvalidSignature
	}
	parsed, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return "", ErrOSSCallbackInvalidSignature
	}
	decodedPath, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return "", ErrOSSCallbackInvalidSignature
	}
	queryString := ""
	if parsed.RawQuery != "" {
		queryString = "?" + parsed.RawQuery
	}
	return decodedPath + queryString + "\n" + string(body), nil
}

func parseOSSCallbackPublicKey(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, errors.New("invalid pem")
	}
	if parsed, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if publicKey, ok := parsed.(*rsa.PublicKey); ok {
			return publicKey, nil
		}
	}
	if publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return publicKey, nil
	}
	return nil, errors.New("invalid rsa public key")
}
