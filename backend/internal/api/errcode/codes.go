package errcode

// 业务错误码（与 HTTP 状态码分离，统一在 body.code + body.message 中返回）
// 规则：0=成功；1xxxx=参数/请求；2xxxx=认证/授权；3xxxx=业务/资源；5xxxx=服务端
const (
	// --- 1xxxx 参数/请求 ---
	ErrParamInvalid    = 1001 // 参数错误
	ErrParamMissing    = 1002 // 参数缺失
	ErrParamFormat     = 1003 // 参数格式错误
	ErrInvalidID       = 1004 // 无效的 ID（如菜单ID、用户ID、角色ID等，可配合 message 说明）

	// --- 2xxxx 认证/授权 ---
	ErrUnauthorized    = 2001 // 未登录或 token 无效
	ErrTokenExpired    = 2002 // token 已过期
	ErrForbidden       = 2003 // 无权限
	ErrAPIKeyMissing   = 2004 // 缺少 API Key
	ErrTokenBadFormat  = 2005 // Token 格式错误

	// --- 3xxxx 业务/资源 ---
	ErrNotFound        = 3001 // 资源不存在（通用）
	ErrUserNotFound    = 3002 // 用户不存在
	ErrTenantNotFound  = 3003 // 团队不存在
	ErrMenuNotFound    = 3004 // 菜单不存在
	ErrRoleNotFound    = 3005 // 角色不存在
	ErrNoManagedTeam   = 3006 // 您暂无管理的团队
	ErrRoleCodeExists  = 3007 // 角色编码已存在
	ErrMemberExists    = 3008 // 该用户已在团队中
	ErrMemberNotFound  = 3009 // 成员不在团队中
	ErrTeamRoleNotFound = 3010 // 团队角色不存在或无权操作
	ErrMenuSystemProtected = 3011 // 系统默认菜单不可删除
	ErrInvalidParent   = 3012 // 无效的上级（如不能将上级设为自己或子级）
	ErrConflict        = 3013 // 业务冲突（通用，如重复创建）
	ErrUsernameExists  = 3014 // 用户名已存在
	ErrSystemRoleProtected = 3016 // 系统角色不可删除
	ErrProductNotFound = 3017 // 商品不存在
	ErrGlobalRolePermissionReadOnly = 3018 // 全局角色权限不可在此修改

	// --- 5xxxx 服务端 ---
	ErrInternal        = 5001 // 内部错误（通用）
	ErrDatabase        = 5002 // 数据库错误
	ErrExternal        = 5003 // 外部服务错误
)

// defaultMessages 错误码默认说明（可被 Handler 层覆盖为更具体文案）
var defaultMessages = map[int]string{
	ErrParamInvalid:       "参数错误",
	ErrParamMissing:       "参数缺失",
	ErrParamFormat:        "参数格式错误",
	ErrInvalidID:         "无效的 ID",
	ErrUnauthorized:      "未登录或 token 无效",
	ErrTokenExpired:      "token 已过期",
	ErrForbidden:         "无权限",
	ErrAPIKeyMissing:     "缺少 API Key",
	ErrTokenBadFormat:    "Token 格式错误",
	ErrNotFound:          "资源不存在",
	ErrUserNotFound:      "用户不存在",
	ErrTenantNotFound:    "团队不存在",
	ErrMenuNotFound:      "菜单不存在",
	ErrRoleNotFound:      "角色不存在",
	ErrNoManagedTeam:     "您暂无管理的团队",
	ErrRoleCodeExists:    "角色编码已存在",
	ErrMemberExists:      "该用户已在团队中",
	ErrMemberNotFound:    "成员不在团队中",
	ErrTeamRoleNotFound:  "角色不存在或无权操作",
	ErrMenuSystemProtected: "系统默认菜单不可删除",
	ErrInvalidParent:     "无效的上级",
	ErrConflict:          "业务冲突",
	ErrUsernameExists:    "用户名已存在",
	ErrSystemRoleProtected: "系统角色不可删除",
	ErrProductNotFound:   "商品不存在",
	ErrGlobalRolePermissionReadOnly: "全局角色权限不可在此修改",
	ErrInternal:          "服务器内部错误，请稍后重试",
	ErrDatabase:          "数据库错误",
	ErrExternal:          "外部服务错误",
}

// defaultHTTPStatus 错误码建议的 HTTP 状态码（用于 Handler 设置 c.JSON(status, body)）
var defaultHTTPStatus = map[int]int{
	ErrParamInvalid:   400,
	ErrParamMissing:   400,
	ErrParamFormat:    400,
	ErrInvalidID:     400,
	ErrUnauthorized:  401,
	ErrTokenExpired:  401,
	ErrAPIKeyMissing: 401,
	ErrTokenBadFormat: 401,
	ErrForbidden:     403,
	ErrNotFound:      404,
	ErrUserNotFound:  404,
	ErrTenantNotFound: 404,
	ErrMenuNotFound:  404,
	ErrRoleNotFound:  404,
	ErrNoManagedTeam: 404,
	ErrRoleCodeExists: 409,
	ErrMemberExists:  409,
	ErrMemberNotFound: 404,
	ErrTeamRoleNotFound: 404,
	ErrMenuSystemProtected: 403,
	ErrInvalidParent: 400,
	ErrConflict:      409,
	ErrUsernameExists: 409,
	ErrSystemRoleProtected: 403,
	ErrProductNotFound: 404,
	ErrGlobalRolePermissionReadOnly: 403,
	ErrInternal:      500,
	ErrDatabase:      500,
	ErrExternal:      500,
}

// Message 返回错误码对应的默认说明
func Message(code int) string {
	if msg, ok := defaultMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// HTTPStatus 返回错误码对应的建议 HTTP 状态码
func HTTPStatus(code int) int {
	if status, ok := defaultHTTPStatus[code]; ok {
		return status
	}
	return 500
}
