package workspacerolebinding

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

type PersonalRoleRow struct {
	UserID               uuid.UUID
	BindingWorkspaceID   uuid.UUID
	BindingWorkspaceType string
	Role                 models.Role
}

func HasPersonalRoleCodesByUserID(db *gorm.DB, userID uuid.UUID, roleCodes []string, onlyActive bool) (bool, error) {
	if db == nil || userID == uuid.Nil {
		return false, nil
	}
	normalized := normalizeRoleCodes(roleCodes)
	if len(normalized) == 0 {
		return false, nil
	}
	var count int64
	query := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN workspaces ON workspaces.id = workspace_role_bindings.workspace_id").
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.user_id = ?", userID).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("workspaces.workspace_type = ? AND workspaces.owner_user_id = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypePersonal, userID).
		Where("roles.code IN ? AND roles.deleted_at IS NULL", normalized).
		Where("NOT EXISTS (SELECT 1 FROM role_scopes rs WHERE rs.role_id = roles.id AND rs.deleted_at IS NULL AND rs.scope_type <> ?)", models.ScopeTypeGlobal)
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func HasCollaborationWorkspaceRoleCodesByCollaborationWorkspaceAndUser(db *gorm.DB, collaborationWorkspaceID, userID uuid.UUID, roleCodes []string, onlyActive bool) (bool, error) {
	if db == nil || collaborationWorkspaceID == uuid.Nil || userID == uuid.Nil {
		return false, nil
	}
	workspace, err := GetCollaborationWorkspaceByCollaborationWorkspaceID(db, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	normalized := normalizeRoleCodes(roleCodes)
	if len(normalized) == 0 {
		return false, nil
	}
	var count int64
	query := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.workspace_id = ? AND workspace_role_bindings.user_id = ?", workspace.ID, userID).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("roles.code IN ? AND roles.deleted_at IS NULL", normalized)
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetPersonalWorkspaceByUserID(db *gorm.DB, userID uuid.UUID) (*models.Workspace, error) {
	if db == nil || userID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	var workspace models.Workspace
	err := db.Where("workspace_type = ? AND owner_user_id = ? AND deleted_at IS NULL", models.WorkspaceTypePersonal, userID).
		First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func GetCollaborationWorkspaceByCollaborationWorkspaceID(db *gorm.DB, collaborationWorkspaceID uuid.UUID) (*models.Workspace, error) {
	if db == nil || collaborationWorkspaceID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	var workspace models.Workspace
	err := db.Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, collaborationWorkspaceID).
		First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func GetWorkspaceByCollaborationWorkspaceID(db *gorm.DB, collaborationWorkspaceID uuid.UUID) (*models.Workspace, error) {
	return GetCollaborationWorkspaceByCollaborationWorkspaceID(db, collaborationWorkspaceID)
}

func EnsurePersonalWorkspace(tx *gorm.DB, userID uuid.UUID) (*models.Workspace, error) {
	if tx == nil || userID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	workspace, err := GetPersonalWorkspaceByUserID(tx, userID)
	if err == nil {
		if ensureErr := ensureWorkspaceMember(tx, workspace.ID, userID, models.WorkspaceMemberOwner, models.WorkspaceStatusActive); ensureErr != nil {
			return nil, ensureErr
		}
		return workspace, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var user models.User
	if err := tx.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, err
	}

	workspace = &models.Workspace{
		WorkspaceType: models.WorkspaceTypePersonal,
		Name:          buildPersonalWorkspaceName(user),
		Code:          buildPersonalWorkspaceCode(user),
		OwnerUserID:   uuidPtr(user.ID),
		Status:        models.WorkspaceStatusActive,
		Meta: models.MetaJSON{
			"legacy_source": "user",
		},
	}
	if err := tx.Create(workspace).Error; err != nil {
		return nil, err
	}
	if err := ensureWorkspaceMember(tx, workspace.ID, userID, models.WorkspaceMemberOwner, models.WorkspaceStatusActive); err != nil {
		return nil, err
	}
	return workspace, nil
}

func ListPersonalRoleIDsByUserID(db *gorm.DB, userID uuid.UUID, onlyActive bool) ([]uuid.UUID, error) {
	if db == nil || userID == uuid.Nil {
		return []uuid.UUID{}, nil
	}
	var roleIDs []uuid.UUID
	query := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN workspaces ON workspaces.id = workspace_role_bindings.workspace_id").
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.user_id = ?", userID).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("workspaces.workspace_type = ? AND workspaces.owner_user_id = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypePersonal, userID).
		Where("roles.deleted_at IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM role_scopes rs WHERE rs.role_id = roles.id AND rs.deleted_at IS NULL AND rs.scope_type <> ?)", models.ScopeTypeGlobal)
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	if err := query.Distinct("workspace_role_bindings.role_id").Pluck("workspace_role_bindings.role_id", &roleIDs).Error; err != nil {
		return nil, err
	}
	return roleIDs, nil
}

func ListCollaborationWorkspaceRoleIDsByCollaborationWorkspaceAndUser(db *gorm.DB, collaborationWorkspaceID, userID uuid.UUID, onlyActive bool) ([]uuid.UUID, error) {
	if db == nil || collaborationWorkspaceID == uuid.Nil || userID == uuid.Nil {
		return []uuid.UUID{}, nil
	}
	workspace, err := GetCollaborationWorkspaceByCollaborationWorkspaceID(db, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}
	return listRoleIDsByWorkspaceAndUser(db, workspace.ID, userID, onlyActive)
}

func ReplacePersonalRoleBindings(tx *gorm.DB, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if tx == nil || userID == uuid.Nil {
		return nil
	}
	workspace, err := EnsurePersonalWorkspace(tx, userID)
	if err != nil {
		return err
	}
	return replaceRoleBindingsByWorkspaceAndUser(tx, workspace.ID, userID, roleIDs)
}

func ReplaceCollaborationWorkspaceRoleBindings(tx *gorm.DB, collaborationWorkspaceID, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if tx == nil || collaborationWorkspaceID == uuid.Nil || userID == uuid.Nil {
		return nil
	}
	workspace, err := GetCollaborationWorkspaceByCollaborationWorkspaceID(tx, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return replaceRoleBindingsByWorkspaceAndUser(tx, workspace.ID, userID, roleIDs)
}

func ListCollaborationWorkspaceRoleIDsByUser(db *gorm.DB, collaborationWorkspaceID, userID uuid.UUID, onlyActive bool) ([]uuid.UUID, error) {
	return ListCollaborationWorkspaceRoleIDsByCollaborationWorkspaceAndUser(db, collaborationWorkspaceID, userID, onlyActive)
}

func HasCollaborationWorkspaceRoleCodesByUser(db *gorm.DB, collaborationWorkspaceID, userID uuid.UUID, roleCodes []string, onlyActive bool) (bool, error) {
	return HasCollaborationWorkspaceRoleCodesByCollaborationWorkspaceAndUser(db, collaborationWorkspaceID, userID, roleCodes, onlyActive)
}

func LoadPersonalRoleRows(db *gorm.DB, userIDs []uuid.UUID, onlyActive bool) ([]PersonalRoleRow, error) {
	if db == nil || len(userIDs) == 0 {
		return []PersonalRoleRow{}, nil
	}

	type row struct {
		UserID               uuid.UUID `gorm:"column:user_id"`
		BindingWorkspaceID   uuid.UUID `gorm:"column:binding_workspace_id"`
		BindingWorkspaceType string    `gorm:"column:binding_workspace_type"`
		ID                   uuid.UUID `gorm:"column:id"`
		Code                 string    `gorm:"column:code"`
		Name                 string    `gorm:"column:name"`
		Description          string    `gorm:"column:description"`
		Status               string    `gorm:"column:status"`
		SortOrder            int       `gorm:"column:sort_order"`
		IsSystem             bool      `gorm:"column:is_system"`
		CreatedAt            time.Time `gorm:"column:created_at"`
		UpdatedAt            time.Time `gorm:"column:updated_at"`
	}

	var rows []row
	query := db.Table("workspace_role_bindings").
		Select(`
			workspaces.owner_user_id AS user_id,
			workspaces.id AS binding_workspace_id,
			workspaces.workspace_type AS binding_workspace_type,
			roles.id,
			roles.code,
			roles.name,
			roles.description,
			roles.status,
			roles.sort_order,
			roles.is_system,
			roles.created_at,
			roles.updated_at
		`).
		Joins("JOIN workspaces ON workspaces.id = workspace_role_bindings.workspace_id").
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspaces.workspace_type = ? AND workspaces.owner_user_id IN ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypePersonal, userIDs).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("roles.deleted_at IS NULL").
		Where(`
			NOT EXISTS (
				SELECT 1
				FROM role_scopes
				WHERE role_scopes.role_id = roles.id
					AND role_scopes.deleted_at IS NULL
					AND role_scopes.scope_type <> ?
			)
		`, models.ScopeTypeGlobal)
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]PersonalRoleRow, 0, len(rows))
	for _, item := range rows {
		result = append(result, PersonalRoleRow{
			UserID:               item.UserID,
			BindingWorkspaceID:   item.BindingWorkspaceID,
			BindingWorkspaceType: item.BindingWorkspaceType,
			Role: models.Role{
				ID:          item.ID,
				Code:        item.Code,
				Name:        item.Name,
				Description: item.Description,
				Status:      item.Status,
				SortOrder:   item.SortOrder,
				IsSystem:    item.IsSystem,
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
			},
		})
	}
	return result, nil
}

