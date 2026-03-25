package featurepackage

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

var (
	ErrFeaturePackageNotFound = errors.New("feature package not found")
	ErrFeaturePackageExists   = errors.New("feature package already exists")
	ErrFeaturePackageBuiltin  = errors.New("feature package is builtin")
)

type Service interface {
	List(req *dto.FeaturePackageListRequest) ([]user.FeaturePackage, int64, error)
	GetPackageStats(packageIDs []uuid.UUID) (map[uuid.UUID]int64, map[uuid.UUID]int64, map[uuid.UUID]int64, error)
	Get(id uuid.UUID) (*user.FeaturePackage, error)
	Create(req *dto.FeaturePackageCreateRequest) (*user.FeaturePackage, error)
	Update(id uuid.UUID, req *dto.FeaturePackageUpdateRequest) error
	Delete(id uuid.UUID) error
	GetPackageChildren(id uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error)
	SetPackageChildren(id uuid.UUID, childPackageIDs []uuid.UUID) error
	GetPackageKeys(id uuid.UUID) ([]uuid.UUID, []user.PermissionKey, error)
	SetPackageKeys(id uuid.UUID, actionIDs []uuid.UUID) error
	GetPackageMenus(id uuid.UUID) ([]uuid.UUID, []user.Menu, error)
	SetPackageMenus(id uuid.UUID, menuIDs []uuid.UUID) error
	GetPackageTeams(id uuid.UUID) ([]uuid.UUID, error)
	SetPackageTeams(id uuid.UUID, teamIDs []uuid.UUID, grantedBy *uuid.UUID) error
	GetTeamPackages(teamID uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error)
	SetTeamPackages(teamID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
}

type service struct {
	packageRepo       user.FeaturePackageRepository
	packageBundleRepo user.FeaturePackageBundleRepository
	packageActionRepo user.FeaturePackageKeyRepository
	packageMenuRepo   user.FeaturePackageMenuRepository
	teamPackageRepo   user.TeamFeaturePackageRepository
	rolePackageRepo   user.RoleFeaturePackageRepository
	actionRepo        user.PermissionKeyRepository
	menuRepo          user.MenuRepository
	tenantRepo        user.TenantRepository
	boundaryService   teamboundary.Service
	refresher         permissionrefresh.Service
}

func NewService(
	packageRepo user.FeaturePackageRepository,
	packageBundleRepo user.FeaturePackageBundleRepository,
	packageActionRepo user.FeaturePackageKeyRepository,
	packageMenuRepo user.FeaturePackageMenuRepository,
	teamPackageRepo user.TeamFeaturePackageRepository,
	rolePackageRepo user.RoleFeaturePackageRepository,
	actionRepo user.PermissionKeyRepository,
	menuRepo user.MenuRepository,
	tenantRepo user.TenantRepository,
	boundaryService teamboundary.Service,
	refresher permissionrefresh.Service,
) Service {
	return &service{
		packageRepo:       packageRepo,
		packageBundleRepo: packageBundleRepo,
		packageActionRepo: packageActionRepo,
		packageMenuRepo:   packageMenuRepo,
		teamPackageRepo:   teamPackageRepo,
		rolePackageRepo:   rolePackageRepo,
		actionRepo:        actionRepo,
		menuRepo:          menuRepo,
		tenantRepo:        tenantRepo,
		boundaryService:   boundaryService,
		refresher:         refresher,
	}
}

func (s *service) List(req *dto.FeaturePackageListRequest) ([]user.FeaturePackage, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	return s.packageRepo.List((req.Current-1)*req.Size, req.Size, &user.FeaturePackageListParams{
		Keyword:     strings.TrimSpace(req.Keyword),
		PackageKey:  strings.TrimSpace(req.PackageKey),
		PackageType: normalizePackageType(req.PackageType),
		Name:        strings.TrimSpace(req.Name),
		ContextType: normalizeContextType(req.ContextType),
		Status:      strings.TrimSpace(req.Status),
	})
}

func (s *service) GetPackageStats(packageIDs []uuid.UUID) (map[uuid.UUID]int64, map[uuid.UUID]int64, map[uuid.UUID]int64, error) {
	actionCounts, err := s.packageActionRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	menuCounts, err := s.packageMenuRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	teamCounts, err := s.teamPackageRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	return actionCounts, menuCounts, teamCounts, nil
}

func (s *service) Get(id uuid.UUID) (*user.FeaturePackage, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *service) Create(req *dto.FeaturePackageCreateRequest) (*user.FeaturePackage, error) {
	packageKey := strings.TrimSpace(req.PackageKey)
	if packageKey == "" {
		return nil, errors.New("package_key 不能为空")
	}
	if _, err := s.packageRepo.GetByPackageKey(packageKey); err == nil {
		return nil, ErrFeaturePackageExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	item := &user.FeaturePackage{
		PackageKey:  packageKey,
		PackageType: normalizePackageTypeDefault(req.PackageType, "base"),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		ContextType: normalizeContextTypeDefault(req.ContextType, "team"),
		Status:      normalizeStatus(req.Status),
		SortOrder:   req.SortOrder,
	}
	if err := s.packageRepo.Create(item); err != nil {
		return nil, err
	}
	return s.packageRepo.GetByID(item.ID)
}

func (s *service) Update(id uuid.UUID, req *dto.FeaturePackageUpdateRequest) error {
	current, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
	}
	updates := map[string]interface{}{
		"updated_at": time.Now(),
		"sort_order": req.SortOrder,
	}
	if packageKey := strings.TrimSpace(req.PackageKey); packageKey != "" && packageKey != current.PackageKey {
		existing, getErr := s.packageRepo.GetByPackageKey(packageKey)
		if getErr == nil && existing != nil && existing.ID != id {
			return ErrFeaturePackageExists
		}
		if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return getErr
		}
		updates["package_key"] = packageKey
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		updates["name"] = name
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if packageType := normalizePackageType(req.PackageType); packageType != "" {
		if packageType != current.PackageType {
			if packageType == "bundle" {
				actionIDs, getActionErr := s.packageActionRepo.GetKeyIDsByPackageID(id)
				if getActionErr != nil {
					return getActionErr
				}
				if len(actionIDs) > 0 {
					return errors.New("存在已绑定功能权限，请先清空后再改为组合包")
				}
				menuIDs, getMenuErr := s.packageMenuRepo.GetMenuIDsByPackageID(id)
				if getMenuErr != nil {
					return getMenuErr
				}
				if len(menuIDs) > 0 {
					return errors.New("存在已绑定菜单，请先清空后再改为组合包")
				}
			}
			if packageType == "base" {
				childPackageIDs, childErr := s.packageBundleRepo.GetChildPackageIDs(id)
				if childErr != nil {
					return childErr
				}
				if len(childPackageIDs) > 0 {
					return errors.New("组合包仍包含基础包，请先清空组合关系后再改为基础包")
				}
			}
		}
		updates["package_type"] = packageType
	}
	if contextType := normalizeContextType(req.ContextType); contextType != "" {
		updates["context_type"] = contextType
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		updates["status"] = normalizeStatus(status)
	}
	if err := s.packageRepo.UpdateWithMap(id, updates); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshByPackage(id)
	}
	return nil
}

