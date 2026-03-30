package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	spacesvc "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

const (
	defaultDemoPassword = "Demo123456"
	defaultDemoTeamName = "演示团队"
	defaultDemoSpaceKey = "ops"
)

type demoUserSpec struct {
	Username      string
	Email         string
	Nickname      string
	IsSuperAdmin  bool
	PlatformRoles []string
}

type demoInitializer struct {
	db           *gorm.DB
	logger       *zap.Logger
	spaceService spacesvc.Service
	refresher    permissionrefresh.Service
	passwordHash string
	spaceKey     string
	teamName     string

	platformAdmin systemmodels.User
	teamAdmin     systemmodels.User
	member        systemmodels.User
	team          systemmodels.Tenant

	platformSender        systemmodels.MessageSender
	platformManagerSender systemmodels.MessageSender
	teamSender            systemmodels.MessageSender
	teamManagerSender     systemmodels.MessageSender

	platformTemplate systemmodels.MessageTemplate
	teamTemplate     systemmodels.MessageTemplate

	platformGroup systemmodels.MessageRecipientGroup
	teamGroup     systemmodels.MessageRecipientGroup
}

func main() {
	var (
		passwordFlag    = flag.String("password", defaultDemoPassword, "演示账号统一密码")
		teamNameFlag    = flag.String("team-name", defaultDemoTeamName, "演示团队名称")
		spaceKeyFlag    = flag.String("space-key", defaultDemoSpaceKey, "演示菜单空间标识")
		allowProduction = flag.Bool("allow-production", false, "允许在 production 环境执行")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if strings.EqualFold(strings.TrimSpace(cfg.Env), "production") && !*allowProduction {
		log.Fatalf("init-demo 仅允许在非生产环境执行；如需强制执行，请显式传入 -allow-production")
	}

	appLogger, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	_, err = database.Init(&cfg.DB, database.RuntimeOptions{
		Env:      cfg.Env,
		LogLevel: cfg.Log.Level,
	})
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	hash, err := password.Hash(*passwordFlag)
	if err != nil {
		appLogger.Fatal("Failed to hash demo password", zap.Error(err))
	}

	boundaryService := teamboundary.NewService(database.DB)
	platformService := platformaccess.NewService(database.DB)
	roleSnapshotService := platformroleaccess.NewService(database.DB)
	refresher := permissionrefresh.NewService(database.DB, boundaryService, platformService, roleSnapshotService)
	spaceService := spacesvc.NewService(database.DB, refresher, appLogger)

	initializer := &demoInitializer{
		db:           database.DB,
		logger:       appLogger,
		spaceService: spaceService,
		refresher:    refresher,
		passwordHash: hash,
		spaceKey:     normalizeSpaceKey(*spaceKeyFlag),
		teamName:     strings.TrimSpace(*teamNameFlag),
	}

	if err := initializer.Run(); err != nil {
		appLogger.Fatal("Demo initialization failed", zap.Error(err))
	}

	fmt.Println()
	fmt.Println("✅ 演示数据初始化完成")
	fmt.Println()
	fmt.Println("账号：")
	fmt.Printf("  平台管理员: %s / %s\n", initializer.platformAdmin.Username, *passwordFlag)
	fmt.Printf("  团队管理员: %s / %s\n", initializer.teamAdmin.Username, *passwordFlag)
	fmt.Printf("  普通成员: %s / %s\n", initializer.member.Username, *passwordFlag)
	fmt.Println()
	fmt.Println("演示数据：")
	fmt.Printf("  团队: %s\n", initializer.team.Name)
	fmt.Printf("  菜单空间: %s\n", initializer.spaceKey)
	fmt.Printf("  平台模板: %s\n", initializer.platformTemplate.Name)
	fmt.Printf("  团队模板: %s\n", initializer.teamTemplate.Name)
	fmt.Printf("  平台接收组: %s\n", initializer.platformGroup.Name)
	fmt.Printf("  团队接收组: %s\n", initializer.teamGroup.Name)
	fmt.Println()
	fmt.Println("建议：")
	fmt.Println("  1. 先执行 go test ./... 和 pnpm exec vue-tsc --noEmit")
	fmt.Println("  2. 再用三类账号分别登录回归菜单、页面、菜单空间和消息链路")
	fmt.Println()
}

func (s *demoInitializer) Run() error {
	if err := s.ensureDemoUsers(); err != nil {
		return err
	}
	if err := s.ensureDemoTeam(); err != nil {
		return err
	}
	if err := s.ensureDemoTeamFeaturePackages(); err != nil {
		return err
	}
	if err := s.refreshSnapshots(); err != nil {
		return err
	}
	if err := s.ensureDemoMenuSpace(); err != nil {
		return err
	}
	if err := s.ensureDemoMessagingAssets(); err != nil {
		return err
	}
	if err := s.seedDemoMessages(); err != nil {
		return err
	}
	return nil
}

func (s *demoInitializer) ensureDemoUsers() error {
	specs := []demoUserSpec{
		{
			Username:      "platform_admin_demo",
			Email:         "platform_admin_demo@gg.demo",
			Nickname:      "平台演示管理员",
			IsSuperAdmin:  true,
			PlatformRoles: []string{"admin"},
		},
		{
			Username:      "team_admin_demo",
			Email:         "team_admin_demo@gg.demo",
			Nickname:      "团队演示管理员",
			IsSuperAdmin:  false,
			PlatformRoles: nil,
		},
		{
			Username:      "member_demo",
			Email:         "member_demo@gg.demo",
			Nickname:      "普通演示成员",
			IsSuperAdmin:  false,
			PlatformRoles: nil,
		},
	}

	for _, spec := range specs {
		user, err := s.ensureUser(spec)
		if err != nil {
			return err
		}
		switch spec.Username {
		case "platform_admin_demo":
			s.platformAdmin = *user
		case "team_admin_demo":
			s.teamAdmin = *user
		case "member_demo":
			s.member = *user
		}
		for _, roleCode := range spec.PlatformRoles {
			if err := s.ensurePlatformRoleAssignment(user.ID, roleCode); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *demoInitializer) ensureUser(spec demoUserSpec) (*systemmodels.User, error) {
	var user systemmodels.User
	err := s.db.Where("username = ?", spec.Username).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = systemmodels.User{
			Email:          spec.Email,
			Username:       spec.Username,
			PasswordHash:   s.passwordHash,
			Nickname:       spec.Nickname,
			Status:         "active",
			IsSuperAdmin:   spec.IsSuperAdmin,
			RegisterSource: "seed",
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"email":           spec.Email,
		"password_hash":   s.passwordHash,
		"nickname":        spec.Nickname,
		"status":          "active",
		"is_super_admin":  spec.IsSuperAdmin,
		"register_source": "seed",
	}).Error; err != nil {
		return nil, err
	}
	user.Email = spec.Email
	user.PasswordHash = s.passwordHash
	user.Nickname = spec.Nickname
	user.Status = "active"
	user.IsSuperAdmin = spec.IsSuperAdmin
	user.RegisterSource = "seed"
	return &user, nil
}

func (s *demoInitializer) ensurePlatformRoleAssignment(userID uuid.UUID, roleCode string) error {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return nil
	}
	var role systemmodels.Role
	if err := s.db.Where("tenant_id IS NULL AND code = ? AND deleted_at IS NULL", roleCode).First(&role).Error; err != nil {
		return err
	}
	var count int64
	if err := s.db.Model(&systemmodels.UserRole{}).
		Where("user_id = ? AND role_id = ? AND tenant_id IS NULL", userID, role.ID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return s.db.Create(&systemmodels.UserRole{
		UserID:   userID,
		RoleID:   role.ID,
		TenantID: nil,
	}).Error
}

func (s *demoInitializer) ensureDemoTeam() error {
	teamName := s.teamName
	if teamName == "" {
		teamName = defaultDemoTeamName
	}
	var team systemmodels.Tenant
	err := s.db.Where("name = ? AND deleted_at IS NULL", teamName).First(&team).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		team = systemmodels.Tenant{
			Name:       teamName,
			Remark:     "系统收尾回归演示团队",
			Plan:       "demo",
			OwnerID:    s.teamAdmin.ID,
			MaxMembers: 20,
			Status:     "active",
		}
		if err := s.db.Create(&team).Error; err != nil {
			return err
		}
	} else {
		if err := s.db.Model(&team).Updates(map[string]interface{}{
			"remark":      "系统收尾回归演示团队",
			"plan":        "demo",
			"owner_id":    s.teamAdmin.ID,
			"max_members": 20,
			"status":      "active",
		}).Error; err != nil {
			return err
		}
		team.OwnerID = s.teamAdmin.ID
		team.Status = "active"
	}

	s.team = team
	if err := s.ensureTenantMember(team.ID, s.teamAdmin.ID, "team_admin", &s.platformAdmin.ID); err != nil {
		return err
	}
	if err := s.ensureTenantMemberRemoved(team.ID, s.member.ID); err != nil {
		return err
	}
	return nil
}

func (s *demoInitializer) ensureTenantMember(tenantID, userID uuid.UUID, roleCode string, invitedBy *uuid.UUID) error {
	var member systemmodels.TenantMember
	err := s.db.Where("tenant_id = ? AND user_id = ? AND deleted_at IS NULL", tenantID, userID).First(&member).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	now := time.Now()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		member = systemmodels.TenantMember{
			TenantID:  tenantID,
			UserID:    userID,
			RoleCode:  roleCode,
			Status:    "active",
			JoinedAt:  now,
			InvitedBy: invitedBy,
		}
		return s.db.Create(&member).Error
	}
	return s.db.Model(&member).Updates(map[string]interface{}{
		"role_code":  roleCode,
		"status":     "active",
		"invited_by": invitedBy,
	}).Error
}

