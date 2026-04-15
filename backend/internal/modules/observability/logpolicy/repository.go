package logpolicy

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context, tenantID, pipeline string, enabled *bool) ([]LogPolicy, error)
	Get(ctx context.Context, tenantID string, id uuid.UUID) (*LogPolicy, error)
	Create(ctx context.Context, policy *LogPolicy) error
	Update(ctx context.Context, policy *LogPolicy) error
	Delete(ctx context.Context, tenantID string, id uuid.UUID) error
	ListEnabled(ctx context.Context, tenantID, pipeline string) ([]LogPolicy, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) List(ctx context.Context, tenantID, pipeline string, enabled *bool) ([]LogPolicy, error) {
	targetTenant := normalizeTenantID(strings.TrimSpace(tenantID))
	query := r.db.WithContext(ctx).Model(&LogPolicy{}).Where("tenant_id = ?", targetTenant)
	if p := strings.TrimSpace(pipeline); p != "" {
		query = query.Where("pipeline = ?", p)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	var items []LogPolicy
	if err := query.Order("priority DESC, created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormRepository) Get(ctx context.Context, tenantID string, id uuid.UUID) (*LogPolicy, error) {
	targetTenant := normalizeTenantID(strings.TrimSpace(tenantID))
	var item LogPolicy
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", targetTenant, id).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, policy *LogPolicy) error {
	if policy == nil {
		return errors.New("logpolicy: nil policy")
	}
	policy.TenantID = normalizeTenantID(strings.TrimSpace(policy.TenantID))
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *GormRepository) Update(ctx context.Context, policy *LogPolicy) error {
	if policy == nil {
		return errors.New("logpolicy: nil policy")
	}
	targetTenant := normalizeTenantID(strings.TrimSpace(policy.TenantID))
	updates := map[string]any{
		"pipeline":    policy.Pipeline,
		"match_field": policy.MatchField,
		"pattern":     policy.Pattern,
		"decision":    policy.Decision,
		"sample_rate": policy.SampleRate,
		"priority":    policy.Priority,
		"enabled":     policy.Enabled,
		"note":        policy.Note,
	}
	tx := r.db.WithContext(ctx).
		Model(&LogPolicy{}).
		Where("tenant_id = ? AND id = ?", targetTenant, policy.ID).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) Delete(ctx context.Context, tenantID string, id uuid.UUID) error {
	targetTenant := normalizeTenantID(strings.TrimSpace(tenantID))
	tx := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", targetTenant, id).
		Delete(&LogPolicy{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) ListEnabled(ctx context.Context, tenantID, pipeline string) ([]LogPolicy, error) {
	enabled := true
	return r.List(ctx, tenantID, pipeline, &enabled)
}

func EnsureCompliancePolicies(ctx context.Context, repo Repository) error {
	if repo == nil {
		return nil
	}
	current, err := repo.List(ctx, DefaultTenantID, PipelineAudit, nil)
	if err != nil {
		return err
	}
	byPattern := make(map[string]*LogPolicy, len(current))
	for i := range current {
		item := current[i]
		if item.MatchField != MatchFieldAction {
			continue
		}
		byPattern[item.Pattern] = &item
	}

	for _, pattern := range ComplianceLockedPatterns {
		if existing, ok := byPattern[pattern]; ok {
			existing.Decision = DecisionAllow
			existing.SampleRate = nil
			existing.Priority = 999
			existing.Enabled = true
			existing.Note = "compliance lock builtin"
			if err := repo.Update(ctx, existing); err != nil {
				return err
			}
			continue
		}
		policy := &LogPolicy{
			TenantID:   DefaultTenantID,
			Pipeline:   PipelineAudit,
			MatchField: MatchFieldAction,
			Pattern:    pattern,
			Decision:   DecisionAllow,
			SampleRate: nil,
			Priority:   999,
			Enabled:    true,
			Note:       "compliance lock builtin",
		}
		if err := repo.Create(ctx, policy); err != nil {
			return err
		}
	}
	return nil
}
