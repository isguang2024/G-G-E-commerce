package page

import "github.com/gin-gonic/gin"

func BuildRuntimePageMaps(items []Record) []gin.H {
	return buildRuntimePageRecords(items)
}
