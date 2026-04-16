package featurepackage

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/pkg/appscope"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
	"github.com/maben/backend/internal/pkg/workspacefeaturebinding"
)

func (s *service) GetCollaborationWorkspacePackages(collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, []user.FeaturePackage, error) {
	packageIDs, err := appscope.PackageIDsByCollaborationWorkspace(s.db, collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.packageRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, nil, err
	}
	return filterPackagesForApp(packageIDs, items, normalizeAppKey(appKey))
}

func (s *service) GetPackageCollaborationWorkspaces(id uuid.UUID, appKey string) ([]uuid.UUID, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, ErrFeaturePackageNotFound
	}
	workspaceCollaborationWorkspaceIDs, err := s.getWorkspaceCollaborationWorkspaceIDsByPackageID(id, appKey)
	if err != nil {
		return nil, err
	}
	legacyCollaborationWorkspaceIDs, err := s.collaborationWorkspaceFeaturePackageRepo.GetCollaborationWorkspaceIDsByPackageID(id)
	if err != nil {
		return nil, err
	}
	return mergeUUIDSlice(workspaceCollaborationWorkspaceIDs, legacyCollaborationWorkspaceIDs), nil
}

func (s *service) SetPackageCollaborationWorkspaces(id uuid.UUID, collaborationWorkspaceIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	if !packageBelongsToApp(item, appKey) {
		return nil, ErrFeaturePackageNotFound
	}
	currentCollaborationWorkspaceIDs, err := s.GetPackageCollaborationWorkspaces(id, appKey)
	if err != nil {
		return nil, err
	}
	desired := make(map[uuid.UUID]struct{}, len(collaborationWorkspaceIDs))
	affected := make(map[uuid.UUID]struct{}, len(currentCollaborationWorkspaceIDs)+len(collaborationWorkspaceIDs))
	for _, collaborationWorkspaceID := range currentCollaborationWorkspaceIDs {
		affected[collaborationWorkspaceID] = struct{}{}
	}
	collaborationWorkspaceMap, err := s.getCollaborationWorkspaceMap(collaborationWorkspaceIDs)
	if err != nil {
		return nil, err
	}
	for _, collaborationWorkspaceID := range collaborationWorkspaceIDs {
		if _, ok := collaborationWorkspaceMap[collaborationWorkspaceID]; !ok {
			return nil, errors.New("存在无效的协作空间")
		}
		desired[collaborationWorkspaceID] = struct{}{}
		affected[collaborationWorkspaceID] = struct{}{}
	}
	for collaborationWorkspaceID := range affected {
		packageIDs, err := appscope.PackageIDsByCollaborationWorkspace(s.db, collaborationWorkspaceID, item.AppKey)
		if err != nil {
			return nil, err
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
		if _, ok := desired[collaborationWorkspaceID]; ok {
			if _, exists := seen[id]; !exists {
				nextPackageIDs = append(nextPackageIDs, id)
			}
		}
		if err := appscope.ReplaceCollaborationWorkspacePackagesInApp(s.db, collaborationWorkspaceID, item.AppKey, nextPackageIDs, grantedBy); err != nil {
			return nil, err
		}
		if err := s.replaceWorkspacePackagesForCollaborationWorkspaceID(collaborationWorkspaceID, item.AppKey, nextPackageIDs); err != nil {
			return nil, err
		}
		if s.refresher != nil {
			if err := s.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := s.boundaryService.RefreshSnapshot(collaborationWorkspaceID, item.AppKey); err != nil {
			return nil, err
		}
	}
	stats := permissionrefresh.RefreshStats{
		CollaborationWorkspaceCount: len(affected),
	}
	_ = s.saveVersionSnapshot(id, "set_collaboration_workspaces", grantedBy, "")
	_ = s.recordRiskAudit("feature_package", id.String(), "set_collaboration_workspaces", nil, ginLikeMap("collaboration_workspace_ids", collaborationWorkspaceIDs), refreshStatsSummary(stats), grantedBy, "")
	return &stats, nil
}

func (s *service) SetCollaborationWorkspacePackages(collaborationWorkspaceID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID, appKey string) (*permissionrefresh.RefreshStats, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	packageMap, err := s.getPackageMap(packageIDs)
	if err != nil {
		return nil, err
	}
	for _, packageID := range packageIDs {
		item, ok := packageMap[packageID]
		if !ok {
			return nil, ErrFeaturePackageNotFound
		}
		if !packageBelongsToApp(&item, normalizedAppKey) {
			return nil, errors.New("功能包不在当前 App 的适用范围内")
		}
	}
	if err := appscope.ReplaceCollaborationWorkspacePackagesInApp(s.db, collaborationWorkspaceID, normalizedAppKey, packageIDs, grantedBy); err != nil {
		return nil, err
	}
	if err := s.replaceWorkspacePackagesForCollaborationWorkspaceID(collaborationWorkspaceID, normalizedAppKey, packageIDs); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			return nil, err
		}
		stats := permissionrefresh.RefreshStats{CollaborationWorkspaceCount: 1}
		_ = s.recordRiskAudit("collaboration_workspace_feature_package", collaborationWorkspaceID.String(), "set_collaboration_workspace_packages", nil, ginLikeMap("package_ids", packageIDs), refreshStatsSummary(stats), grantedBy, "")
		return &stats, nil
	}
	_, err = s.boundaryService.RefreshSnapshot(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	stats := permissionrefresh.RefreshStats{CollaborationWorkspaceCount: 1}
	_ = s.recordRiskAudit("collaboration_workspace_feature_package", collaborationWorkspaceID.String(), "set_collaboration_workspace_packages", nil, ginLikeMap("package_ids", packageIDs), refreshStatsSummary(stats), grantedBy, "")
	return &stats, nil
}

func (s *service) GetRelationTree(workspaceScope, keyword string) (*FeaturePackageRelationTree, error) {
	packages, err := s.ListOptions(&dto.FeaturePackageListRequest{
		WorkspaceScope: normalizeWorkspaceScope(workspaceScope),
		Keyword:        strings.TrimSpace(keyword),
	})
	if err != nil {
		return nil, err
	}
	nodes := make(map[uuid.UUID]FeaturePackageRelationNode, len(packages))
	for _, pkg := range packages {
		nodes[pkg.ID] = FeaturePackageRelationNode{
			ID:             pkg.ID,
			PackageKey:     pkg.PackageKey,
			Name:           pkg.Name,
			PackageType:    pkg.PackageType,
			WorkspaceScope: pkg.WorkspaceScope,
			AppKeys:        pkg.AppKeys,
			Status:         pkg.Status,
		}
	}
	if len(nodes) == 0 {
		return &FeaturePackageRelationTree{
			Roots:             []FeaturePackageRelationNode{},
			CycleDependencies: [][]string{},
			IsolatedBaseKeys:  []string{},
		}, nil
	}

	type relationRow struct {
		PackageID      uuid.UUID
		ChildPackageID uuid.UUID
	}
	rows := make([]relationRow, 0)
	packageIDs := make([]uuid.UUID, 0, len(nodes))
	for id := range nodes {
		packageIDs = append(packageIDs, id)
	}
	if err := s.db.Model(&user.FeaturePackageBundle{}).
		Select("package_id", "child_package_id").
		Where("package_id IN ? AND child_package_id IN ?", packageIDs, packageIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	childrenMap := make(map[uuid.UUID][]uuid.UUID, len(rows))
	parentCount := make(map[uuid.UUID]int, len(nodes))
	for _, row := range rows {
		childrenMap[row.PackageID] = append(childrenMap[row.PackageID], row.ChildPackageID)
		parentCount[row.ChildPackageID]++
	}
	for id, node := range nodes {
		node.ReferenceCount = parentCount[id]
		nodes[id] = node
	}

	visited := make(map[uuid.UUID]bool, len(nodes))
	pathMark := make(map[uuid.UUID]bool, len(nodes))
	cycleSet := make(map[string]struct{})
	cycles := make([][]string, 0)
	var buildTree func(id uuid.UUID, path []uuid.UUID) FeaturePackageRelationNode
	buildTree = func(id uuid.UUID, path []uuid.UUID) FeaturePackageRelationNode {
		node := nodes[id]
		if pathMark[id] {
			cycle := append(path, id)
			keys := make([]string, 0, len(cycle))
			for _, item := range cycle {
				if value, ok := nodes[item]; ok {
					keys = append(keys, value.PackageKey)
				}
			}
			signature := strings.Join(keys, " -> ")
			if signature != "" {
				if _, exists := cycleSet[signature]; !exists {
					cycleSet[signature] = struct{}{}
					cycles = append(cycles, keys)
				}
			}
			return node
		}
		pathMark[id] = true
		visited[id] = true
		children := childrenMap[id]
		if len(children) > 0 {
			node.Children = make([]FeaturePackageRelationNode, 0, len(children))
			for _, childID := range children {
				if _, ok := nodes[childID]; !ok {
					continue
				}
				node.Children = append(node.Children, buildTree(childID, append(path, id)))
			}
		}
		pathMark[id] = false
		return node
	}

	roots := make([]FeaturePackageRelationNode, 0)
	for id := range nodes {
		if parentCount[id] == 0 {
			roots = append(roots, buildTree(id, nil))
		}
	}
	for id := range nodes {
		if visited[id] {
			continue
		}
		roots = append(roots, buildTree(id, nil))
	}

	isolatedBaseKeys := make([]string, 0)
	for id, node := range nodes {
		if node.PackageType != "base" {
			continue
		}
		if len(childrenMap[id]) > 0 {
			continue
		}
		if parentCount[id] > 0 {
			continue
		}
		isolatedBaseKeys = append(isolatedBaseKeys, node.PackageKey)
	}

	return &FeaturePackageRelationTree{
		Roots:             roots,
		CycleDependencies: cycles,
		IsolatedBaseKeys:  isolatedBaseKeys,
	}, nil
}

func (s *service) GetImpactPreview(id uuid.UUID) (*FeaturePackageImpactPreview, error) {
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}

	result := &FeaturePackageImpactPreview{PackageID: id}
	type countRow struct {
		Total int64
	}

	var row countRow
	if err := s.db.Model(&user.RoleFeaturePackage{}).Where("package_id = ? AND enabled = ?", id, true).Distinct("role_id").Count(&row.Total).Error; err != nil {
		return nil, err
	}
	result.RoleCount = row.Total
	workspaceCollaborationWorkspaceIDs, err := workspacefeaturebinding.ListCollaborationWorkspaceIDsByPackageIDs(s.db, []uuid.UUID{id}, item.AppKey)
	if err != nil {
		return nil, err
	}
	var legacyCollaborationWorkspaceIDs []uuid.UUID
	if err := s.db.Model(&user.CollaborationWorkspaceFeaturePackage{}).Where("package_id = ? AND enabled = ?", id, true).Distinct("collaboration_workspace_id").Pluck("collaboration_workspace_id", &legacyCollaborationWorkspaceIDs).Error; err != nil {
		return nil, err
	}
	result.CollaborationWorkspaceCount = int64(len(mergeUUIDSlice(workspaceCollaborationWorkspaceIDs, legacyCollaborationWorkspaceIDs)))
	workspaceUserIDs, err := workspacefeaturebinding.ListPlatformUserIDsByPackageIDs(s.db, []uuid.UUID{id}, item.AppKey)
	if err != nil {
		return nil, err
	}
	var legacyUserIDs []uuid.UUID
	if err := s.db.Model(&user.UserFeaturePackage{}).Where("package_id = ? AND enabled = ?", id, true).Distinct("user_id").Pluck("user_id", &legacyUserIDs).Error; err != nil {
		return nil, err
	}
	result.UserCount = int64(len(mergeUUIDSlice(workspaceUserIDs, legacyUserIDs)))
	if err := s.db.Model(&user.FeaturePackageMenu{}).Where("package_id = ?", id).Count(&row.Total).Error; err != nil {
		return nil, err
	}
	result.MenuCount = row.Total
	if err := s.db.Model(&user.FeaturePackageKey{}).Where("package_id = ?", id).Count(&row.Total).Error; err != nil {
		return nil, err
	}
	result.ActionCount = row.Total
	return result, nil
}

func (s *service) getWorkspacePackageIDsByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var workspace models.Workspace
	if err := s.db.
		Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, collaborationWorkspaceID).
		First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}

	var packageIDs []uuid.UUID
	if err := s.db.Model(&models.WorkspaceFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
		Where("workspace_feature_packages.workspace_id = ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", workspace.ID, true).
		Where("feature_packages.deleted_at IS NULL").
		Where(func(tx *gorm.DB) *gorm.DB {
			requestedAppKey := normalizeAppKey(appKey)
			if requestedAppKey == "" {
				return tx
			}
			return tx.Where(
				"feature_packages.app_key = ? OR COALESCE(jsonb_array_length(feature_packages.app_keys), 0) = 0 OR jsonb_exists(feature_packages.app_keys, ?)",
				requestedAppKey,
				requestedAppKey,
			)
		}).
		Distinct("workspace_feature_packages.package_id").
		Pluck("package_id", &packageIDs).Error; err != nil {
		return nil, err
	}
	return packageIDs, nil
}

