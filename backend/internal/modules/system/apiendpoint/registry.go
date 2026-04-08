package apiendpoint

// registry.go contains the route-annotation registry that was formerly in
// internal/pkg/apiregistry. All functionality is identical; the package was
// inlined here to remove the external dependency.

import (
	"crypto/sha1"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// RouteMeta holds annotation metadata for a single registered route.
type RouteMeta struct {
	Code           string
	Summary        string
	FeatureKind    string
	CategoryCode   string
	ContextScope   string
	Source         string
	PermissionKeys []string
}

// MetaBuilder is a fluent builder for RouteMeta.
type MetaBuilder struct {
	meta RouteMeta
}

// RuntimeRoute is a gin route enriched with optional RouteMeta.
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

// Annotate stores meta for the given method+path pair.
func Annotate(method, fullPath string, meta RouteMeta) {
	routeMetaRegistry.Store(routeKey(method, fullPath), meta)
}

// Lookup retrieves previously annotated meta, if any.
func Lookup(method, fullPath string) (RouteMeta, bool) {
	value, ok := routeMetaRegistry.Load(routeKey(method, fullPath))
	if !ok {
		return RouteMeta{}, false
	}
	meta, ok := value.(RouteMeta)
	return meta, ok
}

// Registrar wraps a *gin.RouterGroup and records RouteMeta on every registration.
type Registrar struct {
	group        *gin.RouterGroup
	categoryHint string
}

// RequireActionFunc is the signature for a single-permission middleware factory.
type RequireActionFunc func(permissionKey string, legacy ...string) gin.HandlerFunc

// RequireAnyActionFunc is the signature for a multi-permission middleware factory.
type RequireAnyActionFunc func(permissionKeys ...string) gin.HandlerFunc

// NewRegistrar creates a Registrar for the given group with an optional category hint.
func NewRegistrar(group *gin.RouterGroup, categoryHint string) *Registrar {
	return &Registrar{
		group:        group,
		categoryHint: normalizeCategoryCode(categoryHint),
	}
}

// Meta starts building a RouteMeta with the given summary.
func Meta(summary string) *MetaBuilder {
	return (&MetaBuilder{}).WithSummary(summary)
}

// Meta on a Registrar pre-binds the registrar's category hint.
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

// MetaWithPermission builds a RouteMeta annotated with a single permission key.
func MetaWithPermission(summary, permissionKey string) *RouteMeta {
	return MetaWithPermissions(summary, []string{permissionKey})
}

// MetaWithPermissions builds a RouteMeta annotated with multiple permission keys.
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

// SyncRoutes persists all annotated gin routes to the api_endpoints table.
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

// CollectRuntimeRoutes filters and enriches gin routes with meta from the registry.
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

// ResolveRouteCode returns the stable code for a route, using meta.Code > fixed map > derived UUID.
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

// deriveStableEndpointCode produces a deterministic UUID-v5 code for a route.
func deriveStableEndpointCode(method, path string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
	return uuid.NewHash(sha1.New(), uuid.NameSpaceURL, []byte("api-endpoint:"+normalized), 5).String()
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

// fixed_codes.go content inlined below

var fixedManagedRouteCodes = map[string]string{
	routeKey("POST", "/api/v1/auth/login"):                                                         "5b230c5b-3719-536e-8a8d-521d9cdcf2ee",
	routeKey("POST", "/api/v1/auth/register"):                                                      "b414f517-d1c9-5ce2-8be1-dd21bbe02162",
	routeKey("POST", "/api/v1/auth/refresh"):                                                       "dbc31465-a050-5002-8faa-14677502fdba",
	routeKey("GET", "/api/v1/user/info"):                                                           "b9a86484-8b9b-5df1-85ca-9faad03d9278",
	routeKey("GET", "/api/v1/api-endpoints/unregistered"):                                          "b72ab1ce-9038-5f64-b722-29dbb9f6195a",
	routeKey("GET", "/api/v1/api-endpoints/overview"):                                              "0eac7d2d-84ae-563d-9f2e-77c76468ad62",
	routeKey("GET", "/api/v1/api-endpoints/stale"):                                                 "0255725b-7771-5a20-bf7e-0409e76fb9bc",
	routeKey("GET", "/api/v1/api-endpoints"):                                                       "35c24e21-0df0-5759-a534-4383aaa85635",
	routeKey("GET", "/api/v1/api-endpoints/categories"):                                            "a81b0405-c351-539c-bf34-2903e803b45b",
	routeKey("POST", "/api/v1/api-endpoints/sync"):                                                 "a8e40cdb-57ee-560d-bd59-2ece4f958f7c",
	routeKey("POST", "/api/v1/api-endpoints/cleanup-stale"):                                        "d98eb86a-6b72-57c5-b030-5f6370c5d1df",
	routeKey("POST", "/api/v1/api-endpoints"):                                                      "1c6178fa-5d7f-5c2f-9038-c8ca4128969a",
	routeKey("PUT", "/api/v1/api-endpoints/:id"):                                                   "3c1d6736-294b-5efd-9eae-7d1e817c8bc6",
	routeKey("PUT", "/api/v1/api-endpoints/:id/context-scope"):                                     "aa5b8825-a2d0-5f4b-a066-af33f6772953",
	routeKey("POST", "/api/v1/api-endpoints/categories"):                                           "7736788e-401b-519f-8e7d-4374d95479c1",
	routeKey("PUT", "/api/v1/api-endpoints/categories/:id"):                                        "57e41a94-d1c4-55d6-980a-56de0f2e39d1",
	routeKey("GET", "/api/v1/feature-packages"):                                                    "f31d4da0-cdeb-51f7-bc62-4d56c31131b3",
	routeKey("GET", "/api/v1/feature-packages/options"):                                            "54c082ac-36e7-5c60-a4bb-bcf6e25d6741",
	routeKey("GET", "/api/v1/feature-packages/:id"):                                                "a3c596d6-c853-5518-a1ad-261669123185",
	routeKey("POST", "/api/v1/feature-packages"):                                                   "6c01ee9c-3fdb-5790-b7fb-d928921bef86",
	routeKey("PUT", "/api/v1/feature-packages/:id"):                                                "63df4132-cacf-5199-9fe1-3508b53072aa",
	routeKey("DELETE", "/api/v1/feature-packages/:id"):                                             "9ef26c95-df63-524b-956e-f487a262ee1a",
	routeKey("GET", "/api/v1/feature-packages/:id/children"):                                       "fa750ec5-8fed-5e99-b2d3-b7d371320b99",
	routeKey("PUT", "/api/v1/feature-packages/:id/children"):                                       "7c76b816-69f7-57eb-aefa-4189a53197c6",
	routeKey("GET", "/api/v1/feature-packages/:id/actions"):                                        "76386948-bd46-5e74-a72f-42322e2bccd1",
	routeKey("PUT", "/api/v1/feature-packages/:id/actions"):                                        "da34b45b-e242-5ae1-ba55-ed87d63f9fc7",
	routeKey("GET", "/api/v1/feature-packages/:id/menus"):                                          "0d4a9756-3bb0-55e9-89a3-617b0218e4cb",
	routeKey("PUT", "/api/v1/feature-packages/:id/menus"):                                          "3772b7e3-d306-52dd-be53-758252a8b69d",
	routeKey("GET", "/api/v1/feature-packages/:id/collaboration-workspaces"):                       "47dc6ed6-b98f-5b32-94dc-c579101866f3",
	routeKey("PUT", "/api/v1/feature-packages/:id/collaboration-workspaces"):                       "b9d3cb8b-7b22-5fd0-99ff-624b50b22f0c",
	routeKey("GET", "/api/v1/feature-packages/collaboration-workspaces/:collaborationWorkspaceId"): "2d47541d-93ea-57f1-a4f3-78d65eec509c",
	routeKey("PUT", "/api/v1/feature-packages/collaboration-workspaces/:collaborationWorkspaceId"): "ed783123-f7d6-56dd-93b1-cfd722244aeb",
	routeKey("POST", "/api/v1/media/upload"):                                                       "0a2fc86e-7d75-5b9e-8dfa-58c7f0f682ff",
	routeKey("GET", "/api/v1/media"):                                                               "8e6d404e-70f7-5bcd-b345-4416dd6b08e9",
	routeKey("DELETE", "/api/v1/media/:id"):                                                        "a5e6bd3d-4b01-5ed7-8ca6-89d3ade408fb",
	routeKey("GET", "/api/v1/menus/tree"):                                                          "75d2c43b-9b1d-51a4-9f19-6424790199bc",
	routeKey("POST", "/api/v1/menus"):                                                              "81181b51-bf3c-5d17-b267-f45854c1e3d2",
	routeKey("PUT", "/api/v1/menus/:id"):                                                           "57df0753-4746-513e-b95e-5931ace6a738",
	routeKey("DELETE", "/api/v1/menus/:id"):                                                        "aeb45da1-adb6-55ea-b9a9-373c22c86430",
	routeKey("GET", "/api/v1/menus/groups"):                                                        "a06fb092-5445-5c7a-912b-3ee297b3f30a",
	routeKey("POST", "/api/v1/menus/groups"):                                                       "237cd251-c8b5-5c04-a3fb-4f011b1ee8f5",
	routeKey("PUT", "/api/v1/menus/groups/:id"):                                                    "0fe95701-a138-5f2d-92ea-7150d6c81344",
	routeKey("DELETE", "/api/v1/menus/groups/:id"):                                                 "22f1c0a4-9c24-523a-8467-9f1b87cde2bd",
	routeKey("POST", "/api/v1/menus/backups"):                                                      "8ca9ec4a-157d-5f4e-b736-c9b6e01a97e4",
	routeKey("GET", "/api/v1/menus/backups"):                                                       "c29f0703-f85b-5094-884c-f3c4cf126bf2",
	routeKey("DELETE", "/api/v1/menus/backups/:id"):                                                "58dcb60d-2633-50f0-ad4d-dab3c4b58fae",
	routeKey("POST", "/api/v1/menus/backups/:id/restore"):                                          "bcea597c-e4ff-5ca3-936d-ac192e9ffa0a",
	routeKey("GET", "/api/v1/pages/menu-options"):                                                  "2f52862e-ffca-57c8-a17e-11eb33aec1f1",
	routeKey("GET", "/api/v1/pages/options"):                                                       "2d684f87-a499-5b14-b552-f3b98728403f",
	routeKey("GET", "/api/v1/pages/unregistered"):                                                  "703b290c-fe5e-52f3-b3c8-5f505632d7a1",
	routeKey("POST", "/api/v1/pages/sync"):                                                         "c9a1ed80-c4ed-5782-ba36-f656eefca523",
	routeKey("GET", "/api/v1/pages"):                                                               "1fda8e29-519b-57d8-9ce1-ba49791e5146",
	routeKey("GET", "/api/v1/pages/:id/breadcrumb-preview"):                                        "bb9f62bf-0dbc-5643-b37a-942d6d7498e7",
	routeKey("GET", "/api/v1/pages/:id"):                                                           "5048ab31-2e3a-5564-ae16-df8cbd448a6b",
	routeKey("POST", "/api/v1/pages"):                                                              "7beaf333-9fb7-5ece-9e82-3b64b96e95bd",
	routeKey("PUT", "/api/v1/pages/:id"):                                                           "d89bd4c4-db13-5ee3-aa22-da2f9457d744",
	routeKey("DELETE", "/api/v1/pages/:id"):                                                        "8bb34dee-4a5c-5fe0-9d5a-842e0a126b25",
	routeKey("GET", "/api/v1/pages/runtime"):                                                       "6f736621-280a-555a-a39e-13716868b635",
	routeKey("GET", "/api/v1/pages/runtime/public"):                                                "d954acbd-0cf1-5bfb-a736-c0b0cf528937",
	routeKey("GET", "/api/v1/permission-actions"):                                                  "1210af17-787f-5a45-9342-01fa86d7c121",
	routeKey("GET", "/api/v1/permission-actions/options"):                                          "a343aa71-a4c9-559c-a1f9-4469a90dee78",
	routeKey("GET", "/api/v1/permission-actions/groups"):                                           "eebffccf-83cd-508a-881a-04b3757f613d",
	routeKey("GET", "/api/v1/permission-actions/:id"):                                              "925b7654-4a2e-5111-9b0d-6ce31120f960",
	routeKey("GET", "/api/v1/permission-actions/:id/endpoints"):                                    "1b26c240-9809-53df-a12b-b7e26d3ae57c",
	routeKey("POST", "/api/v1/permission-actions/:id/endpoints"):                                   "f90b19bf-c4a2-5a16-9732-fbd38ebb207d",
	routeKey("DELETE", "/api/v1/permission-actions/:id/endpoints/:endpointCode"):                   "f38619ed-06b7-56b9-9a9f-959827bd17d0",
	routeKey("POST", "/api/v1/permission-actions/groups"):                                          "84f12bba-19a6-5909-8c40-d2221db43ead",
	routeKey("PUT", "/api/v1/permission-actions/groups/:id"):                                       "555da252-5e52-5e52-909e-18b5acceed0c",
	routeKey("POST", "/api/v1/permission-actions"):                                                 "166c3dd6-aae6-50f1-a70b-546f0b85f31d",
	routeKey("PUT", "/api/v1/permission-actions/:id"):                                              "83507ef9-261c-5989-b7df-ac66179db5cc",
	routeKey("DELETE", "/api/v1/permission-actions/:id"):                                           "91dd41f0-2fa5-5fa2-a004-938053e19eb7",
	routeKey("GET", "/api/v1/roles"):                                                               "7fddad33-1730-58c7-a4c4-df07df409861",
	routeKey("GET", "/api/v1/roles/options"):                                                       "fdde79ca-01ed-5915-954a-ea9471e1859b",
	routeKey("GET", "/api/v1/roles/:id"):                                                           "c6b77d7f-f557-5437-8fbc-fca4ce63ce54",
	routeKey("GET", "/api/v1/roles/:id/packages"):                                                  "4ecc7b01-2bf8-5a1c-a34f-bb68eec1010a",
	routeKey("PUT", "/api/v1/roles/:id/packages"):                                                  "0554e8eb-0a71-5d8e-8fe9-97711f5c0bc1",
	routeKey("GET", "/api/v1/roles/:id/menus"):                                                     "312ae5cc-f248-5161-9f37-3422a4506c4d",
	routeKey("PUT", "/api/v1/roles/:id/menus"):                                                     "aa4d0009-16d0-5cec-acee-32b0891512af",
	routeKey("GET", "/api/v1/roles/:id/actions"):                                                   "f0c1b2dd-85b8-507a-990d-5861dfd98557",
	routeKey("PUT", "/api/v1/roles/:id/actions"):                                                   "b3303d81-776a-590b-8f37-8b8ab5845784",
	routeKey("GET", "/api/v1/roles/:id/data-permissions"):                                          "1c315cee-4818-5dc0-b3a0-f072ea03e6f8",
	routeKey("PUT", "/api/v1/roles/:id/data-permissions"):                                          "38ef7e86-e0ae-5f9d-9df4-98211cb44167",
	routeKey("POST", "/api/v1/roles"):                                                              "43239127-c7da-54da-9f1f-e1be2e1a771a",
	routeKey("PUT", "/api/v1/roles/:id"):                                                           "673cacbc-4919-5e37-a4b3-56dab25e270f",
	routeKey("DELETE", "/api/v1/roles/:id"):                                                        "d347d4f4-24ec-59aa-9060-818897521e8e",
	routeKey("GET", "/api/v1/system/view-pages"):                                                   "63cd7d02-c887-5811-a9fe-0d2c9a76c23f",
	routeKey("GET", "/api/v1/system/fast-enter"):                                                   "6c72d4ff-d692-51f0-b83c-9705694f7395",
	routeKey("PUT", "/api/v1/system/fast-enter"):                                                   "c01cb465-b9fa-5618-b4b4-f5a1e1456754",
	routeKey("GET", "/api/v1/system/menu-spaces/current"):                                          "39193781-b4a8-4511-8777-a58249229e90",
	routeKey("GET", "/api/v1/system/menu-spaces"):                                                  "375bfb66-a4ba-4b9c-965b-2baf51f24e28",
	routeKey("POST", "/api/v1/system/menu-spaces"):                                                 "f681b958-a4ba-4c19-8200-88792a66e733",
	routeKey("POST", "/api/v1/system/menu-spaces/:spaceKey/initialize-default"):                    "377612fb-0659-4dc1-b701-be408671e990",
	routeKey("GET", "/api/v1/system/menu-space-host-bindings"):                                     "c19810c8-74f1-4d68-810f-da8a7818cc88",
	routeKey("POST", "/api/v1/system/menu-space-host-bindings"):                                    "7e7b5065-c0c6-4f30-9ffb-6a43720e6fd1",
	routeKey("GET", "/api/v1/messages/inbox/summary"):                                              "48e9db7c-6422-58d9-bb8b-d1fe8fdf6a51",
	routeKey("GET", "/api/v1/messages/inbox"):                                                      "25756b8c-b9d7-55e2-a22f-2c85eefc1ad9",
	routeKey("GET", "/api/v1/messages/inbox/:deliveryId"):                                          "6b13930f-b05f-5669-84a8-38ac7c4ff1bf",
	routeKey("POST", "/api/v1/messages/inbox/:deliveryId/read"):                                    "0c4eaf75-b037-5fde-97ff-49f9d8abf917",
	routeKey("POST", "/api/v1/messages/inbox/read-all"):                                            "9805ce58-631f-5dbf-a4a1-9410eb8344f4",
	routeKey("POST", "/api/v1/messages/inbox/:deliveryId/todo-action"):                             "c1a3247d-4d68-59b2-99dc-a7f9553ea4ce",
	routeKey("GET", "/api/v1/messages/dispatch/options"):                                           "1f68f2ff-8f30-558b-8f2c-cbe1d8d41d70",
	routeKey("POST", "/api/v1/messages/dispatch"):                                                  "f6d7c84d-dd74-5298-a906-3ff2fab36302",
	routeKey("GET", "/api/v1/messages/templates"):                                                  "80e46f95-a91c-4078-9dac-d5b2546d5d1e",
	routeKey("POST", "/api/v1/messages/templates"):                                                  "5cb0d7af-f09e-45f0-ad8d-6e3b7107efb3",
	routeKey("PUT", "/api/v1/messages/templates/:templateId"):                                      "547b0c98-dea3-42a1-954f-156b2e575ba1",
	routeKey("GET", "/api/v1/messages/senders"):                                                    "0e27da4e-413b-4658-b67b-5ae0ebfb6fbd",
	routeKey("POST", "/api/v1/messages/senders"):                                                   "0efcd269-61ee-406c-8bb9-d0c94b4138e3",
	routeKey("PUT", "/api/v1/messages/senders/:senderId"):                                          "d7f89dd1-f3df-4fef-8cf4-c6018b98c91d",
	routeKey("GET", "/api/v1/messages/recipient-groups"):                                           "9dd16f54-cf91-4e01-b08d-b7408ee664da",
	routeKey("POST", "/api/v1/messages/recipient-groups"):                                          "ac922f75-983e-4909-9efc-37d1f236f2bd",
	routeKey("PUT", "/api/v1/messages/recipient-groups/:groupId"):                                  "f3ca31d2-e384-4fa0-a6b8-93ea1b7b6e6e",
	routeKey("GET", "/api/v1/messages/records"):                                                    "5e23c436-99a5-43f6-a75e-72ed4efa3d33",
	routeKey("GET", "/api/v1/messages/records/:recordId"):                                          "9fd5da11-c7ef-4700-bb4b-abff40183258",
	routeKey("GET", "/api/v1/collaboration-workspaces/mine"):                                       "6445519a-1913-5e1b-ab8a-3972f3bb8ee4",
	routeKey("GET", "/api/v1/collaboration-workspaces/current"):                                    "f173a24e-480a-5379-918a-96a84a271b5a",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/members"):                            "119ba807-e132-5836-8a5b-c212dfd3ab45",
	routeKey("POST", "/api/v1/collaboration-workspaces/current/members"):                           "61ca1da6-f2fa-5e8d-9a64-dd1507691227",
	routeKey("DELETE", "/api/v1/collaboration-workspaces/current/members/:userId"):                 "c0ae5cc7-3549-50af-baeb-a25222a8dc02",
	routeKey("PUT", "/api/v1/collaboration-workspaces/current/members/:userId/role"):               "ec40cc4d-12c9-5e4f-914f-fb5123639aeb",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/members/:userId/roles"):              "57590a32-e816-55ad-9e55-9e0bf76f2812",
	routeKey("PUT", "/api/v1/collaboration-workspaces/current/members/:userId/roles"):              "4d43b70f-8b51-528c-904e-1e47e0a19730",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/roles"):                              "9567e6bc-ecae-578d-950d-f4ea233b7c9f",
	routeKey("POST", "/api/v1/collaboration-workspaces/current/roles"):                             "bc563c06-277f-5486-850a-ce08ef9ad5be",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/boundary/roles"):                     "b5ee7364-a5ef-5aec-9d74-494518905696",
	routeKey("PUT", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId"):             "000c3656-70bb-5ede-b130-b6cbc04f84ad",
	routeKey("DELETE", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId"):          "9d685085-a494-5e76-873c-d1fddcf8ebe1",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId/packages"):    "15814b0e-aaac-508d-ad0a-506e43c54862",
	routeKey("PUT", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId/packages"):    "8df6a884-acc3-5180-83fd-6f3a3aaa1d08",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId/menus"):       "ba9ac32a-ac3c-5aff-aced-272aaee22774",
	routeKey("PUT", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId/menus"):       "11293aad-695d-5ff1-9845-5c31466f3240",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId/actions"):     "511fa5fa-64b8-5ff3-b462-0e57562ff3c6",
	routeKey("PUT", "/api/v1/collaboration-workspaces/current/boundary/roles/:roleId/actions"):     "510d8fec-46fa-5ada-883c-fb34408a15b9",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/boundary/packages"):                  "d0e41c84-de57-57fe-b520-07238619a1ae",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/menus"):                              "f24e2f3e-f138-5827-aba3-6e0d4ea54213",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/menu-origins"):                       "1e5da814-9057-5e47-b313-71d92a3bbf91",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/actions"):                            "a52455a3-4498-5b0c-aaac-44e9243c3c53",
	routeKey("GET", "/api/v1/collaboration-workspaces/current/action-origins"):                     "c5b608a9-db39-58cf-a229-5be78f2beec5",
	routeKey("GET", "/api/v1/collaboration-workspaces"):                                            "d1fdb2f1-2c0e-5566-8b5f-43fd6bcbd802",
	routeKey("GET", "/api/v1/collaboration-workspaces/:id"):                                        "f9345c10-7465-54e4-8da4-1d908c505b54",
	routeKey("POST", "/api/v1/collaboration-workspaces"):                                           "b868c796-b65d-5361-b074-3ac0e0e4966f",
	routeKey("PUT", "/api/v1/collaboration-workspaces/:id"):                                        "d0d31f82-a24e-5f48-a463-2a9f3f30c890",
	routeKey("DELETE", "/api/v1/collaboration-workspaces/:id"):                                     "cd049d6e-ff6a-5501-a35d-4eadfd29f4c6",
	routeKey("GET", "/api/v1/collaboration-workspaces/:id/menus"):                                  "4257eda7-4b5c-53da-a308-9fc95fc2a603",
	routeKey("GET", "/api/v1/collaboration-workspaces/:id/menu-origins"):                           "43dfa344-c205-5962-9eb2-89383c08abaf",
	routeKey("PUT", "/api/v1/collaboration-workspaces/:id/menus"):                                  "0ccab0b0-624f-5fcf-89db-520d3111777b",
	routeKey("GET", "/api/v1/collaboration-workspaces/:id/actions"):                                "05cf49d3-7aca-599b-b5f8-58856c829b92",
	routeKey("GET", "/api/v1/collaboration-workspaces/:id/action-origins"):                         "72a1e7f0-62eb-54d9-8cd1-de0a40de2ca0",
	routeKey("PUT", "/api/v1/collaboration-workspaces/:id/actions"):                                "56b0e47f-e051-5269-95e9-5543aa771fc7",
	routeKey("GET", "/api/v1/collaboration-workspaces/:id/members"):                                "6ca647cd-8af1-5e46-878d-f71cee512d9b",
	routeKey("POST", "/api/v1/collaboration-workspaces/:id/members"):                               "43b06c81-9cc8-5dc5-9ddb-0a236548aabc",
	routeKey("DELETE", "/api/v1/collaboration-workspaces/:id/members/:userId"):                     "d5ae2913-de00-5022-8d0d-e4fed60d4674",
	routeKey("PUT", "/api/v1/collaboration-workspaces/:id/members/:userId/role"):                   "5e4b6e1f-c67c-5138-91bf-c17ca2519764",
	routeKey("GET", "/api/v1/collaboration-workspaces/options"):                                    "20ab7465-6ceb-5442-afc8-e2bebd61f43a",
	routeKey("GET", "/api/v1/users"):                                                               "4e9f6360-5aec-540c-886e-262f79ec896b",
	routeKey("GET", "/api/v1/users/:id"):                                                           "849f35ed-3a45-5515-add3-8e4a407b84cf",
	routeKey("GET", "/api/v1/users/:id/collaboration-workspaces"):                                  "e8a90bf9-e376-5661-8682-968de329f563",
	routeKey("GET", "/api/v1/users/:id/packages"):                                                  "ee41fcf8-5180-5576-8554-3316fd398ee0",
	routeKey("PUT", "/api/v1/users/:id/packages"):                                                  "ecccafd4-44b1-587b-b65e-c71d19889c35",
	routeKey("GET", "/api/v1/users/:id/menus"):                                                     "b68f2108-10b8-5d3e-b39d-57df5e29e8d7",
	routeKey("PUT", "/api/v1/users/:id/menus"):                                                     "810c2922-9e10-5c16-9c34-834bf9f8df80",
	routeKey("GET", "/api/v1/users/:id/permissions"):                                               "c3778011-36aa-5a9c-8a29-daac2f09f56b",
	routeKey("GET", "/api/v1/users/:id/permission-diagnosis"):                                      "6ac1b8eb-ed9f-5e58-bd36-d2ab0ff42a81",
	routeKey("POST", "/api/v1/users/:id/permission-refresh"):                                       "1a7772ed-0a52-5de9-9673-c323cf97989d",
	routeKey("POST", "/api/v1/users"):                                                              "79faec10-1c4b-5ddd-a075-c6f37ce1962e",
	routeKey("PUT", "/api/v1/users/:id"):                                                           "5dcdb2a9-25f9-5ec0-9c0c-73ee78c89da8",
	routeKey("DELETE", "/api/v1/users/:id"):                                                        "5d1d2d00-6e01-5ae6-8a3a-c5a969a909bf",
	routeKey("POST", "/api/v1/users/:id/roles"):                                                    "0495a5d2-e19d-5920-8745-75cb75bcf1eb",
}

func lookupFixedRouteCode(method, fullPath string) string {
	return fixedManagedRouteCodes[routeKey(method, fullPath)]
}

var fixedManagedCodeSet = buildFixedManagedCodeSet()

func buildFixedManagedCodeSet() map[string]struct{} {
	result := make(map[string]struct{}, len(fixedManagedRouteCodes))
	for _, code := range fixedManagedRouteCodes {
		target := strings.TrimSpace(code)
		if target == "" {
			continue
		}
		result[target] = struct{}{}
	}
	return result
}

// IsFixedManagedRouteCode reports whether code is a known fixed route code.
func IsFixedManagedRouteCode(code string) bool {
	_, ok := fixedManagedCodeSet[strings.TrimSpace(code)]
	return ok
}
