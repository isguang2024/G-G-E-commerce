package permissionrefresh

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

type Service interface {
	RefreshTeam(teamID uuid.UUID) error
	RefreshTeams(teamIDs []uuid.UUID) error
	RefreshAllTeams() error
	RefreshPlatformUser(userID uuid.UUID) error
	RefreshPlatformUsers(userIDs []uuid.UUID) error
	RefreshAllPlatformUsers() error
	RefreshPlatformRole(roleID uuid.UUID) error
	RefreshPlatformRoles(roleIDs []uuid.UUID) error
	RefreshAllPlatformRoles() error
	RefreshByPackage(packageID uuid.UUID) error
	RefreshByPackages(packageIDs []uuid.UUID) error
	RefreshByPackageWithStats(packageID uuid.UUID) (RefreshStats, error)
	RefreshByPackagesWithStats(packageIDs []uuid.UUID) (RefreshStats, error)
	RefreshByMenu(menuID uuid.UUID) error
}

type RefreshStats struct {
	RequestedPackageCount int       `json:"requested_package_count"`
	ImpactedPackageCount  int       `json:"impacted_package_count"`
	RoleCount             int       `json:"role_count"`
	TeamCount             int       `json:"team_count"`
	UserCount             int       `json:"user_count"`
	ElapsedMilliseconds   int64     `json:"elapsed_milliseconds"`
	FinishedAt            time.Time `json:"finished_at"`
}

type service struct {
	db              *gorm.DB
	boundaryService teamboundary.Service
	platformService platformaccess.Service
	roleService     platformroleaccess.Service
}

func NewService(db *gorm.DB, boundaryService teamboundary.Service, platformService platformaccess.Service, roleService platformroleaccess.Service) Service {
	return &service{
		db:              db,
		boundaryService: boundaryService,
		platformService: platformService,
		roleService:     roleService,
	}
}

func (s *service) RefreshTeam(teamID uuid.UUID) error {
	if teamID == uuid.Nil || s.boundaryService == nil {
		return nil
	}
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	for _, appKey := range appKeys {
		if _, refreshErr := s.boundaryService.RefreshSnapshot(teamID, appKey); refreshErr != nil {
			return refreshErr
		}
	}
	return nil
}

