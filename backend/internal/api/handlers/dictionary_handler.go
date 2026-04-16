package handlers

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/dictionary"
)

// dictionaryAPIHandler 负责 /dict-* 相关 OpenAPI 操作。
//
// 这是 god handler 拆分示范（新5）：APIHandler 原先把 40+ 域的字段混在一个 struct
// 里，改成每个域独立 sub-handler，通过匿名指针嵌入到 APIHandler —— ogen 通过方法
// 提升仍然满足 gen.Handler 接口，但域内代码只能看到自己真正需要的依赖。
//
// 后续批次按同样模式迁移：
//  1. 在本目录新建 {domain}_handler.go，定义 {domain}APIHandler 结构体 + 构造函数；
//  2. 把 {domain}.go 里所有 receiver 从 *APIHandler 改成 *{domain}APIHandler；
//  3. workspace.go 中删除原字段、改为嵌入 *{domain}APIHandler，并在 NewAPIHandler
//     里调用对应构造函数。
type dictionaryAPIHandler struct {
	dictSvc *dictionary.Service
	logger  *zap.Logger
}

// newDictionaryAPIHandler 构建数据字典域 sub-handler。
func newDictionaryAPIHandler(db *gorm.DB, logger *zap.Logger) *dictionaryAPIHandler {
	return &dictionaryAPIHandler{
		dictSvc: dictionary.NewService(db, logger),
		logger:  logger,
	}
}
