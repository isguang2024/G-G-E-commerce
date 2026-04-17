// storage_admin.go — ogen handlers for the upload configuration center
// (storage providers / buckets / upload keys / upload key rules).
//
// All operations live under the single permission key `system.upload.config.manage`
// declared in api/openapi/domains/storage_admin/paths.yaml. The handler layer is
// intentionally thin: it converts ogen request DTOs to upload.AdminAPI Save inputs,
// invokes the service, and maps GORM models back to ogen Summary/Detail responses.
//
// Secrets returned from the service are already masked (see service.MaskSecret); we
// simply project the access_key_encrypted column into the access_key_masked field.
package handlers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/upload"
	pkgLogger "github.com/maben/backend/internal/pkg/logger"
)

// ---------- providers ----------

func (h *storageAdminAPIHandler) ListStorageProviders(ctx context.Context, params gen.ListStorageProvidersParams) (gen.ListStorageProvidersRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	items, err := h.uploadSvc.Admin().ListProviders(ctx, pkgLogger.TenantFromContext(ctx))
	if err != nil {
		h.logger.Error("list storage providers failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	if statusFilter, ok := params.Status.Get(); ok {
		items = filterByStatus(items, string(statusFilter))
	}
	records := make([]gen.StorageProviderSummary, 0, len(items))
	for i := range items {
		records = append(records, mapStorageProviderSummary(&items[i]))
	}
	return &gen.StorageProviderListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

func (h *storageAdminAPIHandler) GetStorageProvider(ctx context.Context, params gen.GetStorageProviderParams) (gen.GetStorageProviderRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	item, err := h.uploadSvc.Admin().GetProvider(ctx, pkgLogger.TenantFromContext(ctx), params.ID)
	if err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "Provider 不存在"), nil
		}
		h.logger.Error("get storage provider failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	resp := mapStorageProviderSummary(item)
	return &resp, nil
}

func (h *storageAdminAPIHandler) CreateStorageProvider(ctx context.Context, req *gen.StorageProviderSaveRequest) (gen.CreateStorageProviderRes, error) {
	if h.uploadSvc == nil {
		return &gen.CreateStorageProviderBadRequest{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.CreateStorageProviderBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	input := storageProviderInputFromRequest(req)
	created, err := h.uploadSvc.Admin().CreateProvider(ctx, pkgLogger.TenantFromContext(ctx), input)
	if err != nil {
		h.logger.Error("create storage provider failed", zap.Error(err))
		return &gen.CreateStorageProviderBadRequest{Code: 400, Message: err.Error()}, nil
	}
	resp := mapStorageProviderSummary(created)
	return &resp, nil
}

func (h *storageAdminAPIHandler) UpdateStorageProvider(ctx context.Context, req *gen.StorageProviderSaveRequest, params gen.UpdateStorageProviderParams) (gen.UpdateStorageProviderRes, error) {
	if h.uploadSvc == nil {
		return &gen.UpdateStorageProviderBadRequest{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.UpdateStorageProviderBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	input := storageProviderInputFromRequest(req)
	updated, err := h.uploadSvc.Admin().UpdateProvider(ctx, pkgLogger.TenantFromContext(ctx), params.ID, input)
	if err != nil {
		if isRecordNotFound(err) {
			return &gen.UpdateStorageProviderNotFound{Code: 404, Message: "Provider 不存在"}, nil
		}
		h.logger.Error("update storage provider failed", zap.Error(err))
		return &gen.UpdateStorageProviderBadRequest{Code: 400, Message: err.Error()}, nil
	}
	resp := mapStorageProviderSummary(updated)
	return &resp, nil
}

func (h *storageAdminAPIHandler) DeleteStorageProvider(ctx context.Context, params gen.DeleteStorageProviderParams) (gen.DeleteStorageProviderRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	if err := h.uploadSvc.Admin().DeleteProvider(ctx, pkgLogger.TenantFromContext(ctx), params.ID); err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "Provider 不存在"), nil
		}
		h.logger.Error("delete storage provider failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	return ok(), nil
}

func (h *storageAdminAPIHandler) TestStorageProvider(ctx context.Context, params gen.TestStorageProviderParams) (gen.TestStorageProviderRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	result, err := h.uploadSvc.Admin().TestProvider(ctx, pkgLogger.TenantFromContext(ctx), params.ID)
	if err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "Provider 不存在"), nil
		}
		h.logger.Error("test storage provider failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	resp := &gen.StorageProviderTestResponse{Ok: result.OK}
	if result.Message != "" {
		resp.Message = gen.NewOptString(result.Message)
	}
	if result.LatencyMs > 0 {
		resp.LatencyMs = gen.NewOptInt(int(result.LatencyMs))
	}
	return resp, nil
}

// ---------- buckets ----------

func (h *storageAdminAPIHandler) ListStorageBuckets(ctx context.Context, params gen.ListStorageBucketsParams) (gen.ListStorageBucketsRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	items, err := h.uploadSvc.Admin().ListBuckets(ctx, pkgLogger.TenantFromContext(ctx), optUUIDPtr(params.ProviderID))
	if err != nil {
		h.logger.Error("list storage buckets failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	if statusFilter, ok := params.Status.Get(); ok {
		items = filterByStatus(items, string(statusFilter))
	}
	records := make([]gen.StorageBucketSummary, 0, len(items))
	for i := range items {
		records = append(records, mapStorageBucketSummary(&items[i]))
	}
	return &gen.StorageBucketListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

func (h *storageAdminAPIHandler) GetStorageBucket(ctx context.Context, params gen.GetStorageBucketParams) (gen.GetStorageBucketRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	item, err := h.uploadSvc.Admin().GetBucket(ctx, pkgLogger.TenantFromContext(ctx), params.ID)
	if err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "Bucket 不存在"), nil
		}
		h.logger.Error("get storage bucket failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	resp := mapStorageBucketSummary(item)
	return &resp, nil
}

func (h *storageAdminAPIHandler) CreateStorageBucket(ctx context.Context, req *gen.StorageBucketSaveRequest) (gen.CreateStorageBucketRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	if req == nil {
		return errResp(400, "请求体不能为空"), nil
	}
	input := storageBucketInputFromRequest(req)
	created, err := h.uploadSvc.Admin().CreateBucket(ctx, pkgLogger.TenantFromContext(ctx), input)
	if err != nil {
		h.logger.Error("create storage bucket failed", zap.Error(err))
		return errResp(400, err.Error()), nil
	}
	resp := mapStorageBucketSummary(created)
	return &resp, nil
}

func (h *storageAdminAPIHandler) UpdateStorageBucket(ctx context.Context, req *gen.StorageBucketSaveRequest, params gen.UpdateStorageBucketParams) (gen.UpdateStorageBucketRes, error) {
	if h.uploadSvc == nil {
		return &gen.UpdateStorageBucketBadRequest{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.UpdateStorageBucketBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	input := storageBucketInputFromRequest(req)
	updated, err := h.uploadSvc.Admin().UpdateBucket(ctx, pkgLogger.TenantFromContext(ctx), params.ID, input)
	if err != nil {
		if isRecordNotFound(err) {
			return &gen.UpdateStorageBucketNotFound{Code: 404, Message: "Bucket 不存在"}, nil
		}
		h.logger.Error("update storage bucket failed", zap.Error(err))
		return &gen.UpdateStorageBucketBadRequest{Code: 400, Message: err.Error()}, nil
	}
	resp := mapStorageBucketSummary(updated)
	return &resp, nil
}

func (h *storageAdminAPIHandler) DeleteStorageBucket(ctx context.Context, params gen.DeleteStorageBucketParams) (gen.DeleteStorageBucketRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	if err := h.uploadSvc.Admin().DeleteBucket(ctx, pkgLogger.TenantFromContext(ctx), params.ID); err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "Bucket 不存在"), nil
		}
		h.logger.Error("delete storage bucket failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	return ok(), nil
}

// ---------- upload keys ----------

func (h *storageAdminAPIHandler) ListUploadKeys(ctx context.Context, params gen.ListUploadKeysParams) (gen.ListUploadKeysRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	items, err := h.uploadSvc.Admin().ListUploadKeys(ctx, pkgLogger.TenantFromContext(ctx), optUUIDPtr(params.BucketID))
	if err != nil {
		h.logger.Error("list upload keys failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	if statusFilter, ok := params.Status.Get(); ok {
		items = filterByStatus(items, string(statusFilter))
	}
	records := make([]gen.UploadKeySummary, 0, len(items))
	for i := range items {
		records = append(records, mapUploadKeySummary(&items[i]))
	}
	return &gen.UploadKeyListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

func (h *storageAdminAPIHandler) GetUploadKey(ctx context.Context, params gen.GetUploadKeyParams) (gen.GetUploadKeyRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	item, rules, err := h.uploadSvc.Admin().GetUploadKey(ctx, pkgLogger.TenantFromContext(ctx), params.ID)
	if err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "UploadKey 不存在"), nil
		}
		h.logger.Error("get upload key failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	detail := mapUploadKeyDetail(item, rules)
	return &detail, nil
}

func (h *storageAdminAPIHandler) CreateUploadKey(ctx context.Context, req *gen.UploadKeySaveRequest) (gen.CreateUploadKeyRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	if req == nil {
		return errResp(400, "请求体不能为空"), nil
	}
	input := uploadKeyInputFromRequest(req)
	created, err := h.uploadSvc.Admin().CreateUploadKey(ctx, pkgLogger.TenantFromContext(ctx), input)
	if err != nil {
		h.logger.Error("create upload key failed", zap.Error(err))
		return errResp(400, err.Error()), nil
	}
	resp := mapUploadKeySummary(created)
	return &resp, nil
}

func (h *storageAdminAPIHandler) UpdateUploadKey(ctx context.Context, req *gen.UploadKeySaveRequest, params gen.UpdateUploadKeyParams) (gen.UpdateUploadKeyRes, error) {
	if h.uploadSvc == nil {
		return &gen.UpdateUploadKeyBadRequest{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.UpdateUploadKeyBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	input := uploadKeyInputFromRequest(req)
	updated, err := h.uploadSvc.Admin().UpdateUploadKey(ctx, pkgLogger.TenantFromContext(ctx), params.ID, input)
	if err != nil {
		if isRecordNotFound(err) {
			return &gen.UpdateUploadKeyNotFound{Code: 404, Message: "UploadKey 不存在"}, nil
		}
		h.logger.Error("update upload key failed", zap.Error(err))
		return &gen.UpdateUploadKeyBadRequest{Code: 400, Message: err.Error()}, nil
	}
	resp := mapUploadKeySummary(updated)
	return &resp, nil
}

func (h *storageAdminAPIHandler) DeleteUploadKey(ctx context.Context, params gen.DeleteUploadKeyParams) (gen.DeleteUploadKeyRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	if err := h.uploadSvc.Admin().DeleteUploadKey(ctx, pkgLogger.TenantFromContext(ctx), params.ID); err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "UploadKey 不存在"), nil
		}
		h.logger.Error("delete upload key failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	return ok(), nil
}

// ---------- upload key rules ----------

func (h *storageAdminAPIHandler) ListUploadKeyRules(ctx context.Context, params gen.ListUploadKeyRulesParams) (gen.ListUploadKeyRulesRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	rules, err := h.uploadSvc.Admin().ListRulesByUploadKey(ctx, pkgLogger.TenantFromContext(ctx), params.ID)
	if err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "UploadKey 不存在"), nil
		}
		h.logger.Error("list upload key rules failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	records := make([]gen.UploadKeyRuleSummary, 0, len(rules))
	for i := range rules {
		records = append(records, mapUploadKeyRuleSummary(&rules[i]))
	}
	return &gen.UploadKeyRuleListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

func (h *storageAdminAPIHandler) CreateUploadKeyRule(ctx context.Context, req *gen.UploadKeyRuleSaveRequest, params gen.CreateUploadKeyRuleParams) (gen.CreateUploadKeyRuleRes, error) {
	if h.uploadSvc == nil {
		return &gen.CreateUploadKeyRuleBadRequest{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.CreateUploadKeyRuleBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	input := uploadRuleInputFromRequest(req)
	created, err := h.uploadSvc.Admin().CreateRule(ctx, pkgLogger.TenantFromContext(ctx), params.ID, input)
	if err != nil {
		if isRecordNotFound(err) {
			return &gen.CreateUploadKeyRuleNotFound{Code: 404, Message: "UploadKey 不存在"}, nil
		}
		h.logger.Error("create upload key rule failed", zap.Error(err))
		return &gen.CreateUploadKeyRuleBadRequest{Code: 400, Message: err.Error()}, nil
	}
	resp := mapUploadKeyRuleSummary(created)
	return &resp, nil
}

func (h *storageAdminAPIHandler) UpdateUploadKeyRule(ctx context.Context, req *gen.UploadKeyRuleSaveRequest, params gen.UpdateUploadKeyRuleParams) (gen.UpdateUploadKeyRuleRes, error) {
	if h.uploadSvc == nil {
		return &gen.UpdateUploadKeyRuleBadRequest{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.UpdateUploadKeyRuleBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	input := uploadRuleInputFromRequest(req)
	updated, err := h.uploadSvc.Admin().UpdateRule(ctx, pkgLogger.TenantFromContext(ctx), params.RuleId, input)
	if err != nil {
		if isRecordNotFound(err) {
			return &gen.UpdateUploadKeyRuleNotFound{Code: 404, Message: "Rule 不存在"}, nil
		}
		h.logger.Error("update upload key rule failed", zap.Error(err))
		return &gen.UpdateUploadKeyRuleBadRequest{Code: 400, Message: err.Error()}, nil
	}
	resp := mapUploadKeyRuleSummary(updated)
	return &resp, nil
}

func (h *storageAdminAPIHandler) DeleteUploadKeyRule(ctx context.Context, params gen.DeleteUploadKeyRuleParams) (gen.DeleteUploadKeyRuleRes, error) {
	if h.uploadSvc == nil {
		return errResp(500, "上传服务未就绪"), nil
	}
	if err := h.uploadSvc.Admin().DeleteRule(ctx, pkgLogger.TenantFromContext(ctx), params.RuleId); err != nil {
		if isRecordNotFound(err) {
			return errResp(404, "Rule 不存在"), nil
		}
		h.logger.Error("delete upload key rule failed", zap.Error(err))
		return errResp(500, err.Error()), nil
	}
	return ok(), nil
}

// ---------- request → service input ----------

func storageProviderInputFromRequest(req *gen.StorageProviderSaveRequest) upload.ProviderSaveInput {
	input := upload.ProviderSaveInput{
		ProviderKey: req.ProviderKey,
		Name:        req.Name,
		Driver:      string(req.Driver),
		Endpoint:    optString(req.Endpoint),
		Region:      optString(req.Region),
		BaseURL:     optString(req.BaseURL),
		AccessKey:   optString(req.AccessKey),
		SecretKey:   optString(req.SecretKey),
		IsDefault:   optBool(req.IsDefault),
	}
	if req.Status.Set {
		input.Status = string(req.Status.Value)
	}
	if req.Extra.Set {
		input.Extra = jxRawMapToMeta(map[string]jx.Raw(req.Extra.Value))
	}
	return input
}

func storageBucketInputFromRequest(req *gen.StorageBucketSaveRequest) upload.BucketSaveInput {
	input := upload.BucketSaveInput{
		ProviderID:    req.ProviderID,
		BucketKey:     req.BucketKey,
		Name:          req.Name,
		BucketName:    req.BucketName,
		BasePath:      optString(req.BasePath),
		PublicBaseURL: optString(req.PublicBaseURL),
		IsPublic:      optBool(req.IsPublic),
	}
	if req.Status.Set {
		input.Status = string(req.Status.Value)
	}
	if req.Extra.Set {
		input.Extra = jxRawMapToMeta(map[string]jx.Raw(req.Extra.Value))
	}
	return input
}

func uploadKeyInputFromRequest(req *gen.UploadKeySaveRequest) upload.UploadKeySaveInput {
	input := upload.UploadKeySaveInput{
		BucketID:                 req.BucketID,
		Key:                      req.Key,
		Name:                     req.Name,
		PathTemplate:             optString(req.PathTemplate),
		DefaultRuleKey:           optString(req.DefaultRuleKey),
		MaxSizeBytes:             optInt64(req.MaxSizeBytes),
		AllowedMimeTypes:         models.StringList(req.AllowedMimeTypes),
		PermissionKey:            optString(req.PermissionKey),
		FallbackKey:              optString(req.FallbackKey),
		ClientAccept:             models.StringList(req.ClientAccept),
		DirectSizeThresholdBytes: optInt64(req.DirectSizeThresholdBytes),
	}
	if req.UploadMode.Set {
		input.UploadMode = string(req.UploadMode.Value)
	}
	if req.IsFrontendVisible.Set {
		input.IsFrontendVisible = req.IsFrontendVisible.Value
	}
	if req.Visibility.Set {
		input.Visibility = string(req.Visibility.Value)
	}
	if req.Status.Set {
		input.Status = string(req.Status.Value)
	}
	if req.ExtraSchema.Set {
		input.ExtraSchema = jxRawMapToMeta(map[string]jx.Raw(req.ExtraSchema.Value))
	}
	if req.Meta.Set {
		input.Meta = jxRawMapToMeta(map[string]jx.Raw(req.Meta.Value))
	}
	return input
}

func uploadRuleInputFromRequest(req *gen.UploadKeyRuleSaveRequest) upload.UploadRuleSaveInput {
	input := upload.UploadRuleSaveInput{
		RuleKey:          req.RuleKey,
		Name:             req.Name,
		SubPath:          optString(req.SubPath),
		MaxSizeBytes:     optInt64(req.MaxSizeBytes),
		AllowedMimeTypes: models.StringList(req.AllowedMimeTypes),
		ProcessPipeline:  models.StringList(req.ProcessPipeline),
		ClientAccept:     models.StringList(req.ClientAccept),
		IsDefault:        optBool(req.IsDefault),
	}
	if req.FilenameStrategy.Set {
		input.FilenameStrategy = string(req.FilenameStrategy.Value)
	}
	if req.ModeOverride.Set {
		input.ModeOverride = string(req.ModeOverride.Value)
	}
	if req.VisibilityOverride.Set {
		input.VisibilityOverride = string(req.VisibilityOverride.Value)
	}
	if req.Status.Set {
		input.Status = string(req.Status.Value)
	}
	if req.ExtraSchema.Set {
		input.ExtraSchema = jxRawMapToMeta(map[string]jx.Raw(req.ExtraSchema.Value))
	}
	if req.Meta.Set {
		input.Meta = jxRawMapToMeta(map[string]jx.Raw(req.Meta.Value))
	}
	return input
}

// ---------- model → ogen DTO ----------

func mapStorageProviderSummary(item *models.StorageProvider) gen.StorageProviderSummary {
	out := gen.StorageProviderSummary{
		ID:          item.ID,
		ProviderKey: item.ProviderKey,
		Name:        item.Name,
		Driver:      gen.StorageProviderSummaryDriver(coalesceDriver(item.Driver)),
		IsDefault:   item.IsDefault,
		Status:      gen.StorageProviderSummaryStatus(coalesceProviderStatus(item.Status)),
	}
	if item.Endpoint != "" {
		out.Endpoint = gen.NewOptString(item.Endpoint)
	}
	if item.Region != "" {
		out.Region = gen.NewOptString(item.Region)
	}
	if item.BaseURL != "" {
		out.BaseURL = gen.NewOptString(item.BaseURL)
	}
	if item.AccessKeyEncrypted != "" {
		out.AccessKeyMasked = gen.NewOptString(item.AccessKeyEncrypted)
	}
	if item.SecretKeyEncrypted != "" {
		out.SecretKeyMasked = gen.NewOptString(item.SecretKeyEncrypted)
	}
	if extra := metaToProviderSummaryExtra(item.Extra); extra != nil {
		out.Extra = gen.NewOptStorageProviderSummaryExtra(extra)
	}
	if !item.CreatedAt.IsZero() {
		out.CreatedAt = gen.NewOptDateTime(item.CreatedAt)
	}
	if !item.UpdatedAt.IsZero() {
		out.UpdatedAt = gen.NewOptDateTime(item.UpdatedAt)
	}
	return out
}

func mapStorageBucketSummary(item *models.StorageBucket) gen.StorageBucketSummary {
	out := gen.StorageBucketSummary{
		ID:         item.ID,
		ProviderID: item.ProviderID,
		BucketKey:  item.BucketKey,
		Name:       item.Name,
		BucketName: item.BucketName,
		IsPublic:   item.IsPublic,
		Status:     gen.StorageBucketSummaryStatus(coalesceBucketStatus(item.Status)),
	}
	if item.BasePath != "" {
		out.BasePath = gen.NewOptString(item.BasePath)
	}
	if item.PublicBaseURL != "" {
		out.PublicBaseURL = gen.NewOptString(item.PublicBaseURL)
	}
	if extra := metaToBucketSummaryExtra(item.Extra); extra != nil {
		out.Extra = gen.NewOptStorageBucketSummaryExtra(extra)
	}
	if !item.CreatedAt.IsZero() {
		out.CreatedAt = gen.NewOptDateTime(item.CreatedAt)
	}
	if !item.UpdatedAt.IsZero() {
		out.UpdatedAt = gen.NewOptDateTime(item.UpdatedAt)
	}
	return out
}

func mapUploadKeySummary(item *models.UploadKey) gen.UploadKeySummary {
	allowed := []string(item.AllowedMimeTypes)
	if allowed == nil {
		allowed = []string{}
	}
	out := gen.UploadKeySummary{
		ID:                item.ID,
		BucketID:          item.BucketID,
		Key:               item.Key,
		Name:              item.Name,
		AllowedMimeTypes:  allowed,
		UploadMode:        gen.NewOptUploadKeySummaryUploadMode(gen.UploadKeySummaryUploadMode(coalesceUploadMode(item.UploadMode))),
		IsFrontendVisible: gen.OptBool{Value: item.IsFrontendVisible, Set: true},
		ClientAccept:      []string(item.ClientAccept),
		Visibility:        gen.UploadKeySummaryVisibility(coalesceVisibility(item.Visibility)),
		Status:            gen.UploadKeySummaryStatus(coalesceUploadKeyStatus(item.Status)),
	}
	if item.PathTemplate != "" {
		out.PathTemplate = gen.NewOptString(item.PathTemplate)
	}
	if item.DefaultRuleKey != "" {
		out.DefaultRuleKey = gen.NewOptString(item.DefaultRuleKey)
	}
	if item.MaxSizeBytes > 0 {
		out.MaxSizeBytes = gen.NewOptInt64(item.MaxSizeBytes)
	}
	if item.PermissionKey != "" {
		out.PermissionKey = gen.NewOptString(item.PermissionKey)
	}
	if item.FallbackKey != "" {
		out.FallbackKey = gen.NewOptString(item.FallbackKey)
	}
	if item.DirectSizeThresholdBytes > 0 {
		out.DirectSizeThresholdBytes = gen.NewOptInt64(item.DirectSizeThresholdBytes)
	}
	if extraSchema := metaToUploadKeySummaryExtraSchema(item.ExtraSchema); extraSchema != nil {
		out.ExtraSchema = gen.NewOptUploadKeySummaryExtraSchema(extraSchema)
	}
	if meta := metaToUploadKeySummaryMeta(item.Meta); meta != nil {
		out.Meta = gen.NewOptUploadKeySummaryMeta(meta)
	}
	if !item.CreatedAt.IsZero() {
		out.CreatedAt = gen.NewOptDateTime(item.CreatedAt)
	}
	if !item.UpdatedAt.IsZero() {
		out.UpdatedAt = gen.NewOptDateTime(item.UpdatedAt)
	}
	return out
}

func mapUploadKeyDetail(item *models.UploadKey, rules []models.UploadKeyRule) gen.UploadKeyDetail {
	allowed := []string(item.AllowedMimeTypes)
	if allowed == nil {
		allowed = []string{}
	}
	out := gen.UploadKeyDetail{
		ID:                item.ID,
		BucketID:          item.BucketID,
		Key:               item.Key,
		Name:              item.Name,
		AllowedMimeTypes:  allowed,
		UploadMode:        gen.NewOptUploadKeyDetailUploadMode(gen.UploadKeyDetailUploadMode(coalesceUploadMode(item.UploadMode))),
		IsFrontendVisible: gen.OptBool{Value: item.IsFrontendVisible, Set: true},
		ClientAccept:      []string(item.ClientAccept),
		Visibility:        gen.UploadKeyDetailVisibility(coalesceVisibility(item.Visibility)),
		Status:            gen.UploadKeyDetailStatus(coalesceUploadKeyStatus(item.Status)),
		Rules:             make([]gen.UploadKeyRuleSummary, 0, len(rules)),
	}
	if item.PathTemplate != "" {
		out.PathTemplate = gen.NewOptString(item.PathTemplate)
	}
	if item.DefaultRuleKey != "" {
		out.DefaultRuleKey = gen.NewOptString(item.DefaultRuleKey)
	}
	if item.MaxSizeBytes > 0 {
		out.MaxSizeBytes = gen.NewOptInt64(item.MaxSizeBytes)
	}
	if item.PermissionKey != "" {
		out.PermissionKey = gen.NewOptString(item.PermissionKey)
	}
	if item.FallbackKey != "" {
		out.FallbackKey = gen.NewOptString(item.FallbackKey)
	}
	if item.DirectSizeThresholdBytes > 0 {
		out.DirectSizeThresholdBytes = gen.NewOptInt64(item.DirectSizeThresholdBytes)
	}
	if extraSchema := metaToUploadKeyDetailExtraSchema(item.ExtraSchema); extraSchema != nil {
		out.ExtraSchema = gen.NewOptUploadKeyDetailExtraSchema(extraSchema)
	}
	if meta := metaToUploadKeyDetailMeta(item.Meta); meta != nil {
		out.Meta = gen.NewOptUploadKeyDetailMeta(meta)
	}
	if !item.CreatedAt.IsZero() {
		out.CreatedAt = gen.NewOptDateTime(item.CreatedAt)
	}
	if !item.UpdatedAt.IsZero() {
		out.UpdatedAt = gen.NewOptDateTime(item.UpdatedAt)
	}
	for i := range rules {
		out.Rules = append(out.Rules, mapUploadKeyRuleSummary(&rules[i]))
	}
	return out
}

func mapUploadKeyRuleSummary(item *models.UploadKeyRule) gen.UploadKeyRuleSummary {
	allowed := []string(item.AllowedMimeTypes)
	if allowed == nil {
		allowed = []string{}
	}
	pipeline := []string(item.ProcessPipeline)
	if pipeline == nil {
		pipeline = []string{}
	}
	out := gen.UploadKeyRuleSummary{
		ID:               item.ID,
		UploadKeyID:      item.UploadKeyID,
		RuleKey:          item.RuleKey,
		Name:             item.Name,
		FilenameStrategy: gen.UploadKeyRuleSummaryFilenameStrategy(coalesceFilenameStrategy(item.FilenameStrategy)),
		AllowedMimeTypes: allowed,
		ProcessPipeline:  pipeline,
		ClientAccept:     []string(item.ClientAccept),
		IsDefault:        item.IsDefault,
		Status:           gen.UploadKeyRuleSummaryStatus(coalesceUploadKeyStatus(item.Status)),
	}
	if item.SubPath != "" {
		out.SubPath = gen.NewOptString(item.SubPath)
	}
	if item.MaxSizeBytes > 0 {
		out.MaxSizeBytes = gen.NewOptInt64(item.MaxSizeBytes)
	}
	if item.ModeOverride != "" {
		out.ModeOverride = gen.NewOptUploadKeyRuleSummaryModeOverride(gen.UploadKeyRuleSummaryModeOverride(coalesceRuleModeOverride(item.ModeOverride)))
	}
	if item.VisibilityOverride != "" {
		out.VisibilityOverride = gen.NewOptUploadKeyRuleSummaryVisibilityOverride(gen.UploadKeyRuleSummaryVisibilityOverride(coalesceRuleVisibilityOverride(item.VisibilityOverride)))
	}
	if extraSchema := metaToUploadRuleSummaryExtraSchema(item.ExtraSchema); extraSchema != nil {
		out.ExtraSchema = gen.NewOptUploadKeyRuleSummaryExtraSchema(extraSchema)
	}
	if meta := metaToUploadRuleSummaryMeta(item.Meta); meta != nil {
		out.Meta = gen.NewOptUploadKeyRuleSummaryMeta(meta)
	}
	if !item.CreatedAt.IsZero() {
		out.CreatedAt = gen.NewOptDateTime(item.CreatedAt)
	}
	if !item.UpdatedAt.IsZero() {
		out.UpdatedAt = gen.NewOptDateTime(item.UpdatedAt)
	}
	return out
}

// ---------- helpers ----------

// errResp returns the shared *gen.Error2 envelope used by every storage_admin
// operation whose error responses reference the common errors.yaml schema.
// The same pointer satisfies all of those operations' Res interfaces because
// Error2 implements every relevant marker method (see oas_schemas_gen.go).
func errResp(code int, msg string) *gen.Error2 {
	return &gen.Error2{Code: code, Message: msg}
}

func isRecordNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	if errors.Is(err, upload.ErrRecordNotFound) {
		return true
	}
	return false
}

func optInt64(o gen.OptInt64) int64 {
	if !o.Set {
		return 0
	}
	return o.Value
}

func optUUIDPtr(o gen.OptUUID) *uuid.UUID {
	if v, ok := o.Get(); ok {
		return &v
	}
	return nil
}

// jxRawMapToMeta converts the ogen jx.Raw map (used for arbitrary JSON object
// fields like extra/meta) into the GORM-friendly MetaJSON map[string]any.
func jxRawMapToMeta(src map[string]jx.Raw) models.MetaJSON {
	if len(src) == 0 {
		return nil
	}
	out := make(models.MetaJSON, len(src))
	for k, raw := range src {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			zap.L().Warn("jxRawMapToMeta: unmarshal failed, field dropped",
				zap.String("key", k), zap.Error(err))
			continue
		}
		out[k] = v
	}
	return out
}

func metaToJxRaw(meta models.MetaJSON) map[string]jx.Raw {
	if len(meta) == 0 {
		return nil
	}
	out := make(map[string]jx.Raw, len(meta))
	for k, v := range meta {
		raw, err := json.Marshal(v)
		if err != nil {
			continue
		}
		out[k] = jx.Raw(raw)
	}
	return out
}

func metaToProviderSummaryExtra(meta models.MetaJSON) gen.StorageProviderSummaryExtra {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.StorageProviderSummaryExtra(raw)
}

func metaToBucketSummaryExtra(meta models.MetaJSON) gen.StorageBucketSummaryExtra {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.StorageBucketSummaryExtra(raw)
}

func metaToUploadKeySummaryMeta(meta models.MetaJSON) gen.UploadKeySummaryMeta {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.UploadKeySummaryMeta(raw)
}

func metaToUploadKeySummaryExtraSchema(meta models.MetaJSON) gen.UploadKeySummaryExtraSchema {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.UploadKeySummaryExtraSchema(raw)
}

func metaToUploadKeyDetailMeta(meta models.MetaJSON) gen.UploadKeyDetailMeta {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.UploadKeyDetailMeta(raw)
}

func metaToUploadKeyDetailExtraSchema(meta models.MetaJSON) gen.UploadKeyDetailExtraSchema {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.UploadKeyDetailExtraSchema(raw)
}

func metaToUploadRuleSummaryMeta(meta models.MetaJSON) gen.UploadKeyRuleSummaryMeta {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.UploadKeyRuleSummaryMeta(raw)
}

func metaToUploadRuleSummaryExtraSchema(meta models.MetaJSON) gen.UploadKeyRuleSummaryExtraSchema {
	raw := metaToJxRaw(meta)
	if raw == nil {
		return nil
	}
	return gen.UploadKeyRuleSummaryExtraSchema(raw)
}

func coalesceDriver(v string) string {
	if v == string(gen.StorageProviderSummaryDriverAliyunOss) {
		return v
	}
	return string(gen.StorageProviderSummaryDriverLocal)
}

func coalesceProviderStatus(v string) string {
	switch v {
	case string(gen.StorageProviderSummaryStatusReady),
		string(gen.StorageProviderSummaryStatusDisabled),
		string(gen.StorageProviderSummaryStatusError):
		return v
	}
	return string(gen.StorageProviderSummaryStatusReady)
}

func coalesceBucketStatus(v string) string {
	if v == string(gen.StorageBucketSummaryStatusDisabled) {
		return v
	}
	return string(gen.StorageBucketSummaryStatusReady)
}

func coalesceUploadKeyStatus(v string) string {
	if v == string(gen.UploadKeySummaryStatusDisabled) {
		return v
	}
	return string(gen.UploadKeySummaryStatusReady)
}

func coalesceVisibility(v string) string {
	if v == string(gen.UploadKeySummaryVisibilityPrivate) {
		return v
	}
	return string(gen.UploadKeySummaryVisibilityPublic)
}

func coalesceFilenameStrategy(v string) string {
	if v == string(gen.UploadKeyRuleSummaryFilenameStrategyOriginal) {
		return v
	}
	return string(gen.UploadKeyRuleSummaryFilenameStrategyUUID)
}

func coalesceUploadMode(v string) string {
	switch v {
	case string(gen.UploadKeySummaryUploadModeDirect),
		string(gen.UploadKeySummaryUploadModeRelay):
		return v
	default:
		return string(gen.UploadKeySummaryUploadModeAuto)
	}
}

func coalesceRuleModeOverride(v string) string {
	switch v {
	case string(gen.UploadKeyRuleSummaryModeOverrideDirect),
		string(gen.UploadKeyRuleSummaryModeOverrideRelay):
		return v
	default:
		return string(gen.UploadKeyRuleSummaryModeOverrideInherit)
	}
}

func coalesceRuleVisibilityOverride(v string) string {
	switch v {
	case string(gen.UploadKeyRuleSummaryVisibilityOverridePublic),
		string(gen.UploadKeyRuleSummaryVisibilityOverridePrivate):
		return v
	default:
		return string(gen.UploadKeyRuleSummaryVisibilityOverrideInherit)
	}
}

type statusGetter interface{ GetStatus() string }

func filterByStatus[T statusGetter](items []T, status string) []T {
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if item.GetStatus() == status {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

