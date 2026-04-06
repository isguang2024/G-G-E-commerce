package permissionrefresh

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacefeaturebinding"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

type Service interface {
	RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
	RefreshCollaborationWorkspaces(collaborationWorkspaceIDs []uuid.UUID) error
	RefreshAllCollaborationWorkspaces() error
	RefreshPersonalWorkspaceUser(userID uuid.UUID) error
	RefreshPersonalWorkspaceUsers(userIDs []uuid.UUID) error
	RefreshAllPersonalWorkspaceUsers() error
	RefreshPersonalWorkspaceRole(roleID uuid.UUID) error
	RefreshPersonalWorkspaceRoles(roleIDs []uuid.UUID) error
	RefreshAllPersonalWorkspaceRoles() error
	RefreshByPackage(packageID uuid.UUID) error
	RefreshByPackages(packageIDs []uuid.UUID) error
	RefreshByPackageWithStats(packageID uuid.UUID) (RefreshStats, error)
	RefreshByPackagesWithStats(packageIDs []uuid.UUID) (RefreshStats, error)
	RefreshByMenu(menuID uuid.UUID) error
}

type RefreshStats struct {
	RequestedPackageCount       int       `json:"requested_package_count"`
	ImpactedPackageCount        int       `json:"impacted_package_count"`
	RoleCount                   int       `json:"role_count"`
	CollaborationWorkspaceCount int       `json:"collaboration_workspace_count"`
	UserCount                   int       `json:"user_count"`
	ElapsedMilliseconds         int64     `json:"elapsed_milliseconds"`
	FinishedAt                  time.Time `json:"finished_at"`
}

type service struct {
	db                           *gorm.DB
	boundaryService              collaborationworkspaceboundary.Service
	personalWorkspaceService     platformaccess.Service
	personalWorkspaceRoleService platformroleaccess.Service
}

func NewService(db *gorm.DB, boundaryService collaborationworkspaceboundary.Service, platformService platformaccess.Service, roleService platformroleaccess.Service) Service {
	return &service{
		db:                           db,
		boundaryService:              boundaryService,
		personalWorkspaceService:     platformService,
		personalWorkspaceRoleService: roleService,
	}
}

func (s *service) RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error {
	if collaborationWorkspaceID == uuid.Nil || s.boundaryService == nil {
		return nil
	}
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	for _, appKey := range appKeys {
		if _, refreshErr := s.boundaryService.RefreshSnapshot(collaborationWorkspaceID, appKey); refreshErr != nil {
			return refreshErr
		}
	}
	return nil
}

