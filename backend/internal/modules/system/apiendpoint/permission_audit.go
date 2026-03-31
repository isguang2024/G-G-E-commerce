package apiendpoint

import (
	"strings"

	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
)

const (
	permissionPatternNone               = "none"
	permissionPatternPublic             = "public"
	permissionPatternGlobalJWT          = "global_jwt"
	permissionPatternSelfJWT            = "self_jwt"
	permissionPatternAPIKey             = "api_key"
	permissionPatternSingle             = "single"
	permissionPatternShared             = "shared"
	permissionPatternCrossContextShared = "cross_context_shared"
)

type permissionProfile struct {
	PrimaryKey           string
	Keys                 []string
	Contexts             []string
	BindingMode          string
	SharedAcrossContexts bool
	Note                 string
}

func buildPermissionProfile(endpointPath string, permissionKeys []string) permissionProfile {
	keys := normalizePermissionKeys(permissionKeys)
	contexts := make([]string, 0, len(keys))
	seenContexts := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		context := strings.TrimSpace(permissionkey.FromKey(key).ContextType)
		if context == "" {
			context = "team"
		}
		if _, ok := seenContexts[context]; ok {
			continue
		}
		seenContexts[context] = struct{}{}
		contexts = append(contexts, context)
	}

	profile := permissionProfile{
		Keys:        keys,
		Contexts:    contexts,
		BindingMode: permissionPatternNone,
	}
	if len(keys) > 0 {
		profile.PrimaryKey = keys[0]
	}

	authMode := deriveEndpointAuthMode(endpointPath, keys)
	switch len(keys) {
	case 0:
		switch authMode {
		case "public":
			profile.BindingMode = permissionPatternPublic
			profile.Note = "公开接口，无需权限键"
		case "api_key":
			profile.BindingMode = permissionPatternAPIKey
			profile.Note = "开放接口，无需权限键"
		default:
			if isGlobalJWTEndpoint(endpointPath) {
				profile.BindingMode = permissionPatternGlobalJWT
				profile.Note = "登录态全局接口，无需权限键"
			} else {
				profile.BindingMode = permissionPatternSelfJWT
				profile.Note = "登录态自服务接口，无需权限键"
			}
		}
	case 1:
		profile.BindingMode = permissionPatternSingle
		profile.Note = "单权限接口"
	default:
		profile.BindingMode = permissionPatternShared
		profile.Note = "多权限共享接口，按任一权限放行"
		if len(contexts) > 1 {
			profile.BindingMode = permissionPatternCrossContextShared
			profile.SharedAcrossContexts = true
			profile.Note = "跨上下文共享接口，按任一权限放行"
		}
	}

	return profile
}

func deriveEndpointAuthMode(endpointPath string, permissionKeys []string) string {
	switch {
	case endpointPath == "/health":
		return "public"
	case endpointPath == "/api/v1/auth/login" || endpointPath == "/api/v1/auth/register" || endpointPath == "/api/v1/auth/refresh":
		return "public"
	case endpointPath == "/api/v1/pages/runtime/public":
		return "public"
	case strings.HasPrefix(endpointPath, "/open/v1/"):
		return "api_key"
	case len(permissionKeys) > 0:
		return "permission"
	default:
		return "jwt"
	}
}

func normalizePermissionKeys(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		target := strings.TrimSpace(value)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		result = append(result, target)
	}
	return result
}

func isGlobalJWTEndpoint(endpointPath string) bool {
	switch {
	case endpointPath == "/api/v1/pages/runtime":
		return true
	case endpointPath == "/api/v1/system/fast-enter":
		return true
	case endpointPath == "/api/v1/system/menu-spaces/current":
		return true
	case endpointPath == "/api/v1/menus/tree":
		return true
	case strings.HasPrefix(endpointPath, "/api/v1/runtime/"):
		return true
	default:
		return false
	}
}
