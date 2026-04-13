// system_register.go: 注册体系后台 CRUD handler（入口 / 策略 / 注册记录）。
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/apperr"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
)

// policySubTables 从 DB 加载策略绑定的 role_codes 和 feature_package_keys。
func (h *APIHandler) policySubTables(ctx context.Context, policyCode string) (roleCodes, pkgKeys []string) {
	roleCodes = []string{}
	pkgKeys = []string{}

	var roleLinks []systemmodels.RegisterPolicyRole
	if err := h.db.WithContext(ctx).Where("policy_code = ?", policyCode).Find(&roleLinks).Error; err == nil {
		if len(roleLinks) > 0 {
			roleIDs := make([]uuid.UUID, 0, len(roleLinks))
			for _, r := range roleLinks {
				roleIDs = append(roleIDs, r.RoleID)
			}
			var roles []systemmodels.Role
			if err := h.db.WithContext(ctx).Where("id IN ?", roleIDs).Find(&roles).Error; err == nil {
				for _, r := range roles {
					roleCodes = append(roleCodes, r.Code)
				}
			}
		}
	}

	var pkgLinks []systemmodels.RegisterPolicyFeaturePackage
	if err := h.db.WithContext(ctx).Where("policy_code = ?", policyCode).Find(&pkgLinks).Error; err == nil {
		if len(pkgLinks) > 0 {
			pkgIDs := make([]uuid.UUID, 0, len(pkgLinks))
			for _, p := range pkgLinks {
				pkgIDs = append(pkgIDs, p.PackageID)
			}
			var pkgs []systemmodels.FeaturePackage
			if err := h.db.WithContext(ctx).Where("id IN ?", pkgIDs).Find(&pkgs).Error; err == nil {
				for _, p := range pkgs {
					pkgKeys = append(pkgKeys, p.PackageKey)
				}
			}
		}
	}
	return roleCodes, pkgKeys
}

