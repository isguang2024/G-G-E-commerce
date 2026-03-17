package apiregistry

import (
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

type RouteMeta struct {
	Module                string
	Summary               string
	ResourceCode          string
	ActionCode            string
	ScopeCode             string
	FeatureKind           string
	RequiresTenantContext bool
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
	group  *gin.RouterGroup
	module string
}

func NewRegistrar(group *gin.RouterGroup, module string) *Registrar {
	return &Registrar{group: group, module: module}
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

func (r *Registrar) handle(method, relativePath string, meta *RouteMeta, handlers ...gin.HandlerFunc) gin.IRoutes {
	if meta != nil {
		resolved := *meta
		if strings.TrimSpace(resolved.Module) == "" {
			resolved.Module = r.module
		}
		Annotate(method, joinPath(r.group.BasePath(), relativePath), resolved)
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

	scopeIDs, err := loadScopeIDs(db)
	if err != nil {
		return err
	}
	return syncRoutesInternal(db, scopeIDs, logger, routes)
}

func syncRoutesInternal(
	db *gorm.DB,
	scopeIDs map[string]uuid.UUID,
	logger *zap.Logger,
	routes []gin.RouteInfo,
) error {
	for _, route := range routes {
		if !strings.HasPrefix(route.Path, "/api/v1/") && !strings.HasPrefix(route.Path, "/open/v1/") {
			continue
		}

		meta, _ := Lookup(route.Method, route.Path)
		moduleName := strings.TrimSpace(meta.Module)
		if moduleName == "" {
			moduleName = deriveModuleName(route.Path)
		}

		var scopeID *uuid.UUID
		if scopeCode := strings.TrimSpace(meta.ScopeCode); scopeCode != "" {
			if id, ok := scopeIDs[scopeCode]; ok {
				scopeValue := id
				scopeID = &scopeValue
			}
		}

		endpoint := &models.APIEndpoint{
			Method:                strings.ToUpper(route.Method),
			Path:                  route.Path,
			Module:                moduleName,
			FeatureKind:           normalizeFeatureKind(meta.FeatureKind),
			Handler:               route.Handler,
			Summary:               strings.TrimSpace(meta.Summary),
			ResourceCode:          strings.TrimSpace(meta.ResourceCode),
			ActionCode:            strings.TrimSpace(meta.ActionCode),
			ScopeID:               scopeID,
			RequiresTenantContext: meta.RequiresTenantContext,
			Status:                "normal",
		}
		if err := upsertEndpoint(db, endpoint); err != nil {
			return err
		}

		if err := ensurePermissionAction(db, endpoint, logger); err != nil {
			return err
		}
	}

	logger.Info("API endpoints synced", zap.Int("count", len(routes)))
	return nil
}

func ensurePermissionAction(db *gorm.DB, endpoint *models.APIEndpoint, logger *zap.Logger) error {
	if endpoint == nil || endpoint.ResourceCode == "" || endpoint.ActionCode == "" || endpoint.ScopeID == nil {
		return nil
	}

	var existing models.PermissionAction
	err := db.Where("resource_code = ? AND action_code = ? AND scope_id = ?", endpoint.ResourceCode, endpoint.ActionCode, *endpoint.ScopeID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	name := buildActionName(endpoint)
	description := buildActionDescription(endpoint)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		action := &models.PermissionAction{
			ResourceCode:          endpoint.ResourceCode,
			ActionCode:            endpoint.ActionCode,
			ModuleCode:            normalizeModuleCode(endpoint.Module, endpoint.ResourceCode),
			Category:              endpoint.Module,
			Source:                "api",
			FeatureKind:           normalizeFeatureKind(endpoint.FeatureKind),
			Name:                  name,
			Description:           description,
			ScopeID:               *endpoint.ScopeID,
			RequiresTenantContext: endpoint.RequiresTenantContext,
			Status:                "normal",
		}
		return db.Create(action).Error
	}

	updates := map[string]interface{}{}
	if existing.Source != "api" {
		updates["source"] = "api"
	}
	if normalizeFeatureKind(existing.FeatureKind) != normalizeFeatureKind(endpoint.FeatureKind) {
		updates["feature_kind"] = normalizeFeatureKind(endpoint.FeatureKind)
	}
	if strings.TrimSpace(endpoint.Module) != "" && strings.TrimSpace(existing.Category) != endpoint.Module {
		updates["category"] = endpoint.Module
	}
	if normalizedModuleCode := normalizeModuleCode(endpoint.Module, endpoint.ResourceCode); strings.TrimSpace(existing.ModuleCode) != normalizedModuleCode {
		updates["module_code"] = normalizedModuleCode
	}
	if existing.RequiresTenantContext != endpoint.RequiresTenantContext {
		updates["requires_tenant_context"] = endpoint.RequiresTenantContext
	}
	if strings.TrimSpace(existing.Status) == "" {
		updates["status"] = "normal"
	}
	if strings.TrimSpace(existing.Name) == "" {
		updates["name"] = name
	}
	if strings.TrimSpace(existing.Description) == "" {
		updates["description"] = description
	}
	if len(updates) == 0 {
		return nil
	}
	logger.Info("Permission action metadata updated from API registry",
		zap.String("resource", endpoint.ResourceCode),
		zap.String("action", endpoint.ActionCode))
	return db.Model(&models.PermissionAction{}).Where("id = ?", existing.ID).Updates(updates).Error
}

func loadScopeIDs(db *gorm.DB) (map[string]uuid.UUID, error) {
	var scopes []models.Scope
	if err := db.Find(&scopes).Error; err != nil {
		return nil, err
	}
	result := make(map[string]uuid.UUID, len(scopes))
	for _, scope := range scopes {
		result[scope.Code] = scope.ID
	}
	return result, nil
}

func deriveModuleName(routePath string) string {
	trimmed := strings.TrimPrefix(routePath, "/")
	segments := strings.Split(trimmed, "/")
	if len(segments) >= 3 {
		return segments[2]
	}
	if len(segments) >= 1 {
		return segments[len(segments)-1]
	}
	return "unknown"
}

func normalizeFeatureKind(value string) string {
	switch strings.TrimSpace(value) {
	case "business":
		return "business"
	default:
		return "system"
	}
}

func normalizeModuleCode(moduleName, fallbackResource string) string {
	if trimmed := strings.TrimSpace(moduleName); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallbackResource)
}

func buildActionName(endpoint *models.APIEndpoint) string {
	if endpoint == nil {
		return ""
	}
	if strings.TrimSpace(endpoint.Summary) != "" {
		return endpoint.Summary
	}
	return endpoint.ResourceCode + ":" + endpoint.ActionCode
}

func buildActionDescription(endpoint *models.APIEndpoint) string {
	if endpoint == nil {
		return ""
	}
	if strings.TrimSpace(endpoint.Summary) != "" {
		return endpoint.Summary + "（自动同步自接口注册表）"
	}
	return "自动同步自接口注册表：" + endpoint.Method + " " + endpoint.Path
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

func upsertEndpoint(db *gorm.DB, endpoint *models.APIEndpoint) error {
	if endpoint == nil {
		return nil
	}
	updates := map[string]interface{}{
		"module":                  endpoint.Module,
		"feature_kind":            normalizeFeatureKind(endpoint.FeatureKind),
		"handler":                 endpoint.Handler,
		"summary":                 endpoint.Summary,
		"resource_code":           endpoint.ResourceCode,
		"action_code":             endpoint.ActionCode,
		"scope_id":                endpoint.ScopeID,
		"requires_tenant_context": endpoint.RequiresTenantContext,
		"status":                  endpoint.Status,
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var existing models.APIEndpoint
		err := tx.Where("method = ? AND path = ?", endpoint.Method, endpoint.Path).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Create(endpoint).Error
			}
			return err
		}
		return tx.Model(&existing).Updates(updates).Error
	})
}