func (s *demoInitializer) ensureDemoTeamFeaturePackages() error {
	packageID, err := s.findFeaturePackageID("team.member_admin")
	if err != nil {
		return err
	}
	var grant systemmodels.TeamFeaturePackage
	err = s.db.Where("team_id = ? AND package_id = ?", s.team.ID, packageID).First(&grant).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	now := time.Now()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		grant = systemmodels.TeamFeaturePackage{
			TeamID:    s.team.ID,
			PackageID: packageID,
			Enabled:   true,
			GrantedBy: &s.platformAdmin.ID,
			GrantedAt: &now,
		}
		return s.db.Create(&grant).Error
	}
	return s.db.Model(&grant).Updates(map[string]interface{}{
		"enabled":    true,
		"granted_by": s.platformAdmin.ID,
		"granted_at": now,
	}).Error
}

func (s *demoInitializer) ensureTenantMemberRemoved(tenantID, userID uuid.UUID) error {
	return s.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Delete(&systemmodels.TenantMember{}).Error
}

func (s *demoInitializer) refreshSnapshots() error {
	if err := s.refresher.RefreshAllPlatformUsers(); err != nil {
		return err
	}
	if err := s.refresher.RefreshAllPlatformRoles(); err != nil {
		return err
	}
	return s.refresher.RefreshAllTeams()
}

