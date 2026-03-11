# G-G-E-commerce 项目索引

## 1. 项目概览 (Project Overview)
**G-G-E-commerce** 是一个现代化的前后端分离电子商务平台。后端使用 Go 语言开发，前端基于 Vue 3 框架构建。系统集成了高性能缓存 (Redis)、全文搜索 (Elasticsearch) 和对象存储 (MinIO) 等组件。

---

## 2. 技术栈 (Technology Stack)

### **后端 (Backend)**
- **核心语言:** Go (v1.22+)
- **Web 框架:** [Gin](https://github.com/gin-gonic/gin)
- **数据库:** PostgreSQL
- **ORM 框架:** [GORM](https://gorm.io/)
- **缓存:** Redis (go-redis/v9)
- **搜索引擎:** Elasticsearch (v8.12.0)
- **对象存储:** MinIO
- **配置管理:** Viper
- **身份认证:** JWT (JSON Web Token)
- **日志库:** Uber Zap

### **前端 (Frontend)**
- **核心框架:** Vue 3 (Composition API)
- **构建工具:** Vite
- **状态管理:** Pinia
- **UI 组件库:** Element Plus
- **样式方案:** Tailwind CSS + Sass
- **HTTP 客户端:** Axios
- **数据可视化:** ECharts
- **多语言支持:** Vue I18n

---

## 3. 项目目录结构 (Directory Structure)

### **根目录 (Root)**
- `/backend/`: 后端 Go 项目源代码
- `/frontend/`: 前端 Vue 项目源代码
- `/docker-compose.yml`: 基础设施容器定义 (PostgreSQL, Redis, ES, MinIO)

### **后端目录详情 (`/backend/`)**
- `cmd/`: 应用程序入口
  - `server/`: 主 Web 服务器入口
  - `migrate/`: 数据库自动迁移工具
  - `init-admin/`: 初始管理员账号脚本
- `internal/`: 核心业务逻辑 (私有)
  - `api/`: Handler 接口处理层 (Controller)
  - `service/`: Service 业务逻辑层
  - `repository/`: Repository 数据访问层
  - `model/`: Domain Models 数据库实体定义
  - `config/`: 配置文件加载与解析
  - `pkg/`: 内部通用工具与公共库
- `configs/`: 配置文件目录 (`config.yaml`)
- `docs/`: API 文档 (Swagger)

### **前端目录详情 (`/frontend/`)**
- `src/`: 源代码
  - `api/`: 后端接口封装与 Axios 配置
  - `views/`: 路由页面组件
  - `components/`: 通用 UI 组件
  - `store/`: Pinia 全局状态管理
  - `router/`: Vue Router 路由配置
  - `locales/`: 多语言配置
  - `hooks/`: 自定义组合式 API (Hooks)
  - `assets/`: 静态资源 (图片、基础样式)

---

## 4. 重要文件与入口 (Key Entry Points)

- **后端主程序:** `backend/cmd/server/main.go`
- **前端主程序:** `frontend/src/main.ts`
- **后端配置:** `backend/configs/config.yaml`
- **前端环境配置:** `frontend/.env.development`
- **基础设施部署:** `docker-compose.yml`

---

## 5. 开发常用命令 (Development Commands)

### **后端 (Backend)**
- `make dev`: 启动开发服务器
- `make build`: 编译二进制程序
- `make migrate`: 运行数据库迁移

### **前端 (Frontend)**
- `npm install`: 安装依赖
- `npm run dev`: 启动前端开发服务器
- `npm run build`: 构建生产版本

---
*上次更新日期: 2026-03-10*
