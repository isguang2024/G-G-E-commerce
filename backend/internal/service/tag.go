package service

// TagService 标签服务接口
type TagService interface {
	// TODO: 定义标签服务方法
	// List(tenantID string) ([]*model.Tag, error)
	// Create(tenantID string, req *dto.CreateTagRequest) (*model.Tag, error)
	// Update(tenantID, id string, req *dto.UpdateTagRequest) error
	// Delete(tenantID, id string) error
}

// tagService 标签服务实现
type tagService struct {
	// TODO: 注入 TagRepository
}

// NewTagService 创建标签服务
func NewTagService() TagService {
	return &tagService{}
}
