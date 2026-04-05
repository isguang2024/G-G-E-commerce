package page

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	spaceutil "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

const runtimePageCacheTTL = 24 * time.Hour

type runtimePageCacheSnapshot struct {
	all       []Record
	public    []Record
	menuMap   map[uuid.UUID]runtimeMenuNode
	expiresAt time.Time
}

type runtimePageCacheStore struct {
	mu        sync.RWMutex
	snapshots map[string]*runtimePageCacheSnapshot
}

type CompiledAccessContext struct {
	SpaceKey       string
	Authenticated  bool
	SuperAdmin     bool
	ActionKeys     map[string]struct{}
	VisibleMenuIDs map[uuid.UUID]struct{}
}

func (ctx *CompiledAccessContext) VisibleMenuIDList() []uuid.UUID {
	if ctx == nil || len(ctx.VisibleMenuIDs) == 0 {
		return []uuid.UUID{}
	}
	result := make([]uuid.UUID, 0, len(ctx.VisibleMenuIDs))
	for id := range ctx.VisibleMenuIDs {
		result = append(result, id)
	}
	return result
}

var globalRuntimePageCache runtimePageCacheStore

func InvalidateRuntimeCache() {
	globalRuntimePageCache.invalidate()
}

func (s *service) loadRuntimeRecords(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) ([]Record, error) {
	return globalRuntimePageCache.get(s, normalizeAppKey(appKey), false, host, requestedSpaceKey, userID, tenantID)
}

func (s *service) loadPublicRuntimeRecords(appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) ([]Record, error) {
	return globalRuntimePageCache.get(s, normalizeAppKey(appKey), true, host, requestedSpaceKey, userID, tenantID)
}

func (s *service) buildCompiledAccessContextForSpace(appKey, spaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*CompiledAccessContext, error) {
	resolvedSpaceKey := normalizeSpaceKey(spaceKey)
	menuMap, err := s.loadMenuMap(normalizeAppKey(appKey), resolvedSpaceKey)
	if err != nil {
		return nil, err
	}
	return s.buildRuntimeAccessContext(normalizeAppKey(appKey), userID, tenantID, menuMap, resolvedSpaceKey)
}

func (s *service) loadRuntimeRecordsWithAccess(appKey, spaceKey string, accessCtx *CompiledAccessContext) ([]Record, error) {
	resolvedSpaceKey := normalizeSpaceKey(spaceKey)
	snapshot, err := globalRuntimePageCache.getSnapshot(s, normalizeAppKey(appKey), resolvedSpaceKey)
	if err != nil {
		return nil, err
	}
	if accessCtx == nil {
		return []Record{}, nil
	}
	filtered, err := filterRuntimeRecordsWithAccess(snapshot.all, snapshot.menuMap, accessCtx, resolvedSpaceKey)
	if err != nil {
		return nil, err
	}
	return mergeRuntimeRecordsByOrder(snapshot.all, filtered, snapshot.public), nil
}

