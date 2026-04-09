package register

import (
	"context"
	"fmt"

	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
)

// Resolver 负责把 (host, path) 解析为 EffectiveRegisterContext。
type Resolver struct {
	repo *Repository
}

func NewResolver(repo *Repository) *Resolver { return &Resolver{repo: repo} }

// Resolve 解析注册上下文。算法：
// 1. host+path 命中 register_entries
// 2. 未命中则 fallback 到 entry_code=default
// 3. 加载 policy
// 4. 合并 entry.*bool 覆盖 policy 同名字段
func (r *Resolver) Resolve(ctx context.Context, host, path string) (*EffectiveRegisterContext, error) {
	entry, err := r.repo.FindEntryByHostPath(ctx, host, path)
	if err != nil {
		if !IsNotFound(err) {
			return nil, err
		}
		entry, err = r.repo.FindEntryByCode(ctx, permissionseed.DefaultRegisterEntryCode)
		if err != nil {
			return nil, fmt.Errorf("register: no entry matched and default missing: %w", err)
		}
	}
	policy, err := r.repo.FindPolicyByCode(ctx, entry.PolicyCode)
	if err != nil {
		return nil, fmt.Errorf("register: load policy %q: %w", entry.PolicyCode, err)
	}
	eff := &EffectiveRegisterContext{
		EntryCode:                entry.EntryCode,
		EntryName:                entry.Name,
		EntryAppKey:              entry.AppKey,
		RegisterSource:           entry.RegisterSource,
		PolicyCode:               policy.PolicyCode,
		TargetAppKey:             policy.TargetAppKey,
		TargetNavigationSpaceKey: policy.TargetNavigationSpaceKey,
		TargetHomePath:           policy.TargetHomePath,
		DefaultWorkspaceType:     policy.DefaultWorkspaceType,
		AllowPublicRegister:      policy.AllowPublicRegister,
		RequireInvite:            policy.RequireInvite,
		RequireEmailVerify:       policy.RequireEmailVerify,
		RequireCaptcha:           policy.RequireCaptcha,
		AutoLogin:                policy.AutoLogin,
		CaptchaProvider:          policy.CaptchaProvider,
		CaptchaSiteKey:           policy.CaptchaSiteKey,
	}
	if entry.AllowPublicRegister != nil {
		eff.AllowPublicRegister = *entry.AllowPublicRegister
	}
	if entry.RequireInvite != nil {
		eff.RequireInvite = *entry.RequireInvite
	}
	if entry.RequireEmailVerify != nil {
		eff.RequireEmailVerify = *entry.RequireEmailVerify
	}
	if entry.RequireCaptcha != nil {
		eff.RequireCaptcha = *entry.RequireCaptcha
	}
	if entry.AutoLogin != nil {
		eff.AutoLogin = *entry.AutoLogin
	}
	return eff, nil
}
