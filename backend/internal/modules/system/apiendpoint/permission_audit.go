package apiendpoint

import (
	"strings"

	"github.com/maben/backend/internal/pkg/permissionkey"
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

// buildPermissionProfile builds the audit profile for an endpoint.
// accessModeMap is an optional "METHOD /path" -> access_mode lookup derived
// from the openapi seed; pass nil to fall back to binding-based inference.
func buildPermissionProfile(endpointMethod, endpointPath string, permissionKeys []string, accessModeMap map[string]string) permissionProfile {
	keys := normalizePermissionKeys(permissionKeys)
	contexts := make([]string, 0, len(keys))
	seenContexts := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		context := strings.TrimSpace(permissionkey.FromKey(key).ContextType)
		if context == "" {
			context = "collaboration"
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

	authMode := deriveEndpointAuthMode(endpointMethod, endpointPath, keys, accessModeMap)
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

// deriveEndpointAuthMode returns the effective auth mode for an endpoint.
// Source priority:
//  1. accessModeMap (from openapi_seed, injected by caller) — single source of truth
//  2. Permission-key binding inference (has binding → "permission")
//  3. Hard fallback: "jwt"
//
// Note: accessModeMap is injected to avoid an import cycle between apiendpoint
// and permissionseed. Callers should load it via permissionseed.LoadOpenAPISeed()
// and pass seed.AccessModeByMethodPath().
func deriveEndpointAuthMode(endpointMethod, endpointPath string, permissionKeys []string, accessModeMap map[string]string) string {
	if len(accessModeMap) > 0 {
		key := strings.ToUpper(strings.TrimSpace(endpointMethod)) + " " + strings.TrimSpace(endpointPath)
		if mode, ok := accessModeMap[key]; ok && mode != "" {
			// Normalise: spec uses "authenticated", UI/audit uses "jwt".
			if mode == "authenticated" {
				return "jwt"
			}
			return mode
		}
	}
	// Fallback: infer from binding table.
	if len(permissionKeys) > 0 {
		return "permission"
	}
	return "jwt"
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