func (c *runtimePageCacheStore) get(s *service, appKey string, publicOnly bool, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) ([]Record, error) {
	now := time.Now()
	resolvedSpaceKey := resolveSpaceKeyForRuntime(s, appKey, host, requestedSpaceKey, userID, tenantID)
	cacheKey := runtimePageCacheKey(appKey, resolvedSpaceKey)

	c.mu.RLock()
	snapshot := c.snapshots[cacheKey]
	if snapshot != nil && now.Before(snapshot.expiresAt) {
		records := snapshot.all
		if publicOnly {
			records = snapshot.public
		}
		c.mu.RUnlock()
		if publicOnly {
			return cloneRuntimeRecords(records), nil
		}
		filtered, err := filterRuntimeRecordsForContext(s, appKey, records, snapshot.menuMap, userID, tenantID, resolvedSpaceKey)
		if err != nil {
			return nil, err
		}
		return mergeRuntimeRecordsByOrder(records, filtered, snapshot.public), nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.snapshots == nil {
		c.snapshots = make(map[string]*runtimePageCacheSnapshot)
	}
	snapshot = c.snapshots[cacheKey]
	if snapshot != nil && now.Before(snapshot.expiresAt) {
		records := snapshot.all
		if publicOnly {
			records = snapshot.public
		}
		if publicOnly {
			return cloneRuntimeRecords(records), nil
		}
		filtered, err := filterRuntimeRecordsForContext(s, appKey, records, snapshot.menuMap, userID, tenantID, resolvedSpaceKey)
		if err != nil {
			return nil, err
		}
		return mergeRuntimeRecordsByOrder(records, filtered, snapshot.public), nil
	}

	all, menuMap, err := s.buildRuntimeRecords(appKey, resolvedSpaceKey)
	if err != nil {
		return nil, err
	}
	publicRecords := buildPublicRuntimeRecords(all, menuMap, resolvedSpaceKey)
	c.snapshots[cacheKey] = &runtimePageCacheSnapshot{
		all:       cloneRuntimeRecords(all),
		public:    cloneRuntimeRecords(publicRecords),
		menuMap:   cloneRuntimeMenuMap(menuMap),
		expiresAt: now.Add(runtimePageCacheTTL),
	}

	if publicOnly {
		return cloneRuntimeRecords(publicRecords), nil
	}
	filtered, err := filterRuntimeRecordsForContext(s, appKey, all, menuMap, userID, tenantID, resolvedSpaceKey)
	if err != nil {
		return nil, err
	}
	return mergeRuntimeRecordsByOrder(all, filtered, publicRecords), nil
}

func (c *runtimePageCacheStore) getSnapshot(s *service, appKey string, spaceKey string) (*runtimePageCacheSnapshot, error) {
	resolvedSpaceKey := normalizeSpaceKey(spaceKey)
	cacheKey := runtimePageCacheKey(appKey, resolvedSpaceKey)
	now := time.Now()

	c.mu.RLock()
	snapshot := c.snapshots[cacheKey]
	if snapshot != nil && now.Before(snapshot.expiresAt) {
		c.mu.RUnlock()
		return snapshot, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.snapshots == nil {
		c.snapshots = make(map[string]*runtimePageCacheSnapshot)
	}
	snapshot = c.snapshots[cacheKey]
	if snapshot != nil && now.Before(snapshot.expiresAt) {
		return snapshot, nil
	}

	all, menuMap, err := s.buildRuntimeRecords(appKey, resolvedSpaceKey)
	if err != nil {
		return nil, err
	}
	publicRecords := buildPublicRuntimeRecords(all, menuMap, resolvedSpaceKey)
	snapshot = &runtimePageCacheSnapshot{
		all:       cloneRuntimeRecords(all),
		public:    cloneRuntimeRecords(publicRecords),
		menuMap:   cloneRuntimeMenuMap(menuMap),
		expiresAt: now.Add(runtimePageCacheTTL),
	}
	c.snapshots[cacheKey] = snapshot
	return snapshot, nil
}

func (c *runtimePageCacheStore) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshots = nil
}

func cloneRuntimeRecords(items []Record) []Record {
	if len(items) == 0 {
		return []Record{}
	}
	result := make([]Record, len(items))
	copy(result, items)
	return result
}

func cloneRuntimeMenuMap(items map[uuid.UUID]runtimeMenuNode) map[uuid.UUID]runtimeMenuNode {
	if len(items) == 0 {
		return map[uuid.UUID]runtimeMenuNode{}
	}
	result := make(map[uuid.UUID]runtimeMenuNode, len(items))
	for id, item := range items {
		result[id] = item
	}
	return result
}

func mergeRuntimeRecordsByOrder(all []Record, primary []Record, secondary []Record) []Record {
	if len(all) == 0 {
		return []Record{}
	}
	keep := make(map[string]struct{}, len(primary)+len(secondary))
	for _, item := range primary {
		key := strings.TrimSpace(item.PageKey)
		if key != "" {
			keep[key] = struct{}{}
		}
	}
	for _, item := range secondary {
		key := strings.TrimSpace(item.PageKey)
		if key != "" {
			keep[key] = struct{}{}
		}
	}
	if len(keep) == 0 {
		return []Record{}
	}
	result := make([]Record, 0, len(keep))
	for _, item := range all {
		if _, ok := keep[strings.TrimSpace(item.PageKey)]; ok {
			result = append(result, item)
		}
	}
	return result
}

func buildPublicRuntimeRecords(
	all []Record,
	menuMap map[uuid.UUID]runtimeMenuNode,
	spaceKey string,
) []Record {
	if len(all) == 0 {
		return []Record{}
	}
	if normalized := normalizeSpaceKey(spaceKey); normalized != "" {
		all = filterRecordsBySpace(all, normalized)
	}

	pageMap := make(map[string]Record, len(all))
	for _, item := range all {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		pageMap[pageKey] = item
	}

	included := make(map[string]struct{}, len(all))
	accessCache := make(map[string]string, len(all))
	resolving := make(map[string]struct{})

	var resolveAccessMode func(pageKey string) string
	resolveAccessMode = func(pageKey string) string {
		normalizedKey := strings.TrimSpace(pageKey)
		if normalizedKey == "" {
			return "jwt"
		}
		if cached, ok := accessCache[normalizedKey]; ok {
			return cached
		}
		if _, ok := resolving[normalizedKey]; ok {
			return "jwt"
		}

		item, ok := pageMap[normalizedKey]
		if !ok {
			return "jwt"
		}

		resolving[normalizedKey] = struct{}{}
		defer delete(resolving, normalizedKey)

		mode := normalizeAccessMode(item.AccessMode)
		switch mode {
		case "public", "jwt", "permission":
		case "inherit":
			parentPageKey := strings.TrimSpace(item.ParentPageKey)
			if parentPageKey != "" {
				mode = resolveAccessMode(parentPageKey)
			} else {
				mode = resolveMenuAccessMode(item.ParentMenuID, menuMap)
			}
		default:
			mode = "jwt"
		}
		if mode == "" {
			mode = "jwt"
		}
		accessCache[normalizedKey] = mode
		return mode
	}

	var includeAncestors func(pageKey string)
	includeAncestors = func(pageKey string) {
		normalizedKey := strings.TrimSpace(pageKey)
		if normalizedKey == "" {
			return
		}
		if _, ok := included[normalizedKey]; ok {
			return
		}
		item, ok := pageMap[normalizedKey]
		if !ok {
			return
		}
		included[normalizedKey] = struct{}{}
		if parentPageKey := strings.TrimSpace(item.ParentPageKey); parentPageKey != "" {
			includeAncestors(parentPageKey)
		}
	}

	for _, item := range all {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		if resolveAccessMode(pageKey) == "public" {
			includeAncestors(pageKey)
		}
	}

	result := make([]Record, 0, len(included))
	for _, item := range all {
		if _, ok := included[strings.TrimSpace(item.PageKey)]; ok {
			result = append(result, item)
		}
	}
	return result
}

type runtimeAccessContext = CompiledAccessContext

type runtimeAccessDecision struct {
	allowed       bool
	effectiveMode string
}

type runtimeMenuRequirement struct {
	actions        []string
	matchMode      string
	visibilityMode string
}

func filterRuntimeRecordsForContext(
	s *service,
	appKey string,
	all []Record,
	menuMap map[uuid.UUID]runtimeMenuNode,
	userID *uuid.UUID,
	tenantID *uuid.UUID,
	spaceKey string,
) ([]Record, error) {
	accessCtx, err := s.buildRuntimeAccessContext(normalizeAppKey(appKey), userID, tenantID, menuMap, spaceKey)
	if err != nil {
		return nil, err
	}
	return filterRuntimeRecordsWithAccess(all, menuMap, accessCtx, spaceKey)
}

func filterRuntimeRecordsWithAccess(
	all []Record,
	menuMap map[uuid.UUID]runtimeMenuNode,
	accessCtx *runtimeAccessContext,
	spaceKey string,
) ([]Record, error) {
	if len(all) == 0 {
		return []Record{}, nil
	}
	records := filterRecordsBySpace(all, normalizeSpaceKey(spaceKey))
	if len(records) == 0 {
		return []Record{}, nil
	}
	if accessCtx == nil || !accessCtx.Authenticated {
		return []Record{}, nil
	}

	pageMap := make(map[string]Record, len(records))
	for _, item := range records {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		pageMap[pageKey] = item
	}

	visibleMenusByID, visibleMenusByPath := buildVisibleRuntimeMenuIndex(
		menuMap,
		accessCtx.VisibleMenuIDs,
		spaceKey,
	)

	decisions := make(map[string]runtimeAccessDecision, len(pageMap))
	resolving := make(map[string]struct{}, len(pageMap))

	var resolvePageAccess func(pageKey string) runtimeAccessDecision
	resolvePageAccess = func(pageKey string) runtimeAccessDecision {
		normalizedKey := strings.TrimSpace(pageKey)
		if normalizedKey == "" {
			return runtimeAccessDecision{allowed: false, effectiveMode: "jwt"}
		}
		if cached, ok := decisions[normalizedKey]; ok {
			return cached
		}
		if _, ok := resolving[normalizedKey]; ok {
			return runtimeAccessDecision{allowed: false, effectiveMode: "jwt"}
		}

		item, ok := pageMap[normalizedKey]
		if !ok {
			return runtimeAccessDecision{allowed: false, effectiveMode: "jwt"}
		}

		resolving[normalizedKey] = struct{}{}
		defer delete(resolving, normalizedKey)

		mode := normalizeAccessMode(item.AccessMode)
		var decision runtimeAccessDecision
		switch mode {
		case "public":
			decision = runtimeAccessDecision{allowed: true, effectiveMode: "public"}
		case "jwt":
			decision = runtimeAccessDecision{allowed: accessCtx.Authenticated, effectiveMode: "jwt"}
		case "permission":
			decision = runtimeAccessDecision{
				allowed:       accessCtx.hasAction(strings.TrimSpace(item.PermissionKey)),
				effectiveMode: "permission",
			}
		case "inherit":
			parentPageKey := strings.TrimSpace(item.ParentPageKey)
			if parentPageKey != "" {
				parentDecision := resolvePageAccess(parentPageKey)
				if !parentDecision.allowed || parentDecision.effectiveMode == "public" {
					decision = runtimeAccessDecision{
						allowed:       false,
						effectiveMode: parentDecision.effectiveMode,
					}
				} else {
					decision = runtimeAccessDecision{
						allowed:       true,
						effectiveMode: parentDecision.effectiveMode,
					}
				}
			} else {
				decision = resolveRuntimeMenuInheritedAccess(
					item,
					accessCtx,
					visibleMenusByID,
					visibleMenusByPath,
				)
			}
		default:
			decision = runtimeAccessDecision{allowed: accessCtx.Authenticated, effectiveMode: "jwt"}
		}

		decisions[normalizedKey] = decision
		return decision
	}

	included := make(map[string]struct{}, len(pageMap))
	var includeAncestors func(pageKey string)
	includeAncestors = func(pageKey string) {
		normalizedKey := strings.TrimSpace(pageKey)
		if normalizedKey == "" {
			return
		}
		if _, ok := included[normalizedKey]; ok {
			return
		}
		item, ok := pageMap[normalizedKey]
		if !ok {
			return
		}
		included[normalizedKey] = struct{}{}
		if parentPageKey := strings.TrimSpace(item.ParentPageKey); parentPageKey != "" {
			includeAncestors(parentPageKey)
		}
	}

	for _, item := range records {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		decision := resolvePageAccess(pageKey)
		if !decision.allowed {
			continue
		}
		includeAncestors(pageKey)
	}

	result := make([]Record, 0, len(included))
	for _, item := range records {
		if _, ok := included[strings.TrimSpace(item.PageKey)]; ok {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *service) buildRuntimeAccessContext(
	appKey string,
	userID *uuid.UUID,
	tenantID *uuid.UUID,
	menuMap map[uuid.UUID]runtimeMenuNode,
	spaceKey string,
) (*runtimeAccessContext, error) {
	ctx := &runtimeAccessContext{
		SpaceKey:       normalizeSpaceKey(spaceKey),
		ActionKeys:     make(map[string]struct{}),
		VisibleMenuIDs: make(map[uuid.UUID]struct{}),
	}
	if s == nil || s.db == nil || userID == nil {
		return ctx, nil
	}

	var currentUser models.User
	if err := s.db.Select("id", "status", "is_super_admin").
		Where("id = ?", *userID).
		First(&currentUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx, nil
		}
		return nil, err
	}
	if strings.TrimSpace(currentUser.Status) != "active" {
		return ctx, nil
	}

	ctx.Authenticated = true
	ctx.SuperAdmin = currentUser.IsSuperAdmin
	if currentUser.IsSuperAdmin {
		for id, node := range menuMap {
			if normalizeSpaceKey(node.Menu.SpaceKey) != normalizeSpaceKey(spaceKey) {
				continue
			}
			if !isRuntimeMenuEnabled(node.Menu.Meta) {
				continue
			}
			ctx.VisibleMenuIDs[id] = struct{}{}
		}
		return ctx, nil
	}

	authzService := authorization.NewService(s.db, zap.NewNop())
	actionKeys, err := authzService.GetUserActionKeysInApp(*userID, tenantID, normalizeAppKey(appKey))
	if err != nil {
		return nil, err
	}
	for _, actionKey := range actionKeys {
		normalized := permissionkey.Normalize(actionKey)
		if normalized == "" {
			continue
		}
		ctx.ActionKeys[normalized] = struct{}{}
	}

	visibleMenuIDs, err := s.resolveRuntimeVisibleMenuIDs(normalizeAppKey(appKey), *userID, tenantID, menuMap, spaceKey)
	if err != nil {
		return nil, err
	}
	for _, menuID := range visibleMenuIDs {
		ctx.VisibleMenuIDs[menuID] = struct{}{}
	}
	return ctx, nil
}

func (ctx *runtimeAccessContext) hasAction(permissionKey string) bool {
	normalized := permissionkey.Normalize(permissionKey)
	if normalized == "" {
		return false
	}
	if ctx.SuperAdmin {
		return true
	}
	_, ok := ctx.ActionKeys[normalized]
	return ok
}

func (s *service) resolveRuntimeVisibleMenuIDs(
	appKey string,
	userID uuid.UUID,
	tenantID *uuid.UUID,
	menuMap map[uuid.UUID]runtimeMenuNode,
	spaceKey string,
) ([]uuid.UUID, error) {
	enabledPublicMenuIDs := collectRuntimePublicMenuIDs(menuMap, spaceKey)
	if tenantID == nil {
		snapshot, err := platformaccess.NewService(s.db).GetSnapshot(userID, normalizeAppKey(appKey))
		if err != nil {
			return nil, err
		}
		return finalizeRuntimeVisibleMenuIDs(menuMap, spaceKey, mergeRuntimeUUIDs(snapshot.MenuIDs, enabledPublicMenuIDs)), nil
	}

	roles, err := s.loadRuntimeEffectiveActiveRoles(userID, *tenantID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return finalizeRuntimeVisibleMenuIDs(menuMap, spaceKey, enabledPublicMenuIDs), nil
	}

	roleMap := make(map[uuid.UUID]models.Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}
	boundaryService := teamboundary.NewService(s.db)
	menuSet := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		snapshot, err := boundaryService.GetRoleSnapshot(*tenantID, role.ID, role.TenantID == nil, normalizeAppKey(appKey))
		if err != nil {
			return nil, err
		}
		menuSet = mergeRuntimeUUIDs(menuSet, snapshot.MenuIDs)
	}
	return finalizeRuntimeVisibleMenuIDs(menuMap, spaceKey, mergeRuntimeUUIDs(menuSet, enabledPublicMenuIDs)), nil
}

func (s *service) loadRuntimeEffectiveActiveRoles(userID, tenantID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Model(&models.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.tenant_id = ?", tenantID).
		Where("roles.status = ?", "normal").
		Distinct("roles.*").
		Find(&roles).Error; err != nil {
		return nil, err
	}
	if len(roles) > 0 {
		return roles, nil
	}

	var member models.TenantMember
	if err := s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []models.Role{}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(member.Status) != "active" {
		return []models.Role{}, nil
	}

	var identityRoles []models.Role
	if err := s.db.Where("code = ? AND tenant_id IS NULL AND status = ?", member.RoleCode, "normal").
		Order("sort_order ASC, created_at DESC").
		Find(&identityRoles).Error; err != nil {
		return nil, err
	}
	return identityRoles, nil
}

func buildVisibleRuntimeMenuIndex(
	menuMap map[uuid.UUID]runtimeMenuNode,
	visibleMenuIDs map[uuid.UUID]struct{},
	spaceKey string,
) (map[uuid.UUID]runtimeMenuNode, map[string]runtimeMenuNode) {
	byID := make(map[uuid.UUID]runtimeMenuNode, len(visibleMenuIDs))
	byPath := make(map[string]runtimeMenuNode, len(visibleMenuIDs))
	targetSpaceKey := normalizeSpaceKey(spaceKey)
	for id := range visibleMenuIDs {
		node, ok := menuMap[id]
		if !ok {
			continue
		}
		if normalizeSpaceKey(node.Menu.SpaceKey) != targetSpaceKey {
			continue
		}
		if !isRuntimeMenuEnabled(node.Menu.Meta) {
			continue
		}
		byID[id] = node
		if fullPath := normalizeRoutePath(node.FullPath); fullPath != "" {
			byPath[fullPath] = node
		}
	}
	return byID, byPath
}

func resolveRuntimeMenuInheritedAccess(
	page Record,
	accessCtx *runtimeAccessContext,
	visibleMenusByID map[uuid.UUID]runtimeMenuNode,
	visibleMenusByPath map[string]runtimeMenuNode,
) runtimeAccessDecision {
	if accessCtx == nil || !accessCtx.Authenticated {
		return runtimeAccessDecision{allowed: false, effectiveMode: "jwt"}
	}

	var node runtimeMenuNode
	var ok bool
	activePath := normalizeRoutePath(page.ActiveMenuPath)
	if activePath != "" {
		node, ok = visibleMenusByPath[activePath]
	}
	if !ok && page.ParentMenuID != nil {
		node, ok = visibleMenusByID[*page.ParentMenuID]
	}
	if !ok {
		// 页面显式挂在某个菜单下时，若当前访问图里已经找不到该菜单，
		// 说明菜单已禁用、被空间过滤或当前用户无权访问，此时必须同步拒绝页面深链进入。
		if activePath != "" || page.ParentMenuID != nil {
			return runtimeAccessDecision{allowed: false, effectiveMode: "jwt"}
		}
		return runtimeAccessDecision{allowed: true, effectiveMode: "jwt"}
	}

	mode := runtimeMenuAccessMode(node.Menu.Meta)
	switch mode {
	case "public":
		return runtimeAccessDecision{allowed: false, effectiveMode: "public"}
	case "jwt":
		return runtimeAccessDecision{allowed: true, effectiveMode: "jwt"}
	default:
		requirement := extractRuntimeMenuRequirement(node.Menu.Meta)
		if len(requirement.actions) == 0 {
			return runtimeAccessDecision{allowed: true, effectiveMode: "permission"}
		}
		matched := false
		if requirement.matchMode == "all" {
			matched = true
			for _, action := range requirement.actions {
				if !accessCtx.hasAction(action) {
					matched = false
					break
				}
			}
		} else {
			for _, action := range requirement.actions {
				if accessCtx.hasAction(action) {
					matched = true
					break
				}
			}
		}
		if matched || requirement.visibilityMode == "show" {
			return runtimeAccessDecision{allowed: true, effectiveMode: "permission"}
		}
		return runtimeAccessDecision{allowed: false, effectiveMode: "permission"}
	}
}

func finalizeRuntimeVisibleMenuIDs(
	menuMap map[uuid.UUID]runtimeMenuNode,
	spaceKey string,
	menuIDs []uuid.UUID,
) []uuid.UUID {
	targetSpaceKey := normalizeSpaceKey(spaceKey)
	enabledSet := make(map[uuid.UUID]struct{}, len(menuMap))
	for id, node := range menuMap {
		if normalizeSpaceKey(node.Menu.SpaceKey) != targetSpaceKey {
			continue
		}
		if !isRuntimeMenuEnabled(node.Menu.Meta) {
			continue
		}
		enabledSet[id] = struct{}{}
	}

	result := make([]uuid.UUID, 0, len(menuIDs))
	seen := make(map[uuid.UUID]struct{}, len(menuIDs))
	for _, menuID := range menuIDs {
		if _, ok := enabledSet[menuID]; !ok {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		result = append(result, menuID)
	}
	return result
}

func collectRuntimePublicMenuIDs(menuMap map[uuid.UUID]runtimeMenuNode, spaceKey string) []uuid.UUID {
	targetSpaceKey := normalizeSpaceKey(spaceKey)
	result := make([]uuid.UUID, 0, len(menuMap))
	for id, node := range menuMap {
		if normalizeSpaceKey(node.Menu.SpaceKey) != targetSpaceKey {
			continue
		}
		if !isRuntimeMenuEnabled(node.Menu.Meta) {
			continue
		}
		if isRuntimePublicMenu(node.Menu.Meta) {
			result = append(result, id)
		}
	}
	return result
}

func isRuntimeMenuEnabled(meta models.MetaJSON) bool {
	if meta == nil {
		return true
	}
	if enabled, ok := meta["isEnable"].(bool); ok {
		return enabled
	}
	return true
}

func isRuntimePublicMenu(meta models.MetaJSON) bool {
	if meta == nil {
		return false
	}
	if accessMode := runtimeMenuAccessMode(meta); accessMode == "public" || accessMode == "jwt" {
		return true
	}
	for _, key := range []string{"isPublic", "public", "globalVisible", "publicMenu", "public_menu"} {
		value, ok := meta[key]
		if !ok {
			continue
		}
		flag, ok := value.(bool)
		if ok && flag {
			return true
		}
	}
	return false
}

func runtimeMenuAccessMode(meta models.MetaJSON) string {
	if meta == nil {
		return "permission"
	}
	value, ok := meta["accessMode"].(string)
	if !ok {
		return "permission"
	}
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "public", "jwt", "permission":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "permission"
	}
}

func extractRuntimeMenuRequirement(meta models.MetaJSON) runtimeMenuRequirement {
	actions := make([]string, 0, 2)
	appendAction := func(raw interface{}) {
		value, ok := raw.(string)
		if !ok {
			return
		}
		normalized := permissionkey.Normalize(strings.TrimSpace(value))
		if normalized == "" {
			return
		}
		for _, existing := range actions {
			if existing == normalized {
				return
			}
		}
		actions = append(actions, normalized)
	}

	appendAction(meta["requiredAction"])
	if rawList, ok := meta["requiredActions"].([]interface{}); ok {
		for _, item := range rawList {
			appendAction(item)
		}
	} else if rawList, ok := meta["requiredActions"].([]string); ok {
		for _, item := range rawList {
			appendAction(item)
		}
	}

	matchMode := "any"
	if value, ok := meta["actionMatchMode"].(string); ok && strings.TrimSpace(strings.ToLower(value)) == "all" {
		matchMode = "all"
	}
	visibilityMode := "hide"
	if value, ok := meta["actionVisibilityMode"].(string); ok && strings.TrimSpace(strings.ToLower(value)) == "show" {
		visibilityMode = "show"
	}

	return runtimeMenuRequirement{
		actions:        actions,
		matchMode:      matchMode,
		visibilityMode: visibilityMode,
	}
}

func mergeRuntimeUUIDs(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, id := range group {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func filterRecordsBySpace(records []Record, spaceKey string) []Record {
	if len(records) == 0 {
		return records
	}
	result := make([]Record, 0, len(records))
	for _, item := range records {
		if !isPageVisibleInSpace(item.UIPage, spaceKey) {
			continue
		}
		result = append(result, item)
	}
	return result
}

func normalizeSpaceKey(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return spaceutil.DefaultMenuSpaceKey
	}
	return target
}

func resolveSpaceKeyForRuntime(s *service, appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) string {
	if spaceutil.IsSingleSpaceMode(s.db, normalizeAppKey(appKey)) {
		explicit := normalizeSpaceKey(requestedSpaceKey)
		if explicit != "" && explicit != spaceutil.DefaultMenuSpaceKey && s != nil && s.db != nil {
			if allowed, accessErr := spaceutil.CanAccessSpace(s.db, userID, tenantID, explicit); accessErr == nil && allowed {
				return explicit
			}
		}
		return spaceutil.DefaultMenuSpaceKey
	}
	if s == nil || s.db == nil {
		return normalizeSpaceKey(requestedSpaceKey)
	}
	resolved, _, err := spaceutil.ResolveSpaceKeyByHost(s.db, normalizeAppKey(appKey), host)
	if err != nil {
		return normalizeSpaceKey(requestedSpaceKey)
	}
	trimmed := strings.TrimSpace(requestedSpaceKey)
	if trimmed == "" {
		if allowed, accessErr := spaceutil.CanAccessSpace(s.db, userID, tenantID, resolved); accessErr == nil && allowed {
			return resolved
		}
		return spaceutil.DefaultMenuSpaceKey
	}
	if explicit := normalizeSpaceKey(trimmed); explicit == spaceutil.DefaultMenuSpaceKey {
		return explicit
	}
	if explicit := normalizeSpaceKey(trimmed); explicit != spaceutil.DefaultMenuSpaceKey {
		if allowed, accessErr := spaceutil.CanAccessSpace(s.db, userID, tenantID, explicit); accessErr == nil && allowed {
			return explicit
		}
		if allowed, accessErr := spaceutil.CanAccessSpace(s.db, userID, tenantID, resolved); accessErr == nil && allowed {
			return resolved
		}
		return spaceutil.DefaultMenuSpaceKey
	}
	if allowed, accessErr := spaceutil.CanAccessSpace(s.db, userID, tenantID, resolved); accessErr == nil && allowed {
		return resolved
	}
	return spaceutil.DefaultMenuSpaceKey
}

func runtimePageCacheKey(appKey, spaceKey string) string {
	return normalizeAppKey(appKey) + "::" + normalizeSpaceKey(spaceKey)
}

func resolveMenuAccessMode(
	parentMenuID *uuid.UUID,
	menuMap map[uuid.UUID]runtimeMenuNode,
) string {
	if parentMenuID == nil {
		return "jwt"
	}
	node, ok := menuMap[*parentMenuID]
	if !ok {
		return "jwt"
	}
	raw, ok := node.Menu.Meta["accessMode"]
	if !ok {
		return "permission"
	}
	value, ok := raw.(string)
	if !ok {
		return "permission"
	}
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "public", "jwt", "permission":
		return value
	default:
		return "permission"
	}
}
