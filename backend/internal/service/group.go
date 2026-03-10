package service

// GroupService 分组服务接口
type GroupService interface {
	// TODO: 定义分组服务方法
	// List(tenantID string) ([]*model.Group, error)
	// Create(tenantID string, req *dto.CreateGroupRequest) (*model.Group, error)
	// Update(tenantID, id string, req *dto.UpdateGroupRequest) error
	// Delete(tenantID, id string) error
}

// groupService 分组服务实现
type groupService struct {
	// TODO: 注入 GroupRepository
}

// NewGroupService 创建分组服务
func NewGroupService() GroupService {
	return &groupService{}
}
