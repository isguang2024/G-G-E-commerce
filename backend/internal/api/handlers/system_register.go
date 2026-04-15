// system_register.go: 注册体系后台 CRUD handler（入口 / 策略 / 注册记录）。
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/apperr"
	"github.com/gg-ecommerce/backend/internal/modules/observability/audit"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/register"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
)

// ── helpers ──────────────────────────────────────────────────────────────

func optBoolPtr(o gen.OptNilBool) *bool {
	if !o.Set || o.Null {
		return nil
	}
	v := o.Value
	return &v
}

func boolPtrToOptNil(p *bool) gen.OptNilBool {
	if p == nil {
		return gen.OptNilBool{Set: true, Null: true}
	}
	return gen.OptNilBool{Set: true, Value: *p}
}

func entryToDTO(e *systemmodels.RegisterEntry) *gen.RegisterEntryItem {
	out := &gen.RegisterEntryItem{
		ID:                       e.ID,
		AppKey:                   e.AppKey,
		EntryCode:                e.EntryCode,
		Name:                     e.Name,
		Host:                     gen.NewOptString(e.Host),
		PathPrefix:               gen.NewOptString(e.PathPrefix),
		RegisterSource:           gen.NewOptString(e.RegisterSource),
		LoginPageKey:             e.LoginPageKey,
		Status:                   e.Status,
		AllowPublicRegister:      e.AllowPublicRegister,
		RequireInvite:            e.RequireInvite,
		RequireEmailVerify:       e.RequireEmailVerify,
		RequireCaptcha:           e.RequireCaptcha,
		AutoLogin:                e.AutoLogin,
		IsSystemReserved:         e.IsSystemReserved,
		TargetURL:                gen.NewOptString(e.TargetURL),
		TargetAppKey:             gen.NewOptString(e.TargetAppKey),
		TargetNavigationSpaceKey: gen.NewOptString(e.TargetNavigationSpaceKey),
		TargetHomePath:           gen.NewOptString(e.TargetHomePath),
		CaptchaProvider:          gen.NewOptString(e.CaptchaProvider),
		CaptchaSiteKey:           gen.NewOptString(e.CaptchaSiteKey),
		Description:              gen.NewOptString(e.Description),
		RoleCodes:                []string(e.RoleCodes),
		FeaturePackageKeys:       []string(e.FeaturePackageKeys),
		SortOrder:                gen.NewOptInt(e.SortOrder),
		Remark:                   gen.NewOptString(e.Remark),
	}
	return out
}

func metaJSONToRawMap(src systemmodels.MetaJSON) map[string]jx.Raw {
	if len(src) == 0 {
		return map[string]jx.Raw{}
	}
	out := make(map[string]jx.Raw, len(src))
	for key, value := range src {
		if buf, err := json.Marshal(value); err == nil {
			out[key] = jx.Raw(buf)
		}
	}
	return out
}

