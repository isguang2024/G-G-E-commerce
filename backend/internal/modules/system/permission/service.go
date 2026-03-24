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
	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

var (
	ErrPermissionKeyNotFound   = errors.New("permission key not found")
	ErrPermissionKeyExists     = errors.New("permission key already exists")
	ErrPermissionGroupNotFound = errors.New("permission group not found")
	ErrPermissionGroupExists   = errors.New("permission group already exists")
)

type PermissionService interface {
	List(req *dto.PermissionKeyListRequest) ([]user.PermissionKey, int64, error)
	Get(id uuid.UUID) (*user.PermissionKey, error)
	ListGroups(req *dto.PermissionGroupListRequest) ([]user.PermissionGroup, int64, error)
	ListEndpoints(id uuid.UUID) ([]user.APIEndpoint, error)
	CreateGroup(req *dto.PermissionGroupSaveRequest) (*user.PermissionGroup, error)
	UpdateGroup(id uuid.UUID, req *dto.PermissionGroupSaveRequest) error
	Create(req *dto.PermissionKeyCreateRequest) (*user.PermissionKey, error)
	Update(id uuid.UUID, req *dto.PermissionKeyUpdateRequest) error
	Delete(id uuid.UUID) error
}

type permissionService struct {
	groupRepo              user.PermissionGroupRepository
	keyRepo                user.PermissionKeyRepository
	apiEndpointRepo        user.APIEndpointRepository
	apiEndpointBindingRepo user.APIEndpointPermissionBindingRepository
	packageKeyRepo         user.FeaturePackageKeyRepository
	teamPackageRepo        user.TeamFeaturePackageRepository
	roleDisabledActionRepo user.RoleDisabledActionRepository
	teamBlockedActionRepo  user.TeamBlockedActionRepository
	userActionRepo         user.UserActionPermissionRepository
	boundaryService        teamboundary.Service
	refresher              permissionrefresh.Service
}

func NewPermissionService(
	groupRepo user.PermissionGroupRepository,
	keyRepo user.PermissionKeyRepository,
	apiEndpointRepo user.APIEndpointRepository,
	apiEndpointBindingRepo user.APIEndpointPermissionBindingRepository,
	packageKeyRepo user.FeaturePackageKeyRepository,
	teamPackageRepo user.TeamFeaturePackageRepository,
	roleDisabledActionRepo user.RoleDisabledActionRepository,
	teamBlockedActionRepo user.TeamBlockedActionRepository,
	userActionRepo user.UserActionPermissionRepository,
	boundaryService teamboundary.Service,
	refresher permissionrefresh.Service,
) PermissionService {
	return &permissionService{
		groupRepo:              groupRepo,
		keyRepo:                keyRepo,
		apiEndpointRepo:        apiEndpointRepo,
		apiEndpointBindingRepo: apiEndpointBindingRepo,
		packageKeyRepo:         packageKeyRepo,
		teamPackageRepo:        teamPackageRepo,
		roleDisabledActionRepo: roleDisabledActionRepo,
		teamBlockedActionRepo:  teamBlockedActionRepo,
		userActionRepo:         userActionRepo,
		boundaryService:        boundaryService,
		refresher:              refresher,
	}
}

func (s *permissionService) ListEndpoints(id uuid.UUID) ([]user.APIEndpoint, error) {
	item, err := s.keyRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionKeyNotFound
		}
		return nil, err
	}
	permissionKey := canonicalPermissionKey(item.PermissionKey)
	endpointIDs, err := s.apiEndpointBindingRepo.ListEndpointIDsByPermissionKey(permissionKey)
	if err != nil {
		return nil, err
	}
	if len(endpointIDs) == 0 {
		return []user.APIEndpoint{}, nil
	}
	endpoints, _, err := s.apiEndpointRepo.List(0, 5000, &user.APIEndpointListParams{
		Status: "normal",
	})
	if err != nil {
		return nil, err
	}
	endpointIDSet := make(map[uuid.UUID]struct{}, len(endpointIDs))
	for _, endpointID := range endpointIDs {
		endpointIDSet[endpointID] = struct{}{}
	}
	result := make([]user.APIEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if _, ok := endpointIDSet[endpoint.ID]; !ok {
			continue
		}
		result = append(result, endpoint)
	}
	return result, nil
}

func canonicalPermissionKey(permissionKey string) string {
	return permissionkey.Normalize(permissionKey)
}

