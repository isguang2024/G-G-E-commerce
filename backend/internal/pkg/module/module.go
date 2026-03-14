package module

import "github.com/gin-gonic/gin"

type Module interface {
	Init() error
	RegisterRoutes(rg *gin.RouterGroup)
}

type Registry struct {
	modules []Module
}

var registry *Registry

func init() {
	registry = &Registry{
		modules: make([]Module, 0),
	}
}

func GetRegistry() *Registry {
	return registry
}

func (r *Registry) Register(m Module) {
	r.modules = append(r.modules, m)
}

func (r *Registry) GetModules() []Module {
	return r.modules
}
