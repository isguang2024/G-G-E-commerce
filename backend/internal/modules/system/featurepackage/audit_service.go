package featurepackage

import (
	"strings"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

func (s *service) ListRiskAudits(id uuid.UUID, current, size int) ([]user.RiskOperationAudit, int64, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	query := s.db.Model(&user.RiskOperationAudit{}).
		Where("object_type = ? AND object_id = ?", "feature_package", id.String())
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]user.RiskOperationAudit, 0)
	if err := query.Order("created_at DESC").Offset((current - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) recordRiskAudit(
	objectType string,
	objectID string,
	operationType string,
	beforeSummary map[string]interface{},
	afterSummary map[string]interface{},
	impactSummary map[string]interface{},
	operatorID *uuid.UUID,
	requestID string,
) error {
	item := &user.RiskOperationAudit{
		ObjectType:    strings.TrimSpace(objectType),
		ObjectID:      strings.TrimSpace(objectID),
		OperationType: strings.TrimSpace(operationType),
		BeforeSummary: beforeSummary,
		AfterSummary:  afterSummary,
		ImpactSummary: impactSummary,
		OperatorID:    operatorID,
		RequestID:     strings.TrimSpace(requestID),
	}
	return s.db.Create(item).Error
}
