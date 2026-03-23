package permission

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

var (
	ErrPermissionActionNotFound = errors.New("permission action not found")
	ErrPermissionActionExists   = errors.New("permission action already exists")
)

type PermissionService interface {
	List(req *dto.PermissionActionListRequest) ([]user.PermissionAction, int64, error)
	Get(id uuid.UUID) (*user.PermissionAction, error)
	Create(req *dto.PermissionActionCreateRequest) (*user.PermissionAction, error)
	Update(id uuid.UUID, req *dto.PermissionActionUpdateRequest) error
	Delete(id uuid.UUID) error
}

type permissionService struct {
	actionRepo             user.PermissionActionRepository
	packageActionRepo      user.FeaturePackageActionRepository
	teamPackageRepo        user.TeamFeaturePackageRepository
	roleDisabledActionRepo user.RoleDisabledActionRepository
	teamBlockedActionRepo  user.TeamBlockedActionRepository
	userActionRepo         user.UserActionPermissionRepository
	boundaryService        teamboundary.Service
	refresher              permissionrefresh.Service
}

func NewPermissionService(
	actionRepo user.PermissionActionRepository,
	packageActionRepo user.FeaturePackageActionRepository,
	teamPackageRepo user.TeamFeaturePackageRepository,
	roleDisabledActionRepo user.RoleDisabledActionRepository,
	teamBlockedActionRepo user.TeamBlockedActionRepository,
	userActionRepo user.UserActionPermissionRepository,
	boundaryService teamboundary.Service,
	refresher permissionrefresh.Service,
) PermissionService {
	return &permissionService{
		actionRepo:             actionRepo,
		packageActionRepo:      packageActionRepo,
		teamPackageRepo:        teamPackageRepo,
		roleDisabledActionRepo: roleDisabledActionRepo,
		teamBlockedActionRepo:  teamBlockedActionRepo,
		userActionRepo:         userActionRepo,
		boundaryService:        boundaryService,
		refresher:              refresher,
	}
}

func (s *permissionService) List(req *dto.PermissionActionListRequest) ([]user.PermissionAction, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	params := &user.PermissionActionListParams{
		Keyword:       strings.TrimSpace(req.Keyword),
		PermissionKey: strings.TrimSpace(req.PermissionKey),
		Name:          req.Name,
		ResourceCode:  req.ResourceCode,
		ActionCode:    req.ActionCode,
		ModuleCode:    strings.TrimSpace(req.ModuleCode),
		ContextType:   normalizeContextType(req.ContextType, ""),
		Source:        strings.TrimSpace(req.Source),
		FeatureKind:   normalizeFeatureKind(req.FeatureKind, ""),
		Status:        req.Status,
	}
	return s.actionRepo.List((req.Current-1)*req.Size, req.Size, params)
}

func (s *permissionService) Get(id uuid.UUID) (*user.PermissionAction, error) {
	action, err := s.actionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionActionNotFound
		}
		return nil, err
	}
	return action, nil
}