func (s *service) RefreshTeams(teamIDs []uuid.UUID) error {
	for _, teamID := range dedupeUUIDs(teamIDs) {
		if err := s.RefreshTeam(teamID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) RefreshAllTeams() error {
	if s.db == nil {
		return nil
	}
	type tenantIDOnly struct {
		ID uuid.UUID
	}
	return s.db.Model(&models.Tenant{}).
		Select("id").
		FindInBatches(&[]tenantIDOnly{}, 200, func(tx *gorm.DB, _ int) error {
			rows, ok := tx.Statement.Dest.(*[]tenantIDOnly)
			if !ok || len(*rows) == 0 {
				return nil
			}
			teamIDs := make([]uuid.UUID, 0, len(*rows))
			for _, row := range *rows {
				if row.ID != uuid.Nil {
					teamIDs = append(teamIDs, row.ID)
				}
			}
			return s.RefreshTeams(teamIDs)
		}).Error
}

func (s *service) RefreshPlatformUser(userID uuid.UUID) error {
	if userID == uuid.Nil || s.platformService == nil {
		return nil
	}
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	for _, appKey := range appKeys {
		if _, refreshErr := s.platformService.RefreshSnapshot(userID, appKey); refreshErr != nil {
			return refreshErr
		}
	}
	return nil
}

func (s *service) RefreshPlatformUsers(userIDs []uuid.UUID) error {
	for _, userID := range dedupeUUIDs(userIDs) {
		if err := s.RefreshPlatformUser(userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) RefreshAllPlatformUsers() error {
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
			return s.RefreshPlatformUsers(userIDs)
		}).Error
}

func (s *service) RefreshPlatformRole(roleID uuid.UUID) error {
	if roleID == uuid.Nil {
		return nil
	}
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	if s.roleService != nil {
		for _, appKey := range appKeys {
			if _, refreshErr := s.roleService.RefreshSnapshot(roleID, appKey); refreshErr != nil {
				return refreshErr
			}
		}
	}
	userIDs, err := s.getPlatformUserIDsByRoleIDs([]uuid.UUID{roleID})
	if err != nil {
		return err
	}
	return s.RefreshPlatformUsers(userIDs)
}

func (s *service) RefreshPlatformRoles(roleIDs []uuid.UUID) error {
	dedupedRoleIDs := dedupeUUIDs(roleIDs)
	appKeys, err := s.listAppKeys()
	if err != nil {
		return err
	}
	if s.roleService != nil {
		for _, appKey := range appKeys {
			for _, roleID := range dedupedRoleIDs {
				if roleID == uuid.Nil {
					continue
				}
				if _, refreshErr := s.roleService.RefreshSnapshot(roleID, appKey); refreshErr != nil {
					return refreshErr
				}
			}
		}
	}
	userIDs, err := s.getPlatformUserIDsByRoleIDs(dedupedRoleIDs)
	if err != nil {
		return err
	}
	return s.RefreshPlatformUsers(userIDs)
}

func (s *service) RefreshAllPlatformRoles() error {
	if s.db == nil {
		return nil
	}
	type roleIDOnly struct {
		ID uuid.UUID
	}
	return s.db.Model(&models.Role{}).
		Select("id").
		Where("tenant_id IS NULL").
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
			return s.RefreshPlatformRoles(roleIDs)
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

	teamIDs, err := s.getTeamIDsByPackageIDs(impactedPackageIDs)
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
		if binding.TenantID == nil {
			platformRoleIDs = append(platformRoleIDs, binding.RoleID)
			continue
		}
		teamIDs = append(teamIDs, *binding.TenantID)
	}
	dedupedTeamIDs := dedupeUUIDs(teamIDs)
	dedupedRoleIDs := dedupeUUIDs(platformRoleIDs)

	roleUserIDs, err := s.getPlatformUserIDsByRoleIDs(dedupedRoleIDs)
	if err != nil {
		return stats, err
	}
	dedupedUserIDs := dedupeUUIDs(append(userIDs, roleUserIDs...))
	stats.TeamCount = len(dedupedTeamIDs)
	stats.RoleCount = len(dedupedRoleIDs)
	stats.UserCount = len(dedupedUserIDs)

	if err := s.RefreshPlatformRoles(dedupedRoleIDs); err != nil {
		return stats, err
	}
	if err := s.RefreshTeams(dedupedTeamIDs); err != nil {
		return stats, err
	}
	if err := s.RefreshPlatformUsers(dedupedUserIDs); err != nil {
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
	RoleID   uuid.UUID
	TenantID *uuid.UUID
}

type roleBindingRow struct {
	RoleID   uuid.UUID
	TenantID sql.NullString
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

func (s *service) getTeamIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var teamIDs []uuid.UUID
	err := s.db.Model(&models.TeamFeaturePackage{}).
		Where("package_id IN ? AND enabled = ?", packageIDs, true).
		Distinct("team_id").
		Pluck("team_id", &teamIDs).Error
	return teamIDs, err
}

func (s *service) getRoleBindingsByPackageIDs(packageIDs []uuid.UUID) ([]roleBinding, error) {
	if len(packageIDs) == 0 {
		return []roleBinding{}, nil
	}
	var rows []roleBindingRow
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Select("roles.id AS role_id, roles.tenant_id AS tenant_id").
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
		if row.TenantID.Valid && row.TenantID.String != "" {
			tenantID, parseErr := uuid.Parse(row.TenantID.String)
			if parseErr != nil {
				return nil, parseErr
			}
			binding.TenantID = &tenantID
		}
		result = append(result, binding)
	}
	return result, nil
}

func (s *service) getPlatformUserIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var userIDs []uuid.UUID
	err := s.db.Model(&models.UserFeaturePackage{}).
		Where("package_id IN ? AND enabled = ?", packageIDs, true).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

func (s *service) getPlatformUserIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var userIDs []uuid.UUID
	err := s.db.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.role_id IN ?", roleIDs).
		Where("user_roles.tenant_id IS NULL").
		Where("roles.tenant_id IS NULL").
		Distinct("user_roles.user_id").
		Pluck("user_roles.user_id", &userIDs).Error
	return userIDs, err
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