func (s *service) RefreshCollaborationWorkspaces(collaborationWorkspaceIDs []uuid.UUID) error {
	for _, collaborationWorkspaceID := range dedupeUUIDs(collaborationWorkspaceIDs) {
		if err := s.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) RefreshAllCollaborationWorkspaces() error {
	if s.db == nil {
		return nil
	}
	type collaborationWorkspaceIDOnly struct {
		ID uuid.UUID
	}
	return s.db.Model(&models.CollaborationWorkspace{}).
		Select("id").
		FindInBatches(&[]collaborationWorkspaceIDOnly{}, 200, func(tx *gorm.DB, _ int) error {
			rows, ok := tx.Statement.Dest.(*[]collaborationWorkspaceIDOnly)
			if !ok || len(*rows) == 0 {
				return nil
			}
			collaborationWorkspaceIDs := make([]uuid.UUID, 0, len(*rows))
			for _, row := range *rows {
				if row.ID != uuid.Nil {
					collaborationWorkspaceIDs = append(collaborationWorkspaceIDs, row.ID)
				}
			}
			return s.RefreshCollaborationWorkspaces(collaborationWorkspaceIDs)
		}).Error
}

func (s *service) RefreshPersonalWorkspaceUser(userID uuid.UUID) error {
	if userID == uuid.Nil || s.personalWorkspaceService == nil {
		return nil
	}
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	for _, appKey := range appKeys {
		if _, refreshErr := s.personalWorkspaceService.RefreshSnapshot(userID, appKey); refreshErr != nil {
			return refreshErr
		}
	}
	return nil
}

func (s *service) RefreshPersonalWorkspaceUsers(userIDs []uuid.UUID) error {
	for _, userID := range dedupeUUIDs(userIDs) {
		if err := s.RefreshPersonalWorkspaceUser(userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) RefreshAllPersonalWorkspaceUsers() error {
	if s.db == nil {
		return nil
	}
	type userIDOnly struct {
		ID uuid.UUID
	}
	return s.db.Model(&models.User{}).
		Select("id").
		FindInBatches(&[]userIDOnly{}, 200, func(tx *gorm.DB, _ int) error {
			rows, ok := tx.Statement.Dest.(*[]userIDOnly)
			if !ok || len(*rows) == 0 {
				return nil
			}
			userIDs := make([]uuid.UUID, 0, len(*rows))
			for _, row := range *rows {
				if row.ID != uuid.Nil {
					userIDs = append(userIDs, row.ID)
				}
			}
			return s.RefreshPersonalWorkspaceUsers(userIDs)
		}).Error
}

func (s *service) RefreshPersonalWorkspaceRole(roleID uuid.UUID) error {
	if roleID == uuid.Nil {
		return nil
	}
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	if s.personalWorkspaceRoleService != nil {
		for _, appKey := range appKeys {
			if _, refreshErr := s.personalWorkspaceRoleService.RefreshSnapshot(roleID, appKey); refreshErr != nil {
				return refreshErr
			}
		}
	}
	userIDs, err := s.getPlatformUserIDsByRoleIDs([]uuid.UUID{roleID})
	if err != nil {
		return err
	}
	return s.RefreshPersonalWorkspaceUsers(userIDs)
}

func (s *service) RefreshPersonalWorkspaceRoles(roleIDs []uuid.UUID) error {
	dedupedRoleIDs := dedupeUUIDs(roleIDs)
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	if s.personalWorkspaceRoleService != nil {
		for _, appKey := range appKeys {
			for _, roleID := range dedupedRoleIDs {
				if roleID == uuid.Nil {
					continue
				}
				if _, refreshErr := s.personalWorkspaceRoleService.RefreshSnapshot(roleID, appKey); refreshErr != nil {
					return refreshErr
				}
			}
		}
	}
	userIDs, err := s.getPlatformUserIDsByRoleIDs(dedupedRoleIDs)
	if err != nil {
		return err
	}
	return s.RefreshPersonalWorkspaceUsers(userIDs)
}

func (s *service) RefreshAllPersonalWorkspaceRoles() error {
	if s.db == nil {
		return nil
	}
	type roleIDOnly struct {
		ID uuid.UUID
	}
	return s.db.Model(&models.Role{}).
		Select("id").
		Where("collaboration_workspace_id IS NULL").
		FindInBatches(&[]roleIDOnly{}, 200, func(tx *gorm.DB, _ int) error {
			rows, ok := tx.Statement.Dest.(*[]roleIDOnly)
			if !ok || len(*rows) == 0 {
				return nil
			}
			roleIDs := make([]uuid.UUID, 0, len(*rows))
			for _, row := range *rows {
				if row.ID != uuid.Nil {
					roleIDs = append(roleIDs, row.ID)
				}
			}
			return s.RefreshPersonalWorkspaceRoles(roleIDs)
		}).Error
}

func (s *service) RefreshByPackage(packageID uuid.UUID) error {
	return s.RefreshByPackages([]uuid.UUID{packageID})
}

func (s *service) RefreshByPackages(packageIDs []uuid.UUID) error {
	_, err := s.RefreshByPackagesWithStats(packageIDs)
	return err
}

func (s *service) RefreshByPackageWithStats(packageID uuid.UUID) (RefreshStats, error) {
	return s.RefreshByPackagesWithStats([]uuid.UUID{packageID})
}

func (s *service) RefreshByPackagesWithStats(packageIDs []uuid.UUID) (RefreshStats, error) {
	startedAt := time.Now()
	stats := RefreshStats{
		RequestedPackageCount: len(dedupeUUIDs(packageIDs)),
	}
	impactedPackageIDs, err := s.collectImpactedPackageIDs(packageIDs)
	if err != nil {
		return stats, err
	}
	stats.ImpactedPackageCount = len(impactedPackageIDs)
	if len(impactedPackageIDs) == 0 {
		stats.ElapsedMilliseconds = time.Since(startedAt).Milliseconds()
		stats.FinishedAt = time.Now()
		return stats, nil
	}

	collaborationWorkspaceIDs, err := s.getCollaborationWorkspaceIDsByPackageIDs(impactedPackageIDs)
	if err != nil {
		return stats, err
	}
	roleBindings, err := s.getRoleBindingsByPackageIDs(impactedPackageIDs)
	if err != nil {
		return stats, err
	}
	userIDs, err := s.getPlatformUserIDsByPackageIDs(impactedPackageIDs)
	if err != nil {
		return stats, err
	}

	platformRoleIDs := make([]uuid.UUID, 0, len(roleBindings))
	for _, binding := range roleBindings {
		if binding.CollaborationWorkspaceID == nil {
			platformRoleIDs = append(platformRoleIDs, binding.RoleID)
			continue
		}
		collaborationWorkspaceIDs = append(collaborationWorkspaceIDs, *binding.CollaborationWorkspaceID)
	}
	dedupedCollaborationWorkspaceIDs := dedupeUUIDs(collaborationWorkspaceIDs)
	dedupedRoleIDs := dedupeUUIDs(platformRoleIDs)

	roleUserIDs, err := s.getPlatformUserIDsByRoleIDs(dedupedRoleIDs)
	if err != nil {
		return stats, err
	}
	dedupedUserIDs := dedupeUUIDs(append(userIDs, roleUserIDs...))
	stats.CollaborationWorkspaceCount = len(dedupedCollaborationWorkspaceIDs)
	stats.RoleCount = len(dedupedRoleIDs)
	stats.UserCount = len(dedupedUserIDs)

	if err := s.RefreshPersonalWorkspaceRoles(dedupedRoleIDs); err != nil {
		return stats, err
	}
	if err := s.RefreshCollaborationWorkspaces(dedupedCollaborationWorkspaceIDs); err != nil {
		return stats, err
	}
	if err := s.RefreshPersonalWorkspaceUsers(dedupedUserIDs); err != nil {
		return stats, err
	}

	stats.ElapsedMilliseconds = time.Since(startedAt).Milliseconds()
	stats.FinishedAt = time.Now()
	return stats, nil
}

func (s *service) RefreshByMenu(menuID uuid.UUID) error {
	if menuID == uuid.Nil {
		return nil
	}
	packageIDs, err := s.getPackageIDsByMenuID(menuID)
	if err != nil {
		return err
	}
	return s.RefreshByPackages(packageIDs)
}

type roleBinding struct {
	RoleID                   uuid.UUID
	CollaborationWorkspaceID *uuid.UUID
}

type roleBindingRow struct {
	RoleID                   uuid.UUID
	CollaborationWorkspaceID sql.NullString
}

func (s *service) collectImpactedPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	current := dedupeUUIDs(packageIDs)
	if len(current) == 0 {
		return []uuid.UUID{}, nil
	}

	result := append([]uuid.UUID{}, current...)
	seen := make(map[uuid.UUID]struct{}, len(current))
	for _, item := range current {
		seen[item] = struct{}{}
	}

	queue := current
	for len(queue) > 0 {
		var parentIDs []uuid.UUID
		if err := s.db.Model(&models.FeaturePackageBundle{}).
			Where("child_package_id IN ?", queue).
			Distinct("package_id").
			Pluck("package_id", &parentIDs).Error; err != nil {
			return nil, err
		}
		nextQueue := make([]uuid.UUID, 0, len(parentIDs))
		for _, parentID := range parentIDs {
			if _, ok := seen[parentID]; ok {
				continue
			}
			seen[parentID] = struct{}{}
			result = append(result, parentID)
			nextQueue = append(nextQueue, parentID)
		}
		queue = nextQueue
	}
	return result, nil
}

func (s *service) getCollaborationWorkspaceIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	workspaceCollaborationWorkspaceIDs, err := workspacefeaturebinding.ListCollaborationWorkspaceIDsByPackageIDs(s.db, packageIDs, "")
	if err != nil {
		return nil, err
	}
	var collaborationWorkspaceIDs []uuid.UUID
	err = s.db.Model(&models.CollaborationWorkspaceFeaturePackage{}).
		Where("package_id IN ? AND enabled = ?", packageIDs, true).
		Distinct("collaboration_workspace_id").
		Pluck("collaboration_workspace_id", &collaborationWorkspaceIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(append(workspaceCollaborationWorkspaceIDs, collaborationWorkspaceIDs...)), nil
}

func (s *service) getRoleBindingsByPackageIDs(packageIDs []uuid.UUID) ([]roleBinding, error) {
	if len(packageIDs) == 0 {
		return []roleBinding{}, nil
	}
	var rows []roleBindingRow
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Select("roles.id AS role_id, roles.collaboration_workspace_id AS collaboration_workspace_id").
		Joins("JOIN roles ON roles.id = role_feature_packages.role_id").
		Where("role_feature_packages.package_id IN ? AND role_feature_packages.enabled = ?", packageIDs, true).
		Distinct().
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]roleBinding, 0, len(rows))
	for _, row := range rows {
		binding := roleBinding{RoleID: row.RoleID}
		if row.CollaborationWorkspaceID.Valid && row.CollaborationWorkspaceID.String != "" {
			collaborationWorkspaceID, parseErr := uuid.Parse(row.CollaborationWorkspaceID.String)
			if parseErr != nil {
				return nil, parseErr
			}
			binding.CollaborationWorkspaceID = &collaborationWorkspaceID
		}
		result = append(result, binding)
	}
	return result, nil
}