func (s *demoInitializer) ensureDemoMenuSpace() error {
	if s.spaceKey == "" || s.spaceKey == systemmodels.DefaultMenuSpaceKey {
		s.spaceKey = defaultDemoSpaceKey
	}
	if _, err := s.spaceService.SaveSpace(&spacesvc.SaveSpaceRequest{
		SpaceKey:        s.spaceKey,
		Name:            "运营空间",
		Description:     "系统收尾回归专用的非默认菜单空间示例",
		DefaultHomePath: "/dashboard/console",
		Status:          "normal",
		AccessMode:      "platform_admin",
		Meta:            map[string]interface{}{"demo": true},
	}); err != nil {
		return err
	}

	records, err := s.spaceService.ListSpaces()
	if err != nil {
		return err
	}
	for _, item := range records {
		if normalizeSpaceKey(item.SpaceKey) != s.spaceKey {
			continue
		}
		if item.MenuCount > 0 || item.PageCount > 0 {
			return nil
		}
		_, err = s.spaceService.InitializeFromDefault(s.spaceKey, false, &s.platformAdmin.ID)
		return err
	}
	return nil
}

func (s *demoInitializer) ensureDemoMessagingAssets() error {
	platformSender, err := s.ensureMessageSender("platform", nil, "平台", "平台默认发送人", true)
	if err != nil {
		return err
	}
	s.platformSender = *platformSender

	platformManagerSender, err := s.ensureMessageSender("platform", nil, "平台管理", "平台治理与系统消息发送身份", false)
	if err != nil {
		return err
	}
	s.platformManagerSender = *platformManagerSender

	teamSender, err := s.ensureMessageSender("tenant", &s.team.ID, "团队", "团队默认发送人", true)
	if err != nil {
		return err
	}
	s.teamSender = *teamSender

	teamManagerSender, err := s.ensureMessageSender("tenant", &s.team.ID, "团队管理", "团队管理员发送身份", false)
	if err != nil {
		return err
	}
	s.teamManagerSender = *teamManagerSender

	platformTemplate, err := s.ensureMessageTemplate(systemmodels.MessageTemplate{
		TemplateKey:     "demo.wrapup.platform.notice",
		Name:            "系统收尾平台演示模板",
		Description:     "用于平台管理员回归消息中心、发送记录和内部链接跳转。",
		MessageType:     "notice",
		OwnerScope:      "platform",
		AudienceType:    "specified_users",
		TitleTemplate:   "{{title}}",
		SummaryTemplate: "{{summary}}",
		ContentTemplate: "{{content}}",
		ActionType:      "none",
		Status:          "normal",
		Meta:            systemmodels.MetaJSON{"demo": true},
	})
	if err != nil {
		return err
	}
	s.platformTemplate = *platformTemplate

	teamTemplate, err := s.ensureMessageTemplate(systemmodels.MessageTemplate{
		TemplateKey:     "demo.wrapup.team.notice",
		Name:            "系统收尾团队演示模板",
		Description:     "用于团队管理员回归团队消息链路。",
		MessageType:     "message",
		OwnerScope:      "tenant",
		OwnerTenantID:   &s.team.ID,
		AudienceType:    "recipient_group",
		TitleTemplate:   "{{title}}",
		SummaryTemplate: "{{summary}}",
		ContentTemplate: "{{content}}",
		ActionType:      "none",
		Status:          "normal",
		Meta:            systemmodels.MetaJSON{"demo": true},
	})
	if err != nil {
		return err
	}
	s.teamTemplate = *teamTemplate

	platformGroup, err := s.ensureRecipientGroup("platform", nil, "演示平台接收组", "平台回归用接收组，包含指定用户与角色规则示例。")
	if err != nil {
		return err
	}
	s.platformGroup = *platformGroup
	if err := s.replaceRecipientGroupTargets(platformGroup.ID, []systemmodels.MessageRecipientGroupTarget{
		{TargetType: "user", UserID: &s.teamAdmin.ID, SortOrder: 1, Meta: systemmodels.MetaJSON{"demo": true}},
		{TargetType: "user", UserID: &s.member.ID, SortOrder: 2, Meta: systemmodels.MetaJSON{"demo": true}},
		{TargetType: "role", RoleCode: "admin", SortOrder: 3, Meta: systemmodels.MetaJSON{"demo": true}},
	}); err != nil {
		return err
	}

	teamGroup, err := s.ensureRecipientGroup("tenant", &s.team.ID, "演示团队接收组", "团队回归用接收组，包含团队成员、角色和功能包规则示例。")
	if err != nil {
		return err
	}
	s.teamGroup = *teamGroup
	if err := s.replaceRecipientGroupTargets(teamGroup.ID, []systemmodels.MessageRecipientGroupTarget{
		{TargetType: "tenant_users", TenantID: &s.team.ID, SortOrder: 1, Meta: systemmodels.MetaJSON{"demo": true}},
		{TargetType: "role", RoleCode: "team_admin", SortOrder: 2, Meta: systemmodels.MetaJSON{"demo": true}},
		{TargetType: "feature_package", PackageKey: "team.member_admin", SortOrder: 3, Meta: systemmodels.MetaJSON{"demo": true}},
	}); err != nil {
		return err
	}

	return nil
}

