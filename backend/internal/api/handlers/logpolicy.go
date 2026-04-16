package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/logpolicy"
	"github.com/maben/backend/internal/pkg/logger"
)

func (h *logPolicyAPIHandler) ListLogPolicies(ctx context.Context, params gen.ListLogPoliciesParams) (gen.ListLogPoliciesRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.ListLogPoliciesUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.policyRepo == nil {
		return &gen.ListLogPoliciesInternalServerError{Code: 500, Message: "日志策略服务未就绪"}, nil
	}

	var pipeline string
	if params.Pipeline.Set {
		pipeline = string(params.Pipeline.Value)
	}
	var enabled *bool
	if params.Enabled.Set {
		val := params.Enabled.Value
		enabled = &val
	}

	items, err := h.policyRepo.List(ctx, logger.TenantFromContext(ctx), pipeline, enabled)
	if err != nil {
		h.logger.Error("list log policies failed", zap.Error(err))
		return &gen.ListLogPoliciesInternalServerError{Code: 500, Message: "查询日志策略失败"}, nil
	}

	current, size := observabilityPagination(params.Current, params.Size)
	total := len(items)
	start := (current - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}

	records := make([]gen.LogPolicyItem, 0, end-start)
	for i := start; i < end; i++ {
		records = append(records, toGenLogPolicyItem(&items[i]))
	}
	return &gen.LogPolicyList{
		Records: records,
		Total:   int64(total),
		Current: current,
		Size:    size,
	}, nil
}

func (h *logPolicyAPIHandler) CreateLogPolicy(ctx context.Context, req *gen.LogPolicyCreateRequest) (gen.CreateLogPolicyRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.CreateLogPolicyUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.policyRepo == nil {
		return &gen.CreateLogPolicyInternalServerError{Code: 500, Message: "日志策略服务未就绪"}, nil
	}
	if req == nil {
		return &gen.CreateLogPolicyBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}

	pipeline := string(req.Pipeline)
	matchField := string(req.MatchField)
	decision := string(req.Decision)
	pattern := strings.TrimSpace(req.Pattern)
	if pattern == "" {
		return &gen.CreateLogPolicyBadRequest{Code: 400, Message: "pattern 不能为空"}, nil
	}

	var sampleRate *int
	if req.SampleRate.Set {
		rate := req.SampleRate.Value
		sampleRate = &rate
	}
	if err := validatePolicyInput(pipeline, matchField, decision, sampleRate); err != nil {
		return &gen.CreateLogPolicyBadRequest{Code: 400, Message: err.Error()}, nil
	}

	enabled := true
	if req.Enabled.Set {
		enabled = req.Enabled.Value
	}
	priority := 0
	if req.Priority.Set {
		priority = req.Priority.Value
	}
	tenantID := logger.TenantFromContext(ctx)
	note := strings.TrimSpace(optString(req.Note))

	policy := &logpolicy.LogPolicy{
		TenantID:   tenantID,
		Pipeline:   pipeline,
		MatchField: matchField,
		Pattern:    pattern,
		Decision:   decision,
		SampleRate: sampleRate,
		Priority:   priority,
		Enabled:    enabled,
		Note:       note,
	}
	if actorID, ok := userIDFromContext(ctx); ok && actorID != uuid.Nil {
		policy.CreatedBy = &actorID
	}

	if err := h.policyRepo.Create(ctx, policy); err != nil {
		if isConflictError(err) {
			return &gen.CreateLogPolicyConflict{Code: 409, Message: "日志策略已存在"}, nil
		}
		h.logger.Error("create log policy failed", zap.Error(err))
		return &gen.CreateLogPolicyInternalServerError{Code: 500, Message: "创建日志策略失败"}, nil
	}

	if err := h.refreshPolicyEngine(ctx); err != nil {
		return &gen.CreateLogPolicyInternalServerError{Code: 500, Message: "策略刷新失败"}, nil
	}
	h.recordPolicyAudit(ctx, "observability.policy.create", policy)
	item := toGenLogPolicyItem(policy)
	return &item, nil
}

