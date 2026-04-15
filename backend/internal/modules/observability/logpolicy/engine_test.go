package logpolicy

import (
	"context"
	"fmt"
	"testing"
)

func TestEngineDecide_ExactMatch(t *testing.T) {
	repo := newMemoryRepository()
	_ = repo.Create(context.Background(), &LogPolicy{
		TenantID:   DefaultTenantID,
		Pipeline:   PipelineAudit,
		MatchField: MatchFieldAction,
		Pattern:    "system.user.update",
		Decision:   DecisionDeny,
		Priority:   10,
		Enabled:    true,
	})

	e := NewEngine(repo, nil)
	if err := e.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	decision := e.Decide(PipelineAudit, map[string]string{
		"action": "system.user.update",
	})
	if decision.Decision != DecisionDeny {
		t.Fatalf("decision = %s, want deny", decision.Decision)
	}
}

func TestEngineDecide_PrefixMatch(t *testing.T) {
	repo := newMemoryRepository()
	_ = repo.Create(context.Background(), &LogPolicy{
		TenantID:   DefaultTenantID,
		Pipeline:   PipelineTelemetry,
		MatchField: MatchFieldEvent,
		Pattern:    "http.*",
		Decision:   DecisionDeny,
		Priority:   10,
		Enabled:    true,
	})
	e := NewEngine(repo, nil)
	_ = e.Refresh(context.Background())

	decision := e.Decide(PipelineTelemetry, map[string]string{
		"event": "http.error",
	})
	if decision.Decision != DecisionDeny {
		t.Fatalf("decision = %s, want deny", decision.Decision)
	}
}

func TestEngineDecide_WildcardMatch(t *testing.T) {
	repo := newMemoryRepository()
	_ = repo.Create(context.Background(), &LogPolicy{
		TenantID:   DefaultTenantID,
		Pipeline:   PipelineTelemetry,
		MatchField: MatchFieldEvent,
		Pattern:    "*",
		Decision:   DecisionSample,
		Priority:   1,
		Enabled:    true,
	})
	e := NewEngine(repo, nil)
	_ = e.Refresh(context.Background())

	decision := e.Decide(PipelineTelemetry, map[string]string{
		"event": "any.event",
	})
	if decision.Decision != DecisionSample || decision.SampleRate != 100 {
		t.Fatalf("decision = %+v, want sample rate 100", decision)
	}
}

func TestEngineDecide_PriorityOrder(t *testing.T) {
	repo := newMemoryRepository()
	_ = repo.Create(context.Background(), &LogPolicy{
		TenantID:   DefaultTenantID,
		Pipeline:   PipelineAudit,
		MatchField: MatchFieldAction,
		Pattern:    "system.user.*",
		Decision:   DecisionAllow,
		Priority:   1,
		Enabled:    true,
	})
	_ = repo.Create(context.Background(), &LogPolicy{
		TenantID:   DefaultTenantID,
		Pipeline:   PipelineAudit,
		MatchField: MatchFieldAction,
		Pattern:    "system.user.update",
		Decision:   DecisionDeny,
		Priority:   100,
		Enabled:    true,
	})
	e := NewEngine(repo, nil)
	_ = e.Refresh(context.Background())

	decision := e.Decide(PipelineAudit, map[string]string{
		"action": "system.user.update",
	})
	if decision.Decision != DecisionDeny {
		t.Fatalf("decision = %s, want deny by high-priority rule", decision.Decision)
	}
}

func TestShouldKeepBySample_Distribution(t *testing.T) {
	const total = 10000
	kept := 0
	for i := 0; i < total; i++ {
		if ShouldKeepBySample(50, fmt.Sprintf("sample-%d", i)) {
			kept++
		}
	}
	ratio := float64(kept) / float64(total)
	if ratio < 0.45 || ratio > 0.55 {
		t.Fatalf("sample ratio = %.4f, want around 0.50", ratio)
	}
}

func TestEngineDecide_ComplianceLockOverridesDeny(t *testing.T) {
	repo := newMemoryRepository()
	_ = repo.Create(context.Background(), &LogPolicy{
		TenantID:   DefaultTenantID,
		Pipeline:   PipelineAudit,
		MatchField: MatchFieldAction,
		Pattern:    "observability.policy.*",
		Decision:   DecisionDeny,
		Priority:   999,
		Enabled:    true,
	})
	e := NewEngine(repo, nil)
	_ = e.Refresh(context.Background())

	decision := e.Decide(PipelineAudit, map[string]string{
		"action": "observability.policy.delete",
	})
	if decision.Decision != DecisionAllow {
		t.Fatalf("decision = %s, want allow by compliance lock", decision.Decision)
	}
}
