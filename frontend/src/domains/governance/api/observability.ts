import { v5Client, unwrap, type V5Query } from './_shared'

export type AuditLogListQuery = V5Query<'/observability/audit-logs', 'get'>
export type TelemetryLogListQuery = V5Query<'/observability/telemetry-logs', 'get'>
export type AuditLogStatsQuery = V5Query<'/observability/audit-logs/stats', 'get'>

export async function fetchListAuditLogs(query: AuditLogListQuery) {
  return unwrap(v5Client.GET('/observability/audit-logs', { params: { query } }))
}

export async function fetchGetAuditLog(id: number) {
  return unwrap(v5Client.GET('/observability/audit-logs/{id}', { params: { path: { id } } }))
}

export async function fetchListTelemetryLogs(query: TelemetryLogListQuery) {
  return unwrap(v5Client.GET('/observability/telemetry-logs', { params: { query } }))
}

export async function fetchGetTelemetryLog(id: number) {
  return unwrap(v5Client.GET('/observability/telemetry-logs/{id}', { params: { path: { id } } }))
}

// fetchAuditLogStats: 按 group_by=action|outcome|hour 做聚合统计。
// 供 dashboard widget 和运维仪表盘使用，权限与 list 一致（audit.read）。
export async function fetchAuditLogStats(query: AuditLogStatsQuery) {
  return unwrap(v5Client.GET('/observability/audit-logs/stats', { params: { query } }))
}

// fetchObservabilityTrace: 按 request_id 拉取一次请求涉及的 audit_logs + telemetry_logs
// 权限走 audit.read（后端已约定），前端不做二次授权判断。
export async function fetchObservabilityTrace(requestID: string) {
  return unwrap(
    v5Client.GET('/observability/trace/{request_id}', {
      params: { path: { request_id: requestID } }
    })
  )
}
