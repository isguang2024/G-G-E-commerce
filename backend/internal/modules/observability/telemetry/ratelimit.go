package telemetry

import (
	"sync"
	"time"
)

// tokenBucket 是一个极简的令牌桶。单条 ingest 请求消耗 1 个 token；
// 单条 log entry 不单独扣 token，避免 hot path 上的 map 重复加锁 —
// 入口端已经限制了 maxItems=100/batch，200 req/s × 100 entry = 20k entry/s，
// DB 写入能兜住。
//
// 不用 golang.org/x/time/rate 是因为：
//  1. 我们不需要 Wait / Reserve 等复杂 API；
//  2. 不想为了 telemetry 这一个用例新增依赖；
//  3. 本实现只在一条路径上用，语义足够清晰。
type tokenBucket struct {
	mu           sync.Mutex
	tokens       float64
	capacity     float64
	refillPerSec float64
	lastRefill   time.Time
	lastSeen     time.Time
}

func newTokenBucket(capacity, refillPerSec float64) *tokenBucket {
	now := time.Now()
	return &tokenBucket{
		tokens:       capacity,
		capacity:     capacity,
		refillPerSec: refillPerSec,
		lastRefill:   now,
		lastSeen:     now,
	}
}

// allow 返回本次请求是否可以通过。consume=1 代表一条 ingest 请求。
func (b *tokenBucket) allow(consume float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.refillPerSec
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastRefill = now
	}
	b.lastSeen = now
	if b.tokens < consume {
		return false
	}
	b.tokens -= consume
	return true
}

// rateLimiter 维护 session_id / ip 两个维度的桶。超过 idleTTL 未使用
// 的桶会被 GC 回收，防止长尾 session 把内存撑爆。
//
// 并发安全：bucketMap 的 LoadOrStore 保证同一 key 不会重复创建；
// 单个 bucket 自己用 Mutex 保护 tokens 字段。
type rateLimiter struct {
	buckets      sync.Map // key: string -> *tokenBucket
	capacity     float64
	refillPerSec float64
	idleTTL      time.Duration
}

func newRateLimiter(capacity, refillPerSec float64, idleTTL time.Duration) *rateLimiter {
	return &rateLimiter{
		capacity:     capacity,
		refillPerSec: refillPerSec,
		idleTTL:      idleTTL,
	}
}

// allow 对 key 维度做限流判定。空 key 直接放行（调用方已处理"无身份"的兜底）。
func (r *rateLimiter) allow(key string) bool {
	if key == "" {
		return true
	}
	v, ok := r.buckets.Load(key)
	if !ok {
		b := newTokenBucket(r.capacity, r.refillPerSec)
		actual, loaded := r.buckets.LoadOrStore(key, b)
		v = actual
		if !loaded {
			// 新桶成功落地，跳过后续再次读取。
			return b.allow(1)
		}
	}
	return v.(*tokenBucket).allow(1)
}

// gc 清理长期空闲的 bucket。调用方通过后台 ticker 触发。
func (r *rateLimiter) gc(now time.Time) {
	r.buckets.Range(func(k, v any) bool {
		b := v.(*tokenBucket)
		b.mu.Lock()
		stale := now.Sub(b.lastSeen) > r.idleTTL
		b.mu.Unlock()
		if stale {
			r.buckets.Delete(k)
		}
		return true
	})
}
