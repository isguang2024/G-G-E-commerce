package space

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/pkg/workspacerolebinding"
)

const (
	DefaultMenuSpaceKey          = models.DefaultMenuSpaceKey
	requestMenuSpaceKeyQuery     = "menu_space_key"
	requestForwardedHostHeader   = "X-Forwarded-Host"
	requestRealHostHeader        = "X-Real-Host"
	spaceAccessModeAll           = "all"
	spaceAccessModePersonal      = "personal_workspace_admin"
	spaceAccessModeCollaboration = "collaboration_admin"
	spaceAccessModeRoleCodes     = "role_codes"
	spaceModeSingle              = "single"
	spaceModeMulti               = "multi"
)

type SpaceAccessProfile struct {
	Mode             string   `json:"mode"`
	AllowedRoleCodes []string `json:"allowed_role_codes"`
}

func NormalizeMenuSpaceKey(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return DefaultMenuSpaceKey
	}
	return target
}

func NormalizeSpaceMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case spaceModeMulti:
		return spaceModeMulti
	default:
		return spaceModeSingle
	}
}

func CurrentSpaceMode(db *gorm.DB, appKey string) string {
	if db == nil {
		return spaceModeSingle
	}
	var app models.App
	err := db.Where("app_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey)).First(&app).Error
	if err != nil {
		return spaceModeSingle
	}
	return NormalizeSpaceMode(app.SpaceMode)
}

func SaveCurrentSpaceMode(db *gorm.DB, appKey, mode string) (string, error) {
	normalized := NormalizeSpaceMode(mode)
	if db == nil {
		return normalized, nil
	}
	return normalized, db.Model(&models.App{}).
		Where("app_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey)).
		Updates(map[string]interface{}{"space_mode": normalized, "status": "normal"}).Error
}

func IsSingleSpaceMode(db *gorm.DB, appKey string) bool {
	return CurrentSpaceMode(db, appKey) == spaceModeSingle
}

func NormalizeHost(value string) string {
	target := strings.TrimSpace(strings.ToLower(value))
	if target == "" {
		return ""
	}
	if idx := strings.Index(target, ","); idx >= 0 {
		target = strings.TrimSpace(target[:idx])
	}
	if idx := strings.Index(target, "://"); idx >= 0 {
		if parsed, err := url.Parse(target); err == nil && parsed.Host != "" {
			target = parsed.Host
		}
	}
	if host, _, err := net.SplitHostPort(target); err == nil {
		target = host
	}
	target = strings.TrimSuffix(target, ".")
	return target
}

func RequestHost(c *gin.Context) string {
	if c == nil {
		return ""
	}
	for _, header := range []string{requestForwardedHostHeader, requestRealHostHeader} {
		if value := NormalizeHost(c.GetHeader(header)); value != "" {
			return value
		}
	}
	if c.Request == nil {
		return ""
	}
	return NormalizeHost(c.Request.Host)
}

func RequestMenuSpaceKey(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value := strings.TrimSpace(c.Query(requestMenuSpaceKeyQuery)); value != "" {
		return NormalizeMenuSpaceKey(value)
	}
	return ""
}

func ResolveMenuSpaceKey(db *gorm.DB, c *gin.Context) string {
	if db == nil {
		return DefaultMenuSpaceKey
	}
	appKey := currentContextAppKey(c)
	userID := currentContextUserID(c)
	collaborationWorkspaceID := ResolveContextCollaborationWorkspaceID(db, c)
	key, _, err := ResolveCurrentMenuSpaceKey(db, appKey, RequestHost(c), RequestMenuSpaceKey(c), userID, collaborationWorkspaceID)
	if err == nil && strings.TrimSpace(key) != "" {
		return key
	}
	return DefaultMenuSpaceKey
}

func ResolveCurrentMenuSpaceKey(db *gorm.DB, appKey, host, requestedMenuSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) (string, string, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	defaultMenuSpaceKey, err := loadAppDefaultMenuSpaceKey(db, normalizedAppKey)
	if err != nil {
		return DefaultMenuSpaceKey, "", err
	}
	if defaultMenuSpaceKey == "" {
		defaultMenuSpaceKey = DefaultMenuSpaceKey
	}

	if IsSingleSpaceMode(db, normalizedAppKey) {
		explicit := NormalizeMenuSpaceKey(requestedMenuSpaceKey)
		if explicit != "" && explicit != defaultMenuSpaceKey {
			ok, existsErr := spaceExists(db, normalizedAppKey, explicit)
			if existsErr != nil {
				return defaultMenuSpaceKey, "", existsErr
			}
			if ok {
				allowed, accessErr := CanAccessSpace(
					db,
					userID,
					collaborationWorkspaceID,
					normalizedAppKey,
					explicit,
				)
				if accessErr != nil {
					return defaultMenuSpaceKey, "", accessErr
				}
				if allowed {
					return explicit, "single_mode_explicit", nil
				}
			}
		}
		return defaultMenuSpaceKey, "single_mode_default", nil
	}

	if explicit := NormalizeMenuSpaceKey(requestedMenuSpaceKey); explicit != "" {
		ok, existsErr := spaceExists(db, normalizedAppKey, explicit)
		if existsErr != nil {
			return defaultMenuSpaceKey, "", existsErr
		}
		if ok {
			allowed, accessErr := CanAccessSpace(
				db,
				userID,
				collaborationWorkspaceID,
				normalizedAppKey,
				explicit,
			)
			if accessErr != nil {
				return defaultMenuSpaceKey, "", accessErr
			}
			if allowed {
				return explicit, "explicit", nil
			}
		}
	}

	resolvedByHost, source, hostErr := ResolveMenuSpaceKeyByHost(db, normalizedAppKey, host)
	if hostErr != nil {
		return defaultMenuSpaceKey, "", hostErr
	}
	if strings.TrimSpace(resolvedByHost) != "" {
		allowed, accessErr := CanAccessSpace(
			db,
			userID,
			collaborationWorkspaceID,
			normalizedAppKey,
			resolvedByHost,
		)
		if accessErr != nil {
			return defaultMenuSpaceKey, "", accessErr
		}
		if allowed {
			return resolvedByHost, source, nil
		}
	}

	if allowed, accessErr := CanAccessSpace(
		db,
		userID,
		collaborationWorkspaceID,
		normalizedAppKey,
		defaultMenuSpaceKey,
	); accessErr == nil && allowed {
		return defaultMenuSpaceKey, "app_default", nil
	}

	return defaultMenuSpaceKey, "fallback_default", nil
}

func currentContextUserID(c *gin.Context) *uuid.UUID {
	if c == nil {
		return nil
	}
	raw, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &id
}

func currentContextCollaborationWorkspaceID(c *gin.Context) *uuid.UUID {
	if c == nil {
		return nil
	}
	raw, ok := c.Get("collaboration_workspace_id")
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &id
}

func currentContextAuthWorkspaceID(c *gin.Context) *uuid.UUID {
	if c == nil {
		return nil
	}
	raw, ok := c.Get("auth_workspace_id")
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &id
}

func currentContextAuthWorkspaceType(c *gin.Context) string {
	if c == nil {
		return ""
	}
	raw, ok := c.Get("auth_workspace_type")
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

// ResolveContextCollaborationWorkspaceID 优先根据 canonical auth workspace 解析当前协作空间，
// 仅在旧链路尚未完全下线时回退到 collaboration_workspace_id。
func ResolveContextCollaborationWorkspaceID(db *gorm.DB, c *gin.Context) *uuid.UUID {
	if c == nil {
		return nil
	}
	if db != nil && currentContextAuthWorkspaceType(c) == models.WorkspaceTypeCollaboration {
		if workspaceID := currentContextAuthWorkspaceID(c); workspaceID != nil {
			var workspace models.Workspace
			if err := db.Select("collaboration_workspace_id").
				Where("id = ? AND workspace_type = ? AND deleted_at IS NULL", *workspaceID, models.WorkspaceTypeCollaboration).
				First(&workspace).Error; err == nil {
				if workspace.CollaborationWorkspaceID != nil && *workspace.CollaborationWorkspaceID != uuid.Nil {
					return workspace.CollaborationWorkspaceID
				}
			}
		}
	}
	return currentContextCollaborationWorkspaceID(c)
}

func ExtractSpaceAccessProfile(meta models.MetaJSON) SpaceAccessProfile {
	mode := strings.TrimSpace(toMetaString(meta, "access_mode", "accessMode"))
	switch mode {
	case spaceAccessModePersonal, spaceAccessModeCollaboration, spaceAccessModeRoleCodes:
	default:
		mode = spaceAccessModeAll
	}
	return SpaceAccessProfile{
		Mode:             mode,
		AllowedRoleCodes: normalizeMetaStringList(meta["allowed_role_codes"], meta["allowedRoleCodes"]),
	}
}

func CanAccessSpace(
	db *gorm.DB,
	userID *uuid.UUID,
	collaborationWorkspaceID *uuid.UUID,
	appKey string,
	spaceKey string,
) (bool, error) {
	if db == nil {
		return NormalizeMenuSpaceKey(spaceKey) == DefaultMenuSpaceKey, nil
	}
	normalizedAppKey := normalizeAppKey(appKey)
	var record models.MenuSpace
	if err := db.Select("menu_space_key", "status", "meta").
		Where(
			"app_key = ? AND menu_space_key = ? AND deleted_at IS NULL",
			normalizedAppKey,
			NormalizeMenuSpaceKey(spaceKey),
		).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if record.Status == "disabled" {
		return false, nil
	}
	profile := ExtractSpaceAccessProfile(record.Meta)
	switch profile.Mode {
	case spaceAccessModeAll:
		return true, nil
	case spaceAccessModePersonal:
		if userID == nil {
			return false, nil
		}
		return hasPersonalWorkspaceRoleCode(db, *userID, []string{"admin"})
	case spaceAccessModeCollaboration:
		if userID == nil {
			return false, nil
		}
		return hasCollaborationWorkspaceAdminRole(db, *userID, collaborationWorkspaceID)
	case spaceAccessModeRoleCodes:
		if userID == nil || len(profile.AllowedRoleCodes) == 0 {
			return false, nil
		}
		return hasAnyRoleCode(db, *userID, collaborationWorkspaceID, profile.AllowedRoleCodes)
	default:
		return true, nil
	}
}

func hasPersonalWorkspaceRoleCode(db *gorm.DB, userID uuid.UUID, roleCodes []string) (bool, error) {
	return workspacerolebinding.HasPersonalRoleCodesByUserID(db, userID, roleCodes, true)
}

func hasCollaborationWorkspaceAdminRole(db *gorm.DB, userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) (bool, error) {
	if collaborationWorkspaceID == nil {
		return false, nil
	}
	return workspacerolebinding.HasCollaborationWorkspaceRoleCodesByUser(
		db, *collaborationWorkspaceID, userID,
		[]string{"collaboration_admin"}, true,
	)
}

func hasAnyRoleCode(db *gorm.DB, userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, roleCodes []string) (bool, error) {
	normalized := normalizeStringSlice(roleCodes)
	if len(normalized) == 0 {
		return false, nil
	}
	personalMatched, err := hasPersonalWorkspaceRoleCode(db, userID, normalized)
	if err != nil || personalMatched {
		return personalMatched, err
	}
	if collaborationWorkspaceID != nil {
		return workspacerolebinding.HasCollaborationWorkspaceRoleCodesByUser(db, *collaborationWorkspaceID, userID, normalized, true)
	}
	return false, nil
}

func normalizeStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func toMetaString(meta models.MetaJSON, keys ...string) string {
	for _, key := range keys {
		if value, ok := meta[key]; ok {
			if text, ok := value.(string); ok {
				return text
			}
		}
	}
	return ""
}

func normalizeMetaStringList(values ...interface{}) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, raw := range values {
		switch typed := raw.(type) {
		case []string:
			for _, item := range typed {
				value := strings.TrimSpace(item)
				if value == "" {
					continue
				}
				if _, ok := seen[value]; ok {
					continue
				}
				seen[value] = struct{}{}
				result = append(result, value)
			}
		case []interface{}:
			for _, item := range typed {
				text, ok := item.(string)
				if !ok {
					continue
				}
				value := strings.TrimSpace(text)
				if value == "" {
					continue
				}
				if _, ok := seen[value]; ok {
					continue
				}
				seen[value] = struct{}{}
				result = append(result, value)
			}
		}
	}
	return result
}

func ResolveMenuSpaceKeyByHost(db *gorm.DB, appKey, host string) (string, string, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	if IsSingleSpaceMode(db, normalizedAppKey) {
		return DefaultMenuSpaceKey, "single_mode_default", nil
	}
	if db == nil {
		return DefaultMenuSpaceKey, "fallback_default", nil
	}
	normalizedHost := NormalizeHost(host)
	if normalizedHost != "" {
		var appBinding models.AppHostBinding
		if err := db.Where("host = ? AND status = ? AND deleted_at IS NULL", normalizedHost, "normal").First(&appBinding).Error; err == nil {
			key := NormalizeMenuSpaceKey(appBinding.DefaultMenuSpaceKey)
			if normalizeAppKey(appBinding.AppKey) == normalizedAppKey {
				ok, err := spaceExists(db, normalizedAppKey, key)
				if err == nil && ok {
					return key, "app_host_binding", nil
				}
			}
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultMenuSpaceKey, "", err
		}

	}
	defaultMenuSpaceKey, err := loadAppDefaultMenuSpaceKey(db, normalizedAppKey)
	if err != nil {
		return DefaultMenuSpaceKey, "", err
	}
	if defaultMenuSpaceKey == "" {
		defaultMenuSpaceKey = DefaultMenuSpaceKey
	}
	ok, err := spaceExists(db, normalizedAppKey, defaultMenuSpaceKey)
	if err != nil {
		return defaultMenuSpaceKey, "", err
	}
	if ok {
		return defaultMenuSpaceKey, "app_default", nil
	}
	return defaultMenuSpaceKey, "fallback_default", nil
}

func spaceExists(db *gorm.DB, appKey, spaceKey string) (bool, error) {
	var count int64
	if err := db.Model(&models.MenuSpace{}).
		Where("app_key = ? AND menu_space_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey), NormalizeMenuSpaceKey(spaceKey)).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func loadAppDefaultMenuSpaceKey(db *gorm.DB, appKey string) (string, error) {
	if db == nil {
		return DefaultMenuSpaceKey, nil
	}
	var app models.App
	err := db.Select("default_menu_space_key").Where("app_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey)).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultMenuSpaceKey, nil
	}
	if err != nil {
		return DefaultMenuSpaceKey, err
	}
	return NormalizeMenuSpaceKey(app.DefaultMenuSpaceKey), nil
}

func normalizeAppKey(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return models.DefaultAppKey
	}
	return target
}

func currentContextAppKey(c *gin.Context) string {
	if c == nil {
		return models.DefaultAppKey
	}
	return normalizeAppKey(c.GetString("app_key"))
}
