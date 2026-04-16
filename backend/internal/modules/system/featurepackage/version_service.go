package featurepackage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/appscope"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
)

func (s *service) ListVersions(id uuid.UUID, current, size int) ([]user.FeaturePackageVersion, int64, error) {
	if _, err := s.packageRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrFeaturePackageNotFound
		}
		return nil, 0, err
	}
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	query := s.db.Model(&user.FeaturePackageVersion{}).Where("package_id = ?", id)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]user.FeaturePackageVersion, 0)
	if err := query.Order("version_no DESC").Offset((current - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) Rollback(id uuid.UUID, versionID uuid.UUID, operatorID *uuid.UUID, requestID string) (*permissionrefresh.RefreshStats, error) {
	pkg, err := s.packageRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeaturePackageNotFound
		}
		return nil, err
	}
	var version user.FeaturePackageVersion
	if err := s.db.Where("id = ? AND package_id = ?", versionID, id).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("版本不存在")
		}
		return nil, err
	}
	snapshot := version.Snapshot
	parseUUIDList := func(key string) []uuid.UUID {
		raw, ok := snapshot[key]
		if !ok {
			return []uuid.UUID{}
		}
		switch values := raw.(type) {
		case []interface{}:
			result := make([]uuid.UUID, 0, len(values))
			for _, item := range values {
				parsed, parseErr := uuid.Parse(strings.TrimSpace(fmt.Sprint(item)))
				if parseErr == nil {
					result = append(result, parsed)
				}
			}
			return result
		case []string:
			result := make([]uuid.UUID, 0, len(values))
			for _, item := range values {
				parsed, parseErr := uuid.Parse(strings.TrimSpace(item))
				if parseErr == nil {
					result = append(result, parsed)
				}
			}
			return result
		default:
			return []uuid.UUID{}
		}
	}

	childIDs := parseUUIDList("child_package_ids")
	actionIDs := parseUUIDList("action_ids")
	menuIDs := parseUUIDList("menu_ids")
	collaborationWorkspaceIDs := parseUUIDList("collaboration_workspace_ids")
	updates := map[string]interface{}{
		"package_key":     strings.TrimSpace(fmt.Sprint(snapshot["package_key"])),
		"package_type":    normalizePackageTypeDefault(fmt.Sprint(snapshot["package_type"]), pkg.PackageType),
		"name":            strings.TrimSpace(fmt.Sprint(snapshot["name"])),
		"description":     strings.TrimSpace(fmt.Sprint(snapshot["description"])),
		"workspace_scope": normalizeWorkspaceScopeDefault(fmt.Sprint(snapshot["workspace_scope"]), pkg.WorkspaceScope),
		"context_type":    normalizeContextTypeDefault(fmt.Sprint(snapshot["context_type"]), pkg.ContextType),
		"app_keys":        snapshot["app_keys"],
		"status":          normalizeStatus(fmt.Sprint(snapshot["status"])),
		"sort_order":      intFromSnapshot(snapshot["sort_order"], pkg.SortOrder),
		"updated_at":      time.Now(),
	}
	if err := s.packageRepo.UpdateWithMap(id, updates); err != nil {
		return nil, err
	}
	if err := s.packageBundleRepo.ReplaceChildPackages(id, childIDs); err != nil {
		return nil, err
	}
	if err := s.packageActionRepo.ReplacePackageKeys(id, actionIDs); err != nil {
		return nil, err
	}
	if err := s.packageMenuRepo.ReplacePackageMenus(id, menuIDs); err != nil {
		return nil, err
	}
	if err := s.syncPackageCollaborationWorkspacesBySet(id, collaborationWorkspaceIDs, operatorID); err != nil {
		return nil, err
	}

	var stats permissionrefresh.RefreshStats
	if s.refresher != nil {
		ref, refreshErr := s.refresher.RefreshByPackageWithStats(id)
		if refreshErr != nil {
			return nil, refreshErr
		}
		stats = ref
	}
	_ = s.saveVersionSnapshot(id, "rollback", operatorID, requestID)
	_ = s.recordRiskAudit("feature_package", id.String(), "rollback", nil, map[string]interface{}{
		"rollback_version_id": versionID.String(),
		"rollback_version_no": version.VersionNo,
	}, refreshStatsSummary(stats), operatorID, requestID)
	return &stats, nil
}

