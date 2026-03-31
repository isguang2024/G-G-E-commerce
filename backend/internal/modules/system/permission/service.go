package permission

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

var (
	ErrPermissionKeyNotFound    = errors.New("permission key not found")
	ErrPermissionKeyExists      = errors.New("permission key already exists")
	ErrPermissionContextInvalid = errors.New("permission context invalid")
	ErrPermissionGroupNotFound  = errors.New("permission group not found")
	ErrPermissionGroupExists    = errors.New("permission group already exists")
	ErrAPIEndpointNotFound      = errors.New("api endpoint not found")
)

type PermissionService interface {
	List(req *dto.PermissionKeyListRequest) ([]PermissionListItem, int64, PermissionAuditSummary, error)
	ListOptions(req *dto.PermissionKeyListRequest) ([]user.PermissionKey, error)
	Get(id uuid.UUID) (*user.PermissionKey, error)
	ListGroups(req *dto.PermissionGroupListRequest) ([]user.PermissionGroup, int64, error)
	ListEndpoints(id uuid.UUID) ([]user.APIEndpoint, error)
	CleanupUnused() (*CleanupUnusedResult, error)
	AddEndpoint(id uuid.UUID, endpointCode string) error
	RemoveEndpoint(id uuid.UUID, endpointCode string) error
	CreateGroup(req *dto.PermissionGroupSaveRequest) (*user.PermissionGroup, error)
	UpdateGroup(id uuid.UUID, req *dto.PermissionGroupSaveRequest) error
	Create(req *dto.PermissionKeyCreateRequest) (*user.PermissionKey, error)
	Update(id uuid.UUID, req *dto.PermissionKeyUpdateRequest) error
	Delete(id uuid.UUID) error
}

type CleanupUnusedResult struct {
	DeletedCount int
	DeletedKeys  []string
}

type permissionService struct {
	db                     *gorm.DB
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
	db *gorm.DB,
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
		db:                     db,
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
	endpointCodes, err := s.apiEndpointBindingRepo.ListEndpointCodesByPermissionKey(permissionKey)
	if err != nil {
		return nil, err
	}
	if len(endpointCodes) == 0 {
		return []user.APIEndpoint{}, nil
	}
	endpoints, err := s.apiEndpointRepo.GetByCodes(endpointCodes)
	if err != nil {
		return nil, err
	}
	result := make([]user.APIEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if endpoint.Status != "normal" {
			continue
		}
		result = append(result, endpoint)
	}
	return result, nil
}

func (s *permissionService) CleanupUnused() (*CleanupUnusedResult, error) {
	var actions []user.PermissionKey
	if err := s.db.
		Model(&user.PermissionKey{}).
		Where("is_builtin = ?", false).
		Preload("ModuleGroup").
		Preload("FeatureGroup").
		Order("sort_order ASC, created_at DESC").
		Find(&actions).Error; err != nil {
		return nil, err
	}

	auditProfiles, err := s.buildPermissionAuditProfiles(actions)
	if err != nil {
		return nil, err
	}

	result := &CleanupUnusedResult{
		DeletedKeys: make([]string, 0),
	}
	for _, action := range actions {
		profile := auditProfiles[action.ID]
		if profile.UsagePattern != permissionUsagePatternUnused {
			continue
		}
		if err := s.Delete(action.ID); err != nil {
			return nil, err
		}
		result.DeletedCount++
		result.DeletedKeys = append(result.DeletedKeys, canonicalPermissionKey(action.PermissionKey))
	}
	return result, nil
}

func (s *permissionService) AddEndpoint(id uuid.UUID, endpointCode string) error {
	item, err := s.keyRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionKeyNotFound
		}
		return err
	}
	targetCode := strings.TrimSpace(endpointCode)
	if targetCode == "" {
		return ErrAPIEndpointNotFound
	}
	if _, err := s.apiEndpointRepo.GetByCode(targetCode); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAPIEndpointNotFound
		}
		return err
	}
	if err := s.apiEndpointBindingRepo.AddByPermissionKey(canonicalPermissionKey(item.PermissionKey), targetCode); err != nil {
		return err
	}
	return s.refreshByPermissionKeyID(item.ID)
}

func (s *permissionService) RemoveEndpoint(id uuid.UUID, endpointCode string) error {
	item, err := s.keyRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionKeyNotFound
		}
		return err
	}
	if err := s.apiEndpointBindingRepo.RemoveByPermissionKey(canonicalPermissionKey(item.PermissionKey), strings.TrimSpace(endpointCode)); err != nil {
		return err
	}
	return s.refreshByPermissionKeyID(item.ID)
}

func canonicalPermissionKey(permissionKey string) string {
	return permissionkey.Normalize(permissionKey)
}

