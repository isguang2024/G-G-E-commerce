package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/service"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{userService: userService, logger: logger}
}

// List 用户列表（分页）
func (h *UserHandler) List(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	// 手动获取查询参数（如果绑定失败，作为备用）
	if req.UserPhone == "" {
		req.UserPhone = c.Query("userPhone")
	}
	if req.UserEmail == "" {
		req.UserEmail = c.Query("userEmail")
	}
	if req.UserName == "" {
		req.UserName = c.Query("userName")
	}
	if req.Status == "" {
		req.Status = c.Query("status")
	}
	if req.RoleID == "" {
		req.RoleID = c.Query("roleId")
	}
	// 调试日志：输出接收到的查询参数
	h.logger.Info("User list request",
		zap.String("userName", req.UserName),
		zap.String("userPhone", req.UserPhone),
		zap.String("userEmail", req.UserEmail),
		zap.String("status", req.Status),
		zap.String("roleId", req.RoleID),
		zap.Int("current", req.Current),
		zap.Int("size", req.Size))
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	list, total, err := h.userService.List(&req)
	if err != nil {
		h.logger.Error("User list failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, u := range list {
		roles := make([]gin.H, 0)
		for _, r := range u.Roles {
			roles = append(roles, gin.H{"id": r.ID.String(), "code": r.Code, "name": r.Name})
		}
		records = append(records, gin.H{
			"id":         u.ID.String(),
			"userName":   u.Username,
			"userEmail":  u.Email,
			"nickName":   u.Nickname,
			"userPhone":  u.Phone,
			"systemRemark": u.SystemRemark,
			"lastLoginTime": formatNullableTime(u.LastLoginAt),
			"lastLoginIP": u.LastLoginIP,
			"status":     u.Status,
			"avatar":     u.AvatarURL,
			"createTime": u.CreatedAt.Format("2006-01-02 15:04:05"),
			"updateTime": u.UpdatedAt.Format("2006-01-02 15:04:05"),
			"userRoles":  roleCodes(u.Roles),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

// Get 用户详情
func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	user, err := h.userService.Get(id)
	if err != nil {
		if err == service.ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}
	roles := make([]gin.H, 0)
	for _, r := range user.Roles {
		roles = append(roles, gin.H{"id": r.ID.String(), "code": r.Code, "name": r.Name})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"id":         user.ID.String(),
		"userName":   user.Username,
		"userEmail":  user.Email,
		"nickName":   user.Nickname,
		"userPhone":  user.Phone,
		"systemRemark": user.SystemRemark,
		"lastLoginTime": formatNullableTime(user.LastLoginAt),
		"lastLoginIP": user.LastLoginIP,
		"status":     user.Status,
		"avatar":     user.AvatarURL,
		"createTime": user.CreatedAt.Format("2006-01-02 15:04:05"),
		"updateTime": user.UpdatedAt.Format("2006-01-02 15:04:05"),
		"roles":      roles,
		"userRoles":  roleCodes(user.Roles),
	}))
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	user, err := h.userService.Create(&req)
	if err != nil {
		if err == service.ErrUserExists {
			status, resp := errcode.Response(errcode.ErrUsernameExists)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create user failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建用户失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": user.ID.String()}))
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.userService.Update(id, &req); err != nil {
		if err == service.ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update user failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新用户失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	if err := h.userService.Delete(id); err != nil {
		if err == service.ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete user failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除用户失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// AssignRoles 分配角色
func (h *UserHandler) AssignRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	var req dto.UserAssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.userService.AssignRoles(id, req.RoleIDs); err != nil {
		if err == service.ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Assign roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "分配角色失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func roleCodes(roles []model.Role) []string {
	codes := make([]string, 0, len(roles))
	for _, r := range roles {
		codes = append(codes, r.Code)
	}
	return codes
}

func formatNullableTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