func (s *service) saveVersionSnapshot(packageID uuid.UUID, changeType string, operatorID *uuid.UUID, requestID string) error {
	pkg, err := s.packageRepo.GetByID(packageID)
	if err != nil {
		return err
	}
	childIDs, err := s.packageBundleRepo.GetChildPackageIDs(packageID)
	if err != nil {
		return err
	}
	actionIDs, err := s.packageActionRepo.GetKeyIDsByPackageID(packageID)
	if err != nil {
		return err
	}
	menuIDs, err := s.packageMenuRepo.GetMenuIDsByPackageID(packageID)
	if err != nil {
		return err
	}
	collaborationWorkspaceIDs, err := s.collaborationWorkspaceFeaturePackageRepo.GetCollaborationWorkspaceIDsByPackageID(packageID)
	if err != nil {
		return err
	}

	var maxVersion int64
	if err := s.db.Model(&user.FeaturePackageVersion{}).Where("package_id = ?", packageID).Select("COALESCE(MAX(version_no), 0)").Scan(&maxVersion).Error; err != nil {
		return err
	}
	item := &user.FeaturePackageVersion{
		PackageID:  packageID,
		VersionNo:  int(maxVersion) + 1,
		ChangeType: strings.TrimSpace(changeType),
		Snapshot: map[string]interface{}{
			"package_id":                  packageID.String(),
			"package_key":                 pkg.PackageKey,
			"package_type":                pkg.PackageType,
			"name":                        pkg.Name,
			"description":                 pkg.Description,
			"workspace_scope":             pkg.WorkspaceScope,
			"context_type":                pkg.ContextType,
			"app_keys":                    pkg.AppKeys,
			"status":                      pkg.Status,
			"sort_order":                  pkg.SortOrder,
			"child_package_ids":           uuidSliceToStrings(childIDs),
			"action_ids":                  uuidSliceToStrings(actionIDs),
			"menu_ids":                    uuidSliceToStrings(menuIDs),
			"collaboration_workspace_ids": uuidSliceToStrings(collaborationWorkspaceIDs),
			"snapshot_createdAt":          time.Now().Format(time.RFC3339),
		},
		OperatorID: operatorID,
		RequestID:  strings.TrimSpace(requestID),
	}
	return s.db.Create(item).Error
}

func (s *service) syncPackageCollaborationWorkspacesBySet(id uuid.UUID, collaborationWorkspaceIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	currentCollaborationWorkspaceIDs, err := s.collaborationWorkspaceFeaturePackageRepo.GetCollaborationWorkspaceIDsByPackageID(id)
	if err != nil {
		return err
	}
	item, err := s.packageRepo.GetByID(id)
	if err != nil {
		return err
	}
	desired := make(map[uuid.UUID]struct{}, len(collaborationWorkspaceIDs))
	affected := make(map[uuid.UUID]struct{}, len(currentCollaborationWorkspaceIDs)+len(collaborationWorkspaceIDs))
	for _, collaborationWorkspaceID := range currentCollaborationWorkspaceIDs {
		affected[collaborationWorkspaceID] = struct{}{}
	}
	for _, collaborationWorkspaceID := range collaborationWorkspaceIDs {
		desired[collaborationWorkspaceID] = struct{}{}
		affected[collaborationWorkspaceID] = struct{}{}
	}
	for collaborationWorkspaceID := range affected {
		packageIDs, packageErr := appscope.PackageIDsByCollaborationWorkspace(s.db, collaborationWorkspaceID, item.AppKey)
		if packageErr != nil {
			return packageErr
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
			nextPackageIDs = append(nextPackageIDs, id)
		}
		if err := appscope.ReplaceCollaborationWorkspacePackagesInApp(s.db, collaborationWorkspaceID, item.AppKey, nextPackageIDs, grantedBy); err != nil {
			return err
		}
	}
	return nil
}

func mergeRefreshStats(left, right permissionrefresh.RefreshStats) permissionrefresh.RefreshStats {
	result := permissionrefresh.RefreshStats{
		RequestedPackageCount:       left.RequestedPackageCount + right.RequestedPackageCount,
		ImpactedPackageCount:        left.ImpactedPackageCount + right.ImpactedPackageCount,
		RoleCount:                   left.RoleCount + right.RoleCount,
		CollaborationWorkspaceCount: left.CollaborationWorkspaceCount + right.CollaborationWorkspaceCount,
		UserCount:                   left.UserCount + right.UserCount,
		ElapsedMilliseconds:         left.ElapsedMilliseconds + right.ElapsedMilliseconds,
	}
	if right.FinishedAt.After(left.FinishedAt) {
		result.FinishedAt = right.FinishedAt
	} else {
		result.FinishedAt = left.FinishedAt
	}
	return result
}

func packageSummary(item *user.FeaturePackage) map[string]interface{} {
	if item == nil {
		return nil
	}
	return map[string]interface{}{
		"id":              item.ID.String(),
		"package_key":     item.PackageKey,
		"package_type":    item.PackageType,
		"name":            item.Name,
		"workspace_scope": item.WorkspaceScope,
		"context_type":    item.ContextType,
		"app_keys":        item.AppKeys,
		"status":          item.Status,
		"sort_order":      item.SortOrder,
	}
}

func packageSummaryFromUpdates(base *user.FeaturePackage, updates map[string]interface{}) map[string]interface{} {
	if base == nil {
		return nil
	}
	result := packageSummary(base)
	for key, value := range updates {
		if key == "updated_at" {
			continue
		}
		result[key] = value
	}
	return result
}

func refreshStatsSummary(stats permissionrefresh.RefreshStats) map[string]interface{} {
	return map[string]interface{}{
		"requested_package_count":       stats.RequestedPackageCount,
		"impacted_package_count":        stats.ImpactedPackageCount,
		"role_count":                    stats.RoleCount,
		"collaboration_workspace_count": stats.CollaborationWorkspaceCount,
		"user_count":                    stats.UserCount,
		"elapsed_milliseconds":          stats.ElapsedMilliseconds,
	}
}

func uuidSliceToStrings(items []uuid.UUID) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.String())
	}
	return result
}

func ginLikeMap(key string, values []uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		key: uuidSliceToStrings(values),
	}
}

func intFromSnapshot(value interface{}, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		parsed := strings.TrimSpace(v)
		if parsed == "" {
			return fallback
		}
		var num int
		_, err := fmt.Sscanf(parsed, "%d", &num)
		if err != nil {
			return fallback
		}
		return num
	default:
		return fallback
	}
}