func rawMapToMetaJSON(src map[string]jx.Raw) systemmodels.MetaJSON {
	if len(src) == 0 {
		return systemmodels.MetaJSON{}
	}
	out := make(systemmodels.MetaJSON, len(src))
	for key, value := range src {
		if len(value) == 0 {
			out[key] = nil
			continue
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(value), &decoded); err != nil {
			out[key] = string(value)
			continue
		}
		out[key] = decoded
	}
	return out
}
func applyEntryUpsert(e *systemmodels.RegisterEntry, req *gen.RegisterEntryUpsertRequest) error {
	e.AppKey = req.AppKey
	e.EntryCode = req.EntryCode
	e.Name = req.Name
	if req.Host.Set {
		e.Host = req.Host.Value
	}
	if req.PathPrefix.Set {
		e.PathPrefix = req.PathPrefix.Value
	}
	if req.RegisterSource.Set {
		e.RegisterSource = req.RegisterSource.Value
	}
	if req.LoginPageKey.Set {
		e.LoginPageKey = req.LoginPageKey.Value
	}
	if req.Status.Set {
		e.Status = req.Status.Value
	} else if e.Status == "" {
		e.Status = "enabled"
	}
	if req.AllowPublicRegister.Set {
		e.AllowPublicRegister = req.AllowPublicRegister.Value
	}
	if req.RequireInvite.Set {
		e.RequireInvite = req.RequireInvite.Value
	}
	if req.RequireEmailVerify.Set {
		e.RequireEmailVerify = req.RequireEmailVerify.Value
	}
	if req.RequireCaptcha.Set {
		e.RequireCaptcha = req.RequireCaptcha.Value
	}
	if req.AutoLogin.Set {
		e.AutoLogin = req.AutoLogin.Value
	}
	if req.IsSystemReserved.Set {
		e.IsSystemReserved = req.IsSystemReserved.Value
	}
	if req.TargetURL.Set {
		if v := strings.TrimSpace(req.TargetURL.Value); v != "" && !register.IsSafeRedirectURL(v) {
			return fmt.Errorf("target_url 不安全，仅允许 http(s) 或相对路径")
		}
		e.TargetURL = req.TargetURL.Value
	}
	if req.TargetAppKey.Set {
		e.TargetAppKey = req.TargetAppKey.Value
	}
	if req.TargetNavigationSpaceKey.Set {
		e.TargetNavigationSpaceKey = req.TargetNavigationSpaceKey.Value
	}
	if req.TargetHomePath.Set {
		e.TargetHomePath = req.TargetHomePath.Value
	}
	if req.CaptchaProvider.Set {
		e.CaptchaProvider = req.CaptchaProvider.Value
	}
	if req.CaptchaSiteKey.Set {
		e.CaptchaSiteKey = req.CaptchaSiteKey.Value
	}
	if req.Description.Set {
		e.Description = req.Description.Value
	}
	if req.RoleCodes != nil {
		e.RoleCodes = systemmodels.StringList(req.RoleCodes)
	}
	if req.FeaturePackageKeys != nil {
		e.FeaturePackageKeys = systemmodels.StringList(req.FeaturePackageKeys)
	}
	if req.SortOrder.Set {
		e.SortOrder = req.SortOrder.Value
	}
	if req.Remark.Set {
		e.Remark = req.Remark.Value
	}
	return nil
}

func loginPageTemplateToDTO(item *systemmodels.LoginPageTemplate) *gen.LoginPageTemplateItem {
	if item == nil {
		return nil
	}
	return &gen.LoginPageTemplateItem{
		ID:          item.ID,
		TenantID:    item.TenantID,
		TemplateKey: item.TemplateKey,
		Name:        item.Name,
		Scene:       item.Scene,
		AppScope:    item.AppScope,
		Status:      item.Status,
		IsDefault:   item.IsDefault,
		Config:      gen.LoginPageTemplateItemConfig(metaJSONToRawMap(item.Config)),
		Meta:        gen.LoginPageTemplateItemMeta(metaJSONToRawMap(item.Meta)),
		CreatedAt:   gen.NewOptDateTime(item.CreatedAt),
		UpdatedAt:   gen.NewOptDateTime(item.UpdatedAt),
	}
}

func applyLoginPageTemplateUpsert(
	item *systemmodels.LoginPageTemplate,
	req *gen.LoginPageTemplateUpsertRequest,
) {
	item.TenantID = "default"
	if req.TenantID.Set && req.TenantID.Value != "" {
		item.TenantID = req.TenantID.Value
	}
	item.TemplateKey = req.TemplateKey
	item.Name = req.Name
	if req.Scene.Set {
		item.Scene = req.Scene.Value
	} else if item.Scene == "" {
		item.Scene = "auth_family"
	}
	if req.AppScope.Set {
		item.AppScope = req.AppScope.Value
	} else if item.AppScope == "" {
		item.AppScope = "shared"
	}
	if req.Status.Set {
		item.Status = req.Status.Value
	} else if item.Status == "" {
		item.Status = "normal"
	}
	if req.IsDefault.Set {
		item.IsDefault = req.IsDefault.Value
	}
	if req.Config.Set {
		item.Config = rawMapToMetaJSON(req.Config.Value)
	}
	if req.Meta.Set {
		item.Meta = rawMapToMetaJSON(req.Meta.Value)
	}
}

// ── register entries CRUD ────────────────────────────────────────────────

