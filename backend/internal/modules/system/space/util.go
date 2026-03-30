package space

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const (
	DefaultMenuSpaceKey        = models.DefaultMenuSpaceKey
	requestSpaceKeyQuery       = "space_key"
	requestSpaceKeyHeader      = "X-Space-Key"
	requestSpaceAltHeader      = "X-Menu-Space"
	requestForwardedHostHeader = "X-Forwarded-Host"
	requestRealHostHeader      = "X-Real-Host"
	spaceAccessModeAll         = "all"
	spaceAccessModePlatform    = "platform_admin"
	spaceAccessModeTeam        = "team_admin"
	spaceAccessModeRoleCodes   = "role_codes"
	spaceModeSingle            = "single"
	spaceModeMulti             = "multi"
	menuSpaceModeSettingKey    = "system.menu_space.mode"
)

type SpaceAccessProfile struct {
	Mode             string   `json:"mode"`
	AllowedRoleCodes []string `json:"allowed_role_codes"`
}

func NormalizeSpaceKey(value string) string {
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

func CurrentSpaceMode(db *gorm.DB) string {
	if db == nil {
		return spaceModeSingle
	}
	var setting models.SystemSetting
	err := db.Where("key = ? AND deleted_at IS NULL", menuSpaceModeSettingKey).First(&setting).Error
	if err != nil {
		return spaceModeSingle
	}
	return NormalizeSpaceMode(toMetaString(setting.Value, "mode"))
}

func SaveCurrentSpaceMode(db *gorm.DB, mode string) (string, error) {
	normalized := NormalizeSpaceMode(mode)
	if db == nil {
		return normalized, nil
	}
	payload := models.MetaJSON{"mode": normalized}
	var setting models.SystemSetting
	err := db.Where("key = ? AND deleted_at IS NULL", menuSpaceModeSettingKey).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			record := models.SystemSetting{
				Key:    menuSpaceModeSettingKey,
				Value:  payload,
				Status: "normal",
			}
			return normalized, db.Create(&record).Error
		}
		return normalized, err
	}
	return normalized, db.Model(&setting).Updates(map[string]interface{}{
		"value":  payload,
		"status": "normal",
	}).Error
}

