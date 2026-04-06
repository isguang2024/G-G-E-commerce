// 兼容层：tenant 已预留给未来多租户系统。
// 当前协作主语义统一走 collaboration workspace。
import { useCollaborationWorkspaceStore } from './collaboration-workspace'

export const useTenantStore = useCollaborationWorkspaceStore

export {
  hasPersonalWorkspaceAccessByUserInfo,
  hasPlatformAccessByUserInfo,
  useCollaborationWorkspaceStore,
  type AppContextMode
} from './collaboration-workspace'