func (h *APIHandler) ListRegisterEntries(ctx context.Context) (gen.ListRegisterEntriesRes, error) {
	var rows []systemmodels.RegisterEntry
	if err := h.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]gen.RegisterEntryItem, 0, len(rows))
	for i := range rows {
		records = append(records, *entryToDTO(&rows[i]))
	}
	return &gen.RegisterEntryList{Records: records, Total: len(records)}, nil
}

// checkEntryCodeUnique 验证 entry_code 唯一性。若 excludeID 非空，则排除自身（编辑场景）。
// 冲突时返回 FieldError，mapper 会翻译为 400 + details.entry_code = Reason。
func (h *APIHandler) checkEntryCodeUnique(ctx context.Context, code string, excludeID *uuid.UUID) error {
	if code == "" {
		return nil
	}
	q := h.db.WithContext(ctx).Model(&systemmodels.RegisterEntry{}).Where("entry_code = ?", code)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return &apperr.FieldError{
			Field:  "entry_code",
			Reason: "入口 Code 已存在",
			Msg:    "入口 Code 已存在，请更换",
			Code:   apperr.CodeConflict,
		}
	}
	return nil
}

func (h *APIHandler) CreateRegisterEntry(ctx context.Context, req *gen.RegisterEntryUpsertRequest) (gen.CreateRegisterEntryRes, error) {
	if req == nil {
		return nil, errors.New("请求体为空")
	}
	if strings.TrimSpace(req.EntryCode) == "" {
		return nil, &apperr.FieldError{Field: "entry_code", Reason: "不能为空", Msg: "请填写入口 Code"}
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, &apperr.FieldError{Field: "name", Reason: "不能为空", Msg: "请填写名称"}
	}
	if strings.TrimSpace(req.AppKey) == "" {
		return nil, &apperr.FieldError{Field: "app_key", Reason: "不能为空", Msg: "请选择所属 App"}
	}
	if err := h.checkEntryCodeUnique(ctx, req.EntryCode, nil); err != nil {
		return nil, err
	}
	var entry systemmodels.RegisterEntry
	if err := applyEntryUpsert(&entry, req); err != nil {
		return nil, &apperr.ParamError{Msg: err.Error()}
	}
	if err := h.db.WithContext(ctx).Create(&entry).Error; err != nil {
		h.logger.Error("create register entry failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.entry.create",
			ResourceType: "register_entry",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata: map[string]any{
				"entry_code": req.EntryCode,
				"app_key":    req.AppKey,
			},
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.register.entry.create",
		ResourceType: "register_entry",
		ResourceID:   entry.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata: map[string]any{
			"entry_code": entry.EntryCode,
			"app_key":    req.AppKey,
		},
	})
	return entryToDTO(&entry), nil
}

func (h *APIHandler) UpdateRegisterEntry(ctx context.Context, req *gen.RegisterEntryUpsertRequest, params gen.UpdateRegisterEntryParams) (gen.UpdateRegisterEntryRes, error) {
	if req == nil {
		return nil, errors.New("请求体为空")
	}
	var entry systemmodels.RegisterEntry
	if err := h.db.WithContext(ctx).Where("id = ?", params.ID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.Error{Code: 404, Message: "注册入口不存在"}, nil
		}
		return nil, err
	}
	// 系统保留入口：不允许修改 entry_code 和 is_system_reserved
	if entry.IsSystemReserved {
		if req.EntryCode != entry.EntryCode {
			return nil, &apperr.FieldError{Field: "entry_code", Reason: "系统保留入口不可修改", Msg: "系统保留入口不可修改 entry_code"}
		}
		if req.IsSystemReserved.Set && !req.IsSystemReserved.Value {
			return nil, &apperr.FieldError{Field: "is_system_reserved", Reason: "系统保留入口不可取消保留标记", Msg: "系统保留入口不可取消保留标记"}
		}
	}
	if strings.TrimSpace(req.EntryCode) == "" {
		return nil, &apperr.FieldError{Field: "entry_code", Reason: "不能为空", Msg: "请填写入口 Code"}
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, &apperr.FieldError{Field: "name", Reason: "不能为空", Msg: "请填写名称"}
	}
	if strings.TrimSpace(req.AppKey) == "" {
		return nil, &apperr.FieldError{Field: "app_key", Reason: "不能为空", Msg: "请选择所属 App"}
	}
	if err := h.checkEntryCodeUnique(ctx, req.EntryCode, &entry.ID); err != nil {
		return nil, err
	}
	if err := applyEntryUpsert(&entry, req); err != nil {
		return nil, &apperr.ParamError{Msg: err.Error()}
	}
	if err := h.db.WithContext(ctx).Save(&entry).Error; err != nil {
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.entry.update",
			ResourceType: "register_entry",
			ResourceID:   entry.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.register.entry.update",
		ResourceType: "register_entry",
		ResourceID:   entry.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata:     map[string]any{"entry_code": entry.EntryCode},
	})
	return entryToDTO(&entry), nil
}

