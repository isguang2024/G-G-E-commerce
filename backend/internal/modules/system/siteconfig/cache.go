package siteconfig

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/maben/backend/internal/config"
)

const (
	siteConfigCacheTTLDefault = 5 * time.Minute
	siteConfigInvalidateChan  = "siteconfig:config:invalidate"
	globalAppKeyPlaceholder   = "_global"
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

type siteConfigInvalidation struct {
	Keys     []string `json:"keys,omitempty"`
	Prefixes []string `json:"prefixes,omitempty"`
}

// newResolvedConfigCache 创建缓存实例；Redis 不可用时返回 nil，调用方降级为无缓存。
func newResolvedConfigCache(cfg *config.Config, logger *zap.Logger) (*resolvedConfigCache, error) {
	if cfg == nil {
		return nil, nil
	}
	if strings.TrimSpace(cfg.Redis.Host) == "" || cfg.Redis.Port <= 0 {
		return nil, nil
	}
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
		return nil, fmt.Errorf("connect site config cache redis: %w", err)
	}
	return &resolvedConfigCache{
		logger:     logger,
		ttl:        siteConfigCacheTTLDefault,
		redis:      client,
		localItems: make(map[string]localCacheItem),
	}, nil
}

// newLocalResolvedConfigCache 仅创建本地缓存（测试或 Redis 不可用时使用）。
func newLocalResolvedConfigCache(ttl time.Duration) *resolvedConfigCache {
	if ttl <= 0 {
		ttl = siteConfigCacheTTLDefault
	}
	return &resolvedConfigCache{
		ttl:        ttl,
		localItems: make(map[string]localCacheItem),
	}
}

// Start 启动 Pub/Sub 订阅协程，用于跨实例本地缓存失效。
func (c *resolvedConfigCache) Start(ctx context.Context) {
	if c == nil || c.redis == nil {
		return
	}
	c.startOnce.Do(func() {
		go c.consumeInvalidation(ctx)
	})
}

func (c *resolvedConfigCache) consumeInvalidation(ctx context.Context) {
	pubsub := c.redis.Subscribe(ctx, siteConfigInvalidateChan)
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
			c.logWarn("read site config cache failed", zap.String("key", key), zap.Error(err))
		}
		return false
	}
	c.setLocal(key, payload)
	if err := json.Unmarshal(payload, dest); err != nil {
		c.logWarn("decode site config cache failed", zap.String("key", key), zap.Error(err))
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
		c.logWarn("encode site config cache failed", zap.String("key", key), zap.Error(err))
		return
	}
	c.setLocal(key, payload)
	if c.redis == nil {
		return
	}
	if err := c.redis.Set(ctx, key, payload, c.ttl).Err(); err != nil {
		c.logWarn("write site config cache failed", zap.String("key", key), zap.Error(err))
	}
}

// Invalidate 同时清理本地、Redis 并通过 Pub/Sub 广播给其它实例。
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
				c.logWarn("delete site config cache keys failed", zap.Strings("keys", keys), zap.Error(err))
			}
		}
		for _, prefix := range prefixes {
			if err := c.deleteRemotePrefix(ctx, prefix); err != nil {
				c.logWarn("delete site config cache prefix failed", zap.String("prefix", prefix), zap.Error(err))
			}
		}
	}
	c.publishInvalidation(ctx, siteConfigInvalidation{
		Keys:     keys,
		Prefixes: prefixes,
	})
}

func (c *resolvedConfigCache) publishInvalidation(ctx context.Context, event siteConfigInvalidation) {
	if c == nil || c.redis == nil {
		return
	}
	if len(event.Keys) == 0 && len(event.Prefixes) == 0 {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		c.logWarn("encode site config invalidation payload failed", zap.Error(err))
		return
	}
	if err := c.redis.Publish(ctx, siteConfigInvalidateChan, payload).Err(); err != nil {
		c.logWarn("publish site config invalidation failed", zap.Error(err))
	}
}

func (c *resolvedConfigCache) handleInvalidationPayload(payload string) {
	if c == nil || strings.TrimSpace(payload) == "" {
		return
	}
	var event siteConfigInvalidation
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		c.logWarn("decode site config invalidation payload failed", zap.Error(err))
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

// ---- 缓存键工具 ----

func normalizeAppKey(appKey string) string {
	appKey = strings.TrimSpace(appKey)
	if appKey == "" {
		return globalAppKeyPlaceholder
	}
	return appKey
}

func normalizeTenantID(tenantID string) string {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "default"
	}
	return tenantID
}

func siteConfigKeyCache(tenantID, appKey, configKey string) string {
	return fmt.Sprintf("siteconfig:cfg:%s:%s:%s", normalizeTenantID(tenantID), normalizeAppKey(appKey), configKey)
}

func siteConfigResolvedCache(tenantID, appKey string, keys, setCodes []string) string {
	hash := fingerprintResolveInput(keys, setCodes)
	return fmt.Sprintf("siteconfig:resolved:%s:%s:%s", normalizeTenantID(tenantID), normalizeAppKey(appKey), hash)
}

func siteConfigSetsCache(tenantID string) string {
	return fmt.Sprintf("siteconfig:sets:%s", normalizeTenantID(tenantID))
}

func siteConfigCfgPrefix(tenantID string) string {
	return fmt.Sprintf("siteconfig:cfg:%s:", normalizeTenantID(tenantID))
}

func siteConfigResolvedPrefix(tenantID string) string {
	return fmt.Sprintf("siteconfig:resolved:%s:", normalizeTenantID(tenantID))
}

func siteConfigResolvedPrefixApp(tenantID, appKey string) string {
	return fmt.Sprintf("siteconfig:resolved:%s:%s:", normalizeTenantID(tenantID), normalizeAppKey(appKey))
}

func fingerprintResolveInput(keys, setCodes []string) string {
	sortedKeys := append([]string(nil), compactStrings(keys)...)
	sort.Strings(sortedKeys)
	sortedSets := append([]string(nil), compactStrings(setCodes)...)
	sort.Strings(sortedSets)
	h := sha1.New()
	_, _ = h.Write([]byte(strings.Join(sortedKeys, ",")))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write([]byte(strings.Join(sortedSets, ",")))
	return hex.EncodeToString(h.Sum(nil))[:16]
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