func ListPlatformUserIDsByRoleIDs(db *gorm.DB, roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if db == nil || len(roleIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var userIDs []uuid.UUID
	err := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN workspaces ON workspaces.id = workspace_role_bindings.workspace_id").
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.role_id IN ?", roleIDs).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("workspaces.workspace_type = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypePersonal).
		Where("roles.deleted_at IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM role_scopes rs WHERE rs.role_id = roles.id AND rs.deleted_at IS NULL AND rs.scope_type <> ?)", models.ScopeTypeGlobal).
		Distinct("workspaces.owner_user_id").
		Pluck("workspaces.owner_user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

func ListPlatformUserIDsByRoleCodes(db *gorm.DB, roleCodes []string, onlyActive bool) ([]uuid.UUID, error) {
	if db == nil {
		return []uuid.UUID{}, nil
	}
	normalized := normalizeRoleCodes(roleCodes)
	if len(normalized) == 0 {
		return []uuid.UUID{}, nil
	}
	var userIDs []uuid.UUID
	query := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN workspaces ON workspaces.id = workspace_role_bindings.workspace_id").
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("workspaces.workspace_type = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypePersonal).
		Where("roles.code IN ? AND roles.deleted_at IS NULL", normalized).
		Where("NOT EXISTS (SELECT 1 FROM role_scopes rs WHERE rs.role_id = roles.id AND rs.deleted_at IS NULL AND rs.scope_type <> ?)", models.ScopeTypeGlobal)
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	err := query.Distinct("workspaces.owner_user_id").Pluck("workspaces.owner_user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(userIDs), nil
}

func ListUserIDsByCollaborationWorkspaceRoleCodes(db *gorm.DB, collaborationWorkspaceID uuid.UUID, roleCodes []string, onlyActive bool) ([]uuid.UUID, error) {
	if db == nil || collaborationWorkspaceID == uuid.Nil {
		return []uuid.UUID{}, nil
	}
	workspace, err := GetWorkspaceByCollaborationWorkspaceID(db, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}
	normalized := normalizeRoleCodes(roleCodes)
	if len(normalized) == 0 {
		return []uuid.UUID{}, nil
	}
	var userIDs []uuid.UUID
	query := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.workspace_id = ?", workspace.ID).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("roles.code IN ? AND roles.deleted_at IS NULL", normalized)
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	err = query.Distinct("workspace_role_bindings.user_id").Pluck("workspace_role_bindings.user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(userIDs), nil
}

func ListUserIDsByCollaborationWorkspaceRoleIDs(db *gorm.DB, collaborationWorkspaceID uuid.UUID, roleIDs []uuid.UUID, onlyActive bool) ([]uuid.UUID, error) {
	if db == nil || collaborationWorkspaceID == uuid.Nil || len(roleIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	workspace, err := GetWorkspaceByCollaborationWorkspaceID(db, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}
	var userIDs []uuid.UUID
	query := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.workspace_id = ?", workspace.ID).
		Where("workspace_role_bindings.role_id IN ?", dedupeUUIDs(roleIDs)).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("roles.deleted_at IS NULL")
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	err = query.Distinct("workspace_role_bindings.user_id").Pluck("workspace_role_bindings.user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(userIDs), nil
}

func listRoleIDsByWorkspaceAndUser(db *gorm.DB, workspaceID, userID uuid.UUID, onlyActive bool) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	query := db.Model(&models.WorkspaceRoleBinding{}).
		Joins("JOIN roles ON roles.id = workspace_role_bindings.role_id").
		Where("workspace_role_bindings.workspace_id = ? AND workspace_role_bindings.user_id = ?", workspaceID, userID).
		Where("workspace_role_bindings.enabled = ? AND workspace_role_bindings.deleted_at IS NULL", true).
		Where("roles.deleted_at IS NULL")
	if onlyActive {
		query = query.Where("roles.status = ?", "normal")
	}
	if err := query.Distinct("workspace_role_bindings.role_id").Pluck("workspace_role_bindings.role_id", &roleIDs).Error; err != nil {
		return nil, err
	}
	return roleIDs, nil
}

func replaceRoleBindingsByWorkspaceAndUser(tx *gorm.DB, workspaceID, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if err := tx.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Delete(&models.WorkspaceRoleBinding{}).Error; err != nil {
		return err
	}
	items := make([]models.WorkspaceRoleBinding, 0, len(roleIDs))
	seen := make(map[uuid.UUID]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID == uuid.Nil {
			continue
		}
		if _, exists := seen[roleID]; exists {
			continue
		}
		seen[roleID] = struct{}{}
		items = append(items, models.WorkspaceRoleBinding{
			WorkspaceID: workspaceID,
			UserID:      userID,
			RoleID:      roleID,
			Enabled:     true,
		})
	}
	if len(items) == 0 {
		return nil
	}
	return tx.Create(&items).Error
}

func ensureWorkspaceMember(tx *gorm.DB, workspaceID, userID uuid.UUID, memberType, status string) error {
	var existing models.WorkspaceMember
	err := tx.Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", workspaceID, userID).First(&existing).Error
	if err == nil {
		return tx.Model(&existing).Updates(map[string]interface{}{
			"member_type": memberType,
			"status":      status,
			"updated_at":  tx.NowFunc(),
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&models.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		MemberType:  memberType,
		Status:      status,
	}).Error
}

func buildPersonalWorkspaceName(user models.User) string {
	for _, candidate := range []string{
		strings.TrimSpace(user.Nickname),
		strings.TrimSpace(user.Username),
		strings.TrimSpace(user.Email),
	} {
		if candidate != "" {
			return candidate + " Personal Workspace"
		}
	}
	return "Personal Workspace"
}

func buildPersonalWorkspaceCode(user models.User) string {
	base := normalizeWorkspaceCodeComponent(firstNonEmpty(
		strings.TrimSpace(user.Username),
		strings.TrimSpace(user.Email),
		user.ID.String(),
	))
	if base == "" {
		base = user.ID.String()
	}
	return "personal-" + base
}

func normalizeWorkspaceCodeComponent(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return ""
	}
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	target = reg.ReplaceAllString(target, "-")
	return strings.Trim(target, "-")
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

func uuidPtr(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	target := value
	return &target
}

func normalizeRoleCodes(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func dedupeUUIDs(items []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == uuid.Nil {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
