package apiregistry

import (
	"crypto/sha1"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type RouteMeta struct {
	Code           string
	Summary        string
	FeatureKind    string
	CategoryCode   string
	ContextScope   string
	Source         string
	PermissionKeys []string
}

type MetaBuilder struct {
	meta RouteMeta
}

type RuntimeRoute struct {
	Method    string
	Path      string
	Handler   string
	HasMeta   bool
	RouteMeta RouteMeta
	IsOpenAPI bool
	IsManaged bool
}

var routeMetaRegistry sync.Map

func Annotate(method, fullPath string, meta RouteMeta) {
	routeMetaRegistry.Store(routeKey(method, fullPath), meta)
}

func Lookup(method, fullPath string) (RouteMeta, bool) {
	value, ok := routeMetaRegistry.Load(routeKey(method, fullPath))
	if !ok {
		return RouteMeta{}, false
	}
	meta, ok := value.(RouteMeta)
	return meta, ok
}

type Registrar struct {
	group        *gin.RouterGroup
	categoryHint string
}

type RequireActionFunc func(permissionKey string, legacy ...string) gin.HandlerFunc
type RequireAnyActionFunc func(permissionKeys ...string) gin.HandlerFunc

func NewRegistrar(group *gin.RouterGroup, categoryHint string) *Registrar {
	return &Registrar{
		group:        group,
		categoryHint: normalizeCategoryCode(categoryHint),
	}
}

func Meta(summary string) *MetaBuilder {
	return (&MetaBuilder{}).WithSummary(summary)
}

func (r *Registrar) Meta(summary string) *MetaBuilder {
	builder := Meta(summary)
	if r == nil || r.categoryHint == "" {
		return builder
	}
	return builder.BindCategory(r.categoryHint)
}

func (r *Registrar) GET(relativePath string, meta *RouteMeta, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.handle(http.MethodGet, relativePath, meta, handlers...)
}

func (r *Registrar) POST(relativePath string, meta *RouteMeta, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.handle(http.MethodPost, relativePath, meta, handlers...)
}

func (r *Registrar) PUT(relativePath string, meta *RouteMeta, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.handle(http.MethodPut, relativePath, meta, handlers...)
}

func (r *Registrar) DELETE(relativePath string, meta *RouteMeta, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.handle(http.MethodDelete, relativePath, meta, handlers...)
}

func (r *Registrar) GETAction(relativePath, summary, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.GET(relativePath, r.metaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) GETActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.GET(relativePath, r.metaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) GETProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.GET(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) POSTAction(relativePath, summary, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.POST(relativePath, r.metaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) POSTActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.POST(relativePath, r.metaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) POSTProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.POST(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) PUTAction(relativePath, summary, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.PUT(relativePath, r.metaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) PUTActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.PUT(relativePath, r.metaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) PUTProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.PUT(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) DELETEAction(relativePath, summary, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.DELETE(relativePath, r.metaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) DELETEActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.DELETE(relativePath, r.metaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) DELETEProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.DELETE(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func MetaWithPermission(summary, permissionKey string) *RouteMeta {
	return MetaWithPermissions(summary, []string{permissionKey})
}

func MetaWithPermissions(summary string, permissionKeys []string) *RouteMeta {
	return Meta(summary).
		BindPermissionKeys(permissionKeys...).
		BindContextScope("optional").
		BindSource("sync").
		Build()
}

func (r *Registrar) metaWithPermission(summary, permissionKey string) *RouteMeta {
	return r.metaWithPermissions(summary, []string{permissionKey})
}

func (r *Registrar) metaWithPermissions(summary string, permissionKeys []string) *RouteMeta {
	builder := Meta(summary).
		BindPermissionKeys(permissionKeys...).
		BindContextScope("optional").
		BindSource("sync")
	if r != nil && r.categoryHint != "" {
		builder = builder.BindCategory(r.categoryHint)
	}
	return builder.Build()
}

func (b *MetaBuilder) clone() *MetaBuilder {
	if b == nil {
		return &MetaBuilder{}
	}
	cloned := *b
	cloned.meta.PermissionKeys = append([]string(nil), b.meta.PermissionKeys...)
	return &cloned
}

func (b *MetaBuilder) WithSummary(summary string) *MetaBuilder {
	next := b.clone()
	next.meta.Summary = strings.TrimSpace(summary)
	return next
}

func (b *MetaBuilder) BindCode(code string) *MetaBuilder {
	next := b.clone()
	next.meta.Code = strings.TrimSpace(code)
	return next
}

func (b *MetaBuilder) BindGroup(categoryCode string) *MetaBuilder {
	return b.BindCategory(categoryCode)
}

func (b *MetaBuilder) BindCategory(categoryCode string) *MetaBuilder {
	next := b.clone()
	next.meta.CategoryCode = strings.TrimSpace(categoryCode)
	return next
}

func (b *MetaBuilder) BindSource(source string) *MetaBuilder {
	next := b.clone()
	next.meta.Source = strings.TrimSpace(source)
	return next
}

func (b *MetaBuilder) BindFeatureKind(featureKind string) *MetaBuilder {
	next := b.clone()
	next.meta.FeatureKind = strings.TrimSpace(featureKind)
	return next
}

func (b *MetaBuilder) BindContextScope(contextScope string) *MetaBuilder {
	next := b.clone()
	next.meta.ContextScope = strings.TrimSpace(contextScope)
	return next
}

func (b *MetaBuilder) BindPermissionKey(permissionKey string) *MetaBuilder {
	return b.BindPermissionKeys(permissionKey)
}

func (b *MetaBuilder) BindPermissionKeys(permissionKeys ...string) *MetaBuilder {
	next := b.clone()
	next.meta.PermissionKeys = normalizePermissionKeys(append(next.meta.PermissionKeys, permissionKeys...))
	return next
}

func (b *MetaBuilder) Build() *RouteMeta {
	if b == nil {
		return &RouteMeta{
			ContextScope: "optional",
			Source:       "sync",
		}
	}
	meta := b.meta
	meta.Summary = strings.TrimSpace(meta.Summary)
	meta.Code = strings.TrimSpace(meta.Code)
	meta.FeatureKind = strings.TrimSpace(meta.FeatureKind)
	meta.CategoryCode = strings.TrimSpace(meta.CategoryCode)
	meta.ContextScope = normalizeContextScope(meta.ContextScope)
	meta.Source = normalizeSource(meta.Source)
	meta.PermissionKeys = normalizePermissionKeys(meta.PermissionKeys)
	return &meta
}

func appendRequireActionHandler(permissionKey string, requireAction RequireActionFunc, handlers []gin.HandlerFunc) []gin.HandlerFunc {
	if requireAction == nil || strings.TrimSpace(permissionKey) == "" {
		return handlers
	}
	withAuth := make([]gin.HandlerFunc, 0, len(handlers)+1)
	withAuth = append(withAuth, requireAction(permissionKey))
	withAuth = append(withAuth, handlers...)
	return withAuth
}

func appendRequireAnyActionHandler(permissionKeys []string, requireAction RequireAnyActionFunc, handlers []gin.HandlerFunc) []gin.HandlerFunc {
	if requireAction == nil {
		return handlers
	}
	keys := normalizePermissionKeys(permissionKeys)
	if len(keys) == 0 {
		return handlers
	}
	withAuth := make([]gin.HandlerFunc, 0, len(handlers)+1)
	withAuth = append(withAuth, requireAction(keys...))
	withAuth = append(withAuth, handlers...)
	return withAuth
}

func (r *Registrar) handle(method, relativePath string, meta *RouteMeta, handlers ...gin.HandlerFunc) gin.IRoutes {
	fullPath := joinPath(r.group.BasePath(), relativePath)
	if meta != nil {
		normalizedMeta := normalizeRouteMeta(method, fullPath, applyDefaultCategory(meta, r.categoryHint))
		Annotate(method, fullPath, normalizedMeta)
	}

	switch method {
	case http.MethodGet:
		return r.group.GET(relativePath, handlers...)
	case http.MethodPost:
		return r.group.POST(relativePath, handlers...)
	case http.MethodPut:
		return r.group.PUT(relativePath, handlers...)
	case http.MethodDelete:
		return r.group.DELETE(relativePath, handlers...)
	default:
		return r.group.Handle(method, relativePath, handlers...)
	}
}

func applyDefaultCategory(meta *RouteMeta, categoryHint string) *RouteMeta {
	if meta == nil || strings.TrimSpace(meta.CategoryCode) != "" || strings.TrimSpace(categoryHint) == "" {
		return meta
	}
	cloned := *meta
	cloned.CategoryCode = normalizeCategoryCode(categoryHint)
	return &cloned
}

func SyncRoutes(db *gorm.DB, logger *zap.Logger, routes []gin.RouteInfo) error {
	if db == nil {
		return errors.New("database is nil")
	}
	if err := ensureManagedEndpointColumns(db); err != nil {
		return err
	}
	return syncRoutesInternal(db, logger, routes)
}

func ensureManagedEndpointColumns(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	statements := []string{
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS app_scope varchar(20) NOT NULL DEFAULT 'shared'`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS app_key varchar(100) NOT NULL DEFAULT 'platform-admin'`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS feature_kind varchar(20) NOT NULL DEFAULT 'system'`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS handler varchar(255)`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS summary varchar(255)`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS category_id uuid`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS context_scope varchar(20) NOT NULL DEFAULT 'optional'`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS source varchar(20) NOT NULL DEFAULT 'sync'`,
		`ALTER TABLE api_endpoints ADD COLUMN IF NOT EXISTS status varchar(20) NOT NULL DEFAULT 'normal'`,
		`UPDATE api_endpoints SET app_key = 'platform-admin' WHERE COALESCE(TRIM(app_key), '') = ''`,
		`UPDATE api_endpoints SET app_scope = 'shared' WHERE COALESCE(TRIM(app_scope), '') = '' AND (path LIKE '/api/v1/auth/%' OR path = '/api/v1/pages/runtime/public' OR path LIKE '/open/v1/%' OR path = '/health')`,
		`UPDATE api_endpoints SET app_scope = 'app' WHERE COALESCE(TRIM(app_scope), '') = ''`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func CollectRuntimeRoutes(routes []gin.RouteInfo) []RuntimeRoute {
	result := make([]RuntimeRoute, 0, len(routes))
	for _, route := range routes {
		if !isManagedRoute(route.Path) {
			continue
		}
		meta, hasMeta := Lookup(route.Method, route.Path)
		result = append(result, RuntimeRoute{
			Method:    strings.ToUpper(strings.TrimSpace(route.Method)),
			Path:      route.Path,
			Handler:   route.Handler,
			HasMeta:   hasMeta,
			RouteMeta: meta,
			IsOpenAPI: strings.HasPrefix(route.Path, "/open/v1/"),
			IsManaged: isManagedRoute(route.Path),
		})
	}
	return result
}

func syncRoutesInternal(
	db *gorm.DB,
	logger *zap.Logger,
	routes []gin.RouteInfo,
) error {
	for _, route := range routes {
		if !isManagedRoute(route.Path) {
			continue
		}

		meta, hasMeta := Lookup(route.Method, route.Path)
		if !hasMeta {
			continue
		}
		categoryID := resolveCategoryID(db, meta.CategoryCode)

		endpointCode := ResolveRouteCode(route.Method, route.Path, &meta)
		if endpointCode == "" {
			if logger != nil {
				logger.Warn("Skip syncing API endpoint without fixed route code",
					zap.String("method", strings.ToUpper(strings.TrimSpace(route.Method))),
					zap.String("path", strings.TrimSpace(route.Path)),
				)
			}
			continue
		}
		existing, err := findEndpointByCode(db, endpointCode)
		if err != nil {
			return err
		}
		legacy, err := findEndpointByMethodAndPath(db, route.Method, route.Path)
		if err != nil {
			return err
		}

		summary := strings.TrimSpace(meta.Summary)

		featureKind := normalizeFeatureKind(meta.FeatureKind)

		contextScope := normalizeContextScope(meta.ContextScope)

		source := normalizeSource(meta.Source)

		endpoint := &models.APIEndpoint{
			Code:         endpointCode,
			Method:       strings.ToUpper(route.Method),
			Path:         route.Path,
			FeatureKind:  featureKind,
			Handler:      route.Handler,
			Summary:      summary,
			CategoryID:   categoryID,
			ContextScope: contextScope,
			Source:       source,
			Status:       "normal",
		}
		switch {
		case existing != nil:
			if err := updateManagedEndpoint(db, existing.ID, endpoint); err != nil {
				return err
			}
		case legacy != nil:
			if shouldBackfillManagedEndpointCode(legacy) {
				if err := backfillEndpointCode(db, legacy.ID, endpointCode); err != nil {
					return err
				}
			}
			if err := updateManagedEndpoint(db, legacy.ID, endpoint); err != nil {
				return err
			}
		default:
			if err := insertEndpoint(db, endpoint); err != nil {
				return err
			}
		}

		if err := replaceEndpointPermissionBindings(db, endpoint, meta.PermissionKeys); err != nil {
			return err
		}
	}

	logger.Info("API endpoints synced", zap.Int("count", len(routes)))
	return nil
}

func findEndpointByCode(db *gorm.DB, code string) (*models.APIEndpoint, error) {
	target := strings.TrimSpace(code)
	if db == nil || target == "" {
		return nil, nil
	}
	var endpoint models.APIEndpoint
	result := db.Where("code = ?", target).Limit(1).Find(&endpoint)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &endpoint, nil
}

func findEndpointByMethodAndPath(db *gorm.DB, method, path string) (*models.APIEndpoint, error) {
	if db == nil {
		return nil, nil
	}
	var endpoint models.APIEndpoint
	result := db.Where("method = ? AND path = ?", strings.ToUpper(strings.TrimSpace(method)), strings.TrimSpace(path)).Limit(1).Find(&endpoint)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &endpoint, nil
}

func isManagedRoute(path string) bool {
	return strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/open/v1/")
}

func normalizeRouteMeta(method, fullPath string, meta *RouteMeta) RouteMeta {
	if meta == nil {
		return RouteMeta{}
	}
	normalized := *meta
	normalized.Code = ResolveRouteCode(method, fullPath, meta)
	normalized.Summary = strings.TrimSpace(normalized.Summary)
	normalized.FeatureKind = normalizeFeatureKind(normalized.FeatureKind)
	normalized.CategoryCode = strings.TrimSpace(normalized.CategoryCode)
	normalized.ContextScope = normalizeContextScope(normalized.ContextScope)
	normalized.Source = normalizeSource(normalized.Source)
	normalized.PermissionKeys = normalizePermissionKeys(normalized.PermissionKeys)
	return normalized
}

func normalizeFeatureKind(value string) string {
	switch strings.TrimSpace(value) {
	case "business":
		return "business"
	default:
		return "system"
	}
}

func joinPath(basePath, relativePath string) string {
	if relativePath == "" || relativePath == "/" {
		return strings.TrimRight(basePath, "/")
	}
	base := strings.TrimRight(basePath, "/")
	relative := strings.TrimLeft(relativePath, "/")
	if base == "" {
		return "/" + relative
	}
	return base + "/" + relative
}

func routeKey(method, fullPath string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(fullPath)
}

func ResolveRouteCode(method, path string, meta *RouteMeta) string {
	if meta != nil {
		if code := strings.TrimSpace(meta.Code); code != "" {
			return code
		}
	}
	if code := lookupFixedRouteCode(method, path); code != "" {
		return code
	}
	if meta != nil {
		return deriveStableEndpointCode(method, path)
	}
	return ""
}

func deriveStableEndpointCode(method, path string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
	return uuid.NewHash(sha1.New(), uuid.NameSpaceURL, []byte("api-endpoint:"+normalized), 5).String()
}

func shouldInsertManagedEndpoint(existing *models.APIEndpoint) bool {
	return existing == nil
}

func shouldBackfillManagedEndpointCode(existing *models.APIEndpoint) bool {
	if existing == nil {
		return false
	}
	return strings.TrimSpace(existing.Code) == ""
}

func insertEndpoint(db *gorm.DB, endpoint *models.APIEndpoint) error {
	if endpoint == nil {
		return nil
	}
	endpoint.FeatureKind = normalizeFeatureKind(endpoint.FeatureKind)
	endpoint.ContextScope = normalizeContextScope(endpoint.ContextScope)
	endpoint.Source = normalizeSource(endpoint.Source)
	if strings.TrimSpace(endpoint.Status) == "" {
		endpoint.Status = "normal"
	}
	return db.Create(endpoint).Error
}

func updateManagedEndpoint(db *gorm.DB, endpointID uuid.UUID, endpoint *models.APIEndpoint) error {
	if db == nil || endpoint == nil || endpointID == uuid.Nil {
		return nil
	}
	endpoint.FeatureKind = normalizeFeatureKind(endpoint.FeatureKind)
	endpoint.ContextScope = normalizeContextScope(endpoint.ContextScope)
	endpoint.Source = normalizeSource(endpoint.Source)
	if strings.TrimSpace(endpoint.Status) == "" {
		endpoint.Status = "normal"
	}
	updates := map[string]interface{}{
		"code":          endpoint.Code,
		"method":        endpoint.Method,
		"path":          endpoint.Path,
		"feature_kind":  endpoint.FeatureKind,
		"handler":       endpoint.Handler,
		"summary":       endpoint.Summary,
		"category_id":   endpoint.CategoryID,
		"context_scope": endpoint.ContextScope,
		"source":        endpoint.Source,
		"status":        endpoint.Status,
	}
	return db.Model(&models.APIEndpoint{}).Where("id = ?", endpointID).Updates(updates).Error
}

func backfillEndpointCode(db *gorm.DB, endpointID uuid.UUID, code string) error {
	targetCode := strings.TrimSpace(code)
	if db == nil || endpointID == uuid.Nil || targetCode == "" {
		return nil
	}
	return db.Model(&models.APIEndpoint{}).Where("id = ?", endpointID).Update("code", targetCode).Error
}

func resolveCategoryID(db *gorm.DB, code string) *uuid.UUID {
	target := strings.TrimSpace(code)
	if db == nil || target == "" {
		return nil
	}
	var item models.APIEndpointCategory
	if err := db.Where("code = ?", target).First(&item).Error; err != nil {
		return nil
	}
	return &item.ID
}

func replaceEndpointPermissionBindings(db *gorm.DB, endpoint *models.APIEndpoint, permissionKeys []string) error {
	if db == nil || endpoint == nil {
		return nil
	}
	endpointCode := strings.TrimSpace(endpoint.Code)
	if endpointCode == "" {
		return nil
	}
	keys := normalizePermissionKeys(permissionKeys)
	return db.Transaction(func(tx *gorm.DB) error {
		var existing []models.APIEndpointPermissionBinding
		if err := tx.
			Where("endpoint_code = ?", endpointCode).
			Order("sort_order ASC, created_at ASC").
			Find(&existing).Error; err != nil {
			return err
		}
		if samePermissionBindings(existing, keys) {
			return nil
		}
		if err := tx.Unscoped().Where("endpoint_code = ?", endpointCode).Delete(&models.APIEndpointPermissionBinding{}).Error; err != nil {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		items := make([]models.APIEndpointPermissionBinding, 0, len(keys))
		for idx, key := range keys {
			items = append(items, models.APIEndpointPermissionBinding{
				EndpointCode:  endpointCode,
				PermissionKey: key,
				MatchMode:     "ANY",
				SortOrder:     idx,
			})
		}
		return tx.Create(&items).Error
	})
}

func samePermissionBindings(existing []models.APIEndpointPermissionBinding, permissionKeys []string) bool {
	if len(existing) != len(permissionKeys) {
		return false
	}
	for idx, item := range existing {
		if strings.TrimSpace(item.PermissionKey) != permissionKeys[idx] {
			return false
		}
		if strings.TrimSpace(item.MatchMode) != "ANY" {
			return false
		}
		if item.SortOrder != idx {
			return false
		}
	}
	return true
}

func normalizePermissionKeys(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		target := strings.TrimSpace(value)
		if target == "" || slices.Contains(result, target) {
			continue
		}
		result = append(result, target)
	}
	return result
}

func normalizeCategoryCode(value string) string {
	switch strings.TrimSpace(value) {
	case "permission_action", "system_permission":
		return "permission_key"
	case "collaboration", "collaboration_workspace_member", "collaboration_workspace_member_admin", "collaboration_workspace":
		return "collaboration_workspace"
	default:
		return strings.TrimSpace(value)
	}
}

func normalizeContextScope(value string) string {
	switch strings.TrimSpace(value) {
	case "required", "forbidden":
		return strings.TrimSpace(value)
	default:
		return "optional"
	}
}

func normalizeSource(value string) string {
	switch strings.TrimSpace(value) {
	case "seed", "manual":
		return strings.TrimSpace(value)
	default:
		return "sync"
	}
}