func (s *demoInitializer) ensureMessageSender(scopeType string, scopeID *uuid.UUID, name, description string, isDefault bool) (*systemmodels.MessageSender, error) {
	var sender systemmodels.MessageSender
	query := s.db.Where("scope_type = ? AND name = ? AND deleted_at IS NULL", scopeType, strings.TrimSpace(name))
	if scopeID != nil {
		query = query.Where("scope_id = ?", *scopeID)
	} else {
		query = query.Where("scope_id IS NULL")
	}
	err := query.First(&sender).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		sender = systemmodels.MessageSender{
			ScopeType:   scopeType,
			ScopeID:     scopeID,
			Name:        strings.TrimSpace(name),
			Description: strings.TrimSpace(description),
			IsDefault:   isDefault,
			Status:      "normal",
			Meta:        systemmodels.MetaJSON{"demo": true},
		}
		if err := s.db.Create(&sender).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Model(&sender).Updates(map[string]interface{}{
			"description": strings.TrimSpace(description),
			"is_default":  isDefault,
			"status":      "normal",
			"meta":        systemmodels.MetaJSON{"demo": true},
		}).Error; err != nil {
			return nil, err
		}
		sender.Description = strings.TrimSpace(description)
		sender.IsDefault = isDefault
		sender.Status = "normal"
		sender.Meta = systemmodels.MetaJSON{"demo": true}
	}
	if isDefault {
		scopeQuery := s.db.Model(&systemmodels.MessageSender{}).
			Where("id <> ? AND scope_type = ? AND deleted_at IS NULL", sender.ID, scopeType)
		if scopeID != nil {
			scopeQuery = scopeQuery.Where("scope_id = ?", *scopeID)
		} else {
			scopeQuery = scopeQuery.Where("scope_id IS NULL")
		}
		if err := scopeQuery.Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}
	return &sender, nil
}