func (s *service) Delete(id uuid.UUID) error {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
	}
	if item.IsBuiltin {
		return ErrFeaturePackageBuiltin
	}
	teamIDs, err := s.teamPackageRepo.GetTeamIDsByPackageID(id)
	if err != nil {
		return err
	}
	parentPackageIDs, err := s.packageBundleRepo.GetParentPackageIDs(id)
	if err != nil {
		return err
	}
	if err := s.packageActionRepo.DeleteByPackageID(id); err != nil {
		return err
	}
	if err := s.packageMenuRepo.DeleteByPackageID(id); err != nil {
		return err
	}
	if err := s.teamPackageRepo.DeleteByPackageID(id); err != nil {
		return err
	}
	if err := s.rolePackageRepo.DeleteByPackageID(id); err != nil {
		return err
	}
	if err := s.packageBundleRepo.DeleteByPackageID(id); err != nil {
		return err
	}
	if err := s.packageBundleRepo.DeleteByChildPackageID(id); err != nil {
		return err
	}
	if err := s.packageRepo.Delete(id); err != nil {
		return err
	}
	if s.refresher != nil {
		for _, packageID := range parentPackageIDs {
			if refreshErr := s.refresher.RefreshByPackage(packageID); refreshErr != nil {
				return refreshErr
			}
		}
		return s.refresher.RefreshTeams(teamIDs)
	}
	for _, teamID := range teamIDs {
		if _, err := s.boundaryService.RefreshSnapshot(teamID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) GetPackageChildren(id uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrFeaturePackageNotFound
		}
		return nil, nil, err
	}
	if item.PackageType != "bundle" {
		return []uuid.UUID{}, []user.FeaturePackage{}, nil
	}
	childPackageIDs, err := s.packageBundleRepo.GetChildPackageIDs(id)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.packageRepo.GetByIDs(childPackageIDs)
	if err != nil {
		return nil, nil, err
	}
	return childPackageIDs, items, nil
}