func (s *permissionService) List(req *dto.PermissionKeyListRequest) ([]user.PermissionKey, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	var moduleGroupID *uuid.UUID
	if parsed, ok := parseUUID(req.ModuleGroupID); ok {
		moduleGroupID = &parsed
	}
	var featureGroupID *uuid.UUID
	if parsed, ok := parseUUID(req.FeatureGroupID); ok {
		featureGroupID = &parsed
	}
	var isBuiltin *bool
	if parsed, ok := parseBool(req.IsBuiltin); ok {
		isBuiltin = &parsed
	}
	params := &user.PermissionKeyListParams{
		Keyword:        strings.TrimSpace(req.Keyword),
		PermissionKey:  strings.TrimSpace(req.PermissionKey),
		Name:           req.Name,
		ModuleCode:     strings.TrimSpace(req.ModuleCode),
		ModuleGroupID:  moduleGroupID,
		FeatureGroupID: featureGroupID,
		ContextType:    normalizeContextType(req.ContextType, ""),
		FeatureKind:    normalizeFeatureKind(req.FeatureKind, ""),
		Status:         req.Status,
		IsBuiltin:      isBuiltin,
	}
	return s.keyRepo.List((req.Current-1)*req.Size, req.Size, params)
}

func (s *permissionService) ListGroups(req *dto.PermissionGroupListRequest) ([]user.PermissionGroup, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 200
	}
	return s.groupRepo.List((req.Current-1)*req.Size, req.Size, strings.TrimSpace(req.GroupType), strings.TrimSpace(req.Keyword), strings.TrimSpace(req.Status))
}

func (s *permissionService) CreateGroup(req *dto.PermissionGroupSaveRequest) (*user.PermissionGroup, error) {
	groupType := normalizeGroupType(req.GroupType)
	if groupType == "" {
		return nil, errors.New("group_type 仅支持 module|feature")
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return nil, errors.New("code 不能为空")
	}
	if _, err := s.groupRepo.GetByTypeAndCode(groupType, code); err == nil {
		return nil, ErrPermissionGroupExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	item := &user.PermissionGroup{
		ID:          permissionSeedGroupID(groupType, code),
		GroupType:   groupType,
		Code:        code,
		Name:        strings.TrimSpace(req.Name),
		NameEn:      strings.TrimSpace(req.NameEn),
		Description: strings.TrimSpace(req.Description),
		Status:      normalizeStatus(req.Status),
		SortOrder:   req.SortOrder,
		IsBuiltin:   false,
	}
	if err := s.groupRepo.Create(item); err != nil {
		return nil, err
	}
	return s.groupRepo.GetByID(item.ID)
}

func (s *permissionService) UpdateGroup(id uuid.UUID, req *dto.PermissionGroupSaveRequest) error {
	current, err := s.groupRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionGroupNotFound
		}
		return err
	}
	groupType := current.GroupType
	if target := normalizeGroupType(req.GroupType); target != "" {
		groupType = target
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = current.Code
	}
	existing, getErr := s.groupRepo.GetByTypeAndCode(groupType, code)
	if getErr == nil && existing.ID != id {
		return ErrPermissionGroupExists
	}
	if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
		return getErr
	}
	return s.groupRepo.UpdateWithMap(id, map[string]interface{}{
		"group_type":  groupType,
		"code":        code,
		"name":        strings.TrimSpace(req.Name),
		"name_en":     strings.TrimSpace(req.NameEn),
		"description": strings.TrimSpace(req.Description),
		"status":      normalizeStatus(req.Status),
		"sort_order":  req.SortOrder,
		"updated_at":  time.Now(),
	})
}