func (s *demoInitializer) ensureMessageTemplate(template systemmodels.MessageTemplate) (*systemmodels.MessageTemplate, error) {
	var record systemmodels.MessageTemplate
	err := s.db.Where("template_key = ?", template.TemplateKey).First(&record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.db.Create(&template).Error; err != nil {
			return nil, err
		}
		return &template, nil
	}
	updates := map[string]interface{}{
		"name":                   template.Name,
		"description":            template.Description,
		"message_type":           template.MessageType,
		"owner_scope":            template.OwnerScope,
		"owner_tenant_id":        template.OwnerTenantID,
		"audience_type":          template.AudienceType,
		"title_template":         template.TitleTemplate,
		"summary_template":       template.SummaryTemplate,
		"content_template":       template.ContentTemplate,
		"action_type":            template.ActionType,
		"action_target_template": template.ActionTargetTemplate,
		"status":                 template.Status,
		"meta":                   template.Meta,
	}
	if err := s.db.Model(&record).Updates(updates).Error; err != nil {
		return nil, err
	}
	for key, value := range updates {
		switch key {
		case "name":
			record.Name = value.(string)
		case "description":
			record.Description = value.(string)
		case "message_type":
			record.MessageType = value.(string)
		case "owner_scope":
			record.OwnerScope = value.(string)
		case "owner_tenant_id":
			record.OwnerTenantID = template.OwnerTenantID
		case "audience_type":
			record.AudienceType = value.(string)
		case "title_template":
			record.TitleTemplate = value.(string)
		case "summary_template":
			record.SummaryTemplate = value.(string)
		case "content_template":
			record.ContentTemplate = value.(string)
		case "action_type":
			record.ActionType = value.(string)
		case "action_target_template":
			record.ActionTargetTemplate = value.(string)
		case "status":
			record.Status = value.(string)
		case "meta":
			record.Meta = value.(systemmodels.MetaJSON)
		}
	}
	return &record, nil
}

