package system

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	space "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	cachepkg "github.com/gg-ecommerce/backend/internal/pkg/cache"
	"github.com/gg-ecommerce/backend/internal/pkg/module"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

type SystemModule struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

func NewSystemModule(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *SystemModule {
	return &SystemModule{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

func (m *SystemModule) Init() error {
	m.logger.Info("Initializing System module")
	return nil
}

func (m *SystemModule) RegisterRoutes(rg *gin.RouterGroup) {
	var systemCache *cachepkg.Cache
	systemCache, cacheErr := cachepkg.NewCache(
		m.config.Redis.Host,
		m.config.Redis.Port,
		m.config.Redis.Password,
		m.config.Redis.DB,
	)
	if cacheErr != nil {
		m.logger.Warn("Redis cache unavailable, page-association cache disabled", zap.Error(cacheErr))
	}

	fastEnterService := NewFastEnterService(m.db)
	messageService := NewMessageService(m.db, m.logger)
	systemHandler := NewSystemHandler(m.logger, systemCache, fastEnterService, messageService)
	boundaryService := teamboundary.NewService(m.db)
	platformService := platformaccess.NewService(m.db)
	roleSnapshotService := platformroleaccess.NewService(m.db)
	refresher := permissionrefresh.NewService(m.db, boundaryService, platformService, roleSnapshotService)
	menuSpaceService := space.NewService(m.db, refresher, m.logger)
	menuSpaceHandler := space.NewHandler(m.logger, menuSpaceService)
	authzService := authorization.NewService(m.db, m.logger)

	system := rg.Group("/system")
	reg := apiregistry.NewRegistrar(system, "system")
	{
		reg.GETProtected("/view-pages", reg.Meta("获取页面文件映射").BindPermissionKey("system.page_catalog.view").Build(), "system.page_catalog.view", authzService.RequireAction, systemHandler.GetViewPages)
		reg.GET("/fast-enter", reg.Meta("获取快捷入口配置").BindGroup("system").Build(), systemHandler.GetFastEnterConfig)
		reg.PUTProtected("/fast-enter", reg.Meta("更新快捷入口配置").BindGroup("system").BindPermissionKey("system.fast_enter.manage").Build(), "system.fast_enter.manage", authzService.RequireAction, systemHandler.UpdateFastEnterConfig)
		reg.GET("/menu-spaces/current", reg.Meta("获取当前菜单空间").BindGroup("system").Build(), menuSpaceHandler.GetCurrent)
		reg.GETProtected("/menu-spaces", reg.Meta("获取菜单空间列表").BindGroup("system").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuSpaceHandler.List)
		reg.POSTProtected("/menu-spaces", reg.Meta("保存菜单空间").BindGroup("system").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuSpaceHandler.SaveSpace)
		reg.POSTProtected("/menu-spaces/:spaceKey/initialize-default", reg.Meta("从默认空间初始化菜单空间").BindGroup("system").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuSpaceHandler.InitializeFromDefault)
		reg.GETProtected("/menu-space-host-bindings", reg.Meta("获取菜单空间Host绑定").BindGroup("system").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuSpaceHandler.ListHostBindings)
		reg.POSTProtected("/menu-space-host-bindings", reg.Meta("保存菜单空间Host绑定").BindGroup("system").BindPermissionKey("system.menu.manage").Build(), "system.menu.manage", authzService.RequireAction, menuSpaceHandler.SaveHostBinding)
	}

	messages := rg.Group("/messages")
	messageReg := apiregistry.NewRegistrar(messages, "message")
	{
		messageReg.GET("/inbox/summary", messageReg.Meta("获取消息摘要").BindGroup("system").Build(), systemHandler.GetInboxSummary)
		messageReg.GET("/inbox", messageReg.Meta("获取消息列表").BindGroup("system").Build(), systemHandler.ListInbox)
		messageReg.GET("/inbox/:deliveryId", messageReg.Meta("获取消息详情").BindGroup("system").Build(), systemHandler.GetInboxDetail)
		messageReg.POST("/inbox/:deliveryId/read", messageReg.Meta("标记消息已读").BindGroup("system").Build(), systemHandler.MarkInboxRead)
		messageReg.POST("/inbox/read-all", messageReg.Meta("批量标记消息已读").BindGroup("system").Build(), systemHandler.MarkInboxReadAll)
		messageReg.POST("/inbox/:deliveryId/todo-action", messageReg.Meta("处理待办消息").BindGroup("system").Build(), systemHandler.HandleInboxTodo)
		messageDispatchKeys := []string{"message.manage", "team.message.manage"}
		messageReg.GETActions("/dispatch/options", "获取消息发送配置", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.GetMessageDispatchOptions)
		messageReg.POSTActions("/dispatch", "发送站内消息", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.DispatchMessage)
		messageReg.GETActions("/templates", "获取消息模板", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.ListMessageTemplates)
		messageReg.POSTActions("/templates", "新建消息模板", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.SaveMessageTemplate)
		messageReg.PUTActions("/templates/:templateId", "更新消息模板", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.SaveMessageTemplate)
		messageReg.GETActions("/senders", "获取消息发送人", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.ListMessageSenders)
		messageReg.POSTActions("/senders", "新建消息发送人", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.SaveMessageSender)
		messageReg.PUTActions("/senders/:senderId", "更新消息发送人", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.SaveMessageSender)
		messageReg.GETActions("/recipient-groups", "获取消息接收组", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.ListMessageRecipientGroups)
		messageReg.POSTActions("/recipient-groups", "新建消息接收组", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.SaveMessageRecipientGroup)
		messageReg.PUTActions("/recipient-groups/:groupId", "更新消息接收组", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.SaveMessageRecipientGroup)
		messageReg.GETActions("/records", "获取消息发送记录", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.ListDispatchRecords)
		messageReg.GETActions("/records/:recordId", "获取消息发送记录详情", messageDispatchKeys, authzService.RequireAnyAction, systemHandler.GetDispatchRecordDetail)
	}
}

func init() {
	module.GetRegistry().Register(&systemModuleWrapper{})
}

type systemModuleWrapper struct{}

func (w *systemModuleWrapper) Init() error {
	return nil
}

func (w *systemModuleWrapper) RegisterRoutes(rg *gin.RouterGroup) {
}
