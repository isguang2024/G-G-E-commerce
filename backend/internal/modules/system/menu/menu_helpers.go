package menu

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

// menuToRuntimeMap converts a Menu model to the runtime navigation JSON shape.
func menuToRuntimeMap(m *user.Menu) gin.H {
	meta := gin.H{
		"title": m.Title,
	}
	if m.Icon != "" {
		meta["icon"] = m.Icon
	}
	if m.Meta != nil {
		if accessMode := strings.TrimSpace(toStringValue(m.Meta["accessMode"])); accessMode != "" {
			meta["accessMode"] = accessMode
		}
		if link := strings.TrimSpace(toStringValue(m.Meta["link"])); link != "" {
			meta["link"] = link
		}
		if activePath := strings.TrimSpace(toStringValue(m.Meta["activePath"])); activePath != "" {
			meta["activePath"] = activePath
		}
		if roles := filterStringArray(m.Meta["roles"]); len(roles) > 0 {
			meta["roles"] = roles
		}
		copyBool(meta, "isEnable", m.Meta["isEnable"])
		copyTruthyBool(meta, "isHide", m.Meta["isHide"])
		copyTruthyBool(meta, "isIframe", m.Meta["isIframe"])
		copyTruthyBool(meta, "isHideTab", m.Meta["isHideTab"])
		copyTruthyBool(meta, "keepAlive", m.Meta["keepAlive"])
		copyTruthyBool(meta, "fixedTab", m.Meta["fixedTab"])
		copyTruthyBool(meta, "isFullPage", m.Meta["isFullPage"])
	}

	node := gin.H{
		"id":         m.ID.String(),
		"app_key":    m.AppKey,
		"space_key":  m.SpaceKey,
		"kind":       m.Kind,
		"path":       m.Path,
		"name":       m.Name,
		"component":  m.Component,
		"meta":       meta,
		"sort_order": m.SortOrder,
	}
	if m.ParentID != nil {
		node["parent_id"] = m.ParentID.String()
	}
	if len(m.Children) > 0 {
		children := make([]gin.H, 0, len(m.Children))
		for _, ch := range m.Children {
			children = append(children, menuToRuntimeMap(ch))
		}
		node["children"] = children
	}
	return node
}

func copyBool(target gin.H, key string, value any) {
	if flag, ok := value.(bool); ok {
		target[key] = flag
	}
}

func copyTruthyBool(target gin.H, key string, value any) {
	if flag, ok := value.(bool); ok && flag {
		target[key] = true
	}
}

func toStringValue(value any) string {
	text, _ := value.(string)
	return text
}

func filterStringArray(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]string); ok {
			result := make([]string, 0, len(typed))
			for _, item := range typed {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		text, ok := item.(string)
		if !ok {
			continue
		}
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