func (s *demoInitializer) ensureRecipientGroup(scopeType string, scopeID *uuid.UUID, name, description string) (*systemmodels.MessageRecipientGroup, error) {
	var group systemmodels.MessageRecipientGroup
	query := s.db.Where("scope_type = ? AND name = ? AND deleted_at IS NULL", scopeType, strings.TrimSpace(name))
	if scopeID != nil {
		query = query.Where("scope_id = ?", *scopeID)
	} else {
		query = query.Where("scope_id IS NULL")
	}
	err := query.First(&group).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		group = systemmodels.MessageRecipientGroup{
			ScopeType:   scopeType,
			ScopeID:     scopeID,
			Name:        strings.TrimSpace(name),
			Description: strings.TrimSpace(description),
			MatchMode:   "manual",
			Status:      "normal",
			Meta:        systemmodels.MetaJSON{"demo": true},
		}
		if err := s.db.Create(&group).Error; err != nil {
			return nil, err
		}
		return &group, nil
	}
	if err := s.db.Model(&group).Updates(map[string]interface{}{
		"description": strings.TrimSpace(description),
		"match_mode":  "manual",
		"status":      "normal",
		"meta":        systemmodels.MetaJSON{"demo": true},
	}).Error; err != nil {
		return nil, err
	}
	group.Description = strings.TrimSpace(description)
	group.MatchMode = "manual"
	group.Status = "normal"
	group.Meta = systemmodels.MetaJSON{"demo": true}
	return &group, nil
}

func (s *demoInitializer) replaceRecipientGroupTargets(groupID uuid.UUID, targets []systemmodels.MessageRecipientGroupTarget) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&systemmodels.MessageRecipientGroupTarget{}).Error; err != nil {
			return err
		}
		if len(targets) == 0 {
			return nil
		}
		items := make([]systemmodels.MessageRecipientGroupTarget, 0, len(targets))
		for _, item := range targets {
			item.GroupID = groupID
			if item.Meta == nil {
				item.Meta = systemmodels.MetaJSON{"demo": true}
			}
			items = append(items, item)
		}
		return tx.Create(&items).Error
	})
}

