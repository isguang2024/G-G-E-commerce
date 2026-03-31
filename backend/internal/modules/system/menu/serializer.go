package menu

import (
	"github.com/gin-gonic/gin"

	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

func BuildRuntimeTreeMaps(tree []*user.Menu) []gin.H {
	if len(tree) == 0 {
		return []gin.H{}
	}
	result := make([]gin.H, 0, len(tree))
	for _, node := range tree {
		result = append(result, menuToRuntimeMap(node))
	}
	return result
}