func (h *APIHandler) DeleteRegisterEntry(ctx context.Context, params gen.DeleteRegisterEntryParams) (gen.DeleteRegisterEntryRes, error) {
	var entry systemmodels.RegisterEntry
	if err := h.db.WithContext(ctx).Where("id = ?", params.ID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.Error{Code: 404, Message: "注册入口不存在"}, nil
		}
		return nil, err
	}
	if entry.IsSystemReserved {
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.entry.delete",
			ResourceType: "register_entry",
			ResourceID:   entry.ID.String(),
			Outcome:      audit.OutcomeDenied,
			ErrorCode:    "system_reserved",
		})
		return nil, &apperr.ParamError{Msg: "系统保留入口不可删除"}
	}
	if err := h.db.WithContext(ctx).Delete(&entry).Error; err != nil {
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.entry.delete",
			ResourceType: "register_entry",
			ResourceID:   entry.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.register.entry.delete",
		ResourceType: "register_entry",
		ResourceID:   entry.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata:     map[string]any{"entry_code": entry.EntryCode},
	})
	return &gen.DeleteRegisterEntryNoContent{}, nil
}

func (h *APIHandler) ListLoginPageTemplates(ctx context.Context) (*gen.LoginPageTemplateList, error) {
	var rows []systemmodels.LoginPageTemplate
	if err := h.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", "default").
		Order("is_default DESC, updated_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]gen.LoginPageTemplateItem, 0, len(rows))
	for i := range rows {
		if item := loginPageTemplateToDTO(&rows[i]); item != nil {
			records = append(records, *item)
		}
	}
	return &gen.LoginPageTemplateList{Records: records, Total: len(records)}, nil
}

func (h *APIHandler) CreateLoginPageTemplate(
	ctx context.Context,
	req *gen.LoginPageTemplateUpsertRequest,
) (*gen.LoginPageTemplateItem, error) {
	if req == nil {
		return nil, errors.New("请求体为空")
	}
	var item systemmodels.LoginPageTemplate
	applyLoginPageTemplateUpsert(&item, req)
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if item.IsDefault {
			if err := tx.Model(&systemmodels.LoginPageTemplate{}).
				Where("tenant_id = ? AND scene = ? AND deleted_at IS NULL", item.TenantID, item.Scene).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&item).Error
	}); err != nil {
		h.logger.Error("create login page template failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.template.create",
			ResourceType: "login_page_template",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata: map[string]any{
				"template_key": req.TemplateKey,
				"scene":        item.Scene,
			},
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.register.template.create",
		ResourceType: "login_page_template",
		ResourceID:   item.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata: map[string]any{
			"template_key": item.TemplateKey,
			"scene":        item.Scene,
		},
	})
	return loginPageTemplateToDTO(&item), nil
}

