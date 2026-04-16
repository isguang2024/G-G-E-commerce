package menu

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/user"
)

func TestMenuToRuntimeMapPreservesFalseIsEnable(t *testing.T) {
	node := menuToRuntimeMap(&user.Menu{
		ID:        uuid.New(),
		Path:      "collaboration",
		Name:      "TeamManage",
		Component: "/collaboration-workspace/workspaces",
		Title:     "所有协作空间",
		Meta: map[string]interface{}{
			"isEnable": false,
			"isHide":   true,
		},
	})

	meta, ok := node["meta"].(gin.H)
	if !ok {
		t.Fatalf("meta type = %T, want gin.H", node["meta"])
	}

	if value, exists := meta["isEnable"]; !exists {
		t.Fatalf("expected isEnable to be present in runtime meta")
	} else if enabled, ok := value.(bool); !ok || enabled {
		t.Fatalf("isEnable = %#v, want false", value)
	}

	if value, exists := meta["isHide"]; !exists {
		t.Fatalf("expected isHide to be preserved")
	} else if hidden, ok := value.(bool); !ok || !hidden {
		t.Fatalf("isHide = %#v, want true", value)
	}
}