func (s *demoInitializer) seedDemoMessages() error {
	platformBizType := "demo.system_wrapup.platform"
	teamBizType := "demo.system_wrapup.team"
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := deleteMessagesByBizTypes(tx, []string{platformBizType, teamBizType}); err != nil {
			return err
		}

		now := time.Now()
		platformMessage := systemmodels.Message{
			MessageType:          "notice",
			BizType:              platformBizType,
			ScopeType:            "platform",
			SenderID:             &s.platformManagerSender.ID,
			SenderType:           "system",
			SenderUserID:         &s.platformAdmin.ID,
			SenderNameSnapshot:   s.platformManagerSender.Name,
			SenderAvatarSnapshot: s.platformManagerSender.AvatarURL,
			SenderServiceKey:     "demo_initializer",
			AudienceType:         "specified_users",
			AudienceScope:        "platform",
			TargetUserIDs:        []string{s.platformAdmin.ID.String(), s.member.ID.String()},
			TemplateID:           &s.platformTemplate.ID,
			Title:                "平台收尾演示消息",
			Summary:              "用于验证系统页、消息中心、发送记录和内部链接跳转。",
			Content:              `<p>这是一条平台侧演示消息。<a href="#/system/page">打开页面管理</a>，<a href="/system/menu">打开菜单管理</a>，也可以 <a href="#/workspace/inbox">回到消息中心</a>。</p>`,
			Priority:             "normal",
			ActionType:           "none",
			ActionTarget:         "",
			Status:               "published",
			PublishedAt:          &now,
			Meta:                 systemmodels.MetaJSON{"demo": true, "demo_scope": "platform"},
		}
		if err := tx.Create(&platformMessage).Error; err != nil {
			return err
		}
		platformDeliveries := []systemmodels.MessageDelivery{
			{
				MessageID:       platformMessage.ID,
				RecipientUserID: s.platformAdmin.ID,
				BoxType:         "notice",
				DeliveryStatus:  "unread",
				LastActionAt:    &now,
				Meta: systemmodels.MetaJSON{
					"source_rule_type":   "specified_users",
					"source_rule_label":  "指定用户",
					"source_target_id":   s.platformAdmin.ID.String(),
					"source_target_type": "user",
					"source_target_value": func() string {
						if strings.TrimSpace(s.platformAdmin.Nickname) != "" {
							return s.platformAdmin.Nickname
						}
						return s.platformAdmin.Username
					}(),
				},
			},
			{
				MessageID:       platformMessage.ID,
				RecipientUserID: s.member.ID,
				BoxType:         "notice",
				DeliveryStatus:  "unread",
				LastActionAt:    &now,
				Meta: systemmodels.MetaJSON{
					"source_rule_type":   "specified_users",
					"source_rule_label":  "指定用户",
					"source_target_id":   s.member.ID.String(),
					"source_target_type": "user",
					"source_target_value": func() string {
						if strings.TrimSpace(s.member.Nickname) != "" {
							return s.member.Nickname
						}
						return s.member.Username
					}(),
				},
			},
		}
		if err := tx.Create(&platformDeliveries).Error; err != nil {
			return err
		}

		teamMessage := systemmodels.Message{
			MessageType:          "message",
			BizType:              teamBizType,
			ScopeType:            "tenant",
			ScopeID:              &s.team.ID,
			SenderID:             &s.teamManagerSender.ID,
			SenderType:           "team_user",
			SenderUserID:         &s.teamAdmin.ID,
			SenderNameSnapshot:   s.teamManagerSender.Name,
			SenderAvatarSnapshot: s.teamManagerSender.AvatarURL,
			SenderServiceKey:     "demo_initializer",
			AudienceType:         "recipient_group",
			AudienceScope:        "tenant",
			TargetTenantID:       &s.team.ID,
			TargetGroupIDs:       []string{s.teamGroup.ID.String()},
			TemplateID:           &s.teamTemplate.ID,
			Title:                "团队收尾演示消息",
			Summary:              "团队管理员可在消息中心看到这条团队演示消息。",
			Content:              `<p>这是一条团队侧演示消息。<a href="/team/message">打开团队消息发送</a>，<a href="#/workspace/inbox">打开消息中心</a>。</p>`,
			Priority:             "normal",
			ActionType:           "none",
			ActionTarget:         "",
			Status:               "published",
			PublishedAt:          &now,
			Meta:                 systemmodels.MetaJSON{"demo": true, "demo_scope": "team", "source_group_name": s.teamGroup.Name},
		}
		if err := tx.Create(&teamMessage).Error; err != nil {
			return err
		}
		deliveries := []systemmodels.MessageDelivery{
			{
				MessageID:       teamMessage.ID,
				RecipientUserID: s.teamAdmin.ID,
				RecipientTeamID: &s.team.ID,
				BoxType:         "message",
				DeliveryStatus:  "unread",
				LastActionAt:    &now,
				Meta: systemmodels.MetaJSON{
					"source_group_id":     s.teamGroup.ID.String(),
					"source_group_name":   s.teamGroup.Name,
					"source_rule_type":    "tenant_admins",
					"source_rule_label":   "团队管理员",
					"source_target_type":  "tenant_admins",
					"source_target_value": s.team.ID.String(),
				},
			},
		}
		return tx.Create(&deliveries).Error
	})
}

func deleteMessagesByBizTypes(tx *gorm.DB, bizTypes []string) error {
	if len(bizTypes) == 0 {
		return nil
	}
	var ids []uuid.UUID
	if err := tx.Model(&systemmodels.Message{}).Where("biz_type IN ?", bizTypes).Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	if err := tx.Where("message_id IN ?", ids).Delete(&systemmodels.MessageDelivery{}).Error; err != nil {
		return err
	}
	return tx.Where("id IN ?", ids).Delete(&systemmodels.Message{}).Error
}

func (s *demoInitializer) findFeaturePackageID(packageKey string) (uuid.UUID, error) {
	var pkg systemmodels.FeaturePackage
	if err := s.db.Where("package_key = ? AND deleted_at IS NULL", strings.TrimSpace(packageKey)).First(&pkg).Error; err != nil {
		return uuid.Nil, err
	}
	return pkg.ID, nil
}

func normalizeSpaceKey(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return defaultDemoSpaceKey
	}
	return target
}
