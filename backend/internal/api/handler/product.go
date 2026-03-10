package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/service"
)

// ProductHandler 商品处理器
type ProductHandler struct {
	productService service.ProductService
}

// NewProductHandler 创建商品处理器
func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// List 商品列表
func (h *ProductHandler) List(c *gin.Context) {
	// 获取 tenant_id（从中间件设置）
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr.(string))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的租户ID")
		c.JSON(status, resp)
		return
	}

	// 绑定查询参数
	var req dto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
		c.JSON(status, resp)
		return
	}

	// 处理 tag_ids 参数
	if tagIDsStr := c.QueryArray("tag_ids"); len(tagIDsStr) > 0 {
		for _, idStr := range tagIDsStr {
			if id, err := uuid.Parse(idStr); err == nil {
				req.TagIDs = append(req.TagIDs, id)
			}
		}
	}

	// 调用服务层
	products, total, err := h.productService.List(tenantID, &req)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}

	// 计算总页数
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.SuccessResponseWithMeta(products, &dto.Meta{
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalPages: totalPages,
	}))
}

// Get 商品详情
func (h *ProductHandler) Get(c *gin.Context) {
	// 获取 tenant_id
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr.(string))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的租户ID")
		c.JSON(status, resp)
		return
	}

	// 解析 ID
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的商品ID")
		c.JSON(status, resp)
		return
	}

	// 调用服务层
	product, err := h.productService.Get(tenantID, id)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrProductNotFound)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(product))
}

// Create 创建商品
func (h *ProductHandler) Create(c *gin.Context) {
	// 获取 tenant_id 和 user_id
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr.(string))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的租户ID")
		c.JSON(status, resp)
		return
	}

	userIDStr, _ := c.Get("user_id")
	var userID *uuid.UUID
	if userIDStr != nil {
		if id, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &id
		}
	}

	// 绑定请求体
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
		c.JSON(status, resp)
		return
	}

	// 调用服务层
	product, err := h.productService.Create(tenantID, userID, &req)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(product))
}

// Update 更新商品
func (h *ProductHandler) Update(c *gin.Context) {
	// 获取 tenant_id 和 user_id
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr.(string))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的租户ID")
		c.JSON(status, resp)
		return
	}

	userIDStr, _ := c.Get("user_id")
	var userID *uuid.UUID
	if userIDStr != nil {
		if id, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &id
		}
	}

	// 解析 ID
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的商品ID")
		c.JSON(status, resp)
		return
	}

	// 绑定请求体
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
		c.JSON(status, resp)
		return
	}

	// 调用服务层
	if err := h.productService.Update(tenantID, id, userID, &req); err != nil {
		if err.Error() == "product not found" {
			status, resp := errcode.Response(errcode.ErrProductNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// Delete 删除商品
func (h *ProductHandler) Delete(c *gin.Context) {
	// 获取 tenant_id
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr.(string))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的租户ID")
		c.JSON(status, resp)
		return
	}

	// 解析 ID
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的商品ID")
		c.JSON(status, resp)
		return
	}

	// 调用服务层
	if err := h.productService.Delete(tenantID, id); err != nil {
		if err.Error() == "product not found" {
			status, resp := errcode.Response(errcode.ErrProductNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