func (h *APIHandler) UpdateLoginPageTemplate(
	ctx context.Context,
	req *gen.LoginPageTemplateUpsertRequest,
	params gen.UpdateLoginPageTemplateParams,
) (*gen.LoginPageTemplateItem, error) {
	if req == nil {
		return nil, errors.New("请求体为空")
	}
	var item systemmodels.LoginPageTemplate
	if err := h.db.WithContext(ctx).
		Where("tenant_id = ? AND template_key = ? AND deleted_at IS NULL", "default", params.TemplateKey).
		First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	applyLoginPageTemplateUpsert(&item, req)
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if item.IsDefault {
			if err := tx.Model(&systemmodels.LoginPageTemplate{}).
				Where("tenant_id = ? AND scene = ? AND template_key <> ? AND deleted_at IS NULL", item.TenantID, item.Scene, item.TemplateKey).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(&item).Error
	}); err != nil {
		h.logger.Error("update login page template failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.template.update",
			ResourceType: "login_page_template",
			ResourceID:   item.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata: map[string]any{
				"template_key": item.TemplateKey,
			},
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.register.template.update",
		ResourceType: "login_page_template",
		ResourceID:   item.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata: map[string]any{
			"template_key": item.TemplateKey,
			"scene":        item.Scene,
		},
	})
	return loginPageTemplateToDTO(&item), nil
}

func (h *APIHandler) DeleteLoginPageTemplate(
	ctx context.Context,
	params gen.DeleteLoginPageTemplateParams,
) (gen.DeleteLoginPageTemplateRes, error) {
	res := h.db.WithContext(ctx).
		Where("tenant_id = ? AND template_key = ?", "default", params.TemplateKey).
		Delete(&systemmodels.LoginPageTemplate{})
	if res.Error != nil {
		h.logger.Error("delete login page template failed", zap.Error(res.Error))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.template.delete",
			ResourceType: "login_page_template",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(res.Error),
			Metadata: map[string]any{
				"template_key": params.TemplateKey,
			},
		})
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		h.audit.Record(ctx, audit.Event{
			Action:       "system.register.template.delete",
			ResourceType: "login_page_template",
			Outcome:      audit.OutcomeError,
			ErrorCode:    "not_found",
			Metadata: map[string]any{
				"template_key": params.TemplateKey,
			},
		})
		return nil, gorm.ErrRecordNotFound
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.register.template.delete",
		ResourceType: "login_page_template",
		Outcome:      audit.OutcomeSuccess,
		Metadata: map[string]any{
			"template_key": params.TemplateKey,
		},
	})
	return &gen.DeleteLoginPageTemplateNoContent{}, nil
}

// ── register logs ────────────────────────────────────────────────────────

func (h *APIHandler) ListRegisterLogs(ctx context.Context, params gen.ListRegisterLogsParams) (gen.ListRegisterLogsRes, error) {
	q := h.db.WithContext(ctx).Model(&usermodel.User{}).
		Where("register_entry_code <> ''")
	if params.Source.Set {
		q = q.Where("register_source = ?", params.Source.Value)
	}
	if params.EntryCode.Set {
		q = q.Where("register_entry_code = ?", params.EntryCode.Value)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	page := 1
	if params.Page.Set && params.Page.Value > 0 {
		page = params.Page.Value
	}
	size := 20
	if params.PageSize.Set && params.PageSize.Value > 0 {
		size = params.PageSize.Value
	}
	var rows []usermodel.User
	if err := q.Order("created_at DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]gen.RegisterLogItem, 0, len(rows))
	for i := range rows {
		u := rows[i]
		item := gen.RegisterLogItem{
			UserID:             u.ID,
			Username:           u.Username,
			Email:              gen.NewOptString(u.Email),
			RegisterAppKey:     gen.NewOptString(u.RegisterAppKey),
			RegisterEntryCode:  u.RegisterEntryCode,
			RegisterSource:     u.RegisterSource,
			RegisterIP:         gen.NewOptString(u.RegisterIP),
			RegisterUserAgent:  gen.NewOptString(u.RegisterUserAgent),
			AgreementVersion:   gen.NewOptString(u.AgreementVersion),
			CreatedAt:          u.CreatedAt,
		}
		if len(u.RegisterPolicySnapshot) > 0 {
			snap := make(gen.RegisterLogItemPolicySnapshot, len(u.RegisterPolicySnapshot))
			for k, v := range u.RegisterPolicySnapshot {
				b, err := json.Marshal(v)
				if err == nil {
					snap[k] = jx.Raw(b)
				}
			}
			item.PolicySnapshot = gen.NewOptNilRegisterLogItemPolicySnapshot(snap)
		}
		records = append(records, item)
	}
	return &gen.RegisterLogList{Records: records, Total: int(total)}, nil
}

var _ = uuid.Nil