// upsertPolicySubTables 在事务内替换策略子表（roles + feature packages）。
func upsertPolicySubTables(tx *gorm.DB, policyCode string, roleCodes, featurePkgKeys []string) error {
	// 删除旧绑定
	if err := tx.Where("policy_code = ?", policyCode).Delete(&systemmodels.RegisterPolicyRole{}).Error; err != nil {
		return err
	}
	if err := tx.Where("policy_code = ?", policyCode).Delete(&systemmodels.RegisterPolicyFeaturePackage{}).Error; err != nil {
		return err
	}

	// 写角色绑定
	if len(roleCodes) > 0 {
		var roles []systemmodels.Role
		if err := tx.Where("code IN ?", roleCodes).Find(&roles).Error; err != nil {
			return err
		}
		roleLinks := make([]systemmodels.RegisterPolicyRole, 0, len(roles))
		for i, r := range roles {
			roleLinks = append(roleLinks, systemmodels.RegisterPolicyRole{
				PolicyCode: policyCode,
				RoleID:     r.ID,
				SortOrder:  i,
			})
		}
		if len(roleLinks) > 0 {
			if err := tx.Create(&roleLinks).Error; err != nil {
				return err
			}
		}
	}

	// 写功能包绑定
	if len(featurePkgKeys) > 0 {
		var pkgs []systemmodels.FeaturePackage
		if err := tx.Where("package_key IN ?", featurePkgKeys).Find(&pkgs).Error; err != nil {
			return err
		}
		pkgLinks := make([]systemmodels.RegisterPolicyFeaturePackage, 0, len(pkgs))
		for i, p := range pkgs {
			pkgLinks = append(pkgLinks, systemmodels.RegisterPolicyFeaturePackage{
				PolicyCode: policyCode,
				PackageID:  p.ID,
				SortOrder:  i,
			})
		}
		if len(pkgLinks) > 0 {
			if err := tx.Create(&pkgLinks).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

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
		ID:                  e.ID,
		AppKey:              e.AppKey,
		EntryCode:           e.EntryCode,
		Name:                e.Name,
		Host:                gen.NewOptString(e.Host),
		PathPrefix:          gen.NewOptString(e.PathPrefix),
		RegisterSource:      gen.NewOptString(e.RegisterSource),
		PolicyCode:          e.PolicyCode,
		LoginPageKey:        e.LoginPageKey,
		Status:              e.Status,
		AllowPublicRegister: boolPtrToOptNil(e.AllowPublicRegister),
		RequireInvite:       boolPtrToOptNil(e.RequireInvite),
		RequireEmailVerify:  boolPtrToOptNil(e.RequireEmailVerify),
		RequireCaptcha:      boolPtrToOptNil(e.RequireCaptcha),
		AutoLogin:           boolPtrToOptNil(e.AutoLogin),
		SortOrder:           gen.NewOptInt(e.SortOrder),
		Remark:              gen.NewOptString(e.Remark),
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
func applyEntryUpsert(e *systemmodels.RegisterEntry, req *gen.RegisterEntryUpsertRequest) {
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
	e.PolicyCode = req.PolicyCode
	if req.LoginPageKey.Set {
		e.LoginPageKey = req.LoginPageKey.Value
	}
	if req.Status.Set {
		e.Status = req.Status.Value
	} else if e.Status == "" {
		e.Status = "enabled"
	}
	e.AllowPublicRegister = optBoolPtr(req.AllowPublicRegister)
	e.RequireInvite = optBoolPtr(req.RequireInvite)
	e.RequireEmailVerify = optBoolPtr(req.RequireEmailVerify)
	e.RequireCaptcha = optBoolPtr(req.RequireCaptcha)
	e.AutoLogin = optBoolPtr(req.AutoLogin)
	if req.SortOrder.Set {
		e.SortOrder = req.SortOrder.Value
	}
	if req.Remark.Set {
		e.Remark = req.Remark.Value
	}
}

func (h *APIHandler) policyToDTO(ctx context.Context, p *systemmodels.RegisterPolicy) *gen.RegisterPolicyItem {
	roleCodes, pkgKeys := h.policySubTables(ctx, p.PolicyCode)
	item := &gen.RegisterPolicyItem{
		ID:                       p.ID,
		AppKey:                   p.AppKey,
		PolicyCode:               p.PolicyCode,
		Name:                     p.Name,
		Description:              gen.NewOptString(p.Description),
		TargetAppKey:             p.TargetAppKey,
		TargetNavigationSpaceKey: p.TargetNavigationSpaceKey,
		TargetHomePath:           gen.NewOptString(p.TargetHomePath),
		DefaultWorkspaceType:     gen.NewOptString(p.DefaultWorkspaceType),
		Status:                   p.Status,
		AllowPublicRegister:      gen.NewOptBool(p.AllowPublicRegister),
		RequireInvite:            gen.NewOptBool(p.RequireInvite),
		RequireEmailVerify:       gen.NewOptBool(p.RequireEmailVerify),
		RequireCaptcha:           gen.NewOptBool(p.RequireCaptcha),
		AutoLogin:                gen.NewOptBool(p.AutoLogin),
		CaptchaProvider:          gen.NewOptString(p.CaptchaProvider),
		CaptchaSiteKey:           gen.NewOptString(p.CaptchaSiteKey),
		RoleCodes:                roleCodes,
		FeaturePackageKeys:       pkgKeys,
	}
	return item
}

func applyPolicyUpsert(p *systemmodels.RegisterPolicy, req *gen.RegisterPolicyUpsertRequest) {
	p.AppKey = req.AppKey
	p.PolicyCode = req.PolicyCode
	p.Name = req.Name
	if req.Description.Set {
		p.Description = req.Description.Value
	}
	p.TargetAppKey = req.TargetAppKey
	p.TargetNavigationSpaceKey = req.TargetNavigationSpaceKey
	if req.TargetHomePath.Set {
		p.TargetHomePath = req.TargetHomePath.Value
	}
	if req.DefaultWorkspaceType.Set {
		p.DefaultWorkspaceType = req.DefaultWorkspaceType.Value
	}
	if req.Status.Set {
		p.Status = req.Status.Value
	} else if p.Status == "" {
		p.Status = "enabled"
	}
	if req.AllowPublicRegister.Set {
		p.AllowPublicRegister = req.AllowPublicRegister.Value
	}
	if req.RequireInvite.Set {
		p.RequireInvite = req.RequireInvite.Value
	}
	if req.RequireEmailVerify.Set {
		p.RequireEmailVerify = req.RequireEmailVerify.Value
	}
	if req.RequireCaptcha.Set {
		p.RequireCaptcha = req.RequireCaptcha.Value
	}
	if req.AutoLogin.Set {
		p.AutoLogin = req.AutoLogin.Value
	}
	if req.CaptchaProvider.Set {
		p.CaptchaProvider = req.CaptchaProvider.Value
	} else if p.CaptchaProvider == "" {
		p.CaptchaProvider = "none"
	}
	if req.CaptchaSiteKey.Set {
		p.CaptchaSiteKey = req.CaptchaSiteKey.Value
	}
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

func (h *APIHandler) CreateRegisterEntry(ctx context.Context, req *gen.RegisterEntryUpsertRequest) (gen.CreateRegisterEntryRes, error) {
	if req == nil {
		return nil, errors.New("请求体为空")
	}
	var entry systemmodels.RegisterEntry
	applyEntryUpsert(&entry, req)
	if err := h.db.WithContext(ctx).Create(&entry).Error; err != nil {
		h.logger.Error("create register entry failed", zap.Error(err))
		return nil, err
	}
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
	applyEntryUpsert(&entry, req)
	if err := h.db.WithContext(ctx).Save(&entry).Error; err != nil {
		return nil, err
	}
	return entryToDTO(&entry), nil
}

func (h *APIHandler) DeleteRegisterEntry(ctx context.Context, params gen.DeleteRegisterEntryParams) (gen.DeleteRegisterEntryRes, error) {
	res := h.db.WithContext(ctx).Where("id = ?", params.ID).Delete(&systemmodels.RegisterEntry{})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return &gen.Error{Code: 404, Message: "注册入口不存在"}, nil
	}
	return &gen.DeleteRegisterEntryNoContent{}, nil
}

// ── register policies CRUD ───────────────────────────────────────────────

func (h *APIHandler) ListRegisterPolicies(ctx context.Context) (gen.ListRegisterPoliciesRes, error) {
	var rows []systemmodels.RegisterPolicy
	if err := h.db.WithContext(ctx).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]gen.RegisterPolicyItem, 0, len(rows))
	for i := range rows {
		records = append(records, *h.policyToDTO(ctx, &rows[i]))
	}
	return &gen.RegisterPolicyList{Records: records, Total: len(records)}, nil
}

func (h *APIHandler) CreateRegisterPolicy(ctx context.Context, req *gen.RegisterPolicyUpsertRequest) (gen.CreateRegisterPolicyRes, error) {
	if req == nil {
		return nil, errors.New("请求体为空")
	}
	var p systemmodels.RegisterPolicy
	applyPolicyUpsert(&p, req)
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		return upsertPolicySubTables(tx, p.PolicyCode, req.RoleCodes, req.FeaturePackageKeys)
	}); err != nil {
		return nil, err
	}
	return h.policyToDTO(ctx, &p), nil
}

