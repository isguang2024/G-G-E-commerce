package siteconfig

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"go.uber.org/zap"

	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/system/models"
)

// 构造一对共享同一 Redis 的缓存实例，并启动 Pub/Sub 订阅。
func newRedisCachePair(t *testing.T) (*resolvedConfigCache, *resolvedConfigCache, func()) {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}

	port, err := strconv.Atoi(server.Port())
	if err != nil {
		server.Close()
		t.Fatalf("strconv.Atoi(redis port) error = %v", err)
	}
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: server.Host(),
			Port: port,
		},
	}
	cacheA, err := newResolvedConfigCache(cfg, zap.NewNop())
	if err != nil {
		server.Close()
		t.Fatalf("newResolvedConfigCache(cacheA) error = %v", err)
	}
	cacheB, err := newResolvedConfigCache(cfg, zap.NewNop())
	if err != nil {
		server.Close()
		t.Fatalf("newResolvedConfigCache(cacheB) error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cacheA.Start(ctx)
	cacheB.Start(ctx)
	// 等待订阅就绪，避免发布时 B 尚未 subscribe。
	time.Sleep(80 * time.Millisecond)

	cleanup := func() {
		cancel()
		server.Close()
	}
	return cacheA, cacheB, cleanup
}

// waitUntilLocalCacheMissing 轮询等待本地缓存被清除；超时即失败。
func waitUntilLocalCacheMissing(t *testing.T, cache *resolvedConfigCache, key string, dest any) {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(time.Second)
	for {
		if _, ok := cache.getLocal(key); !ok {
			return
		}
		// getLocal 没过期就反复 poll；也允许 Get 触发 TTL 清理路径。
		_ = cache.Get(ctx, key, dest)
		if time.Now().After(deadline) {
			t.Fatalf("local cache for %q should be invalidated within timeout", key)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func sampleResolveResult() *ResolveResult {
	return &ResolveResult{
		Version: "v1",
		Items: map[string]ResolvedItem{
			"site.name": {
				Value:     models.MetaJSON{"value": "GG Demo"},
				Source:    ResolveSourceGlobal,
				ValueType: "string",
			},
		},
	}
}

// TestResolvedConfigCacheKeyInvalidationBroadcast 验证按精确 key 失效时跨实例会同步清本地。
func TestResolvedConfigCacheKeyInvalidationBroadcast(t *testing.T) {
	cacheA, cacheB, cleanup := newRedisCachePair(t)
	defer cleanup()

	ctx := context.Background()
	key := siteConfigResolvedCache("default", ScopeTypeApp, "admin", []string{"site.name"}, nil)

	cacheA.Set(ctx, key, sampleResolveResult())

	// cacheB 首次读通过 Redis 命中，并写入本地。
	var got ResolveResult
	if !cacheB.Get(ctx, key, &got) {
		t.Fatalf("cacheB should read value from redis")
	}
	if _, ok := cacheB.getLocal(key); !ok {
		t.Fatalf("cacheB should have populated local cache after redis hit")
	}

	// cacheA 失效该 key，cacheB 应在短时间内清掉本地副本。
	cacheA.Invalidate(ctx, []string{key}, nil)
	waitUntilLocalCacheMissing(t, cacheB, key, &ResolveResult{})
}

// TestResolvedConfigCachePrefixInvalidationBroadcast 验证按 prefix 失效时跨实例会清理本地 + Redis。
func TestResolvedConfigCachePrefixInvalidationBroadcast(t *testing.T) {
	cacheA, cacheB, cleanup := newRedisCachePair(t)
	defer cleanup()

	ctx := context.Background()
	keyA := siteConfigResolvedCache("default", ScopeTypeApp, "admin", []string{"site.name"}, nil)
	keyB := siteConfigResolvedCache("default", ScopeTypeApp, "admin", []string{"site.logo"}, nil)
	prefix := siteConfigResolvedPrefix("default")

	cacheA.Set(ctx, keyA, sampleResolveResult())
	cacheA.Set(ctx, keyB, sampleResolveResult())

	// cacheB 读两次，确保两个 key 都在 cacheB 的本地。
	var sinkA, sinkB ResolveResult
	if !cacheB.Get(ctx, keyA, &sinkA) || !cacheB.Get(ctx, keyB, &sinkB) {
		t.Fatalf("cacheB should fetch both keys through redis")
	}

	cacheA.Invalidate(ctx, nil, []string{prefix})

	waitUntilLocalCacheMissing(t, cacheB, keyA, &ResolveResult{})
	waitUntilLocalCacheMissing(t, cacheB, keyB, &ResolveResult{})

	// prefix 失效后 Redis 上两个 key 也应被清理（通过 SCAN + DEL）。
	remaining, err := cacheA.redis.Exists(ctx, keyA, keyB).Result()
	if err != nil {
		t.Fatalf("redis Exists error = %v", err)
	}
	if remaining != 0 {
		t.Fatalf("redis keys under prefix should be deleted, got %d remaining", remaining)
	}
}

// TestResolvedConfigCacheLocalFallback 验证纯本地模式（无 Redis）读写与 TTL 过期行为。
func TestResolvedConfigCacheLocalFallback(t *testing.T) {
	cache := newLocalResolvedConfigCache(40 * time.Millisecond)
	ctx := context.Background()
	key := "siteconfig:local:test"

	cache.Set(ctx, key, sampleResolveResult())
	var got ResolveResult
	if !cache.Get(ctx, key, &got) {
		t.Fatalf("local cache should hit right after Set")
	}
	if got.Items["site.name"].Source != ResolveSourceGlobal {
		t.Fatalf("decoded value mismatch: %+v", got.Items["site.name"])
	}

	// 等待 TTL 过期，Get 应未命中并清理本地。
	time.Sleep(80 * time.Millisecond)
	var after ResolveResult
	if cache.Get(ctx, key, &after) {
		t.Fatalf("local cache should expire after ttl")
	}
	if _, ok := cache.getLocal(key); ok {
		t.Fatalf("expired key should be removed from local map")
	}
}

// TestFingerprintResolveInputStableOrdering 验证指纹对顺序不敏感。
func TestFingerprintResolveInputStableOrdering(t *testing.T) {
	a := fingerprintResolveInput([]string{"b", "a"}, []string{"y", "x"})
	b := fingerprintResolveInput([]string{"a", "b"}, []string{"x", "y"})
	if a != b {
		t.Fatalf("fingerprint should be stable across input ordering: %s != %s", a, b)
	}
	c := fingerprintResolveInput([]string{"a", "c"}, []string{"x", "y"})
	if c == a {
		t.Fatalf("fingerprint should differ when keys differ")
	}
}