func (s *permissionService) List(req *dto.PermissionKeyListRequest) ([]PermissionListItem, int64, PermissionAuditSummary, error) {
	if req == nil {
		req = &dto.PermissionKeyListRequest{}
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	var actions []user.PermissionKey
	if err := s.buildPermissionListQuery(req).
		Order("sort_order ASC, created_at DESC").
		Find(&actions).Error; err != nil {
		return nil, 0, PermissionAuditSummary{}, err
	}

	auditProfiles, err := s.buildPermissionAuditProfiles(actions)
	if err != nil {
		return nil, 0, PermissionAuditSummary{}, err
	}

	filtered := make([]PermissionListItem, 0, len(actions))
	summary := PermissionAuditSummary{}
	for _, action := range actions {
		profile := auditProfiles[action.ID]
		if !matchesPermissionAuditFilters(profile, req) {
			continue
		}
		filtered = append(filtered, PermissionListItem{
			PermissionKey: action,
			Audit:         profile,
		})
		accumulatePermissionAuditSummary(&summary, profile)
	}

	total := int64(len(filtered))
	start := (req.Current - 1) * req.Size
	if start >= len(filtered) {
		return []PermissionListItem{}, total, summary, nil
	}
	end := start + req.Size
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, summary, nil
}

func (s *permissionService) ListOptions(req *dto.PermissionKeyListRequest) ([]user.PermissionKey, error) {
	query := s.buildPermissionListQuery(req)
	var actions []user.PermissionKey
	err := query.Order("sort_order ASC, created_at DESC").Find(&actions).Error
	return actions, err
}

func (s *permissionService) buildPermissionAuditProfiles(
	items []user.PermissionKey,
) (map[uuid.UUID]PermissionAuditProfile, error) {
	result := make(map[uuid.UUID]PermissionAuditProfile, len(items))
	if len(items) == 0 {
		return result, nil
	}

	counters, err := s.loadPermissionUsageCounters(items)
	if err != nil {
		return nil, err
	}
	duplicateProfiles, err := s.loadPermissionDuplicateProfiles()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		key := canonicalPermissionKey(item.PermissionKey)
		result[item.ID] = buildPermissionAuditProfile(item, counters[item.ID], duplicateProfiles[key])
	}
	return result, nil
}

func (s *permissionService) loadPermissionUsageCounters(
	items []user.PermissionKey,
) (map[uuid.UUID]permissionUsageCounters, error) {
	result := make(map[uuid.UUID]permissionUsageCounters, len(items))
	if len(items) == 0 {
		return result, nil
	}

	keyToID := make(map[string]uuid.UUID, len(items))
	keyIDs := make([]uuid.UUID, 0, len(items))
	permissionKeys := make([]string, 0, len(items))
	for _, item := range items {
		result[item.ID] = permissionUsageCounters{}
		keyIDs = append(keyIDs, item.ID)
		permissionKey := canonicalPermissionKey(item.PermissionKey)
		if permissionKey == "" {
			continue
		}
		if _, exists := keyToID[permissionKey]; exists {
			continue
		}
		keyToID[permissionKey] = item.ID
		permissionKeys = append(permissionKeys, permissionKey)
	}

	type keyedCountRow struct {
		PermissionKey string
		Total         int64
	}
	type idCountRow struct {
		ActionID uuid.UUID
		Total    int64
	}

	if len(permissionKeys) > 0 {
		var apiRows []keyedCountRow
		if err := s.db.Model(&user.APIEndpointPermissionBinding{}).
			Select("permission_key, COUNT(DISTINCT endpoint_code) AS total").
			Where("permission_key IN ?", permissionKeys).
			Group("permission_key").
			Scan(&apiRows).Error; err != nil {
			return nil, err
		}
		for _, row := range apiRows {
			actionID, ok := keyToID[canonicalPermissionKey(row.PermissionKey)]
			if !ok {
				continue
			}
			counter := result[actionID]
			counter.APICount = row.Total
			result[actionID] = counter
		}

		var pageRows []keyedCountRow
		if err := s.db.Model(&user.UIPage{}).
			Select("permission_key, COUNT(*) AS total").
			Where("permission_key IN ?", permissionKeys).
			Group("permission_key").
			Scan(&pageRows).Error; err != nil {
			return nil, err
		}
		for _, row := range pageRows {
			actionID, ok := keyToID[canonicalPermissionKey(row.PermissionKey)]
			if !ok {
				continue
			}
			counter := result[actionID]
			counter.PageCount = row.Total
			result[actionID] = counter
		}
	}

	if len(keyIDs) == 0 {
		return result, nil
	}

	var packageRows []idCountRow
	if err := s.db.Model(&user.FeaturePackageKey{}).
		Select("action_id, COUNT(DISTINCT package_id) AS total").
		Where("action_id IN ?", keyIDs).
		Group("action_id").
		Scan(&packageRows).Error; err != nil {
		return nil, err
	}
	for _, row := range packageRows {
		counter := result[row.ActionID]
		counter.PackageCount = row.Total
		result[row.ActionID] = counter
	}

	return result, nil
}