func (s *service) getWorkspaceCollaborationWorkspaceIDsByPackageID(packageID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	type row struct {
		SourceCollaborationWorkspaceID *uuid.UUID `gorm:"column:collaboration_workspace_id"`
	}
	rows := make([]row, 0)
	if err := s.db.Model(&models.WorkspaceFeaturePackage{}).
		Select("workspaces.collaboration_workspace_id").
		Joins("JOIN workspaces ON workspaces.id = workspace_feature_packages.workspace_id").
		Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
		Where("workspace_feature_packages.package_id = ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", packageID, true).
		Where("workspaces.workspace_type = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypeCollaboration).
		Where("feature_packages.deleted_at IS NULL").
		Where(func(tx *gorm.DB) *gorm.DB {
			requestedAppKey := normalizeAppKey(appKey)
			if requestedAppKey == "" {
				return tx
			}
			return tx.Where(
				"feature_packages.app_key = ? OR COALESCE(jsonb_array_length(feature_packages.app_keys), 0) = 0 OR jsonb_exists(feature_packages.app_keys, ?)",
				requestedAppKey,
				requestedAppKey,
			)
		}).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	collaborationWorkspaceIDs := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, item := range rows {
		if item.SourceCollaborationWorkspaceID == nil || *item.SourceCollaborationWorkspaceID == uuid.Nil {
			continue
		}
		if _, exists := seen[*item.SourceCollaborationWorkspaceID]; exists {
			continue
		}
		seen[*item.SourceCollaborationWorkspaceID] = struct{}{}
		collaborationWorkspaceIDs = append(collaborationWorkspaceIDs, *item.SourceCollaborationWorkspaceID)
	}
	return collaborationWorkspaceIDs, nil
}

func (s *service) replaceWorkspacePackagesForCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID, appKey string, packageIDs []uuid.UUID) error {
	var workspace models.Workspace
	if err := s.db.
		Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, collaborationWorkspaceID).
		First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	scopedIDs, err := appscope.FilterPackageIDs(s.db, appKey, packageIDs)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workspace_id = ?", workspace.ID).Delete(&models.WorkspaceFeaturePackage{}).Error; err != nil {
			return err
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		items := make([]models.WorkspaceFeaturePackage, 0, len(scopedIDs))
		for _, packageID := range scopedIDs {
			items = append(items, models.WorkspaceFeaturePackage{
				WorkspaceID: workspace.ID,
				PackageID:   packageID,
				Enabled:     true,
			})
		}
		return tx.Create(&items).Error
	})
}

func (s *service) getCollaborationWorkspaceMap(collaborationWorkspaceIDs []uuid.UUID) (map[uuid.UUID]struct{}, error) {
	items, err := s.collaborationWorkspaceRepo.GetByIDs(collaborationWorkspaceIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		result[item.ID] = struct{}{}
	}
	return result, nil
}