func (s *service) SetPackageChildren(id uuid.UUID, childPackageIDs []uuid.UUID) error {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
	}
	if item.PackageType != "bundle" {
		return errors.New("仅组合包允许配置基础包集合")
	}
	childMap, err := s.getPackageMap(childPackageIDs)
	if err != nil {
		return err
	}
	for _, childPackageID := range childPackageIDs {
		if childPackageID == id {
			return errors.New("组合包不能包含自己")
		}
		child, ok := childMap[childPackageID]
		if !ok {
			return ErrFeaturePackageNotFound
		}
		if child.PackageType != "base" {
			return errors.New("组合包只能包含基础包")
		}
		if !contextSupportsChildPackage(item.ContextType, child.ContextType) {
			return errors.New("组合包上下文与基础包上下文不兼容")
		}
	}
	if err := s.packageBundleRepo.ReplaceChildPackages(id, childPackageIDs); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshByPackage(id)
	}
	return nil
}

func (s *service) GetPackageKeys(id uuid.UUID) ([]uuid.UUID, []user.PermissionKey, error) {
	if _, err := s.packageRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrFeaturePackageNotFound
		}
		return nil, nil, err
	}
	actionIDs, err := s.packageActionRepo.GetKeyIDsByPackageID(id)
	if err != nil {
		return nil, nil, err
	}
	actions, err := s.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		return nil, nil, err
	}
	return actionIDs, actions, nil
}

func (s *service) SetPackageKeys(id uuid.UUID, actionIDs []uuid.UUID) error {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
	}
	if item.PackageType == "bundle" {
		return errors.New("组合包不允许直接绑定功能权限")
	}
	if len(actionIDs) > 0 {
		actions, getErr := s.actionRepo.GetByIDs(actionIDs)
		if getErr != nil {
			return getErr
		}
		if len(actions) != len(actionIDs) {
			return errors.New("存在无效的功能权限")
		}
		for _, action := range actions {
			if action.ContextType != "" && item.ContextType != "" && !contextSupportsAction(item.ContextType, action.ContextType) {
				return errors.New("功能包上下文与功能权限上下文不一致")
			}
		}
	}
	if err := s.packageActionRepo.ReplacePackageKeys(id, actionIDs); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshByPackage(id)
	}
	return nil
}

func (s *service) GetPackageMenus(id uuid.UUID) ([]uuid.UUID, []user.Menu, error) {
	if _, err := s.packageRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrFeaturePackageNotFound
		}
		return nil, nil, err
	}
	menuIDs, err := s.packageMenuRepo.GetMenuIDsByPackageID(id)
	if err != nil {
		return nil, nil, err
	}
	menus, err := s.menuRepo.GetByIDs(menuIDs)
	if err != nil {
		return nil, nil, err
	}
	return menuIDs, menus, nil
}

func (s *service) SetPackageMenus(id uuid.UUID, menuIDs []uuid.UUID) error {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
	}
	if item.PackageType == "bundle" {
		return errors.New("组合包不允许直接绑定菜单")
	}
	if item.ContextType == "" {
		return errors.New("功能包上下文无效")
	}
	if len(menuIDs) > 0 {
		menus, getErr := s.menuRepo.GetByIDs(menuIDs)
		if getErr != nil {
			return getErr
		}
		if len(menus) != len(menuIDs) {
			return errors.New("存在无效的菜单")
		}
	}
	if err := s.packageMenuRepo.ReplacePackageMenus(id, menuIDs); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshByPackage(id)
	}
	return nil
}

func (s *service) GetTeamPackages(teamID uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error) {
	packageIDs, err := s.teamPackageRepo.GetPackageIDsByTeamID(teamID)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.packageRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, nil, err
	}
	return packageIDs, items, nil
}

func (s *service) GetPackageTeams(id uuid.UUID) ([]uuid.UUID, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	if !supportsTeamContext(item.ContextType) {
		return []uuid.UUID{}, nil
	}
	return s.teamPackageRepo.GetTeamIDsByPackageID(id)
}

