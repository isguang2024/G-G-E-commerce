package upload

import (
	"container/list"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/maben/backend/internal/config"
)

const (
	secretCipherPrefix = "gge:v1"
	defaultCipherCache = 256
)

var (
	ErrSecretCipherUnavailable = errors.New("upload secret cipher is unavailable")
	ErrSecretCipherMalformed   = errors.New("upload secret payload is malformed")
	ErrSecretCipherKeyNotFound = errors.New("upload secret key not found")
)

type SecretCipher interface {
	Encrypt(ctx context.Context, plaintext string) (string, error)
	Decrypt(ctx context.Context, ciphertext string) (string, error)
	CurrentKeyID() string
}

type secretCipher struct {
	currentKeyID string
	keys         map[string][]byte
	cache        *secretCache
}

type secretCache struct {
	ttl      time.Duration
	capacity int
	mu       sync.Mutex
	items    map[string]*list.Element
	order    *list.List
}

type secretCacheEntry struct {
	key       string
	value     string
	expiresAt time.Time
}

func NewSecretCipher(cfg config.UploadConfig) (SecretCipher, error) {
	keys := parseSecretMasterKeys(cfg.SecretMasterKeys, cfg.SecretCurrentKeyID)
	if len(keys) == 0 {
		return nil, ErrSecretCipherUnavailable
	}
	currentKeyID := strings.TrimSpace(cfg.SecretCurrentKeyID)
	if currentKeyID == "" {
		for keyID := range keys {
			currentKeyID = keyID
			break
		}
	}
	if _, ok := keys[currentKeyID]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrSecretCipherKeyNotFound, currentKeyID)
	}

	return &secretCipher{
		currentKeyID: currentKeyID,
		keys:         keys,
		cache:        newSecretCache(cfg.SecretCacheTTL, defaultCipherCache),
	}, nil
}

func (s *secretCipher) Encrypt(_ context.Context, plaintext string) (string, error) {
	key, ok := s.keys[s.currentKeyID]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrSecretCipherKeyNotFound, s.currentKeyID)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	payload := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("%s:%s:%s", secretCipherPrefix, s.currentKeyID, base64.StdEncoding.EncodeToString(payload)), nil
}

func (s *secretCipher) Decrypt(_ context.Context, ciphertext string) (string, error) {
	if plaintext, ok := s.cache.Get(ciphertext); ok {
		return plaintext, nil
	}

	parts := strings.SplitN(ciphertext, ":", 4)
	if len(parts) != 4 || strings.Join(parts[:2], ":") != secretCipherPrefix {
		return "", ErrSecretCipherMalformed
	}
	keyID := strings.TrimSpace(parts[2])
	payload, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return "", err
	}
	key, ok := s.keys[keyID]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrSecretCipherKeyNotFound, keyID)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", ErrSecretCipherMalformed
	}
	nonce := payload[:gcm.NonceSize()]
	data := payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	result := string(plaintext)
	s.cache.Set(ciphertext, result)
	return result, nil
}

func (s *secretCipher) CurrentKeyID() string {
	return s.currentKeyID
}

func parseSecretMasterKeys(rawKeys []string, currentKeyID string) map[string][]byte {
	keys := make(map[string][]byte)
	fallbackKeyID := strings.TrimSpace(currentKeyID)
	if fallbackKeyID == "" {
		fallbackKeyID = "local"
	}
	for index, item := range rawKeys {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		keyID := fallbackKeyID
		secret := value
		if strings.Contains(value, ":") {
			parts := strings.SplitN(value, ":", 2)
			if trimmedKeyID := strings.TrimSpace(parts[0]); trimmedKeyID != "" {
				keyID = trimmedKeyID
				secret = strings.TrimSpace(parts[1])
			}
		} else if index > 0 {
			keyID = fmt.Sprintf("%s-%d", fallbackKeyID, index+1)
		}
		if secret == "" {
			continue
		}
		sum := sha256.Sum256([]byte(secret))
		keys[keyID] = sum[:]
	}
	return keys
}

func newSecretCache(ttlSeconds, capacity int) *secretCache {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	if capacity <= 0 {
		capacity = defaultCipherCache
	}
	return &secretCache{
		ttl:      time.Duration(ttlSeconds) * time.Second,
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *secretCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.items[key]
	if !ok {
		return "", false
	}
	entry := element.Value.(*secretCacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.order.Remove(element)
		delete(c.items, key)
		return "", false
	}
	c.order.MoveToFront(element)
	return entry.value, true
}

func (c *secretCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		entry := element.Value.(*secretCacheEntry)
		entry.value = value
		entry.expiresAt = time.Now().Add(c.ttl)
		c.order.MoveToFront(element)
		return
	}

	entry := &secretCacheEntry{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	element := c.order.PushFront(entry)
	c.items[key] = element
	if c.order.Len() <= c.capacity {
		return
	}

	tail := c.order.Back()
	if tail == nil {
		return
	}
	c.order.Remove(tail)
	delete(c.items, tail.Value.(*secretCacheEntry).key)
}