func (h *APIHandler) UpdateRegisterPolicy(ctx context.Context, req *gen.RegisterPolicyUpsertRequest, params gen.UpdateRegisterPolicyParams) (gen.UpdateRegisterPolicyRes, error) {
	if req == nil {
		return nil, errors.New("请求体为空")
	}
	var p systemmodels.RegisterPolicy
	if err := h.db.WithContext(ctx).Where("policy_code = ?", params.Code).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.Error{Code: 404, Message: "注册策略不存在"}, nil
		}
		return nil, err
	}
	applyPolicyUpsert(&p, req)
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&p).Error; err != nil {
			return err
		}
		return upsertPolicySubTables(tx, p.PolicyCode, req.RoleCodes, req.FeaturePackageKeys)
	}); err != nil {
		return nil, err
	}
	return h.policyToDTO(ctx, &p), nil
}

func (h *APIHandler) DeleteRegisterPolicy(ctx context.Context, params gen.DeleteRegisterPolicyParams) (gen.DeleteRegisterPolicyRes, error) {
	// 守卫：检查是否有入口仍在引用该策略
	var refCount int64
	if err := h.db.WithContext(ctx).Model(&systemmodels.RegisterEntry{}).
		Where("policy_code = ?", params.Code).Count(&refCount).Error; err != nil {
		return nil, err
	}
	if refCount > 0 {
		return nil, &apperr.ParamError{Msg: fmt.Sprintf("该策略被 %d 个注册入口引用，请先解绑入口后再删除", refCount)}
	}
	res := h.db.WithContext(ctx).Where("policy_code = ?", params.Code).Delete(&systemmodels.RegisterPolicy{})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return &gen.Error{Code: 404, Message: "注册策略不存在"}, nil
	}
	return &gen.DeleteRegisterPolicyNoContent{}, nil
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
		return nil, err
	}
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
		return nil, err
	}
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
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
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
	if params.PolicyCode.Set {
		q = q.Where("register_policy_code = ?", params.PolicyCode.Value)
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
			RegisterPolicyCode: u.RegisterPolicyCode,
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