func (s *service) SetPackageTeams(id uuid.UUID, teamIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
	}
	if !supportsTeamContext(item.ContextType) {
		return errors.New("仅支持为团队功能包配置团队")
	}
	currentTeamIDs, err := s.teamPackageRepo.GetTeamIDsByPackageID(id)
	if err != nil {
		return err
	}
	desired := make(map[uuid.UUID]struct{}, len(teamIDs))
	affected := make(map[uuid.UUID]struct{}, len(currentTeamIDs)+len(teamIDs))
	for _, teamID := range currentTeamIDs {
		affected[teamID] = struct{}{}
	}
	teamMap, err := s.getTeamMap(teamIDs)
	if err != nil {
		return err
	}
	for _, teamID := range teamIDs {
		if _, ok := teamMap[teamID]; !ok {
			return errors.New("存在无效的团队")
		}
		desired[teamID] = struct{}{}
		affected[teamID] = struct{}{}
	}
	for teamID := range affected {
		packageIDs, err := s.teamPackageRepo.GetPackageIDsByTeamID(teamID)
		if err != nil {
			return err
		}
		nextPackageIDs := make([]uuid.UUID, 0, len(packageIDs)+1)
		seen := make(map[uuid.UUID]struct{}, len(packageIDs)+1)
		for _, packageID := range packageIDs {
			if packageID == id {
				continue
			}
			if _, ok := seen[packageID]; ok {
				continue
			}
			seen[packageID] = struct{}{}
			nextPackageIDs = append(nextPackageIDs, packageID)
		}
		if _, ok := desired[teamID]; ok {
			if _, exists := seen[id]; !exists {
				nextPackageIDs = append(nextPackageIDs, id)
			}
		}
		if err := s.teamPackageRepo.ReplaceTeamPackages(teamID, nextPackageIDs, grantedBy); err != nil {
			return err
		}
		if s.refresher != nil {
			if err := s.refresher.RefreshTeam(teamID); err != nil {
				return err
			}
			continue
		}
		if _, err := s.boundaryService.RefreshSnapshot(teamID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) SetTeamPackages(teamID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	packageMap, err := s.getPackageMap(packageIDs)
	if err != nil {
		return err
	}
	for _, packageID := range packageIDs {
		item, ok := packageMap[packageID]
		if !ok {
			return ErrFeaturePackageNotFound
		}
		if !supportsTeamContext(item.ContextType) {
			return errors.New("仅支持为团队分配团队功能包")
		}
	}
	if err := s.teamPackageRepo.ReplaceTeamPackages(teamID, packageIDs, grantedBy); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshTeam(teamID)
	}
	_, err = s.boundaryService.RefreshSnapshot(teamID)
	return err
}

func (s *service) getPackageMap(packageIDs []uuid.UUID) (map[uuid.UUID]user.FeaturePackage, error) {
	items, err := s.packageRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]user.FeaturePackage, len(items))
	for _, item := range items {
		result[item.ID] = item
	}
	return result, nil
}

func (s *service) getTeamMap(teamIDs []uuid.UUID) (map[uuid.UUID]struct{}, error) {
	items, err := s.tenantRepo.GetByIDs(teamIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		result[item.ID] = struct{}{}
	}
	return result, nil
}

func normalizeContextType(value string) string {
	switch strings.ReplaceAll(strings.TrimSpace(value), " ", "") {
	case "platform", "team", "common":
		return strings.ReplaceAll(strings.TrimSpace(value), " ", "")
	default:
		return ""
	}
}

func normalizeContextTypeDefault(value, fallback string) string {
	if normalized := normalizeContextType(value); normalized != "" {
		return normalized
	}
	return fallback
}

func normalizeStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "disabled":
		return "disabled"
	default:
		return "normal"
	}
}

func normalizePackageType(value string) string {
	switch strings.TrimSpace(value) {
	case "base", "bundle":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizePackageTypeDefault(value, fallback string) string {
	if normalized := normalizePackageType(value); normalized != "" {
		return normalized
	}
	return fallback
}

func supportsTeamContext(contextType string) bool {
	return contextType == "team" || contextType == "common"
}

func contextSupportsAction(packageContextType, actionContextType string) bool {
	if packageContextType == "common" {
		return actionContextType == "platform" || actionContextType == "team"
	}
	return packageContextType == actionContextType
}

func contextSupportsChildPackage(bundleContextType, childContextType string) bool {
	switch bundleContextType {
	case "platform":
		return childContextType == "platform" || childContextType == "common"
	case "team":
		return childContextType == "team" || childContextType == "common"
	case "common":
		return childContextType == "platform" || childContextType == "team" || childContextType == "common"
	default:
		return false
	}
}
