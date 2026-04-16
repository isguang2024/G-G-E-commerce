package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/upload"
	pkgLogger "github.com/maben/backend/internal/pkg/logger"
)

func (h *APIHandler) UploadMedia(ctx context.Context, req *gen.UploadMediaReq) (gen.UploadMediaRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.UploadMediaUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.uploadSvc == nil {
		return &gen.UploadMediaInternalServerError{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil || !req.File.Set {
		return &gen.UploadMediaInternalServerError{Code: 500, Message: "缺少上传文件"}, nil
	}
	grantedKeys, err := grantedPermissionKeysFromContext(ctx, h.evaluator, userID)
	if err != nil {
		h.logger.Error("resolve upload permissions failed", zap.Error(err))
		return &gen.UploadMediaInternalServerError{Code: 500, Message: "权限解析失败"}, nil
	}

	file := req.File.Value
	record, err := h.uploadSvc.Upload(ctx, pkgLogger.TenantFromContext(ctx), &userID, upload.UploadInput{
		Key:                   optString(req.Key),
		Rule:                  optString(req.Rule),
		Name:                  file.Name,
		File:                  file.File,
		Size:                  file.Size,
		MimeType:              file.Header.Get("Content-Type"),
		GrantedPermissionKeys: grantedKeys,
	})
	if err != nil {
		if errors.Is(err, upload.ErrUploadPermissionDenied) {
			return &gen.UploadMediaUnauthorized{Code: 401, Message: "无权使用该上传配置"}, nil
		}
		h.logger.Error("upload media failed", zap.Error(err))
		return &gen.UploadMediaInternalServerError{Code: 500, Message: err.Error()}, nil
	}

	return &gen.MediaUploadResponse{
		ID:         record.ID,
		Filename:   record.OriginalFilename,
		StorageKey: record.StorageKey,
		URL:        record.URL,
		MimeType:   record.MimeType,
		Size:       record.Size,
		CreatedAt:  record.CreatedAt,
	}, nil
}

func (h *APIHandler) PrepareMediaUpload(ctx context.Context, req *gen.MediaPrepareUploadRequest) (gen.PrepareMediaUploadRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.PrepareMediaUploadUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.uploadSvc == nil {
		return &gen.PrepareMediaUploadInternalServerError{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.PrepareMediaUploadInternalServerError{Code: 500, Message: "缺少上传参数"}, nil
	}
	grantedKeys, err := grantedPermissionKeysFromContext(ctx, h.evaluator, userID)
	if err != nil {
		h.logger.Error("resolve upload permissions failed", zap.Error(err))
		return &gen.PrepareMediaUploadInternalServerError{Code: 500, Message: "权限解析失败"}, nil
	}

	result, err := h.uploadSvc.PrepareUpload(ctx, pkgLogger.TenantFromContext(ctx), upload.PrepareUploadInput{
		Key:                   optString(req.Key),
		Rule:                  optString(req.Rule),
		Name:                  req.Filename,
		Size:                  req.Size,
		MimeType:              optString(req.MimeType),
		Checksum:              optString(req.Checksum),
		GrantedPermissionKeys: grantedKeys,
	})
	if err != nil {
		if errors.Is(err, upload.ErrUploadPermissionDenied) {
			return &gen.PrepareMediaUploadUnauthorized{Code: 401, Message: "无权使用该上传配置"}, nil
		}
		h.logger.Error("prepare media upload failed", zap.Error(err))
		return &gen.PrepareMediaUploadInternalServerError{Code: 500, Message: err.Error()}, nil
	}

	response := &gen.MediaPrepareUploadResponse{
		Mode:         gen.MediaPrepareUploadResponseMode(result.Mode),
		StorageKey:   result.StorageKey,
		Filename:     result.Filename,
		ContentType:  result.ContentType,
		UploadKey:    result.UploadKey,
		FallbackUsed: result.FallbackUsed,
	}
	if value := result.Method; value != "" {
		response.Method = gen.NewOptString(value)
	}
	if value := result.URL; value != "" {
		response.URL = gen.NewOptString(value)
	}
	if value := result.RelayURL; value != "" {
		response.RelayUrl = gen.NewOptString(value)
	}
	if len(result.Headers) > 0 {
		response.Headers = gen.NewOptMediaPrepareUploadResponseHeaders(gen.MediaPrepareUploadResponseHeaders(result.Headers))
	}
	if len(result.Form) > 0 {
		response.Form = gen.NewOptMediaPrepareUploadResponseForm(gen.MediaPrepareUploadResponseForm(result.Form))
	}
	if value := result.RuleKey; value != "" {
		response.RuleKey = gen.NewOptString(value)
	}
	if value := result.Visibility; value != "" {
		response.Visibility = gen.NewOptMediaPrepareUploadResponseVisibility(gen.MediaPrepareUploadResponseVisibility(value))
	}

	return response, nil
}

func (h *APIHandler) CompleteMediaUpload(ctx context.Context, req *gen.MediaCompleteUploadRequest) (gen.CompleteMediaUploadRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.CompleteMediaUploadUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.uploadSvc == nil {
		return &gen.CompleteMediaUploadInternalServerError{Code: 500, Message: "上传服务未就绪"}, nil
	}
	if req == nil {
		return &gen.CompleteMediaUploadInternalServerError{Code: 500, Message: "缺少上传参数"}, nil
	}
	grantedKeys, err := grantedPermissionKeysFromContext(ctx, h.evaluator, userID)
	if err != nil {
		h.logger.Error("resolve upload permissions failed", zap.Error(err))
		return &gen.CompleteMediaUploadInternalServerError{Code: 500, Message: "权限解析失败"}, nil
	}

	record, err := h.uploadSvc.CompleteDirectUpload(ctx, pkgLogger.TenantFromContext(ctx), &userID, upload.CompleteDirectUploadInput{
		Key:                   optString(req.Key),
		Rule:                  optString(req.Rule),
		Name:                  req.Filename,
		StorageKey:            req.StorageKey,
		Size:                  req.Size,
		MimeType:              optString(req.MimeType),
		Checksum:              optString(req.Checksum),
		ETag:                  optString(req.Etag),
		GrantedPermissionKeys: grantedKeys,
	})
	if err != nil {
		if errors.Is(err, upload.ErrUploadPermissionDenied) {
			return &gen.CompleteMediaUploadUnauthorized{Code: 401, Message: "无权使用该上传配置"}, nil
		}
		h.logger.Error("complete media upload failed", zap.Error(err))
		return &gen.CompleteMediaUploadInternalServerError{Code: 500, Message: err.Error()}, nil
	}

	return &gen.MediaUploadResponse{
		ID:         record.ID,
		Filename:   record.OriginalFilename,
		StorageKey: record.StorageKey,
		URL:        record.URL,
		MimeType:   record.MimeType,
		Size:       record.Size,
		CreatedAt:  record.CreatedAt,
	}, nil
}

func (h *APIHandler) ListVisibleMediaUploadKeys(ctx context.Context) (gen.ListVisibleMediaUploadKeysRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.ListVisibleMediaUploadKeysUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.uploadSvc == nil {
		return &gen.ListVisibleMediaUploadKeysInternalServerError{Code: 500, Message: "上传服务未就绪"}, nil
	}
	grantedKeys, err := grantedPermissionKeysFromContext(ctx, h.evaluator, userID)
	if err != nil {
		h.logger.Error("resolve upload permissions failed", zap.Error(err))
		return &gen.ListVisibleMediaUploadKeysInternalServerError{Code: 500, Message: "权限解析失败"}, nil
	}
	items, err := h.uploadSvc.ListVisibleUploadKeys(ctx, pkgLogger.TenantFromContext(ctx), grantedKeys)
	if err != nil {
		h.logger.Error("list visible media upload keys failed", zap.Error(err))
		return &gen.ListVisibleMediaUploadKeysInternalServerError{Code: 500, Message: err.Error()}, nil
	}
	records := make([]gen.MediaVisibleUploadKey, 0, len(items))
	for i := range items {
		item := items[i]
		rules := make([]gen.MediaVisibleUploadRule, 0, len(item.Rules))
		for j := range item.Rules {
			rule := item.Rules[j]
			rules = append(rules, gen.MediaVisibleUploadRule{
				RuleKey:          rule.RuleKey,
				Name:             rule.Name,
				UploadMode:       gen.MediaVisibleUploadRuleUploadMode(rule.UploadMode),
				Visibility:       gen.MediaVisibleUploadRuleVisibility(rule.Visibility),
				ClientAccept:     append([]string(nil), rule.ClientAccept...),
				MaxSizeBytes:     rule.MaxSizeBytes,
				AllowedMimeTypes: append([]string(nil), rule.AllowedMimeTypes...),
				IsDefault:        rule.IsDefault,
			})
		}
		records = append(records, gen.MediaVisibleUploadKey{
			Key:                      item.Key,
			Name:                     item.Name,
			DefaultRuleKey:           item.DefaultRuleKey,
			UploadMode:               gen.MediaVisibleUploadKeyUploadMode(item.UploadMode),
			Visibility:               gen.MediaVisibleUploadKeyVisibility(item.Visibility),
			ClientAccept:             append([]string(nil), item.ClientAccept...),
			MaxSizeBytes:             item.MaxSizeBytes,
			DirectSizeThresholdBytes: item.DirectSizeThresholdBytes,
			FallbackKey:              item.FallbackKey,
			Rules:                    rules,
		})
	}
	return &gen.MediaVisibleUploadKeyListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

func (h *APIHandler) ListMedia(ctx context.Context) (gen.ListMediaRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.ListMediaUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.uploadSvc == nil {
		return &gen.ListMediaInternalServerError{Code: 500, Message: "上传服务未就绪"}, nil
	}

	items, total, err := h.uploadSvc.List(ctx, pkgLogger.TenantFromContext(ctx), 100)
	if err != nil {
		return &gen.ListMediaInternalServerError{Code: 500, Message: "查询媒体列表失败"}, nil
	}

	records := make([]gen.MediaItem, 0, len(items))
	for i := range items {
		item := items[i]
		records = append(records, gen.MediaItem{
			ID:         item.ID,
			Filename:   item.OriginalFilename,
			StorageKey: item.StorageKey,
			URL:        item.URL,
			MimeType:   item.MimeType,
			Size:       item.Size,
			CreatedAt:  item.CreatedAt,
		})
	}

	return &gen.MediaListResponse{
		Records: records,
		Total:   int(total),
	}, nil
}

func (h *APIHandler) DeleteMedia(ctx context.Context, params gen.DeleteMediaParams) (gen.DeleteMediaRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.DeleteMediaUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if h.uploadSvc == nil {
		return &gen.DeleteMediaInternalServerError{Code: 500, Message: "上传服务未就绪"}, nil
	}

	id, err := uuid.Parse(params.ID)
	if err != nil {
		return &gen.DeleteMediaInternalServerError{Code: 500, Message: "媒体标识非法"}, nil
	}
	if err := h.uploadSvc.Delete(ctx, pkgLogger.TenantFromContext(ctx), id); err != nil {
		if errors.Is(err, upload.ErrRecordNotFound) {
			return &gen.DeleteMediaInternalServerError{Code: 500, Message: "媒体不存在"}, nil
		}
		return &gen.DeleteMediaInternalServerError{Code: 500, Message: "删除媒体失败"}, nil
	}
	return ok(), nil
}