func IsSingleSpaceMode(db *gorm.DB) bool {
	return CurrentSpaceMode(db) == spaceModeSingle
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

func RequestSpaceKey(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value := strings.TrimSpace(c.Query(requestSpaceKeyQuery)); value != "" {
		return NormalizeSpaceKey(value)
	}
	for _, header := range []string{requestSpaceKeyHeader, requestSpaceAltHeader} {
		if value := strings.TrimSpace(c.GetHeader(header)); value != "" {
			return NormalizeSpaceKey(value)
		}
	}
	return ""
}

func ResolveSpaceKey(db *gorm.DB, c *gin.Context) string {
	if db == nil {
		return DefaultMenuSpaceKey
	}
	userID := currentContextUserID(c)
	tenantID := currentContextTenantID(c)
	if IsSingleSpaceMode(db) {
		explicit := RequestSpaceKey(c)
		if explicit == "" {
			return DefaultMenuSpaceKey
		}
		if explicit == DefaultMenuSpaceKey {
			return DefaultMenuSpaceKey
		}
		ok, err := spaceExists(db, explicit)
		if err == nil && ok {
			allowed, accessErr := CanAccessSpace(db, userID, tenantID, explicit)
			if accessErr == nil && allowed {
				return explicit
			}
		}
		return DefaultMenuSpaceKey
	}
	if explicit := RequestSpaceKey(c); explicit != "" && explicit != DefaultMenuSpaceKey {
		ok, err := spaceExists(db, explicit)
		if err == nil && ok {
			allowed, accessErr := CanAccessSpace(db, userID, tenantID, explicit)
			if accessErr == nil && allowed {
				return explicit
			}
		}
	}
	if explicit := RequestSpaceKey(c); explicit == DefaultMenuSpaceKey {
		ok, err := spaceExists(db, explicit)
		if err == nil && ok {
			return explicit
		}
	}
	host := RequestHost(c)
	if host != "" {
		var binding models.MenuSpaceHostBinding
		if err := db.Where("host = ? AND status = ?", host, "normal").First(&binding).Error; err == nil {
			key := NormalizeSpaceKey(binding.SpaceKey)
			ok, spaceErr := spaceExists(db, key)
			if spaceErr == nil && ok {
				allowed, accessErr := CanAccessSpace(db, userID, tenantID, key)
				if accessErr == nil && allowed {
					return key
				}
			}
		}
	}
	if allowed, err := CanAccessSpace(db, userID, tenantID, DefaultMenuSpaceKey); err == nil && allowed {
		return DefaultMenuSpaceKey
	}
	return DefaultMenuSpaceKey
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

func currentContextTenantID(c *gin.Context) *uuid.UUID {
	if c == nil {
		return nil
	}
	raw, ok := c.Get("tenant_id")
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

func ExtractSpaceAccessProfile(meta models.MetaJSON) SpaceAccessProfile {
	mode := strings.TrimSpace(toMetaString(meta, "access_mode", "accessMode"))
	switch mode {
	case spaceAccessModePlatform, spaceAccessModeTeam, spaceAccessModeRoleCodes:
	default:
		mode = spaceAccessModeAll
	}
	return SpaceAccessProfile{
		Mode:             mode,
		AllowedRoleCodes: normalizeMetaStringList(meta["allowed_role_codes"], meta["allowedRoleCodes"]),
	}
}

func CanAccessSpace(db *gorm.DB, userID *uuid.UUID, tenantID *uuid.UUID, spaceKey string) (bool, error) {
	if db == nil {
		return NormalizeSpaceKey(spaceKey) == DefaultMenuSpaceKey, nil
	}
	var record models.MenuSpace
	if err := db.Select("space_key", "status", "meta").
		Where("space_key = ? AND deleted_at IS NULL", NormalizeSpaceKey(spaceKey)).
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
	case spaceAccessModePlatform:
		if userID == nil {
			return false, nil
		}
		return hasPlatformRoleCode(db, *userID, []string{"admin"})
	case spaceAccessModeTeam:
		if userID == nil {
			return false, nil
		}
		return hasTeamAdminRole(db, *userID, tenantID)
	case spaceAccessModeRoleCodes:
		if userID == nil || len(profile.AllowedRoleCodes) == 0 {
			return false, nil
		}
		return hasAnyRoleCode(db, *userID, tenantID, profile.AllowedRoleCodes)
	default:
		return true, nil
	}
}

func hasPlatformRoleCode(db *gorm.DB, userID uuid.UUID, roleCodes []string) (bool, error) {
	var count int64
	err := db.Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.tenant_id IS NULL").
		Where("roles.code IN ?", roleCodes).
		Where("roles.status = ?", "normal").
		Count(&count).Error
	return count > 0, err
}

func hasTeamAdminRole(db *gorm.DB, userID uuid.UUID, tenantID *uuid.UUID) (bool, error) {
	query := db.Table("tenant_members").
		Where("user_id = ?", userID).
		Where("status = ?", "active").
		Where("role_code = ?", "team_admin")
	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func hasAnyRoleCode(db *gorm.DB, userID uuid.UUID, tenantID *uuid.UUID, roleCodes []string) (bool, error) {
	normalized := normalizeStringSlice(roleCodes)
	if len(normalized) == 0 {
		return false, nil
	}
	platformMatched, err := hasPlatformRoleCode(db, userID, normalized)
	if err != nil || platformMatched {
		return platformMatched, err
	}
	if tenantID != nil {
		var teamCount int64
		if err := db.Table("tenant_members").
			Where("user_id = ?", userID).
			Where("tenant_id = ?", *tenantID).
			Where("status = ?", "active").
			Where("role_code IN ?", normalized).
			Count(&teamCount).Error; err != nil {
			return false, err
		}
		if teamCount > 0 {
			return true, nil
		}
		var attachedRoleCount int64
		if err := db.Table("user_roles").
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("user_roles.user_id = ?", userID).
			Where("user_roles.tenant_id = ?", *tenantID).
			Where("roles.code IN ?", normalized).
			Where("roles.status = ?", "normal").
			Count(&attachedRoleCount).Error; err != nil {
			return false, err
		}
		if attachedRoleCount > 0 {
			return true, nil
		}
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

func ResolveSpaceKeyByHost(db *gorm.DB, host string) (string, *models.MenuSpaceHostBinding, error) {
	if IsSingleSpaceMode(db) {
		return DefaultMenuSpaceKey, nil, nil
	}
	if db == nil {
		return DefaultMenuSpaceKey, nil, nil
	}
	normalizedHost := NormalizeHost(host)
	if normalizedHost != "" {
		var binding models.MenuSpaceHostBinding
		if err := db.Where("host = ? AND status = ?", normalizedHost, "normal").First(&binding).Error; err == nil {
			key := NormalizeSpaceKey(binding.SpaceKey)
			ok, err := spaceExists(db, key)
			if err == nil && ok {
				return key, &binding, nil
			}
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultMenuSpaceKey, nil, err
		}
	}
	ok, err := spaceExists(db, DefaultMenuSpaceKey)
	if err != nil {
		return DefaultMenuSpaceKey, nil, err
	}
	if ok {
		return DefaultMenuSpaceKey, nil, nil
	}
	return DefaultMenuSpaceKey, nil, nil
}

func spaceExists(db *gorm.DB, spaceKey string) (bool, error) {
	var count int64
	if err := db.Model(&models.MenuSpace{}).
		Where("space_key = ? AND deleted_at IS NULL", NormalizeSpaceKey(spaceKey)).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
