package register

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

// ErrPublicRegisterDisabled 公开注册被策略/入口禁用。
var ErrPublicRegisterDisabled = errors.New("公开注册未开启")

// Service 注册体系领域服务。
type Service struct {
	db           *gorm.DB
	repo         *Repository
	resolver     *Resolver
	authSvc      auth.AuthService
	workspaceSvc workspace.Service
	logger       *zap.Logger
}

func NewService(
	db *gorm.DB,
	resolver *Resolver,
	authSvc auth.AuthService,
	workspaceSvc workspace.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		db:           db,
		repo:         NewRepository(db),
		resolver:     resolver,
		authSvc:      authSvc,
		workspaceSvc: workspaceSvc,
		logger:       logger,
	}
}

// RegisterInput 注册入参。
type RegisterInput struct {
	Username         string
	Password         string
	ConfirmPassword  string
	Email            string
	Nickname         string
	CaptchaToken     string
	InvitationCode   string
	AgreementVersion string
	Host             string
	Path             string
	IP               string
	UserAgent        string
}

// RegisterResult 注册结果（auto_login 决定返回 Login 或 Pending）。
type RegisterResult struct {
	User    *user.User
	Login   *dto.LoginResponse
	Landing *LandingInfo
	Pending bool
}

// Register 执行注册完整流程（全事务化）。
// 用户创建 / 审计字段 / 角色绑定 / 功能包绑定均在同一 DB 事务内完成；
// 任一步骤失败整体回滚，不会留下半成品账号。
func (s *Service) Register(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
	eff, err := s.resolver.Resolve(ctx, in.Host, in.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve register context: %w", err)
	}
	if !eff.AllowPublicRegister {
		return nil, ErrPublicRegisterDisabled
	}

	// 基础校验（事务外，快速失败）
	if strings.TrimSpace(in.Username) == "" || in.Password == "" {
		return nil, errors.New("用户名和密码必填")
	}
	if in.ConfirmPassword != "" && in.ConfirmPassword != in.Password {
		return nil, errors.New("两次密码不一致")
	}
	if eff.RequireInvite && strings.TrimSpace(in.InvitationCode) == "" {
		return nil, errors.New("当前入口需要邀请码")
	}
	if eff.RequireCaptcha && strings.TrimSpace(in.CaptchaToken) == "" {
		return nil, errors.New("请先完成人机验证")
	}

	var created *user.User

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 在事务内创建用户（含重复用户名/邮箱校验）
		u, err := s.authSvc.CreateUserTx(tx, &dto.RegisterRequest{
			Username: in.Username,
			Password: in.Password,
			Email:    in.Email,
			Nickname: in.Nickname,
		})
		if err != nil {
			return err
		}
		created = u

		// 2. 回写注册来源审计字段
		updates := map[string]interface{}{
			"register_app_key":     eff.EntryAppKey,
			"register_entry_code":  eff.EntryCode,
			"register_policy_code": eff.PolicyCode,
			"register_source":      eff.RegisterSource,
			"register_ip":          in.IP,
			"register_user_agent":  truncate(in.UserAgent, 512),
			"agreement_version":    in.AgreementVersion,
		}
		if err := tx.Model(&user.User{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update audit fields: %w", err)
		}

		// 3. 绑定策略角色（user_roles + workspace binding snapshot，同一 tx）
		roleLinks, err := s.repo.ListPolicyRoles(ctx, eff.PolicyCode)
		if err != nil {
			return fmt.Errorf("list policy roles: %w", err)
		}
		roleCodes := make([]string, 0)
		if len(roleLinks) > 0 {
			roleIDs := make([]uuid.UUID, 0, len(roleLinks))
			for _, link := range roleLinks {
				roleIDs = append(roleIDs, link.RoleID)
			}
			// 更新 workspace binding snapshot（EnsurePersonalWorkspace 会在 tx 内幂等建 workspace）
			if err := workspacerolebinding.ReplacePersonalRoleBindings(tx, u.ID, roleIDs); err != nil {
				return fmt.Errorf("replace personal role bindings: %w", err)
			}
			// 写 user_roles
			userRoles := make([]systemmodels.UserRole, 0, len(roleIDs))
			for _, rid := range roleIDs {
				userRoles = append(userRoles, systemmodels.UserRole{UserID: u.ID, RoleID: rid})
			}
			if err := tx.Create(&userRoles).Error; err != nil {
				return fmt.Errorf("create user_roles: %w", err)
			}
			// 读取 role code，供策略快照使用
			var roles []systemmodels.Role
			if err := tx.Select("code").Where("id IN ?", roleIDs).Find(&roles).Error; err == nil {
				for _, r := range roles {
					roleCodes = append(roleCodes, r.Code)
				}
			}
		}

		// 4. 绑定策略功能包（user_feature_packages）
		pkgLinks, err := s.repo.ListPolicyFeaturePackages(ctx, eff.PolicyCode)
		if err != nil {
			return fmt.Errorf("list policy packages: %w", err)
		}
		pkgKeys := make([]string, 0)
		for _, link := range pkgLinks {
			ufp := systemmodels.UserFeaturePackage{
				AppKey:    eff.TargetAppKey,
				UserID:    u.ID,
				PackageID: link.PackageID,
				Enabled:   true,
			}
			if err := tx.Where("user_id = ? AND package_id = ?", u.ID, link.PackageID).
				FirstOrCreate(&ufp).Error; err != nil {
				return fmt.Errorf("assign user package: %w", err)
			}
		}
		// 读取 package_key，供策略快照使用
		if len(pkgLinks) > 0 {
			pkgIDs := make([]uuid.UUID, 0, len(pkgLinks))
			for _, p := range pkgLinks {
				pkgIDs = append(pkgIDs, p.PackageID)
			}
			var pkgs []systemmodels.FeaturePackage
			if err := tx.Select("package_key").Where("id IN ?", pkgIDs).Find(&pkgs).Error; err == nil {
				for _, p := range pkgs {
					pkgKeys = append(pkgKeys, p.PackageKey)
				}
			}
		}

		// 5. 写入策略快照（冻结注册时刻的有效策略，防止后续策略变更污染历史记录）
		snapshot := systemmodels.MetaJSON{
			"policy_code":                 eff.PolicyCode,
			"target_app_key":              eff.TargetAppKey,
			"target_navigation_space_key": eff.TargetNavigationSpaceKey,
			"target_home_path":            eff.TargetHomePath,
			"allow_public_register":       eff.AllowPublicRegister,
			"require_invite":              eff.RequireInvite,
			"require_email_verify":        eff.RequireEmailVerify,
			"require_captcha":             eff.RequireCaptcha,
			"auto_login":                  eff.AutoLogin,
			"captcha_provider":            eff.CaptchaProvider,
			"role_codes":                  roleCodes,
			"feature_package_keys":        pkgKeys,
		}
		if err := tx.Model(&user.User{}).Where("id = ?", u.ID).
			Update("register_policy_snapshot", snapshot).Error; err != nil {
			return fmt.Errorf("write policy snapshot: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 事务外：确保 personal workspace 存在（幂等，tx 内 workspacerolebinding 已尝试建立）
	if _, err := s.workspaceSvc.EnsurePersonalWorkspaceForUser(created.ID); err != nil {
		s.logger.Warn("ensure personal workspace on register failed", zap.Error(err))
	}

	// 事务成功后生成 token（auto_login）
	landing := &LandingInfo{
		AppKey:             eff.TargetAppKey,
		NavigationSpaceKey: eff.TargetNavigationSpaceKey,
		HomePath:           eff.TargetHomePath,
	}
	result := &RegisterResult{User: created, Landing: landing}
	if eff.AutoLogin {
		loginResp, err := s.authSvc.BuildLoginResponse(created)
		if err != nil {
			return nil, fmt.Errorf("build login response: %w", err)
		}
		result.Login = loginResp
	} else {
		result.Pending = true
	}
	return result, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// 避免 uuid 未使用时的 lint 误报
var _ = uuid.Nil