func (h *logPolicyAPIHandler) UpdateLogPolicy(ctx context.Context, req *gen.LogPolicyUpdateRequest, params gen.UpdateLogPolicyParams) (gen.UpdateLogPolicyRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.UpdateLogPolicyUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.policyRepo == nil {
		return &gen.UpdateLogPolicyInternalServerError{Code: 500, Message: "日志策略服务未就绪"}, nil
	}
	if req == nil {
		return &gen.UpdateLogPolicyBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}

	existing, err := h.policyRepo.Get(ctx, logger.TenantFromContext(ctx), params.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.UpdateLogPolicyNotFound{Code: 404, Message: "日志策略不存在"}, nil
		}
		h.logger.Error("get log policy failed", zap.Error(err))
		return &gen.UpdateLogPolicyInternalServerError{Code: 500, Message: "查询日志策略失败"}, nil
	}
	if isComplianceLockedPolicy(existing) {
		return &gen.UpdateLogPolicyConflict{Code: 409, Message: "compliance lock 策略禁止修改"}, nil
	}

	if req.MatchField.Set {
		existing.MatchField = string(req.MatchField.Value)
	}
	if req.Pattern.Set {
		existing.Pattern = strings.TrimSpace(req.Pattern.Value)
	}
	if req.Decision.Set {
		existing.Decision = string(req.Decision.Value)
	}
	if req.SampleRate.Set {
		if req.SampleRate.Null {
			existing.SampleRate = nil
		} else {
			rate := req.SampleRate.Value
			existing.SampleRate = &rate
		}
	}
	if req.Priority.Set {
		existing.Priority = req.Priority.Value
	}
	if req.Enabled.Set {
		existing.Enabled = req.Enabled.Value
	}
	if req.Note.Set {
		if req.Note.Null {
			existing.Note = ""
		} else {
			existing.Note = strings.TrimSpace(req.Note.Value)
		}
	}

	if strings.TrimSpace(existing.Pattern) == "" {
		return &gen.UpdateLogPolicyBadRequest{Code: 400, Message: "pattern 不能为空"}, nil
	}
	if err := validatePolicyInput(existing.Pipeline, existing.MatchField, existing.Decision, existing.SampleRate); err != nil {
		return &gen.UpdateLogPolicyBadRequest{Code: 400, Message: err.Error()}, nil
	}

	if err := h.policyRepo.Update(ctx, existing); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.UpdateLogPolicyNotFound{Code: 404, Message: "日志策略不存在"}, nil
		}
		if isConflictError(err) {
			return &gen.UpdateLogPolicyConflict{Code: 409, Message: "日志策略冲突"}, nil
		}
		h.logger.Error("update log policy failed", zap.Error(err))
		return &gen.UpdateLogPolicyInternalServerError{Code: 500, Message: "更新日志策略失败"}, nil
	}

	if err := h.refreshPolicyEngine(ctx); err != nil {
		return &gen.UpdateLogPolicyInternalServerError{Code: 500, Message: "策略刷新失败"}, nil
	}
	h.recordPolicyAudit(ctx, "observability.policy.update", existing)

	latest, err := h.policyRepo.Get(ctx, logger.TenantFromContext(ctx), params.ID)
	if err != nil {
		h.logger.Error("reload log policy failed", zap.Error(err))
		return &gen.UpdateLogPolicyInternalServerError{Code: 500, Message: "查询日志策略失败"}, nil
	}
	item := toGenLogPolicyItem(latest)
	return &item, nil
}

func (h *logPolicyAPIHandler) DeleteLogPolicy(ctx context.Context, params gen.DeleteLogPolicyParams) (gen.DeleteLogPolicyRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.DeleteLogPolicyUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.policyRepo == nil {
		return &gen.DeleteLogPolicyInternalServerError{Code: 500, Message: "日志策略服务未就绪"}, nil
	}

	existing, err := h.policyRepo.Get(ctx, logger.TenantFromContext(ctx), params.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.DeleteLogPolicyNotFound{Code: 404, Message: "日志策略不存在"}, nil
		}
		h.logger.Error("get log policy failed", zap.Error(err))
		return &gen.DeleteLogPolicyInternalServerError{Code: 500, Message: "查询日志策略失败"}, nil
	}
	if isComplianceLockedPolicy(existing) {
		return &gen.DeleteLogPolicyConflict{Code: 409, Message: "compliance lock 策略禁止删除"}, nil
	}

	if err := h.policyRepo.Delete(ctx, logger.TenantFromContext(ctx), params.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.DeleteLogPolicyNotFound{Code: 404, Message: "日志策略不存在"}, nil
		}
		h.logger.Error("delete log policy failed", zap.Error(err))
		return &gen.DeleteLogPolicyInternalServerError{Code: 500, Message: "删除日志策略失败"}, nil
	}

	if err := h.refreshPolicyEngine(ctx); err != nil {
		return &gen.DeleteLogPolicyInternalServerError{Code: 500, Message: "策略刷新失败"}, nil
	}
	h.recordPolicyAudit(ctx, "observability.policy.delete", existing)
	return ok(), nil
}

