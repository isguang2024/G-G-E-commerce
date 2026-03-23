package featurepackage

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

var (
	ErrFeaturePackageNotFound = errors.New("feature package not found")
	ErrFeaturePackageExists   = errors.New("feature package already exists")
)

type Service interface {
	List(req *dto.FeaturePackageListRequest) ([]user.FeaturePackage, int64, error)
	GetPackageStats(packageIDs []uuid.UUID) (map[uuid.UUID]int64, map[uuid.UUID]int64, error)
	Get(id uuid.UUID) (*user.FeaturePackage, error)
	Create(req *dto.FeaturePackageCreateRequest) (*user.FeaturePackage, error)
	Update(id uuid.UUID, req *dto.FeaturePackageUpdateRequest) error
	Delete(id uuid.UUID) error
	GetPackageActions(id uuid.UUID) ([]uuid.UUID, []user.PermissionAction, error)
	SetPackageActions(id uuid.UUID, actionIDs []uuid.UUID) error
	GetPackageTeams(id uuid.UUID) ([]uuid.UUID, error)
	SetPackageTeams(id uuid.UUID, teamIDs []uuid.UUID, grantedBy *uuid.UUID) error
	GetTeamPackages(teamID uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error)
	SetTeamPackages(teamID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
}

type service struct {
	packageRepo       user.FeaturePackageRepository
	packageActionRepo user.FeaturePackageActionRepository
	teamPackageRepo   user.TeamFeaturePackageRepository
	actionRepo        user.PermissionActionRepository
	tenantActionRepo  user.TenantActionPermissionRepository
	manualActionRepo  user.TeamManualActionPermissionRepository
	tenantRepo        user.TenantRepository
}

func NewService(
	packageRepo user.FeaturePackageRepository,
	packageActionRepo user.FeaturePackageActionRepository,
	teamPackageRepo user.TeamFeaturePackageRepository,
	actionRepo user.PermissionActionRepository,
	tenantActionRepo user.TenantActionPermissionRepository,
	manualActionRepo user.TeamManualActionPermissionRepository,
	tenantRepo user.TenantRepository,
) Service {
	return &service{
		packageRepo:       packageRepo,
		packageActionRepo: packageActionRepo,
		teamPackageRepo:   teamPackageRepo,
		actionRepo:        actionRepo,
		tenantActionRepo:  tenantActionRepo,
		manualActionRepo:  manualActionRepo,
		tenantRepo:        tenantRepo,
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
		Name:        strings.TrimSpace(req.Name),
		ContextType: normalizeContextType(req.ContextType),
		Status:      strings.TrimSpace(req.Status),
	})
}

func (s *service) GetPackageStats(packageIDs []uuid.UUID) (map[uuid.UUID]int64, map[uuid.UUID]int64, error) {
	actionCounts, err := s.packageActionRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, err
	}
	teamCounts, err := s.teamPackageRepo.CountByPackageIDs(packageIDs)
	if err != nil {
		return nil, nil, err
	}
	return actionCounts, teamCounts, nil
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
	if contextType := normalizeContextType(req.ContextType); contextType != "" {
		updates["context_type"] = contextType
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		updates["status"] = normalizeStatus(status)
	}
	return s.packageRepo.UpdateWithMap(id, updates)
}

func (s *service) Delete(id uuid.UUID) error {
	if _, err := s.packageRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
	}
	if err := s.packageActionRepo.DeleteByPackageID(id); err != nil {
		return err
	}
	if err := s.teamPackageRepo.DeleteByPackageID(id); err != nil {
		return err
	}
	return s.packageRepo.Delete(id)
}

func (s *service) GetPackageActions(id uuid.UUID) ([]uuid.UUID, []user.PermissionAction, error) {
	if _, err := s.packageRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrFeaturePackageNotFound
		}
		return nil, nil, err
	}
	actionIDs, err := s.packageActionRepo.GetActionIDsByPackageID(id)
	if err != nil {
		return nil, nil, err
	}
	actions, err := s.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		return nil, nil, err
	}
	return actionIDs, actions, nil
}

func (s *service) SetPackageActions(id uuid.UUID, actionIDs []uuid.UUID) error {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeaturePackageNotFound
		}
		return err
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
			if action.ContextType != "" && item.ContextType != "" && action.ContextType != item.ContextType {
				return errors.New("功能包上下文与功能权限上下文不一致")
			}
		}
	}
	if err := s.packageActionRepo.ReplacePackageActions(id, actionIDs); err != nil {
		return err
	}
	teamIDs, err := s.teamPackageRepo.GetTeamIDsByPackageID(id)
	if err != nil {
		return err
	}
	for _, teamID := range teamIDs {
		if err := s.refreshTeamActions(teamID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) GetTeamPackages(teamID uuid.UUID) ([]uuid.UUID, []user.FeaturePackage, error) {
	packageIDs, err := s.teamPackageRepo.GetPackageIDsByTeamID(teamID)
	if err != nil {
		return nil, nil, err
	}
	items := make([]user.FeaturePackage, 0, len(packageIDs))
	for _, packageID := range packageIDs {
		item, getErr := s.packageRepo.GetByID(packageID)
		if getErr != nil {
			if errors.Is(getErr, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, nil, getErr
		}
		items = append(items, *item)
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
	if item.ContextType != "team" {
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
	if item.ContextType != "team" {
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
	for _, teamID := range teamIDs {
		if _, err := s.tenantRepo.GetByID(teamID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("存在无效的团队")
			}
			return err
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
		if err := s.refreshTeamActions(teamID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) SetTeamPackages(teamID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	for _, packageID := range packageIDs {
		item, err := s.packageRepo.GetByID(packageID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrFeaturePackageNotFound
			}
			return err
		}
		if item.ContextType != "team" {
			return errors.New("仅支持为团队分配团队功能包")
		}
	}
	if err := s.teamPackageRepo.ReplaceTeamPackages(teamID, packageIDs, grantedBy); err != nil {
		return err
	}
	return s.refreshTeamActions(teamID)
}

func (s *service) refreshTeamActions(teamID uuid.UUID) error {
	packageIDs, err := s.teamPackageRepo.GetPackageIDsByTeamID(teamID)
	if err != nil {
		return err
	}
	allActionIDs := make([]uuid.UUID, 0)
	seenActionIDs := make(map[uuid.UUID]struct{})
	for _, packageID := range packageIDs {
		actionIDs, err := s.packageActionRepo.GetActionIDsByPackageID(packageID)
		if err != nil {
			return err
		}
		for _, actionID := range actionIDs {
			if _, ok := seenActionIDs[actionID]; ok {
				continue
			}
			seenActionIDs[actionID] = struct{}{}
			allActionIDs = append(allActionIDs, actionID)
		}
	}
	manualActionIDs, err := s.manualActionRepo.GetEnabledActionIDsByTenantID(teamID)
	if err != nil {
		return err
	}
	for _, actionID := range manualActionIDs {
		if _, ok := seenActionIDs[actionID]; ok {
			continue
		}
		seenActionIDs[actionID] = struct{}{}
		allActionIDs = append(allActionIDs, actionID)
	}
	return s.tenantActionRepo.ReplaceTenantActions(teamID, allActionIDs)
}

func normalizeContextType(value string) string {
	switch strings.TrimSpace(value) {
	case "platform", "team":
		return strings.TrimSpace(value)
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
