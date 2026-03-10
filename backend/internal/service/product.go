package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/repository"
)

// ProductService 商品服务接口
type ProductService interface {
	List(tenantID uuid.UUID, req *dto.ProductListRequest) ([]*model.Product, int64, error)
	Get(tenantID, id uuid.UUID) (*model.Product, error)
	Create(tenantID uuid.UUID, userID *uuid.UUID, req *dto.CreateProductRequest) (*model.Product, error)
	Update(tenantID, id uuid.UUID, userID *uuid.UUID, req *dto.UpdateProductRequest) error
	Delete(tenantID, id uuid.UUID) error
}

// productService 商品服务实现
type productService struct {
	productRepo repository.ProductRepository
}

// NewProductService 创建商品服务
func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

// List 商品列表
func (s *productService) List(tenantID uuid.UUID, req *dto.ProductListRequest) ([]*model.Product, int64, error) {
	req.Default()
	
	filters := make(map[string]interface{})
	if req.CategoryID != nil {
		filters["category_id"] = *req.CategoryID
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.Keyword != "" {
		filters["keyword"] = req.Keyword
	}
	if req.IsFeatured != nil {
		filters["is_featured"] = *req.IsFeatured
	}
	if req.SortBy != "" {
		filters["sort_by"] = req.SortBy
	}
	if req.Order != "" {
		filters["order"] = req.Order
	}

	return s.productRepo.List(tenantID, req.Page, req.PageSize, filters)
}

// Get 获取商品详情
func (s *productService) Get(tenantID, id uuid.UUID) (*model.Product, error) {
	return s.productRepo.GetByID(tenantID, id)
}

// Create 创建商品
func (s *productService) Create(tenantID uuid.UUID, userID *uuid.UUID, req *dto.CreateProductRequest) (*model.Product, error) {
	// 生成 Slug（如果未提供）
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// 检查 Slug 是否已存在
	_, err := s.productRepo.GetBySlug(tenantID, slug)
	if err == nil {
		return nil, fmt.Errorf("slug already exists: %s", slug)
	}

	product := &model.Product{
		TenantID:       tenantID,
		CreatedBy:      userID,
		Name:           req.Name,
		NameEn:         req.NameEn,
		Slug:           slug,
		Description:    req.Description,
		DescriptionEn: req.DescriptionEn,
		ShortDesc:      req.ShortDesc,
		CategoryID:     req.CategoryID,
		CoverImageID:   req.CoverImageID,
		Price:          req.Price,
		CostPrice:      req.CostPrice,
		SKU:            req.SKU,
		Barcode:        req.Barcode,
		Brand:          req.Brand,
		OriginCountry: req.OriginCountry,
		Weight:         req.Weight,
		Dimensions:     req.Dimensions,
		Attributes:     req.Attributes,
		SEOTitle:       req.SEOTitle,
		SEODescription: req.SEODescription,
		SEOKeywords:    req.SEOKeywords,
		Status:         getStatusOrDefault(req.Status),
		IsFeatured:    req.IsFeatured,
		SortOrder:      req.SortOrder,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// TODO: 处理标签和分组关联

	return product, nil
}

// Update 更新商品
func (s *productService) Update(tenantID, id uuid.UUID, userID *uuid.UUID, req *dto.UpdateProductRequest) error {
	product, err := s.productRepo.GetByID(tenantID, id)
	if err != nil {
		return errors.New("product not found")
	}

	// 更新字段
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.NameEn != nil {
		product.NameEn = *req.NameEn
	}
	if req.Slug != nil {
		// 检查 Slug 是否已被其他商品使用
		existing, err := s.productRepo.GetBySlug(tenantID, *req.Slug)
		if err == nil && existing.ID != id {
			return fmt.Errorf("slug already exists: %s", *req.Slug)
		}
		product.Slug = *req.Slug
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.DescriptionEn != nil {
		product.DescriptionEn = *req.DescriptionEn
	}
	if req.ShortDesc != nil {
		product.ShortDesc = *req.ShortDesc
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.CoverImageID != nil {
		product.CoverImageID = req.CoverImageID
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.CostPrice != nil {
		product.CostPrice = *req.CostPrice
	}
	if req.SKU != nil {
		product.SKU = *req.SKU
	}
	if req.Barcode != nil {
		product.Barcode = *req.Barcode
	}
	if req.Brand != nil {
		product.Brand = *req.Brand
	}
	if req.OriginCountry != nil {
		product.OriginCountry = *req.OriginCountry
	}
	if req.Weight != nil {
		product.Weight = *req.Weight
	}
	if req.Dimensions != nil {
		product.Dimensions = *req.Dimensions
	}
	if req.Attributes != nil {
		product.Attributes = *req.Attributes
	}
	if req.SEOTitle != nil {
		product.SEOTitle = *req.SEOTitle
	}
	if req.SEODescription != nil {
		product.SEODescription = *req.SEODescription
	}
	if req.SEOKeywords != nil {
		product.SEOKeywords = *req.SEOKeywords
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.IsFeatured != nil {
		product.IsFeatured = *req.IsFeatured
	}
	if req.SortOrder != nil {
		product.SortOrder = *req.SortOrder
	}

	product.UpdatedBy = userID

	if err := s.productRepo.Update(product); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	// TODO: 处理标签和分组关联更新

	return nil
}

// Delete 删除商品
func (s *productService) Delete(tenantID, id uuid.UUID) error {
	_, err := s.productRepo.GetByID(tenantID, id)
	if err != nil {
		return errors.New("product not found")
	}

	return s.productRepo.Delete(tenantID, id)
}

// 辅助函数
func generateSlug(name string) string {
	// TODO: 实现 Slug 生成逻辑（转换为小写、替换空格等）
	return name
}

func getStatusOrDefault(status string) string {
	if status == "" {
		return "draft"
	}
	return status
}
