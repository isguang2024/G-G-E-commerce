export interface MetricCard {
  id: string
  label: string
  value: string
  hint?: string
  tone?: 'brand' | 'success' | 'warning' | 'danger' | 'neutral'
}

export interface DashboardSummary {
  currentUserName: string
  currentSpaceLabel: string
  visibleMenuCount: number
  managedPageCount: number
  unreadInboxCount: number
  fastEntryCount: number
  quickLinkCount: number
}

export interface UserCenterProfile {
  id: string
  userName: string
  displayName: string
  email: string
  phone: string
  avatarUrl: string
  status: string
  badges: string[]
  currentTenantId?: string | null
}

export interface PermissionGroupSummary {
  id: string
  code: string
  name: string
  nameEn?: string
  groupType: string
  description: string
  status: string
  sortOrder: number
  isBuiltin: boolean
}

export interface PermissionActionRecord {
  id: string
  permissionKey: string
  name: string
  description: string
  contextType: string
  featureKind: string
  moduleCode: string
  moduleGroupId: string
  featureGroupId: string
  moduleGroup?: PermissionGroupSummary
  featureGroup?: PermissionGroupSummary
  apiCount: number
  pageCount: number
  packageCount: number
  consumerTypes: string[]
  usagePattern: string
  usageNote: string
  duplicatePattern: string
  duplicateGroup: string
  duplicateKeys: string[]
  duplicateNote: string
  status: string
  sortOrder: number
  isBuiltin: boolean
  createdAt: string
  updatedAt: string
}

export interface PermissionActionAuditSummary {
  totalCount: number
  unusedCount: number
  apiOnlyCount: number
  pageOnlyCount: number
  packageOnlyCount: number
  multiConsumerCount: number
  crossContextMirrorCount: number
  suspectedDuplicateCount: number
}

export interface PermissionActionEndpointBinding {
  endpointCode: string
  method: string
  path: string
  summary: string
  authMode: string
}

export interface PermissionActionConsumerDetail {
  type: string
  id: string
  label: string
  description: string
}

export interface ApiEndpointCategory {
  id: string
  code: string
  name: string
  nameEn: string
  description: string
  sortOrder: number
  status: string
}