func (s *service) getPlatformUserIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	workspaceUserIDs, err := workspacefeaturebinding.ListPlatformUserIDsByPackageIDs(s.db, packageIDs, "")
	if err != nil {
		return nil, err
	}
	var userIDs []uuid.UUID
	err = s.db.Model(&models.UserFeaturePackage{}).
		Where("package_id IN ? AND enabled = ?", packageIDs, true).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(append(workspaceUserIDs, userIDs...)), nil
}

func (s *service) getPlatformUserIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	workspaceUserIDs, err := workspacerolebinding.ListPlatformUserIDsByRoleIDs(s.db, roleIDs)
	if err != nil {
		return nil, err
	}
	var userIDs []uuid.UUID
	err = s.db.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.role_id IN ?", roleIDs).
		Where("user_roles.collaboration_workspace_id IS NULL").
		Where("roles.collaboration_workspace_id IS NULL").
		Where("roles.deleted_at IS NULL").
		Distinct("user_roles.user_id").
		Pluck("user_roles.user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(append(workspaceUserIDs, userIDs...)), nil
}

func (s *service) getPackageIDsByMenuID(menuID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.FeaturePackageMenu{}).
		Where("menu_id = ?", menuID).
		Distinct("package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) listAppKeys() ([]string, error) {
	if s.db == nil {
		return []string{models.DefaultAppKey}, nil
	}
	var appKeys []string
	if err := s.db.Model(&models.App{}).
		Where("status = ? AND deleted_at IS NULL", "normal").
		Order("is_default DESC, created_at ASC").
		Pluck("app_key", &appKeys).Error; err != nil {
		return nil, err
	}
	appKeys = dedupeStrings(appKeys)
	if len(appKeys) == 0 {
		return []string{models.DefaultAppKey}, nil
	}
	return appKeys, nil
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

func dedupeStrings(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == "" {
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
