package logpolicy

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memoryRepository struct {
	items map[uuid.UUID]*LogPolicy
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{items: map[uuid.UUID]*LogPolicy{}}
}

func (m *memoryRepository) List(_ context.Context, tenantID, pipeline string, enabled *bool) ([]LogPolicy, error) {
	targetTenant := normalizeTenantID(strings.TrimSpace(tenantID))
	result := make([]LogPolicy, 0)
	for _, item := range m.items {
		if item.TenantID != targetTenant {
			continue
		}
		if pipeline != "" && item.Pipeline != pipeline {
			continue
		}
		if enabled != nil && item.Enabled != *enabled {
			continue
		}
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority == result[j].Priority {
			return result[i].Pattern < result[j].Pattern
		}
		return result[i].Priority > result[j].Priority
	})
	return result, nil
}

func (m *memoryRepository) Get(_ context.Context, tenantID string, id uuid.UUID) (*LogPolicy, error) {
	targetTenant := normalizeTenantID(strings.TrimSpace(tenantID))
	item, ok := m.items[id]
	if !ok || item.TenantID != targetTenant {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *item
	return &cp, nil
}

func (m *memoryRepository) Create(_ context.Context, policy *LogPolicy) error {
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	policy.TenantID = normalizeTenantID(policy.TenantID)
	cp := *policy
	m.items[cp.ID] = &cp
	return nil
}

func (m *memoryRepository) Update(_ context.Context, policy *LogPolicy) error {
	current, ok := m.items[policy.ID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if current.TenantID != normalizeTenantID(policy.TenantID) {
		return gorm.ErrRecordNotFound
	}
	cp := *policy
	m.items[cp.ID] = &cp
	return nil
}

func (m *memoryRepository) Delete(_ context.Context, tenantID string, id uuid.UUID) error {
	item, ok := m.items[id]
	if !ok || item.TenantID != normalizeTenantID(tenantID) {
		return gorm.ErrRecordNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *memoryRepository) ListEnabled(ctx context.Context, tenantID, pipeline string) ([]LogPolicy, error) {
	enabled := true
	return m.List(ctx, tenantID, pipeline, &enabled)
}

func TestRepositorySemantics(t *testing.T) {
	repo := newMemoryRepository()
	ctx := context.Background()

	item := &LogPolicy{
		TenantID:   DefaultTenantID,
		Pipeline:   PipelineAudit,
		MatchField: MatchFieldAction,
		Pattern:    "system.user.create",
		Decision:   DecisionAllow,
		Priority:   10,
		Enabled:    true,
	}
	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.Get(ctx, DefaultTenantID, item.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Pattern != item.Pattern {
		t.Fatalf("Get mismatch: %+v", got)
	}

	rate := 50
	got.Decision = DecisionSample
	got.SampleRate = &rate
	got.Enabled = false
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	enabledRows, err := repo.ListEnabled(ctx, DefaultTenantID, PipelineAudit)
	if err != nil {
		t.Fatalf("ListEnabled failed: %v", err)
	}
	if len(enabledRows) != 0 {
		t.Fatalf("expected no enabled rows, got %d", len(enabledRows))
	}

	allRows, err := repo.List(ctx, DefaultTenantID, PipelineAudit, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(allRows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(allRows))
	}
	if allRows[0].SampleRate == nil || *allRows[0].SampleRate != 50 {
		t.Fatalf("sample rate mismatch: %+v", allRows[0])
	}

	if err := repo.Delete(ctx, DefaultTenantID, item.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := repo.Get(ctx, DefaultTenantID, item.ID); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestEnsureCompliancePolicies(t *testing.T) {
	repo := newMemoryRepository()
	ctx := context.Background()

	if err := EnsureCompliancePolicies(ctx, repo); err != nil {
		t.Fatalf("first ensure failed: %v", err)
	}
	if err := EnsureCompliancePolicies(ctx, repo); err != nil {
		t.Fatalf("second ensure failed: %v", err)
	}

	rows, err := repo.ListEnabled(ctx, DefaultTenantID, PipelineAudit)
	if err != nil {
		t.Fatalf("ListEnabled failed: %v", err)
	}
	if len(rows) != len(ComplianceLockedPatterns) {
		t.Fatalf("expected %d compliance rows, got %d", len(ComplianceLockedPatterns), len(rows))
	}
	for _, row := range rows {
		if row.Decision != DecisionAllow || row.Priority != 999 || row.MatchField != MatchFieldAction {
			t.Fatalf("unexpected row: %+v", row)
		}
	}
}

