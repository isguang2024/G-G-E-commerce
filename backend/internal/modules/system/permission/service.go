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
	"github.com/gg-ecommerce/backend/internal/pkg/workspacefeaturebinding"
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
	GetConsumerDetails(id uuid.UUID) (*PermissionConsumerDetails, error)
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
	GetImpactPreview(id uuid.UUID) (*PermissionImpactPreview, error)
	BatchUpdate(req *PermissionBatchUpdateRequest, operatorID *uuid.UUID, requestID string) (*PermissionBatchUpdateResult, error)
	SaveBatchTemplate(req *PermissionBatchTemplateSaveRequest, operatorID *uuid.UUID) (*user.PermissionBatchTemplate, error)
	ListBatchTemplates() ([]user.PermissionBatchTemplate, error)
	ListRiskAudits(objectID string, current, size int) ([]user.RiskOperationAudit, int64, error)
}

type CleanupUnusedResult struct {
	DeletedCount int
	DeletedKeys  []string
}

type PermissionConsumerDetails struct {
	PermissionKey string                          `json:"permission_key"`
	APIs          []PermissionConsumerAPIItem     `json:"apis"`
	Pages         []PermissionConsumerPageItem    `json:"pages"`
	FeaturePkgs   []PermissionConsumerPackageItem `json:"feature_packages"`
	Roles         []PermissionConsumerRoleItem    `json:"roles"`
}

