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
	group *gin.RouterGroup
}

type RequireActionFunc func(permissionKey string, legacy ...string) gin.HandlerFunc
type RequireAnyActionFunc func(permissionKeys ...string) gin.HandlerFunc

func NewRegistrar(group *gin.RouterGroup, _ string) *Registrar {
	return &Registrar{group: group}
}

func Meta(summary string) *MetaBuilder {
	return (&MetaBuilder{}).WithSummary(summary)
}

func (r *Registrar) Meta(summary string) *MetaBuilder {
	return Meta(summary)
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
	return r.GET(relativePath, MetaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) GETActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.GET(relativePath, MetaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) GETProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.GET(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) POSTAction(relativePath, summary, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.POST(relativePath, MetaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) POSTActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.POST(relativePath, MetaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) POSTProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.POST(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) PUTAction(relativePath, summary, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.PUT(relativePath, MetaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) PUTActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.PUT(relativePath, MetaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) PUTProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.PUT(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) DELETEAction(relativePath, summary, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.DELETE(relativePath, MetaWithPermission(summary, permissionKey), appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func (r *Registrar) DELETEActions(relativePath, summary string, permissionKeys []string, requireAction RequireAnyActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.DELETE(relativePath, MetaWithPermissions(summary, permissionKeys), appendRequireAnyActionHandler(permissionKeys, requireAction, handlers)...)
}

func (r *Registrar) DELETEProtected(relativePath string, meta *RouteMeta, permissionKey string, requireAction RequireActionFunc, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.DELETE(relativePath, meta, appendRequireActionHandler(permissionKey, requireAction, handlers)...)
}

func MetaWithPermission(summary, permissionKey string) *RouteMeta {
	return Meta(summary).
		BindPermissionKey(permissionKey).
		BindContextScope("optional").
		BindSource("sync").
		Build()
}

func MetaWithPermissions(summary string, permissionKeys []string) *RouteMeta {
	return Meta(summary).
		BindPermissionKeys(permissionKeys...).
		BindContextScope("optional").
		BindSource("sync").
		Build()
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
	if meta != nil {
		Annotate(method, joinPath(r.group.BasePath(), relativePath), *meta)
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

func SyncRoutes(db *gorm.DB, logger *zap.Logger, routes []gin.RouteInfo) error {
	if db == nil {
		return errors.New("database is nil")
	}
	return syncRoutesInternal(db, logger, routes)
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
		existing, err := findEndpointByMethodAndPath(db, route.Method, route.Path)
		if err != nil {
			return err
		}
		if !hasMeta && existing == nil {
			continue
		}

		endpointCode := strings.TrimSpace(meta.Code)
		if endpointCode == "" {
			if existing != nil && strings.TrimSpace(existing.Code) != "" {
				endpointCode = strings.TrimSpace(existing.Code)
			} else {
				endpointCode = deriveStableEndpointCode(route.Method, route.Path)
			}
		}

		summary := strings.TrimSpace(meta.Summary)
		if summary == "" && existing != nil {
			summary = strings.TrimSpace(existing.Summary)
		}

		featureKind := normalizeFeatureKind(meta.FeatureKind)
		if strings.TrimSpace(meta.FeatureKind) == "" && existing != nil {
			featureKind = normalizeFeatureKind(existing.FeatureKind)
		}

		categoryID := resolveCategoryID(db, meta.CategoryCode)
		if categoryID == nil && existing != nil {
			categoryID = existing.CategoryID
		}

		contextScope := normalizeContextScope(meta.ContextScope)
		if strings.TrimSpace(meta.ContextScope) == "" && existing != nil {
			contextScope = normalizeContextScope(existing.ContextScope)
		}

		source := normalizeSource(meta.Source)
		if !hasMeta && existing != nil {
			source = normalizeSource(existing.Source)
		}

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
			Status:       resolveSyncedEndpointStatus(existing),
		}
		if existing != nil {
			endpoint.ID = existing.ID
		}
		if err := upsertEndpoint(db, endpoint); err != nil {
			return err
		}

		if hasMeta {
			if err := replaceEndpointPermissionBindings(db, endpoint, meta.PermissionKeys); err != nil {
				return err
			}
		}
	}

	logger.Info("API endpoints synced", zap.Int("count", len(routes)))
	return nil
}

func findEndpointByMethodAndPath(db *gorm.DB, method, path string) (*models.APIEndpoint, error) {
	if db == nil {
		return nil, nil
	}
	var endpoint models.APIEndpoint
	err := db.Where("method = ? AND path = ?", strings.ToUpper(strings.TrimSpace(method)), strings.TrimSpace(path)).First(&endpoint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &endpoint, nil
}

func isManagedRoute(path string) bool {
	return strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/open/v1/")
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

func deriveStableEndpointCode(method, path string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
	return uuid.NewHash(sha1.New(), uuid.NameSpaceURL, []byte("api-endpoint:"+normalized), 5).String()
}

func upsertEndpoint(db *gorm.DB, endpoint *models.APIEndpoint) error {
	if endpoint == nil {
		return nil
	}
	normalizedFeatureKind := normalizeFeatureKind(endpoint.FeatureKind)
	normalizedContextScope := normalizeContextScope(endpoint.ContextScope)
	normalizedSource := normalizeSource(endpoint.Source)
	return db.Transaction(func(tx *gorm.DB) error {
		var existing models.APIEndpoint
		var err error
		if endpoint.Code != "" {
			err = tx.Where("code = ?", endpoint.Code).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = tx.Where("method = ? AND path = ?", endpoint.Method, endpoint.Path).First(&existing).Error
			}
		} else {
			err = tx.Where("method = ? AND path = ?", endpoint.Method, endpoint.Path).First(&existing).Error
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				endpoint.FeatureKind = normalizedFeatureKind
				endpoint.ContextScope = normalizedContextScope
				endpoint.Source = normalizedSource
				return tx.Create(endpoint).Error
			}
			return err
		}
		updates := make(map[string]interface{})
		if endpoint.Code != "" && existing.Code != endpoint.Code {
			updates["code"] = endpoint.Code
		}
		if existing.FeatureKind != normalizedFeatureKind {
			updates["feature_kind"] = normalizedFeatureKind
		}
		if existing.Handler != endpoint.Handler {
			updates["handler"] = endpoint.Handler
		}
		if existing.Summary != endpoint.Summary {
			updates["summary"] = endpoint.Summary
		}
		if !sameUUIDPointer(existing.CategoryID, endpoint.CategoryID) {
			updates["category_id"] = endpoint.CategoryID
		}
		if existing.ContextScope != normalizedContextScope {
			updates["context_scope"] = normalizedContextScope
		}
		if existing.Source != normalizedSource {
			updates["source"] = normalizedSource
		}
		if existing.Status != endpoint.Status {
			updates["status"] = endpoint.Status
		}
		endpoint.ID = existing.ID
		if len(updates) == 0 {
			return nil
		}
		return tx.Model(&existing).Updates(updates).Error
	})
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
	keys := normalizePermissionKeys(permissionKeys)
	return db.Transaction(func(tx *gorm.DB) error {
		var existing []models.APIEndpointPermissionBinding
		if err := tx.
			Where("endpoint_id = ?", endpoint.ID).
			Order("sort_order ASC, created_at ASC").
			Find(&existing).Error; err != nil {
			return err
		}
		if samePermissionBindings(existing, keys) {
			return nil
		}
		if err := tx.Where("endpoint_id = ?", endpoint.ID).Delete(&models.APIEndpointPermissionBinding{}).Error; err != nil {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		items := make([]models.APIEndpointPermissionBinding, 0, len(keys))
		for idx, key := range keys {
			items = append(items, models.APIEndpointPermissionBinding{
				EndpointID:    endpoint.ID,
				PermissionKey: key,
				MatchMode:     "ANY",
				SortOrder:     idx,
			})
		}
		return tx.Create(&items).Error
	})
}

func sameUUIDPointer(left, right *uuid.UUID) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
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

func resolveSyncedEndpointStatus(existing *models.APIEndpoint) string {
	if existing == nil {
		return "normal"
	}
	status := strings.TrimSpace(existing.Status)
	if status == "" {
		return "normal"
	}
	return status
}
