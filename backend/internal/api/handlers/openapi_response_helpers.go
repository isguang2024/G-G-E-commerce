package handlers

import (
	"strings"

	"github.com/gg-ecommerce/backend/api/gen"
	permissionpkg "github.com/gg-ecommerce/backend/internal/modules/system/permission"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

func permissionKeySegments(permissionKey string) (string, string) {
	normalized := permissionKey
	if normalized == "" {
		return "", ""
	}
	if strings.Contains(normalized, ":") {
		parts := strings.SplitN(normalized, ":", 2)
		normalized = strings.Join(parts, ".")
	}
	parts := strings.Split(normalized, ".")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	if len(filtered) <= 1 {
		return "", ""
	}
	return strings.Join(filtered[:len(filtered)-1], "_"), filtered[len(filtered)-1]
}

func collaborationWorkspaceItemFromModel(cw user.CollaborationWorkspace) gen.CollaborationWorkspaceItem {
	return gen.CollaborationWorkspaceItem{
		ID:         cw.ID,
		Name:       cw.Name,
		Remark:     cw.Remark,
		LogoURL:    cw.LogoURL,
		Plan:       cw.Plan,
		OwnerID:    cw.OwnerID,
		MaxMembers: cw.MaxMembers,
		Status:     cw.Status,
		CreatedAt:  cw.CreatedAt,
		UpdatedAt:  cw.UpdatedAt,
	}
}

func collaborationWorkspaceItemsFromModels(items []user.CollaborationWorkspace) []gen.CollaborationWorkspaceItem {
	out := make([]gen.CollaborationWorkspaceItem, 0, len(items))
	for i := range items {
		out = append(out, collaborationWorkspaceItemFromModel(items[i]))
	}
	return out
}

func collaborationWorkspaceMemberItemFromModel(member user.CollaborationWorkspaceMember) gen.CollaborationWorkspaceMemberItem {
	item := gen.CollaborationWorkspaceMemberItem{
		ID:                       member.ID,
		CollaborationWorkspaceID: member.CollaborationWorkspaceID,
		UserID:                   member.UserID,
		RoleCode:                 member.RoleCode,
		Status:                   member.Status,
		JoinedAt:                 member.JoinedAt,
		CreatedAt:                member.CreatedAt,
		UpdatedAt:                member.UpdatedAt,
	}
	if member.RoleID != nil {
		item.RoleID = gen.NewOptNilUUID(*member.RoleID)
	}
	if member.InvitedBy != nil {
		item.InvitedBy = gen.NewOptNilUUID(*member.InvitedBy)
	}
	return item
}

func collaborationWorkspaceMemberItemsFromModels(items []user.CollaborationWorkspaceMember) []gen.CollaborationWorkspaceMemberItem {
	out := make([]gen.CollaborationWorkspaceMemberItem, 0, len(items))
	for i := range items {
		out = append(out, collaborationWorkspaceMemberItemFromModel(items[i]))
	}
	return out
}

func permissionActionRefsFromModels(items []user.PermissionKey) []gen.PermissionActionRef {
	out := make([]gen.PermissionActionRef, 0, len(items))
	for i := range items {
		item := items[i]
		ref := gen.PermissionActionRef{
			ID:        item.ID,
			ActionKey: item.PermissionKey,
			Name:      item.Name,
		}
		if item.Description != "" {
			ref.Description = gen.NewOptNilString(item.Description)
		}
		if item.Status != "" {
			ref.Status = gen.NewOptNilString(item.Status)
		}
		out = append(out, ref)
	}
	return out
}

func permissionActionItemFromPermissionKey(item user.PermissionKey) gen.PermissionActionItem {
	resourceCode, actionCode := permissionKeySegments(item.PermissionKey)
	moduleCode := strings.TrimSpace(item.ModuleCode)
	if item.ModuleGroup != nil && strings.TrimSpace(item.ModuleGroup.Code) != "" {
		moduleCode = strings.TrimSpace(item.ModuleGroup.Code)
	}
	if moduleCode == "" {
		moduleCode = resourceCode
	}
	featureKind := strings.TrimSpace(item.FeatureKind)
	if item.FeatureGroup != nil && strings.TrimSpace(item.FeatureGroup.Code) != "" {
		featureKind = strings.TrimSpace(item.FeatureGroup.Code)
	}
	if featureKind == "" {
		featureKind = "system"
	}
	out := gen.PermissionActionItem{
		ID:                   item.ID,
		ResourceCode:         resourceCode,
		ActionCode:           actionCode,
		ModuleCode:           moduleCode,
		PermissionKey:        item.PermissionKey,
		FeatureKind:          featureKind,
		DataPolicy:           item.DataPolicy,
		Name:                 item.Name,
		Description:          gen.NewOptString(item.Description),
		DataPermissionCode:   item.AppKey,
		DataPermissionName:   item.ContextType,
		Status:               item.Status,
		SortOrder:            int64(item.SortOrder),
		IsBuiltin:            item.IsBuiltin,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
	if item.ModuleGroupID != nil {
		out.ModuleGroupID = gen.NewOptNilUUID(*item.ModuleGroupID)
	}
	if item.FeatureGroupID != nil {
		out.FeatureGroupID = gen.NewOptNilUUID(*item.FeatureGroupID)
	}
	return out
}

func permissionActionOptionItemsFromModels(items []user.PermissionKey) []gen.PermissionActionItem {
	out := make([]gen.PermissionActionItem, 0, len(items))
	for i := range items {
		out = append(out, permissionActionItemFromPermissionKey(items[i]))
	}
	return out
}

func permissionActionItemFromListItem(item permissionpkg.PermissionListItem) gen.PermissionActionItem {
	out := permissionActionItemFromPermissionKey(item.PermissionKey)
	profile := item.PermissionAuditProfile
	out.APICount = profile.APICount
	out.PageCount = profile.PageCount
	out.PackageCount = profile.PackageCount
	out.ConsumerTypes = append([]string{}, profile.ConsumerTypes...)
	out.UsagePattern = profile.UsagePattern
	out.UsageNote = profile.UsageNote
	out.DuplicatePattern = profile.DuplicatePattern
	out.DuplicateGroup = profile.DuplicateGroup
	out.DuplicateKeys = append([]string{}, profile.DuplicateKeys...)
	out.DuplicateNote = profile.DuplicateNote
	return out
}

func permissionActionListItemsFromModels(items []permissionpkg.PermissionListItem) []gen.PermissionActionItem {
	out := make([]gen.PermissionActionItem, 0, len(items))
	for i := range items {
		out = append(out, permissionActionItemFromListItem(items[i]))
	}
	return out
}

func permissionActionAuditSummaryFromModel(summary permissionpkg.PermissionAuditSummary) gen.PermissionActionAuditSummary {
	return gen.PermissionActionAuditSummary{
		TotalCount:              summary.TotalCount,
		UnusedCount:             summary.UnusedCount,
		APIOnlyCount:            summary.APIOnlyCount,
		PageOnlyCount:           summary.PageOnlyCount,
		PackageOnlyCount:        summary.PackageOnlyCount,
		MultiConsumerCount:      summary.MultiConsumerCount,
		CrossContextMirrorCount: summary.CrossContextMirrorCount,
		SuspectedDuplicateCount: summary.SuspectedDuplicateCount,
	}
}

func permissionGroupItemFromModel(item user.PermissionGroup) gen.PermissionGroupItem {
	return gen.PermissionGroupItem{
		ID:          item.ID,
		GroupType:   item.GroupType,
		Code:        item.Code,
		Name:        item.Name,
		NameEn:      gen.NewOptString(item.NameEn),
		Description: gen.NewOptString(item.Description),
		Status:      item.Status,
		SortOrder:   int64(item.SortOrder),
		IsBuiltin:   item.IsBuiltin,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func permissionGroupItemsFromModels(items []user.PermissionGroup) []gen.PermissionGroupItem {
	out := make([]gen.PermissionGroupItem, 0, len(items))
	for i := range items {
		out = append(out, permissionGroupItemFromModel(items[i]))
	}
	return out
}

func permissionBatchTemplateItemFromModel(item user.PermissionBatchTemplate) gen.PermissionActionBatchTemplateItem {
	out := gen.PermissionActionBatchTemplateItem{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Payload:     marshalAnyObject(item.Payload),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
	if item.CreatedBy != nil {
		out.CreatedBy = gen.NewOptNilUUID(*item.CreatedBy)
	}
	return out
}

func permissionBatchTemplateItemsFromModels(items []user.PermissionBatchTemplate) []gen.PermissionActionBatchTemplateItem {
	out := make([]gen.PermissionActionBatchTemplateItem, 0, len(items))
	for i := range items {
		out = append(out, permissionBatchTemplateItemFromModel(items[i]))
	}
	return out
}

func riskAuditItemFromModel(item user.RiskOperationAudit) gen.RiskAuditItem {
	out := gen.RiskAuditItem{
		ID:            item.ID,
		ObjectType:    item.ObjectType,
		ObjectID:      item.ObjectID,
		OperationType: item.OperationType,
		BeforeSummary: marshalAnyObject(item.BeforeSummary),
		AfterSummary:  marshalAnyObject(item.AfterSummary),
		ImpactSummary: marshalAnyObject(item.ImpactSummary),
		RequestID:     item.RequestID,
		CreatedAt:     item.CreatedAt,
	}
	if item.OperatorID != nil {
		out.OperatorID = gen.NewOptNilUUID(*item.OperatorID)
	}
	return out
}

func riskAuditItemsFromModels(items []user.RiskOperationAudit) []gen.RiskAuditItem {
	out := make([]gen.RiskAuditItem, 0, len(items))
	for i := range items {
		out = append(out, riskAuditItemFromModel(items[i]))
	}
	return out
}

func collaborationWorkspaceRoleItemFromModel(item user.Role) gen.CollaborationWorkspaceRoleItem {
	out := gen.CollaborationWorkspaceRoleItem{
		ID:                       item.ID,
		Code:                     item.Code,
		Name:                     item.Name,
		Description:              item.Description,
		Status:                   item.Status,
		IsSystem:                 item.IsSystem,
		IsGlobal:                 item.CollaborationWorkspaceID == nil,
		Priority:                 gen.NewOptInt64(int64(item.Priority)),
		SortOrder:                gen.NewOptInt64(int64(item.SortOrder)),
		CreatedAt:                item.CreatedAt,
		UpdatedAt:                gen.NewOptDateTime(item.UpdatedAt),
	}
	if item.CollaborationWorkspaceID != nil {
		out.CollaborationWorkspaceID = gen.NewNilUUID(*item.CollaborationWorkspaceID)
	}
	return out
}

func collaborationWorkspaceRoleItemsFromModels(items []user.Role) []gen.CollaborationWorkspaceRoleItem {
	out := make([]gen.CollaborationWorkspaceRoleItem, 0, len(items))
	for i := range items {
		out = append(out, collaborationWorkspaceRoleItemFromModel(items[i]))
	}
	return out
}

func permissionActionConsumersResponseFromModel(details *permissionpkg.PermissionConsumerDetails) *gen.PermissionActionConsumersResponse {
	if details == nil {
		return &gen.PermissionActionConsumersResponse{
			PermissionKey:   "",
			Apis:            []gen.PermissionActionConsumerAPIItem{},
			Pages:           []gen.PermissionActionConsumerPageItem{},
			FeaturePackages: []gen.PermissionActionConsumerPackageItem{},
			Roles:           []gen.PermissionActionConsumerRoleItem{},
		}
	}

	apis := make([]gen.PermissionActionConsumerAPIItem, 0, len(details.APIs))
	for _, item := range details.APIs {
		apis = append(apis, gen.PermissionActionConsumerAPIItem{
			Code:    item.Code,
			Method:  item.Method,
			Path:    item.Path,
			Summary: item.Summary,
		})
	}

	pages := make([]gen.PermissionActionConsumerPageItem, 0, len(details.Pages))
	for _, item := range details.Pages {
		pages = append(pages, gen.PermissionActionConsumerPageItem{
			PageKey:    item.PageKey,
			Name:       item.Name,
			RoutePath:  item.RoutePath,
			AccessMode: item.AccessMode,
		})
	}

	packages := make([]gen.PermissionActionConsumerPackageItem, 0, len(details.FeaturePkgs))
	for _, item := range details.FeaturePkgs {
		packages = append(packages, gen.PermissionActionConsumerPackageItem{
			ID:          item.ID,
			PackageKey:  item.PackageKey,
			Name:        item.Name,
			PackageType: item.PackageType,
			ContextType: item.ContextType,
		})
	}

	roles := make([]gen.PermissionActionConsumerRoleItem, 0, len(details.Roles))
	for _, item := range details.Roles {
		roles = append(roles, gen.PermissionActionConsumerRoleItem{
			ID:          item.ID,
			Code:        item.Code,
			Name:        item.Name,
			ContextType: item.ContextType,
		})
	}

	return &gen.PermissionActionConsumersResponse{
		PermissionKey:   details.PermissionKey,
		Apis:            apis,
		Pages:           pages,
		FeaturePackages: packages,
		Roles:           roles,
	}
}

func permissionActionImpactPreviewFromModel(preview *permissionpkg.PermissionImpactPreview) *gen.PermissionActionImpactPreview {
	if preview == nil {
		return &gen.PermissionActionImpactPreview{}
	}
	return &gen.PermissionActionImpactPreview{
		PermissionKey:               preview.PermissionKey,
		APICount:                    preview.APICount,
		PageCount:                   preview.PageCount,
		PackageCount:                preview.PackageCount,
		RoleCount:                   preview.RoleCount,
		CollaborationWorkspaceCount: preview.CollaborationWorkspaceCount,
		UserCount:                   preview.UserCount,
	}
}
