package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/navigation"
)

// navigationAPIHandler 负责 /runtime/navigation。
//
// 与 dictionaryAPIHandler 的区别：navSvc 的构造链依赖 appSvc/menuSvc/pageSvc/
// spaceSvc 等多个协作服务，由 NewAPIHandler 统一编排后注入进来；sub-handler
// 只持有已构造好的 compiler，体现"依赖由外部注入、域 handler 不关心构造"的一面。
type navigationAPIHandler struct {
	navSvc navigation.Compiler
	logger *zap.Logger
}

func newNavigationAPIHandler(navSvc navigation.Compiler, logger *zap.Logger) *navigationAPIHandler {
	return &navigationAPIHandler{
		navSvc: navSvc,
		logger: logger,
	}
}