func (s *permissionService) loadPermissionDuplicateProfiles() (map[string]permissionDuplicateProfile, error) {
	type sourceRow struct {
		ID            uuid.UUID
		PermissionKey string
		ContextType   string
	}

	var rows []sourceRow
	if err := s.db.Model(&user.PermissionKey{}).
		Select("id, permission_key, context_type").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]permissionDuplicateSource, 0, len(rows))
	for _, row := range rows {
		items = append(items, permissionDuplicateSource{
			ID:            row.ID,
			PermissionKey: row.PermissionKey,
			ContextType:   row.ContextType,
		})
	}
	return buildPermissionDuplicateProfiles(items), nil
}

func matchesPermissionAuditFilters(profile PermissionAuditProfile, req *dto.PermissionKeyListRequest) bool {
	if req == nil {
		return true
	}
	if usagePattern := strings.TrimSpace(req.UsagePattern); usagePattern != "" && profile.UsagePattern != usagePattern {
		return false
	}
	if duplicatePattern := strings.TrimSpace(req.DuplicatePattern); duplicatePattern != "" && profile.DuplicatePattern != duplicatePattern {
		return false
	}
	return true
}

func (s *permissionService) buildPermissionListQuery(req *dto.PermissionKeyListRequest) *gorm.DB {
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
	query := s.db.
		Model(&user.PermissionKey{}).
		Joins("LEFT JOIN permission_groups AS module_groups ON module_groups.id = permission_keys.module_group_id AND module_groups.deleted_at IS NULL").
		Joins("LEFT JOIN permission_groups AS feature_groups ON feature_groups.id = permission_keys.feature_group_id AND feature_groups.deleted_at IS NULL").
		Preload("ModuleGroup").
		Preload("FeatureGroup")
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		query = query.Where(
			"(name LIKE ? OR description LIKE ? OR permission_key LIKE ? OR module_code LIKE ? OR feature_kind LIKE ?)",
			keyword, keyword, keyword, keyword, keyword,
		)
	}
	if params.PermissionKey != "" {
		query = query.Where("permission_key LIKE ?", "%"+params.PermissionKey+"%")
	}
	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.ModuleCode != "" {
		query = query.Where("module_code LIKE ?", "%"+params.ModuleCode+"%")
	}
	if params.ModuleGroupID != nil {
		query = query.Where("module_group_id = ?", *params.ModuleGroupID)
	}
	if params.FeatureGroupID != nil {
		query = query.Where("feature_group_id = ?", *params.FeatureGroupID)
	}
	if params.ContextType != "" {
		query = query.Where("context_type = ?", params.ContextType)
	}
	if params.FeatureKind != "" {
		query = query.Where("feature_kind = ?", params.FeatureKind)
	}
	if params.Status != "" {
		query = query.Where(
			`CASE
				WHEN permission_keys.status = 'suspended'
					OR module_groups.status = 'suspended'
					OR feature_groups.status = 'suspended'
				THEN 'suspended'
				ELSE 'normal'
			END = ?`,
			params.Status,
		)
	}
	if params.IsBuiltin != nil {
		query = query.Where("is_builtin = ?", *params.IsBuiltin)
	}
	return query
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
	if err := validatePermissionContext(permissionKey, moduleCode, contextType); err != nil {
		return nil, err
	}
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
	targetModuleCode := normalizeModuleCode(moduleGroup.Code, current.ModuleCode)
	updates["module_code"] = targetModuleCode
	updates["feature_kind"] = normalizeFeatureKind(featureGroup.Code, current.FeatureKind)
	targetContextType := current.ContextType
	if contextType, exists := updates["context_type"]; exists {
		targetContextType = normalizeContextType(fmt.Sprint(contextType), deriveContextType(targetPermissionKey, targetModuleCode))
	} else {
		targetContextType = normalizeContextType(current.ContextType, deriveContextType(targetPermissionKey, targetModuleCode))
		if current.ContextType == "" {
			updates["context_type"] = targetContextType
		}
	}
	if err := validatePermissionContext(targetPermissionKey, targetModuleCode, targetContextType); err != nil {
		return err
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
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.PermissionKey{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if targetPermissionKey != current.PermissionKey {
			if err := tx.Model(&models.APIEndpointPermissionBinding{}).
				Where("permission_key = ?", current.PermissionKey).
				Update("permission_key", targetPermissionKey).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return s.refreshByPermissionKeyID(id)
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
	case strings.HasPrefix(targetKey, "team."):
		return "team"
	case targetModule == "tenant" || targetModule == "role" || targetModule == "user" || targetModule == "menu" || targetModule == "permission_key" || targetModule == "api_endpoint":
		return "platform"
	default:
		return "common"
	}
}

func validatePermissionContext(permissionKey, moduleCode, contextType string) error {
	targetKey := canonicalPermissionKey(permissionKey)
	targetModule := strings.TrimSpace(moduleCode)
	targetContext := normalizeContextType(contextType, deriveContextType(targetKey, targetModule))
	if targetContext == "" {
		return fmt.Errorf("%w: context_type 仅支持 platform、team、common", ErrPermissionContextInvalid)
	}

	if mapping, ok := findCanonicalPermissionMapping(targetKey); ok {
		expectedContext := normalizeContextType(mapping.ContextType, deriveContextType(targetKey, targetModule))
		if expectedContext != "" && targetContext != expectedContext {
			return fmt.Errorf("%w: 内置权限键 %s 必须使用 %s 上下文", ErrPermissionContextInvalid, targetKey, expectedContext)
		}
		return nil
	}

	if expectedContext := deriveReservedContextType(targetKey); expectedContext != "" && targetContext != expectedContext {
		return fmt.Errorf("%w: 权限键 %s 必须使用 %s 上下文", ErrPermissionContextInvalid, targetKey, expectedContext)
	}

	if moduleContext := deriveModuleContextBoundary(targetModule); moduleContext != "" && targetContext != moduleContext {
		return fmt.Errorf("%w: 模块 %s 仅允许使用 %s 上下文", ErrPermissionContextInvalid, targetModule, moduleContext)
	}

	switch targetContext {
	case "platform":
		if !hasPlatformPermissionNamespace(targetKey) && deriveModuleContextBoundary(targetModule) == "" {
			return fmt.Errorf("%w: 平台自定义权限键请使用 system.、platform. 或 tenant. 前缀，或归入平台模块分组", ErrPermissionContextInvalid)
		}
	case "team":
		if !hasTeamPermissionNamespace(targetKey) && deriveModuleContextBoundary(targetModule) == "" {
			return fmt.Errorf("%w: 团队自定义权限键请使用 team. 前缀，或归入团队模块分组", ErrPermissionContextInvalid)
		}
	case "common":
		if deriveReservedContextType(targetKey) != "" || deriveModuleContextBoundary(targetModule) != "" {
			return fmt.Errorf("%w: 平台/团队专属权限键不能标记为 common", ErrPermissionContextInvalid)
		}
	}

	return nil
}

func findCanonicalPermissionMapping(permissionKey string) (permissionkey.Mapping, bool) {
	targetKey := canonicalPermissionKey(permissionKey)
	if targetKey == "" {
		return permissionkey.Mapping{}, false
	}
	for _, mapping := range permissionkey.ListMappings() {
		if canonicalPermissionKey(mapping.Key) == targetKey {
			return permissionkey.FromKey(targetKey), true
		}
	}
	return permissionkey.Mapping{}, false
}

func deriveReservedContextType(permissionKey string) string {
	targetKey := canonicalPermissionKey(permissionKey)
	switch {
	case hasPlatformPermissionNamespace(targetKey):
		return "platform"
	case hasTeamPermissionNamespace(targetKey):
		return "team"
	default:
		return ""
	}
}

func hasPlatformPermissionNamespace(permissionKey string) bool {
	targetKey := canonicalPermissionKey(permissionKey)
	return strings.HasPrefix(targetKey, "system.") ||
		strings.HasPrefix(targetKey, "platform.") ||
		strings.HasPrefix(targetKey, "tenant.")
}

func hasTeamPermissionNamespace(permissionKey string) bool {
	targetKey := canonicalPermissionKey(permissionKey)
	return strings.HasPrefix(targetKey, "team.")
}

func deriveModuleContextBoundary(moduleCode string) string {
	targetModule := strings.TrimSpace(moduleCode)
	switch targetModule {
	case "role", "permission_key", "user", "menu", "menu_backup", "system", "tenant",
		"tenant_member_admin", "api_endpoint", "page", "fast_enter", "message",
		"feature_package", "system_permission", "menu_space", "navigation":
		return "platform"
	case "team_member", "team", "team_message":
		return "team"
	default:
		return ""
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

func (s *permissionService) refreshByPermissionKeyID(keyID uuid.UUID) error {
	packageIDs, err := s.packageKeyRepo.GetPackageIDsByKeyID(keyID)
	if err != nil {
		return err
	}
	if len(packageIDs) == 0 {
		return nil
	}
	if s.refresher != nil {
		return s.refresher.RefreshByPackages(packageIDs)
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
	for teamID := range affectedTeams {
		if _, refreshErr := s.boundaryService.RefreshSnapshot(teamID); refreshErr != nil {
			return refreshErr
		}
	}
	return nil
}
