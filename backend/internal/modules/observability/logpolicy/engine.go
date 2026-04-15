package logpolicy

import (
	"context"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Decision struct {
	Decision   string     `json:"decision"`
	SampleRate int        `json:"sample_rate"`
	Matched    *LogPolicy `json:"matched,omitempty"`
}

type Engine interface {
	Decide(pipeline string, fields map[string]string) Decision
	Refresh(ctx context.Context) error
	Start(ctx context.Context)
}

type compiledRule struct {
	policy       LogPolicy
	matchField   string
	pattern      string
	matchAll     bool
	prefixMatch  bool
	prefixTarget string
}

type engine struct {
	repo Repository
	log  *zap.Logger

	mu    sync.RWMutex
	rules map[string][]compiledRule
}

func NewEngine(repo Repository, log *zap.Logger) Engine {
	if log == nil {
		log = zap.NewNop()
	}
	return &engine{
		repo:  repo,
		log:   log.Named("logpolicy.engine"),
		rules: map[string][]compiledRule{},
	}
}

func (e *engine) Start(ctx context.Context) {
	if err := e.Refresh(ctx); err != nil {
		e.log.Warn("logpolicy.refresh.initial_failed", zap.Error(err))
	}
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := e.Refresh(ctx); err != nil {
					e.log.Warn("logpolicy.refresh.failed", zap.Error(err))
				}
			}
		}
	}()
}

func (e *engine) Refresh(ctx context.Context) error {
	if e.repo == nil {
		return nil
	}
	next := map[string][]compiledRule{
		PipelineAudit:     {},
		PipelineTelemetry: {},
	}
	for _, pipeline := range []string{PipelineAudit, PipelineTelemetry} {
		items, err := e.repo.ListEnabled(ctx, DefaultTenantID, pipeline)
		if err != nil {
			return err
		}
		compiled := make([]compiledRule, 0, len(items))
		for _, item := range items {
			rule := compileRule(item)
			compiled = append(compiled, rule)
		}
		next[pipeline] = compiled
	}

	e.mu.Lock()
	e.rules = next
	e.mu.Unlock()
	return nil
}

func (e *engine) Decide(pipeline string, fields map[string]string) Decision {
	p := normalizePipeline(pipeline)
	if p == "" {
		return Decision{Decision: DecisionAllow}
	}

	normalized := normalizeFields(fields)
	e.mu.RLock()
	rules := e.rules[p]
	e.mu.RUnlock()
	for _, rule := range rules {
		if !rule.match(normalized) {
			continue
		}
		matched := rule.policy
		decision := Decision{
			Decision:   rule.policy.Decision,
			SampleRate: normalizeSampleRate(rule.policy.SampleRate),
			Matched:    &matched,
		}
		if p == PipelineAudit && decision.Decision == DecisionDeny {
			action := normalized[MatchFieldAction]
			if isComplianceLockedAction(action) {
				e.log.Warn("logpolicy.compliance_lock_override",
					zap.String("action", action),
					zap.String("pattern", rule.policy.Pattern),
				)
				decision.Decision = DecisionAllow
				decision.SampleRate = 100
			}
		}
		return decision
	}
	return Decision{Decision: DecisionAllow}
}

func compileRule(item LogPolicy) compiledRule {
	pattern := strings.TrimSpace(item.Pattern)
	rule := compiledRule{
		policy:     item,
		matchField: strings.ToLower(strings.TrimSpace(item.MatchField)),
		pattern:    pattern,
	}
	switch {
	case pattern == "*" || pattern == "":
		rule.matchAll = true
	case strings.HasSuffix(pattern, "*"):
		rule.prefixMatch = true
		rule.prefixTarget = strings.TrimSuffix(pattern, "*")
	default:
	}
	return rule
}

func (r compiledRule) match(fields map[string]string) bool {
	value := fields[r.matchField]
	if r.matchAll {
		return true
	}
	if r.prefixMatch {
		return strings.HasPrefix(value, r.prefixTarget)
	}
	return value == r.pattern
}

func normalizePipeline(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case PipelineAudit:
		return PipelineAudit
	case PipelineTelemetry:
		return PipelineTelemetry
	default:
		return ""
	}
}

func normalizeFields(fields map[string]string) map[string]string {
	result := make(map[string]string, len(fields))
	for k, v := range fields {
		result[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	return result
}

func normalizeSampleRate(rate *int) int {
	if rate == nil {
		return 100
	}
	if *rate <= 0 {
		return 0
	}
	if *rate > 100 {
		return 100
	}
	return *rate
}

func isComplianceLockedAction(action string) bool {
	target := strings.TrimSpace(action)
	for _, pattern := range ComplianceLockedPatterns {
		if pattern == target {
			return true
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(target, prefix) {
				return true
			}
		}
	}
	return false
}

// ShouldKeepBySample 返回给定样本键在 sample_rate 百分比规则下是否保留。
// rate<=0 永不保留，rate>=100 总是保留。
func ShouldKeepBySample(sampleRate int, sampleKey string) bool {
	if sampleRate <= 0 {
		return false
	}
	if sampleRate >= 100 {
		return true
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(sampleKey))
	return int(h.Sum32()%100) < sampleRate
}