type PermissionConsumerAPIItem struct {
	Code    string `json:"code"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Summary string `json:"summary"`
}

type PermissionConsumerPageItem struct {
	PageKey    string `json:"page_key"`
	Name       string `json:"name"`
	RoutePath  string `json:"route_path"`
	AccessMode string `json:"access_mode"`
}

type PermissionConsumerPackageItem struct {
	ID          uuid.UUID `json:"id"`
	PackageKey  string    `json:"package_key"`
	Name        string    `json:"name"`
	PackageType string    `json:"package_type"`
	ContextType string    `json:"context_type"`
}

type PermissionConsumerRoleItem struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	ContextType string    `json:"context_type"`
}

type PermissionImpactPreview struct {
	PermissionKey string `json:"permission_key"`
	APICount      int64  `json:"api_count"`
	PageCount     int64  `json:"page_count"`
	PackageCount  int64  `json:"package_count"`
	RoleCount     int64  `json:"role_count"`
	TeamCount     int64  `json:"team_count"`
	UserCount     int64  `json:"user_count"`
}

type PermissionBatchUpdateRequest struct {
	IDs            []string `json:"ids"`
	Status         *string  `json:"status"`
	ModuleGroupID  *string  `json:"module_group_id"`
	FeatureGroupID *string  `json:"feature_group_id"`
	TemplateName   string   `json:"template_name"`
}

type PermissionBatchUpdateResult struct {
	UpdatedCount int      `json:"updated_count"`
	SkippedIDs   []string `json:"skipped_ids"`
}

type PermissionBatchTemplateSaveRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Payload     map[string]interface{} `json:"payload"`
}

type permissionService struct {
	db                     *gorm.DB
	groupRepo              user.PermissionGroupRepository
	keyRepo                user.PermissionKeyRepository
	apiEndpointRepo        user.APIEndpointRepository
	apiEndpointBindingRepo user.APIEndpointPermissionBindingRepository
	packageKeyRepo         user.FeaturePackageKeyRepository
	teamPackageRepo        user.CollaborationWorkspaceFeaturePackageRepository
	roleDisabledActionRepo user.RoleDisabledActionRepository
	teamBlockedActionRepo  user.CollaborationWorkspaceBlockedActionRepository
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
	teamPackageRepo user.CollaborationWorkspaceFeaturePackageRepository,
	roleDisabledActionRepo user.RoleDisabledActionRepository,
	teamBlockedActionRepo user.CollaborationWorkspaceBlockedActionRepository,
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

func (s *permissionService) GetConsumerDetails(id uuid.UUID) (*PermissionConsumerDetails, error) {
	item, err := s.keyRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionKeyNotFound
		}
		return nil, err
	}
	permissionKey := canonicalPermissionKey(item.PermissionKey)
	details := &PermissionConsumerDetails{
		PermissionKey: permissionKey,
		APIs:          make([]PermissionConsumerAPIItem, 0),
		Pages:         make([]PermissionConsumerPageItem, 0),
		FeaturePkgs:   make([]PermissionConsumerPackageItem, 0),
		Roles:         make([]PermissionConsumerRoleItem, 0),
	}

	apiEndpoints, err := s.ListEndpoints(id)
	if err != nil {
		return nil, err
	}
	for _, endpoint := range apiEndpoints {
		details.APIs = append(details.APIs, PermissionConsumerAPIItem{
			Code:    endpoint.Code,
			Method:  endpoint.Method,
			Path:    endpoint.Path,
			Summary: endpoint.Summary,
		})
	}

	var pages []user.UIPage
	if err := s.db.Model(&user.UIPage{}).
		Select("page_key", "name", "route_path", "access_mode").
		Where("permission_key = ? AND deleted_at IS NULL", permissionKey).
		Order("sort_order ASC, created_at ASC").
		Find(&pages).Error; err != nil {
		return nil, err
	}
	for _, page := range pages {
		details.Pages = append(details.Pages, PermissionConsumerPageItem{
			PageKey:    page.PageKey,
			Name:       page.Name,
			RoutePath:  page.RoutePath,
			AccessMode: page.AccessMode,
		})
	}

	packageIDs, err := s.packageKeyRepo.GetPackageIDsByKeyID(id)
	if err != nil {
		return nil, err
	}
	if len(packageIDs) > 0 {
		packages, getErr := user.NewFeaturePackageRepository(s.db).GetByIDs(packageIDs)
		if getErr != nil {
			return nil, getErr
		}
		for _, pkg := range packages {
			details.FeaturePkgs = append(details.FeaturePkgs, PermissionConsumerPackageItem{
				ID:          pkg.ID,
				PackageKey:  pkg.PackageKey,
				Name:        pkg.Name,
				PackageType: pkg.PackageType,
				ContextType: pkg.ContextType,
			})
		}

		type roleRow struct {
			ID                       uuid.UUID
			Code                     string
			Name                     string
			CollaborationWorkspaceID *uuid.UUID
		}
		var roleRows []roleRow
		if err := s.db.Model(&user.RoleFeaturePackage{}).
			Select("roles.id", "roles.code", "roles.name", "roles.collaboration_workspace_id").
			Joins("JOIN roles ON roles.id = role_feature_packages.role_id").
			Where("role_feature_packages.package_id IN ? AND role_feature_packages.enabled = ?", packageIDs, true).
			Where("roles.deleted_at IS NULL").
			Distinct().
			Find(&roleRows).Error; err != nil {
			return nil, err
		}
		for _, role := range roleRows {
			contextType := "platform"
			if role.CollaborationWorkspaceID != nil {
				contextType = "team"
			}
			details.Roles = append(details.Roles, PermissionConsumerRoleItem{
				ID:          role.ID,
				Code:        role.Code,
				Name:        role.Name,
				ContextType: contextType,
			})
		}
	}

	return details, nil
}

func (s *permissionService) GetImpactPreview(id uuid.UUID) (*PermissionImpactPreview, error) {
	item, err := s.keyRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionKeyNotFound
		}
		return nil, err
	}
	permissionKey := canonicalPermissionKey(item.PermissionKey)
	result := &PermissionImpactPreview{PermissionKey: permissionKey}

	if err := s.db.Model(&user.APIEndpointPermissionBinding{}).Where("permission_key = ?", permissionKey).Distinct("endpoint_code").Count(&result.APICount).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&user.UIPage{}).Where("permission_key = ? AND deleted_at IS NULL", permissionKey).Count(&result.PageCount).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&user.FeaturePackageKey{}).Where("action_id = ?", id).Distinct("package_id").Count(&result.PackageCount).Error; err != nil {
		return nil, err
	}
	packageIDs, err := s.packageKeyRepo.GetPackageIDsByKeyID(id)
	if err != nil {
		return nil, err
	}
	if len(packageIDs) > 0 {
		if err := s.db.Model(&user.RoleFeaturePackage{}).Where("package_id IN ? AND enabled = ?", packageIDs, true).Distinct("role_id").Count(&result.RoleCount).Error; err != nil {
			return nil, err
		}
		workspaceCollaborationWorkspaceIDs, err := workspacefeaturebinding.ListCollaborationWorkspaceIDsByPackageIDs(s.db, packageIDs, "")
		if err != nil {
			return nil, err
		}
		var legacyCollaborationWorkspaceIDs []uuid.UUID
		if err := s.db.Model(&user.CollaborationWorkspaceFeaturePackage{}).Where("package_id IN ? AND enabled = ?", packageIDs, true).Distinct("collaboration_workspace_id").Pluck("collaboration_workspace_id", &legacyCollaborationWorkspaceIDs).Error; err != nil {
			return nil, err
		}
		result.TeamCount = int64(len(mergeUUIDs(workspaceCollaborationWorkspaceIDs, legacyCollaborationWorkspaceIDs)))
		workspaceUserIDs, err := workspacefeaturebinding.ListPlatformUserIDsByPackageIDs(s.db, packageIDs, "")
		if err != nil {
			return nil, err
		}
		var legacyUserIDs []uuid.UUID
		if err := s.db.Model(&user.UserFeaturePackage{}).Where("package_id IN ? AND enabled = ?", packageIDs, true).Distinct("user_id").Pluck("user_id", &legacyUserIDs).Error; err != nil {
			return nil, err
		}
		result.UserCount = int64(len(mergeUUIDs(workspaceUserIDs, legacyUserIDs)))
	}
	return result, nil
}

func (s *permissionService) BatchUpdate(req *PermissionBatchUpdateRequest, operatorID *uuid.UUID, requestID string) (*PermissionBatchUpdateResult, error) {
	if req == nil {
		return nil, errors.New("批量参数不能为空")
	}
	ids := make([]uuid.UUID, 0, len(req.IDs))
	seen := make(map[uuid.UUID]struct{}, len(req.IDs))
	for _, item := range req.IDs {
		parsed, err := uuid.Parse(strings.TrimSpace(item))
		if err != nil {
			continue
		}
		if _, ok := seen[parsed]; ok {
			continue
		}
		seen[parsed] = struct{}{}
		ids = append(ids, parsed)
	}
	if len(ids) == 0 {
		return &PermissionBatchUpdateResult{UpdatedCount: 0, SkippedIDs: req.IDs}, nil
	}

	moduleGroupID := (*uuid.UUID)(nil)
	if req.ModuleGroupID != nil && strings.TrimSpace(*req.ModuleGroupID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.ModuleGroupID))
		if err != nil {
			return nil, errors.New("无效的模块分组ID")
		}
		moduleGroupID = &parsed
	}
	featureGroupID := (*uuid.UUID)(nil)
	if req.FeatureGroupID != nil && strings.TrimSpace(*req.FeatureGroupID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.FeatureGroupID))
		if err != nil {
			return nil, errors.New("无效的功能分组ID")
		}
		featureGroupID = &parsed
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if req.Status != nil {
		targetStatus := normalizeStatus(strings.TrimSpace(*req.Status))
		updates["status"] = targetStatus
	}
	if moduleGroupID != nil {
		updates["module_group_id"] = *moduleGroupID
	}
	if featureGroupID != nil {
		updates["feature_group_id"] = *featureGroupID
	}
	if len(updates) == 1 {
		return &PermissionBatchUpdateResult{UpdatedCount: 0, SkippedIDs: []string{}}, nil
	}

	if err := s.db.Model(&user.PermissionKey{}).Where("id IN ?", ids).Updates(updates).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		_ = s.refreshByPermissionKeyID(id)
		_ = s.recordRiskAudit("permission_action", id.String(), "batch_update", nil, map[string]interface{}{
			"status":           req.Status,
			"module_group_id":  req.ModuleGroupID,
			"feature_group_id": req.FeatureGroupID,
			"template_name":    strings.TrimSpace(req.TemplateName),
		}, nil, operatorID, requestID)
	}
	return &PermissionBatchUpdateResult{
		UpdatedCount: len(ids),
		SkippedIDs:   []string{},
	}, nil
}

func (s *permissionService) SaveBatchTemplate(req *PermissionBatchTemplateSaveRequest, operatorID *uuid.UUID) (*user.PermissionBatchTemplate, error) {
	if req == nil {
		return nil, errors.New("模板参数不能为空")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("模板名称不能为空")
	}
	item := &user.PermissionBatchTemplate{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Payload:     req.Payload,
		CreatedBy:   operatorID,
	}
	var existing user.PermissionBatchTemplate
	if err := s.db.Where("name = ?", name).First(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if createErr := s.db.Create(item).Error; createErr != nil {
			return nil, createErr
		}
		return item, nil
	}
	if err := s.db.Model(&existing).Updates(map[string]interface{}{
		"description": item.Description,
		"payload":     item.Payload,
		"updated_at":  time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *permissionService) ListBatchTemplates() ([]user.PermissionBatchTemplate, error) {
	items := make([]user.PermissionBatchTemplate, 0)
	if err := s.db.Model(&user.PermissionBatchTemplate{}).Order("updated_at DESC, created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *permissionService) ListRiskAudits(objectID string, current, size int) ([]user.RiskOperationAudit, int64, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	query := s.db.Model(&user.RiskOperationAudit{}).Where("object_type = ?", "permission_action")
	if strings.TrimSpace(objectID) != "" {
		query = query.Where("object_id = ?", strings.TrimSpace(objectID))
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]user.RiskOperationAudit, 0)
	if err := query.Order("created_at DESC").Offset((current - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
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
		PermissionKey:         permissionKey,
		AppKey:                derivePermissionAppKey(permissionKey, moduleCode, contextType),
		ModuleCode:            moduleCode,
		ModuleGroupID:         &moduleGroup.ID,
		FeatureGroupID:        &featureGroup.ID,
		ContextType:           contextType,
		FeatureKind:           featureKind,
		DataPolicy:            derivePermissionDataPolicy(permissionKey, moduleCode, contextType),
		AllowedWorkspaceTypes: deriveAllowedWorkspaceTypes(contextType),
		Name:                  strings.TrimSpace(req.Name),
		Description:           strings.TrimSpace(req.Description),
		Status:                status,
		SortOrder:             req.SortOrder,
		IsBuiltin:             false,
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
	updates["app_key"] = derivePermissionAppKey(targetPermissionKey, targetModuleCode, targetContextType)
	updates["data_policy"] = derivePermissionDataPolicy(targetPermissionKey, targetModuleCode, targetContextType)
	updates["allowed_workspace_types"] = deriveAllowedWorkspaceTypes(targetContextType)
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
	_ = s.recordRiskAudit("permission_action", id.String(), "update", map[string]interface{}{
		"permission_key": current.PermissionKey,
		"status":         current.Status,
		"context_type":   current.ContextType,
	}, map[string]interface{}{
		"permission_key": targetPermissionKey,
		"status":         req.Status,
		"context_type":   targetContextType,
	}, nil, nil, "")
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

func derivePermissionAppKey(permissionKey, moduleCode, contextType string) string {
	targetKey := canonicalPermissionKey(permissionKey)
	if targetKey != "" {
		segments := strings.Split(targetKey, ".")
		if len(segments) > 0 {
			switch segments[0] {
			case "system", "platform", "tenant", "team", "common":
				return models.DefaultAppKey
			default:
				if strings.TrimSpace(segments[0]) != "" {
					return strings.TrimSpace(segments[0])
				}
			}
		}
	}

	switch strings.TrimSpace(moduleCode) {
	case "", "role", "user", "tenant", "menu", "menu_backup", "permission_action", "permission_key", "api_endpoint", "feature_package", "workspace", "navigation", "page", "app":
		return models.DefaultAppKey
	default:
		return strings.TrimSpace(moduleCode)
	}
}

func derivePermissionDataPolicy(permissionKey, moduleCode, contextType string) string {
	targetKey := canonicalPermissionKey(permissionKey)
	switch normalizeContextType(contextType, deriveContextType(targetKey, moduleCodeFromPermissionKey(targetKey, moduleCode))) {
	case "team", "common":
		return "auth_workspace"
	default:
		return "none"
	}
}

func deriveAllowedWorkspaceTypes(contextType string) string {
	switch normalizeContextType(contextType, "team") {
	case "platform":
		return "personal"
	case "common":
		return "personal,team"
	default:
		return "team"
	}
}

func moduleCodeFromPermissionKey(permissionKey, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	mapping := permissionkey.FromKey(permissionKey)
	return strings.TrimSpace(mapping.ResourceCode)
}

func deriveContextType(permissionKey, moduleCode string) string {
	targetKey := strings.TrimSpace(permissionKey)
	targetModule := strings.TrimSpace(moduleCode)
	switch {
	case strings.HasPrefix(targetKey, "system."),
		strings.HasPrefix(targetKey, "collaboration_workspace."),
		strings.HasPrefix(targetKey, "platform."),
		targetKey == "collaboration_workspace.manage":
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
			return fmt.Errorf("%w: 平台自定义权限键请使用 system.、platform. 或 collaboration_workspace. 前缀，或归入平台模块分组", ErrPermissionContextInvalid)
		}
	case "team":
		if !hasTeamPermissionNamespace(targetKey) && deriveModuleContextBoundary(targetModule) == "" {
			return fmt.Errorf("%w: 协作空间自定义权限键请使用 team. 前缀，或归入协作空间模块分组", ErrPermissionContextInvalid)
		}
	case "common":
		if deriveReservedContextType(targetKey) != "" || deriveModuleContextBoundary(targetModule) != "" {
			return fmt.Errorf("%w: 平台/协作空间专属权限键不能标记为 common", ErrPermissionContextInvalid)
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
		strings.HasPrefix(targetKey, "collaboration_workspace.")
}

func hasTeamPermissionNamespace(permissionKey string) bool {
	targetKey := canonicalPermissionKey(permissionKey)
	return strings.HasPrefix(targetKey, "team.")
}

func deriveModuleContextBoundary(moduleCode string) string {
	targetModule := strings.TrimSpace(moduleCode)
	switch targetModule {
	case "role", "permission_key", "user", "menu", "menu_backup", "system", "tenant",
		"collaboration_workspace_member_admin", "api_endpoint", "page", "fast_enter", "message",
		"feature_package", "system_permission", "menu_space", "navigation":
		return "platform"
	case "team_member", "team", "collaboration_workspace_message":
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
	item, err := s.keyRepo.GetByID(id)
	if err != nil {
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
		collaborationWorkspaceIDs, teamErr := s.teamPackageRepo.GetCollaborationWorkspaceIDsByPackageID(packageID)
		if teamErr != nil {
			return teamErr
		}
		for _, teamID := range collaborationWorkspaceIDs {
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
		_ = s.recordRiskAudit("permission_action", id.String(), "delete", map[string]interface{}{
			"permission_key": item.PermissionKey,
			"context_type":   item.ContextType,
		}, nil, map[string]interface{}{"package_count": len(packageIDs)}, nil, "")
		return nil
	}
	for teamID := range affectedTeams {
		if _, err := s.boundaryService.RefreshSnapshot(teamID); err != nil {
			return err
		}
	}
	_ = s.recordRiskAudit("permission_action", id.String(), "delete", map[string]interface{}{
		"permission_key": item.PermissionKey,
		"context_type":   item.ContextType,
	}, nil, map[string]interface{}{"team_count": len(affectedTeams)}, nil, "")
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
		collaborationWorkspaceIDs, teamErr := s.teamPackageRepo.GetCollaborationWorkspaceIDsByPackageID(packageID)
		if teamErr != nil {
			return teamErr
		}
		for _, teamID := range collaborationWorkspaceIDs {
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

func (s *permissionService) recordRiskAudit(
	objectType string,
	objectID string,
	operationType string,
	beforeSummary map[string]interface{},
	afterSummary map[string]interface{},
	impactSummary map[string]interface{},
	operatorID *uuid.UUID,
	requestID string,
) error {
	item := &user.RiskOperationAudit{
		ObjectType:    strings.TrimSpace(objectType),
		ObjectID:      strings.TrimSpace(objectID),
		OperationType: strings.TrimSpace(operationType),
		BeforeSummary: beforeSummary,
		AfterSummary:  afterSummary,
		ImpactSummary: impactSummary,
		OperatorID:    operatorID,
		RequestID:     strings.TrimSpace(requestID),
	}
	return s.db.Create(item).Error
}

func mergeUUIDs(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, id := range group {
			if id == uuid.Nil {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}