export interface ApiEndpointRecord {
  id: string
  code: string
  method: string
  path: string
  spec: string
  featureKind: string
  handler: string
  summary: string
  permissionKey: string
  permissionKeys: string[]
  permissionContexts: string[]
  permissionBindingMode: string
  sharedAcrossContexts: boolean
  permissionNote: string
  authMode: string
  categoryId: string
  category?: ApiEndpointCategory
  contextScope: string
  source: string
  dataPermissionCode: string
  dataPermissionName: string
  runtimeExists: boolean
  stale: boolean
  staleReason: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface ApiEndpointOverview {
  totalCount: number
  uncategorizedCount: number
  staleCount: number
  noPermissionCount: number
  sharedPermissionCount: number
  crossContextSharedCount: number
}

export interface ApiUnregisteredRouteRecord {
  method: string
  path: string
  handler: string
  summary: string
  featureKind: string
  suggestedPermissionKey: string
  categoryCode: string
}

export interface ApiScanConfig {
  enabled: boolean
  frequencyMinutes: number
  defaultCategoryId: string
  defaultPermissionKey: string
  markAsNoPermission: boolean
}

export interface FastEnterItem {
  id: string
  name: string
  description: string
  icon: string
  iconColor: string
  enabled: boolean
  order: number
  routeName: string
  link?: string
}

export interface FastEnterConfig {
  applications: FastEnterItem[]
  quickLinks: FastEnterItem[]
  minWidth: number
}

export interface PageDefinition {
  id: string
  pageKey: string
  name: string
  routeName: string
  routePath: string
  component: string
  pageType: string
  source: string
  moduleKey: string
  sortOrder: number
  parentMenuId: string
  parentMenuName: string
  parentPageKey: string
  parentPageName: string
  displayGroupKey: string
  displayGroupName: string
  activeMenuPath: string
  breadcrumbMode: string
  accessMode: string
  permissionKey: string
  inheritPermission: boolean
  keepAlive: boolean
  isFullPage: boolean
  isIframe: boolean
  isHideTab: boolean
  link: string
  spaceKey: string
  spaceKeys: string[]
  spaceScope?: string
  status: string
  meta: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface PageGroupOption {
  id: string
  name: string
  title: string
  path: string
  children: PageGroupOption[]
}

export interface PageBreadcrumbPreviewItem {
  type: string
  title: string
  path: string
  pageKey: string
}

export interface PageUnregisteredRecord {
  filePath: string
  component: string
  pageKey: string
  name: string
  routeName: string
  routePath: string
  pageType: string
  moduleKey: string
  parentMenuId: string
  parentMenuName: string
  activeMenuPath: string
  spaceKey: string
  spaceType: string
  hostKey: string
}

export interface PageSyncResult {
  createdCount: number
  skippedCount: number
  createdKeys: string[]
}

export interface AccessTraceFilter {
  userId?: string
  tenantId?: string
  pageKey?: string
  pageKeys?: string[]
  routePath?: string
  spaceKey?: string
}

export interface AccessTraceRecord {
  label: string
  value: string
}

export interface AccessTraceResult {
  summary: AccessTraceRecord[]
  traceEntries: AccessTraceRecord[]
  recordEntries: AccessTraceRecord[]
  raw: Record<string, unknown>
}

export interface FeaturePackageImpactSummary {
  packageId: string
  roleCount: number
  teamCount: number
  userCount: number
  menuCount: number
  actionCount: number
  metrics: MetricCard[]
}

export interface MenuSpaceRecord {
  id: string
  spaceKey: string
  name: string
  description: string
  defaultHomePath: string
  isDefault: boolean
  status: string
  hostCount: number
  hosts: string[]
  menuCount: number
  pageCount: number
  accessMode: string
  allowedRoleCodes: string[]
  meta: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface MenuSpaceHostBindingRecord {
  id: string
  host: string
  spaceKey: string
  spaceName: string
  description: string
  isDefault: boolean
  status: string
  scheme: string
  routePrefix: string
  authMode: string
  loginHost: string
  callbackHost: string
  cookieScopeMode: string
  cookieDomain: string
  meta: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface RoleRecord {
  id: string
  code: string
  name: string
  description: string
  status: string
  sortOrder: number
  priority: number
  canEditPermission: boolean
  createdAt: string
}

export interface FeaturePackageRecord {
  id: string
  packageKey: string
  packageType: string
  name: string
  description: string
  contextType: string
  isBuiltin: boolean
  actionCount: number
  menuCount: number
  teamCount: number
  status: string
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface RelationSourceRecord {
  entityId: string
  packageIds: string[]
}

export interface SelectionRelation<T> {
  ids: string[]
  items: T[]
  availableIds?: string[]
  hiddenIds?: string[]
  disabledIds?: string[]
  expandedPackageIds?: string[]
  derivedSources?: RelationSourceRecord[]
  inherited?: boolean
}

export interface UserRoleSummary {
  id: string
  code: string
  name: string
  description?: string
}

export interface UserRecord {
  id: string
  userName: string
  nickName: string
  userEmail: string
  userPhone: string
  status: string
  avatar: string
  lastLoginTime: string
  lastLoginIP: string
  registerSource: string
  invitedBy: string
  invitedByName: string
  systemRemark: string
  createdAt: string
  updatedAt: string
  userRoles: string[]
  roleDetails: Array<Pick<UserRoleSummary, 'code' | 'name'>>
  roles?: UserRoleSummary[]
}

export interface UserPermissionDiagnosis {
  context: Record<string, unknown>
  diagnosis: Record<string, unknown> | null
  roles: Record<string, unknown>[]
  snapshot: Record<string, unknown>
  user: Record<string, unknown>
  teamMember?: Record<string, unknown>
  teamPackages?: Record<string, unknown>[]
}

export interface UserDiagnosisRoleSummary {
  title: string
  items: AccessTraceRecord[]
}

export interface UserDiagnosisSummary {
  userItems: AccessTraceRecord[]
  contextItems: AccessTraceRecord[]
  snapshotItems: AccessTraceRecord[]
  diagnosisItems: AccessTraceRecord[]
  roleSummaries: UserDiagnosisRoleSummary[]
  sourcePackageItems: AccessTraceRecord[]
  raw: UserPermissionDiagnosis
}

export interface TeamRecord {
  id: string
  name: string
  remark: string
  logoUrl: string
  plan: string
  maxMembers: number
  status: string
  ownerId: string
  createdAt: string
  updatedAt: string
  adminUsers: Array<{
    userId: string
    userName: string
    nickName: string
  }>
}

export interface TeamMemberRecord {
  id: string
  tenantId: string
  userId: string
  roleCode: string
  role: string
  status: string
  joinedAt: string
  userName: string
  nickName: string
  userEmail: string
  userPhone: string
  avatar: string
}

export interface TeamMemberDetail extends TeamMemberRecord {
  displayName: string
  contactItems: AccessTraceRecord[]
  roleItems: AccessTraceRecord[]
}

export interface TeamBoundaryOriginSummary {
  derivedIds: string[]
  blockedIds: string[]
  derivedSources: RelationSourceRecord[]
}

export interface RoleSavePayload {
  roleName: string
  roleCode: string
  description: string
  sortOrder: number
  priority: number
  status: string
}

export interface UserSavePayload {
  userName: string
  nickName: string
  userEmail: string
  userPhone: string
  password?: string
  status: string
}

export interface FeaturePackageSavePayload {
  packageKey: string
  packageType: string
  name: string
  description: string
  contextType: string
  sortOrder: number
  status: string
}

export interface PermissionActionSavePayload {
  permissionKey: string
  name: string
  description: string
  moduleGroupId?: string
  featureGroupId?: string
  contextType: string
  featureKind: string
  sortOrder: number
  status: string
}

export interface ApiEndpointSavePayload {
  code: string
  method: string
  path: string
  summary: string
  authMode: string
  permissionKey: string
  categoryId?: string
  contextScope: string
  source: string
  status: string
}

export interface PageSavePayload {
  pageKey: string
  name: string
  routeName: string
  routePath: string
  component: string
  pageType: string
  moduleKey: string
  sortOrder: number
  parentMenuId?: string
  parentPageKey?: string
  displayGroupKey?: string
  activeMenuPath?: string
  breadcrumbMode: string
  accessMode: string
  permissionKey?: string
  inheritPermission: boolean
  keepAlive: boolean
  isFullPage: boolean
  isHideTab: boolean
  link?: string
  spaceKey?: string
  meta?: Record<string, unknown>
}

export interface MenuSpaceSavePayload {
  spaceKey: string
  name: string
  description: string
  defaultHomePath: string
  isDefault: boolean
  status: string
  accessMode: string
  allowedRoleCodes: string[]
}

export interface MenuSpaceHostBindingSavePayload {
  host: string
  spaceKey: string
  description: string
  scheme: string
  routePrefix: string
  authMode: string
  loginHost?: string
  callbackHost?: string
  cookieScopeMode: string
  cookieDomain?: string
}

export interface TeamSavePayload {
  name: string
  remark: string
  plan: string
  maxMembers: number
  status: string
}
