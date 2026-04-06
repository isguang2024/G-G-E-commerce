// 兼容层：tenant 已预留给未来多租户系统。
// 当前协作主语义统一走 collaboration workspace。
export {
  hasPlatformAccessByUserInfo,
  useCollaborationWorkspaceStore,
  useTenantStore,
  type AppContextMode
} from './collaboration-workspace'