func (s *permissionService) Get(id uuid.UUID) (*user.PermissionKey, error) {
	item, err := s.keyRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionKeyNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *permissionService) Create(req *dto.PermissionKeyCreateRequest) (*user.PermissionKey, error) {
	permissionKey := permissionkey.Normalize(req.PermissionKey)
	if permissionKey == "" {
		return nil, errors.New("permission_key 不能为空")
	}
	mapping := permissionkey.FromKey(permissionKey)
	resourceCode := strings.TrimSpace(mapping.ResourceCode)
	if resourceCode == "" {
		resourceCode = strings.TrimSpace(req.ModuleCode)
	}
	if resourceCode == "" {
		return nil, errors.New("无法根据 permission_key 推导模块编码")
	}
	if _, err := s.keyRepo.GetByPermissionKey(permissionKey); err == nil {
		return nil, ErrPermissionKeyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	moduleGroup, err := s.resolvePermissionGroup(req.ModuleGroupID, "module", req.ModuleCode, resourceCode)
	if err != nil {
		return nil, err
	}
	featureGroup, err := s.resolvePermissionGroup(req.FeatureGroupID, "feature", req.FeatureKind, "system")
	if err != nil {
		return nil, err
	}
	featureKind := normalizeFeatureKind(featureGroup.Code, "system")
	moduleCode := normalizeModuleCode(moduleGroup.Code, resourceCode)
	contextType := normalizeContextType(req.ContextType, deriveContextType(permissionKey, moduleCode))
	item := &user.PermissionKey{
		PermissionKey:  permissionKey,
		ModuleCode:     moduleCode,
		ModuleGroupID:  &moduleGroup.ID,
		FeatureGroupID: &featureGroup.ID,
		ContextType:    contextType,
		FeatureKind:    featureKind,
		Name:           strings.TrimSpace(req.Name),
		Description:    strings.TrimSpace(req.Description),
		Status:         status,
		SortOrder:      req.SortOrder,
		IsBuiltin:      false,
	}
	if err := s.keyRepo.Create(item); err != nil {
		return nil, err
	}
	return s.keyRepo.GetByID(item.ID)
}

func (s *permissionService) Update(id uuid.UUID, req *dto.PermissionKeyUpdateRequest) error {
	current, err := s.keyRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionKeyNotFound
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
	if contextType := normalizeContextType(req.ContextType, ""); contextType != "" {
		updates["context_type"] = contextType
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		updates["status"] = status
	}
	targetPermissionKey := current.PermissionKey
	if permissionKey := permissionkey.Normalize(req.PermissionKey); permissionKey != "" {
		targetPermissionKey = permissionKey
	}
	if targetPermissionKey != "" {
		mapping := permissionkey.FromKey(targetPermissionKey)
		if mappedResource := strings.TrimSpace(mapping.ResourceCode); mappedResource != "" && strings.TrimSpace(req.ModuleCode) == "" {
			req.ModuleCode = mappedResource
		}
	}
	moduleGroup, err := s.resolvePermissionGroup(req.ModuleGroupID, "module", req.ModuleCode, current.ModuleCode)
	if err != nil {
		return err
	}
	featureGroup, err := s.resolvePermissionGroup(req.FeatureGroupID, "feature", req.FeatureKind, current.FeatureKind)
	if err != nil {
		return err
	}
	updates["module_group_id"] = moduleGroup.ID
	updates["feature_group_id"] = featureGroup.ID
	updates["module_code"] = normalizeModuleCode(moduleGroup.Code, current.ModuleCode)
	updates["feature_kind"] = normalizeFeatureKind(featureGroup.Code, current.FeatureKind)
	if _, exists := updates["context_type"]; !exists && current.ContextType == "" {
		updates["context_type"] = deriveContextType(targetPermissionKey, normalizeModuleCode(moduleGroup.Code, current.ModuleCode))
	}
	if targetPermissionKey != current.PermissionKey {
		existing, getErr := s.keyRepo.GetByPermissionKey(targetPermissionKey)
		if getErr == nil && existing != nil && existing.ID != id {
			return ErrPermissionKeyExists
		}
		if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return getErr
		}
	}
	return s.keyRepo.UpdateWithMap(id, updates)
}

func (s *permissionService) resolvePermissionGroup(idText string, groupType string, codeText string, fallbackCode string) (*user.PermissionGroup, error) {
	if parsed, ok := parseUUID(idText); ok {
		item, err := s.groupRepo.GetByID(parsed)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPermissionGroupNotFound
			}
			return nil, err
		}
		if item.GroupType != groupType {
			return nil, errors.New("分组类型不匹配")
		}
		return item, nil
	}
	code := strings.TrimSpace(codeText)
	if code == "" {
		code = strings.TrimSpace(fallbackCode)
	}
	if code == "" {
		if groupType == "feature" {
			code = "system"
		} else {
			code = "common"
		}
	}
	item, err := s.groupRepo.GetByTypeAndCode(groupType, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionGroupNotFound
		}
		return nil, err
	}
	return item, nil
}

func parseUUID(value string) (uuid.UUID, bool) {
	target := strings.TrimSpace(value)
	if target == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(target)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func parseBool(value string) (bool, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "true", "1":
		return true, true
	case "false", "0":
		return false, true
	default:
		return false, false
	}
}

func normalizeStatus(value string) string {
	if strings.TrimSpace(value) == "suspended" {
		return "suspended"
	}
	return "normal"
}

func normalizeGroupType(value string) string {
	switch strings.TrimSpace(value) {
	case "module", "feature":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func permissionSeedGroupID(groupType, code string) uuid.UUID {
	return permissionseed.StableID("permission-group", groupType+":"+code)
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
	case "platform", "team", "common":
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
	case targetModule == "tenant" || targetModule == "role" || targetModule == "user" || targetModule == "menu" || targetModule == "permission_key" || targetModule == "api_endpoint":
		return "platform"
	default:
		return "team"
	}
}

func normalizeFeatureKind(value, fallback string) string {
	target := strings.TrimSpace(value)
	switch target {
	case "":
		return fallback
	default:
		return target
	}
}

func (s *permissionService) Delete(id uuid.UUID) error {
	if _, err := s.keyRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionKeyNotFound
		}
		return err
	}
	packageIDs, err := s.packageKeyRepo.GetPackageIDsByKeyID(id)
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
	if err := s.packageKeyRepo.DeleteByKeyID(id); err != nil {
		return err
	}
	if err := s.roleDisabledActionRepo.DeleteByKeyID(id); err != nil {
		return err
	}
	if err := s.teamBlockedActionRepo.DeleteByKeyID(id); err != nil {
		return err
	}
	if err := s.userActionRepo.DeleteByKeyID(id); err != nil {
		return err
	}
	if err := s.keyRepo.Delete(id); err != nil {
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