func (s *permissionService) Create(req *dto.PermissionActionCreateRequest) (*user.PermissionAction, error) {
	permissionKey := permissionkey.Normalize(req.PermissionKey)
	if permissionKey == "" {
		permissionKey = permissionkey.FromLegacy(strings.TrimSpace(req.ResourceCode), strings.TrimSpace(req.ActionCode)).Key
	}
	if permissionKey == "" {
		return nil, errors.New("permission_key 不能为空")
	}
	mapping := permissionkey.FromKey(permissionKey)
	resourceCode := strings.TrimSpace(req.ResourceCode)
	if resourceCode == "" {
		resourceCode = strings.TrimSpace(mapping.ResourceCode)
	}
	actionCode := strings.TrimSpace(req.ActionCode)
	if actionCode == "" {
		actionCode = strings.TrimSpace(mapping.ActionCode)
	}
	if resourceCode == "" || actionCode == "" {
		return nil, errors.New("无法根据 permission_key 推导兼容编码")
	}
	if _, err := s.actionRepo.GetByPermissionKey(permissionKey); err == nil {
		return nil, ErrPermissionActionExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	featureKind := normalizeFeatureKind(req.FeatureKind, "business")
	moduleCode := normalizeModuleCode(req.ModuleCode, req.ResourceCode)
	contextType := normalizeContextType(req.ContextType, deriveContextType(permissionKey, moduleCode))
	action := &user.PermissionAction{
		PermissionKey: permissionKey,
		ResourceCode:  resourceCode,
		ActionCode:    actionCode,
		ModuleCode:    moduleCode,
		ContextType:   contextType,
		Source:        "business",
		FeatureKind:   featureKind,
		Name:          strings.TrimSpace(req.Name),
		Description:   strings.TrimSpace(req.Description),
		Status:        status,
		SortOrder:     req.SortOrder,
	}
	if err := s.actionRepo.Create(action); err != nil {
		return nil, err
	}
	return s.actionRepo.GetByID(action.ID)
}

func (s *permissionService) Update(id uuid.UUID, req *dto.PermissionActionUpdateRequest) error {
	current, err := s.actionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionActionNotFound
		}
		return err
	}
	updates := map[string]interface{}{
		"updated_at": time.Now(),
		"sort_order": req.SortOrder,
	}
	if permissionKey := permissionkey.Normalize(req.PermissionKey); permissionKey != "" {
		updates["permission_key"] = permissionKey
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		updates["name"] = name
	}
	if featureKind := normalizeFeatureKind(req.FeatureKind, ""); featureKind != "" {
		updates["feature_kind"] = featureKind
	}
	if contextType := normalizeContextType(req.ContextType, ""); contextType != "" {
		updates["context_type"] = contextType
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		updates["status"] = status
	}
	resourceCode := strings.TrimSpace(req.ResourceCode)
	actionCode := strings.TrimSpace(req.ActionCode)
	targetResourceCode := current.ResourceCode
	targetActionCode := current.ActionCode
	if req.ModuleCode != "" {
		updates["module_code"] = normalizeModuleCode(req.ModuleCode, targetResourceCode)
	}
	targetPermissionKey := current.PermissionKey
	if permissionKey := permissionkey.Normalize(req.PermissionKey); permissionKey != "" {
		targetPermissionKey = permissionKey
	} else if targetPermissionKey == "" {
		targetPermissionKey = permissionkey.FromLegacy(targetResourceCode, targetActionCode).Key
		updates["permission_key"] = targetPermissionKey
	}
	if targetPermissionKey != "" {
		mapping := permissionkey.FromKey(targetPermissionKey)
		if resourceCode == "" {
			resourceCode = strings.TrimSpace(mapping.ResourceCode)
		}
		if actionCode == "" {
			actionCode = strings.TrimSpace(mapping.ActionCode)
		}
	}
	if resourceCode != "" {
		targetResourceCode = resourceCode
		updates["resource_code"] = resourceCode
	}
	if actionCode != "" {
		targetActionCode = actionCode
		updates["action_code"] = actionCode
	}
	if req.ModuleCode == "" && (resourceCode != "" || current.ModuleCode == "") {
		updates["module_code"] = normalizeModuleCode(current.ModuleCode, targetResourceCode)
	}
	if _, exists := updates["context_type"]; !exists && current.ContextType == "" {
		updates["context_type"] = deriveContextType(targetPermissionKey, targetResourceCode)
	}
	if targetPermissionKey != current.PermissionKey {
		existing, getErr := s.actionRepo.GetByPermissionKey(targetPermissionKey)
		if getErr == nil && existing != nil && existing.ID != id {
			return ErrPermissionActionExists
		}
		if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return getErr
		}
	}
	return s.actionRepo.UpdateWithMap(id, updates)
}

func normalizeModuleCode(value, fallbackResource string) string {
	moduleCode := strings.TrimSpace(value)
	if moduleCode != "" {
		return moduleCode
	}
	return strings.TrimSpace(fallbackResource)
}

func normalizeContextType(value, fallback string) string {
	switch strings.TrimSpace(value) {
	case "platform", "team":
		return strings.TrimSpace(value)
	case "":
		return fallback
	default:
		return fallback
	}
}

func deriveContextType(permissionKey, moduleCode string) string {
	targetKey := strings.TrimSpace(permissionKey)
	targetModule := strings.TrimSpace(moduleCode)
	switch {
	case strings.HasPrefix(targetKey, "system."),
		strings.HasPrefix(targetKey, "tenant."),
		strings.HasPrefix(targetKey, "platform."),
		targetKey == "tenant.manage":
		return "platform"
	case strings.HasPrefix(targetKey, "team."),
		strings.HasPrefix(targetKey, "product."),
		strings.HasPrefix(targetKey, "channel."),
		strings.HasPrefix(targetKey, "content."):
		return "team"
	case targetModule == "tenant" || targetModule == "role" || targetModule == "user" || targetModule == "menu" || targetModule == "permission_action" || targetModule == "api_endpoint":
		return "platform"
	default:
		return "team"
	}
}

func normalizeFeatureKind(value, fallback string) string {
	switch strings.TrimSpace(value) {
	case "system", "business":
		return strings.TrimSpace(value)
	case "":
		return fallback
	default:
		return fallback
	}
}

func (s *permissionService) Delete(id uuid.UUID) error {
	if _, err := s.actionRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionActionNotFound
		}
		return err
	}
	packageIDs, err := s.packageActionRepo.GetPackageIDsByActionID(id)
	if err != nil {
		return err
	}
	affectedTeams := make(map[uuid.UUID]struct{})
	for _, packageID := range packageIDs {
		teamIDs, teamErr := s.teamPackageRepo.GetTeamIDsByPackageID(packageID)
		if teamErr != nil {
			return teamErr
		}
		for _, teamID := range teamIDs {
			affectedTeams[teamID] = struct{}{}
		}
	}
	if err := s.packageActionRepo.DeleteByActionID(id); err != nil {
		return err
	}
	if err := s.roleDisabledActionRepo.DeleteByActionID(id); err != nil {
		return err
	}
	if err := s.teamBlockedActionRepo.DeleteByActionID(id); err != nil {
		return err
	}
	if err := s.userActionRepo.DeleteByActionID(id); err != nil {
		return err
	}
	if err := s.actionRepo.Delete(id); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshByPackages(packageIDs); err != nil {
			return err
		}
		return nil
	}
	for teamID := range affectedTeams {
		if _, err := s.boundaryService.RefreshSnapshot(teamID); err != nil {
			return err
		}
	}
	return nil
}
