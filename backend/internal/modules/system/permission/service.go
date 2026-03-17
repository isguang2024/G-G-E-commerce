package permission

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
	actionRepo       user.PermissionActionRepository
	roleActionRepo   user.RoleActionPermissionRepository
	tenantActionRepo user.TenantActionPermissionRepository
	userActionRepo   user.UserActionPermissionRepository
	scopeRepo        user.ScopeRepository
}

func NewPermissionService(
	actionRepo user.PermissionActionRepository,
	roleActionRepo user.RoleActionPermissionRepository,
	tenantActionRepo user.TenantActionPermissionRepository,
	userActionRepo user.UserActionPermissionRepository,
	scopeRepo user.ScopeRepository,
) PermissionService {
	return &permissionService{
		actionRepo:       actionRepo,
		roleActionRepo:   roleActionRepo,
		tenantActionRepo: tenantActionRepo,
		userActionRepo:   userActionRepo,
		scopeRepo:        scopeRepo,
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
		Keyword:               strings.TrimSpace(req.Keyword),
		Name:                  req.Name,
		ResourceCode:          req.ResourceCode,
		ActionCode:            req.ActionCode,
		ModuleCode:            strings.TrimSpace(req.ModuleCode),
		Category:              strings.TrimSpace(req.Category),
		Source:                strings.TrimSpace(req.Source),
		FeatureKind:           normalizeFeatureKind(req.FeatureKind, ""),
		Status:                req.Status,
		RequiresTenantContext: req.RequiresTenantContext,
	}
	if req.ScopeID != "" {
		if parsed, err := uuid.Parse(strings.TrimSpace(req.ScopeID)); err == nil {
			params.ScopeID = &parsed
		}
	}
	params.ScopeCode = strings.TrimSpace(req.ScopeCode)
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
	resourceCode := strings.TrimSpace(req.ResourceCode)
	actionCode := strings.TrimSpace(req.ActionCode)
	if resourceCode == "" || actionCode == "" {
		return nil, errors.New("resource_code 和 action_code 不能为空")
	}
	if _, err := s.actionRepo.GetByResourceAndAction(resourceCode, actionCode); err == nil {
		return nil, ErrPermissionActionExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	scopeID, err := uuid.Parse(strings.TrimSpace(req.ScopeID))
	if err != nil {
		return nil, errors.New("invalid scope_id")
	}
	if _, err := s.scopeRepo.GetByID(scopeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("scope not found")
		}
		return nil, err
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	featureKind := normalizeFeatureKind(req.FeatureKind, "business")
	moduleCode := normalizeModuleCode(req.ModuleCode, req.ResourceCode)
	action := &user.PermissionAction{
		ResourceCode:          resourceCode,
		ActionCode:            actionCode,
		ModuleCode:            moduleCode,
		Category:              strings.TrimSpace(req.Category),
		Source:                "business",
		FeatureKind:           featureKind,
		Name:                  strings.TrimSpace(req.Name),
		Description:           strings.TrimSpace(req.Description),
		ScopeID:               scopeID,
		RequiresTenantContext: req.RequiresTenantContext,
		Status:                status,
		SortOrder:             req.SortOrder,
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
	if name := strings.TrimSpace(req.Name); name != "" {
		updates["name"] = name
	}
	if req.Category != "" {
		updates["category"] = strings.TrimSpace(req.Category)
	}
	if featureKind := normalizeFeatureKind(req.FeatureKind, ""); featureKind != "" {
		updates["feature_kind"] = featureKind
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if scopeIDStr := strings.TrimSpace(req.ScopeID); scopeIDStr != "" {
		scopeID, parseErr := uuid.Parse(scopeIDStr)
		if parseErr != nil {
			return errors.New("invalid scope_id")
		}
		if _, getErr := s.scopeRepo.GetByID(scopeID); getErr != nil {
			if errors.Is(getErr, gorm.ErrRecordNotFound) {
				return errors.New("scope not found")
			}
			return getErr
		}
		updates["scope_id"] = scopeID
	}
	if req.RequiresTenantContext != nil {
		updates["requires_tenant_context"] = *req.RequiresTenantContext
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		updates["status"] = status
	}

	resourceCode := strings.TrimSpace(req.ResourceCode)
	actionCode := strings.TrimSpace(req.ActionCode)
	targetResourceCode := current.ResourceCode
	targetActionCode := current.ActionCode
	if resourceCode != "" {
		targetResourceCode = resourceCode
		updates["resource_code"] = resourceCode
	}
	if actionCode != "" {
		targetActionCode = actionCode
		updates["action_code"] = actionCode
	}
	if req.ModuleCode != "" {
		updates["module_code"] = normalizeModuleCode(req.ModuleCode, targetResourceCode)
	}
	if req.ModuleCode == "" && (resourceCode != "" || current.ModuleCode == "") {
		updates["module_code"] = normalizeModuleCode(current.ModuleCode, targetResourceCode)
	}
	if targetResourceCode != current.ResourceCode || targetActionCode != current.ActionCode {
		existing, getErr := s.actionRepo.GetByResourceAndAction(targetResourceCode, targetActionCode)
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
	if err := s.roleActionRepo.DeleteByActionID(id); err != nil {
		return err
	}
	if err := s.tenantActionRepo.DeleteByActionID(id); err != nil {
		return err
	}
	if err := s.userActionRepo.DeleteByActionID(id); err != nil {
		return err
	}
	return s.actionRepo.Delete(id)
}