func (h *logPolicyAPIHandler) PreviewLogPolicy(ctx context.Context, req *gen.LogPolicyPreviewRequest) (gen.PreviewLogPolicyRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.PreviewLogPolicyUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.policyEngine == nil {
		return &gen.PreviewLogPolicyInternalServerError{Code: 500, Message: "日志策略服务未就绪"}, nil
	}
	if req == nil {
		return &gen.PreviewLogPolicyBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	pipeline := string(req.Pipeline)
	if !isValidPipeline(pipeline) {
		return &gen.PreviewLogPolicyBadRequest{Code: 400, Message: "pipeline 非法"}, nil
	}

	decision := h.policyEngine.Decide(pipeline, req.Fields)
	resp := &gen.LogPolicyPreviewResponse{
		Decision: toGenPreviewDecision(decision.Decision),
		Matched:  decision.Matched != nil,
	}
	if decision.Decision == logpolicy.DecisionSample {
		resp.SampleRate = gen.NewOptNilInt(decision.SampleRate)
	}
	if decision.Matched != nil {
		resp.Policy = gen.NewOptLogPolicyItem(toGenLogPolicyItem(decision.Matched))
	}
	return resp, nil
}

func (h *logPolicyAPIHandler) refreshPolicyEngine(ctx context.Context) error {
	if h.policyEngine == nil {
		return nil
	}
	if err := h.policyEngine.Refresh(ctx); err != nil {
		h.logger.Warn("log policy refresh failed", zap.Error(err))
		return err
	}
	return nil
}

func (h *logPolicyAPIHandler) recordPolicyAudit(ctx context.Context, action string, policy *logpolicy.LogPolicy) {
	if policy == nil {
		return
	}
	h.audit.Record(ctx, audit.Event{
		Action:       action,
		ResourceType: "log_policy",
		ResourceID:   policy.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata: map[string]any{
			"pipeline":    policy.Pipeline,
			"match_field": policy.MatchField,
			"pattern":     policy.Pattern,
			"decision":    policy.Decision,
			"enabled":     policy.Enabled,
			"priority":    policy.Priority,
		},
	})
}

func toGenLogPolicyItem(policy *logpolicy.LogPolicy) gen.LogPolicyItem {
	item := gen.LogPolicyItem{
		ID:               policy.ID,
		TenantID:         policy.TenantID,
		Pipeline:         gen.LogPolicyItemPipeline(policy.Pipeline),
		MatchField:       gen.LogPolicyItemMatchField(policy.MatchField),
		Pattern:          policy.Pattern,
		Decision:         gen.LogPolicyItemDecision(policy.Decision),
		Priority:         policy.Priority,
		Enabled:          policy.Enabled,
		ComplianceLocked: isComplianceLockedPolicy(policy),
		CreatedAt:        policy.CreatedAt,
		UpdatedAt:        policy.UpdatedAt,
	}
	if policy.SampleRate != nil {
		item.SampleRate = gen.NewOptNilInt(*policy.SampleRate)
	}
	if strings.TrimSpace(policy.Note) != "" {
		item.Note = gen.NewOptString(policy.Note)
	}
	if policy.CreatedBy != nil {
		item.CreatedBy = gen.NewOptNilUUID(*policy.CreatedBy)
	}
	return item
}

func toGenPreviewDecision(decision string) gen.LogPolicyPreviewResponseDecision {
	switch decision {
	case logpolicy.DecisionDeny:
		return gen.LogPolicyPreviewResponseDecisionDeny
	case logpolicy.DecisionSample:
		return gen.LogPolicyPreviewResponseDecisionSample
	default:
		return gen.LogPolicyPreviewResponseDecisionAllow
	}
}

func validatePolicyInput(pipeline, matchField, decision string, sampleRate *int) error {
	if !isValidPipeline(pipeline) {
		return errors.New("pipeline 非法")
	}
	if !isValidMatchField(pipeline, matchField) {
		return errors.New("match_field 与 pipeline 不匹配")
	}
	switch decision {
	case logpolicy.DecisionAllow, logpolicy.DecisionDeny:
		if sampleRate != nil {
			return errors.New("allow/deny 策略不允许 sample_rate")
		}
	case logpolicy.DecisionSample:
		if sampleRate == nil {
			return errors.New("sample 策略必须提供 sample_rate")
		}
		if *sampleRate < 1 || *sampleRate > 100 {
			return errors.New("sample_rate 必须在 1-100")
		}
	default:
		return errors.New("decision 非法")
	}
	return nil
}

func isValidPipeline(pipeline string) bool {
	switch pipeline {
	case logpolicy.PipelineAudit, logpolicy.PipelineTelemetry:
		return true
	default:
		return false
	}
}

func isValidMatchField(pipeline, matchField string) bool {
	switch pipeline {
	case logpolicy.PipelineAudit:
		switch matchField {
		case logpolicy.MatchFieldAction, logpolicy.MatchFieldOutcome, logpolicy.MatchFieldResourceType:
			return true
		}
	case logpolicy.PipelineTelemetry:
		switch matchField {
		case logpolicy.MatchFieldLevel, logpolicy.MatchFieldEvent, logpolicy.MatchFieldRoute:
			return true
		}
	}
	return false
}

func isComplianceLockedPolicy(policy *logpolicy.LogPolicy) bool {
	if policy == nil {
		return false
	}
	if policy.Pipeline != logpolicy.PipelineAudit || policy.MatchField != logpolicy.MatchFieldAction {
		return false
	}
	return isComplianceLockedPattern(policy.Pattern)
}

func isComplianceLockedPattern(pattern string) bool {
	target := strings.TrimSpace(pattern)
	for _, locked := range logpolicy.ComplianceLockedPatterns {
		if target == locked {
			return true
		}
		if strings.HasSuffix(locked, "*") {
			if strings.HasPrefix(target, strings.TrimSuffix(locked, "*")) {
				return true
			}
		}
		if strings.HasSuffix(target, "*") {
			if strings.HasPrefix(locked, strings.TrimSuffix(target, "*")) {
				return true
			}
		}
	}
	return false
}

func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "duplicate key") || strings.Contains(lower, "unique constraint")
}

