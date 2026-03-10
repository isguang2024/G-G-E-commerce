package service

// CategoryService 分类服务接口
type CategoryService interface {
	// TODO: 定义分类服务方法
	// GetTree(tenantID string) ([]*model.Category, error)
	// List(tenantID string) ([]*model.Category, error)
	// Get(tenantID, id string) (*model.Category, error)
	// Create(tenantID string, req *dto.CreateCategoryRequest) (*model.Category, error)
	// Update(tenantID, id string, req *dto.UpdateCategoryRequest) error
	// Delete(tenantID, id string) error
}

// categoryService 分类服务实现
type categoryService struct {
	// TODO: 注入 CategoryRepository
}

// NewCategoryService 创建分类服务
func NewCategoryService() CategoryService {
	return &categoryService{}
}
