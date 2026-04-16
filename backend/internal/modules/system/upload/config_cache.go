package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/maben/backend/internal/config"
)

const (
	uploadConfigCacheTTLDefault = 5 * time.Minute
	uploadInvalidationChannel   = "upload:config:invalidate"
)

type resolvedConfigCache struct {
	logger     *zap.Logger
	ttl        time.Duration
	redis      *redis.Client
	startOnce  sync.Once
	localMu    sync.RWMutex
	localItems map[string]localCacheItem
}

type localCacheItem struct {
	payload   []byte
	expiresAt time.Time
}

type uploadCacheInvalidation struct {
	Keys     []string `json:"keys,omitempty"`
	Prefixes []string `json:"prefixes,omitempty"`
}

func newResolvedConfigCache(cfg *config.Config, logger *zap.Logger) (*resolvedConfigCache, error) {
	if cfg == nil {
		return nil, nil
	}
	if strings.TrimSpace(cfg.Redis.Host) == "" || cfg.Redis.Port <= 0 {
		return nil, nil
	}
	cacheTTL := uploadConfigCacheTTLDefault
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     20,
		MinIdleConns: 5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect upload config cache redis: %w", err)
	}
	return &resolvedConfigCache{
		logger:     logger,
		ttl:        cacheTTL,
		redis:      client,
		localItems: make(map[string]localCacheItem),
	}, nil
}

func newLocalResolvedConfigCache(ttl time.Duration) *resolvedConfigCache {
	if ttl <= 0 {
		ttl = uploadConfigCacheTTLDefault
	}
	return &resolvedConfigCache{
		ttl:        ttl,
		localItems: make(map[string]localCacheItem),
	}
}

func (c *resolvedConfigCache) Start(ctx context.Context) {
	if c == nil || c.redis == nil {
		return
	}
	c.startOnce.Do(func() {
		go c.consumeInvalidation(ctx)
	})
}

func (c *resolvedConfigCache) consumeInvalidation(ctx context.Context) {
	pubsub := c.redis.Subscribe(ctx, uploadInvalidationChannel)
	defer func() {
		_ = pubsub.Close()
	}()

	channel := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-channel:
			if !ok {
				return
			}
			c.handleInvalidationPayload(msg.Payload)
		}
	}
}

func (c *resolvedConfigCache) Get(ctx context.Context, key string, dest any) bool {
	if c == nil {
		return false
	}
	if payload, ok := c.getLocal(key); ok {
		if err := json.Unmarshal(payload, dest); err == nil {
			return true
		}
		c.deleteLocalKeys(key)
	}
	if c.redis == nil {
		return false
	}
	payload, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			c.logWarn("read upload config cache failed", zap.String("key", key), zap.Error(err))
		}
		return false
	}
	c.setLocal(key, payload)
	if err := json.Unmarshal(payload, dest); err != nil {
		c.logWarn("decode upload config cache failed", zap.String("key", key), zap.Error(err))
		c.Invalidate(ctx, []string{key}, nil)
		return false
	}
	return true
}

func (c *resolvedConfigCache) Set(ctx context.Context, key string, value any) {
	if c == nil {
		return
	}
	payload, err := json.Marshal(value)
	if err != nil {
		c.logWarn("encode upload config cache failed", zap.String("key", key), zap.Error(err))
		return
	}
	c.setLocal(key, payload)
	if c.redis == nil {
		return
	}
	if err := c.redis.Set(ctx, key, payload, c.ttl).Err(); err != nil {
		c.logWarn("write upload config cache failed", zap.String("key", key), zap.Error(err))
	}
}

func (c *resolvedConfigCache) Invalidate(ctx context.Context, keys []string, prefixes []string) {
	if c == nil {
		return
	}
	keys = compactStrings(keys)
	prefixes = compactStrings(prefixes)
	c.deleteLocalKeys(keys...)
	c.deleteLocalPrefixes(prefixes...)
	if c.redis != nil {
		if len(keys) > 0 {
			if err := c.redis.Del(ctx, keys...).Err(); err != nil {
				c.logWarn("delete upload config cache keys failed", zap.Strings("keys", keys), zap.Error(err))
			}
		}
		for _, prefix := range prefixes {
			if err := c.deleteRemotePrefix(ctx, prefix); err != nil {
				c.logWarn("delete upload config cache prefix failed", zap.String("prefix", prefix), zap.Error(err))
			}
		}
	}
	c.publishInvalidation(ctx, uploadCacheInvalidation{
		Keys:     keys,
		Prefixes: prefixes,
	})
}

func (c *resolvedConfigCache) publishInvalidation(ctx context.Context, event uploadCacheInvalidation) {
	if c == nil || c.redis == nil {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		c.logWarn("encode upload invalidation payload failed", zap.Error(err))
		return
	}
	if err := c.redis.Publish(ctx, uploadInvalidationChannel, payload).Err(); err != nil {
		c.logWarn("publish upload invalidation failed", zap.Error(err))
	}
}

func (c *resolvedConfigCache) handleInvalidationPayload(payload string) {
	if c == nil || strings.TrimSpace(payload) == "" {
		return
	}
	var event uploadCacheInvalidation
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		c.logWarn("decode upload invalidation payload failed", zap.Error(err))
		return
	}
	c.deleteLocalKeys(event.Keys...)
	c.deleteLocalPrefixes(event.Prefixes...)
}

func (c *resolvedConfigCache) getLocal(key string) ([]byte, bool) {
	c.localMu.RLock()
	item, ok := c.localItems[key]
	c.localMu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(item.expiresAt) {
		c.deleteLocalKeys(key)
		return nil, false
	}
	return append([]byte(nil), item.payload...), true
}

func (c *resolvedConfigCache) setLocal(key string, payload []byte) {
	c.localMu.Lock()
	c.localItems[key] = localCacheItem{
		payload:   append([]byte(nil), payload...),
		expiresAt: time.Now().Add(c.ttl),
	}
	c.localMu.Unlock()
}

func (c *resolvedConfigCache) deleteLocalKeys(keys ...string) {
	if len(keys) == 0 {
		return
	}
	c.localMu.Lock()
	for _, key := range keys {
		delete(c.localItems, key)
	}
	c.localMu.Unlock()
}

func (c *resolvedConfigCache) deleteLocalPrefixes(prefixes ...string) {
	if len(prefixes) == 0 {
		return
	}
	c.localMu.Lock()
	for key := range c.localItems {
		for _, prefix := range prefixes {
			if prefix != "" && strings.HasPrefix(key, prefix) {
				delete(c.localItems, key)
				break
			}
		}
	}
	c.localMu.Unlock()
}

func (c *resolvedConfigCache) deleteRemotePrefix(ctx context.Context, prefix string) error {
	if c == nil || c.redis == nil || prefix == "" {
		return nil
	}
	var cursor uint64
	pattern := prefix + "*"
	for {
		keys, next, err := c.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.redis.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func (c *resolvedConfigCache) logWarn(msg string, fields ...zap.Field) {
	if c == nil || c.logger == nil {
		return
	}
	c.logger.Warn(msg, fields...)
}

func uploadKeyCacheKey(tenantID, code string) string {
	return fmt.Sprintf("upload_key:%s:%s", normalizeTenantID(tenantID), strings.TrimSpace(code))
}

func uploadRuleCacheKey(uploadKeyID uuid.UUID) string {
	return "upload_rules:" + uploadKeyID.String()
}

func uploadKeyCachePrefix(tenantID string) string {
	return fmt.Sprintf("upload_key:%s:", normalizeTenantID(tenantID))
}

func compactStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

