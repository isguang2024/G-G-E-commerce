// Package apperr 是 ogen handler 的统一错误翻译层。
//
// 职责分工：
//   - codes.go  — 业务码常量（唯一真源）。前端 error-codes.ts 由 cmd/gen-permissions 从此文件派生。
//   - mapper.go — error → gen.Error 的翻译表，以及挂到 ogen 的 ErrorHandler。
//
// 新增/修改错误码只需改 codes.go + mapper.go，spec 注释引用它，前端自动同步。
package apperr

// 业务码段：
//
//	1xxxx  参数 / 请求错误
//	2xxxx  认证 / 授权错误
//	3xxxx  业务 / 资源错误
//	5xxxx  服务端错误
const (
	// ── 1xxxx 参数 / 请求 ──────────────────────────────────────────────────
	CodeParamInvalid = 1001 // 参数错误（通用）
	CodeParamMissing = 1002 // 参数缺失
	CodeParamFormat  = 1003 // 参数格式错误（JSON 解析失败、类型不匹配等）
	CodeInvalidID    = 1004 // 无效的资源 ID

	// ── 2xxxx 认证 / 授权 ──────────────────────────────────────────────────
	CodeUnauthorized       = 2001 // 未登录或 token 无效
	CodeTokenExpired       = 2002 // token 已过期
	CodeForbidden          = 2003 // 无权限
	CodeAPIKeyMissing      = 2004 // 缺少 API Key
	CodeTokenBadFormat     = 2005 // token 格式错误
	CodeInvalidCredentials = 2006 // 邮箱或密码错误
	CodeUserInactive       = 2007 // 账号已被禁用

	// ── 3xxxx 业务 / 资源 ──────────────────────────────────────────────────
	CodeNotFound                    = 3001 // 资源不存在（通用）
	CodeUserNotFound                = 3002 // 用户不存在
	CodeWorkspaceNotFound           = 3003 // 协作空间不存在
	CodeMenuNotFound                = 3004 // 菜单不存在
	CodeRoleNotFound                = 3005 // 角色不存在
	CodeNoManagedWorkspace          = 3006 // 暂无管理的协作空间
	CodeRoleCodeExists              = 3007 // 角色编码已存在
	CodeWorkspaceRoleNotFound       = 3010 // 协作空间角色不存在或无权操作
	CodeMenuSystemProtected         = 3011 // 系统菜单不可删除
	CodeInvalidParent               = 3012 // 无效的上级
	CodeConflict                    = 3013 // 业务冲突（通用）
	CodeUserExists                  = 3014 // 用户名已存在
	CodeEmailExists                 = 3015 // 邮箱已存在
	CodeSystemRoleProtected         = 3016 // 系统角色不可删除
	CodeProductNotFound             = 3017 // 商品不存在
	CodeGlobalRolePermReadOnly      = 3018 // 全局角色权限不可在此修改
	CodeWorkspaceMemberExists       = 3019 // 该用户已在协作空间中
	CodeWorkspaceMemberNotFound     = 3020 // 成员不在协作空间中
	CodeNoCurrentWorkspace          = 3021 // 暂无协作空间

	// ── 5xxxx 服务端 ───────────────────────────────────────────────────────
	CodeInternal = 5001 // 内部错误（通用）
	CodeDatabase = 5002 // 数据库错误
	CodeExternal = 5003 // 外部服务错误
)
