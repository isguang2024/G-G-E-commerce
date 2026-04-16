// upload-config 管理面 API 客户端封装。
//
// 与媒体上传 SDK（domains/upload）解耦：上传 SDK 走运行时 prepare/complete，
// 这里走配置中心 CRUD —— 只供 system.upload.config.manage 角色调用。

import { v5Client, unwrap } from '@/domains/governance/api/_shared'
import type { components } from '@/api/v5/schema'

// ── 类型别名（直接从 OpenAPI schema 生成的契约）────────────────────────────

export type StorageProviderSummary = components['schemas']['StorageProviderSummary']
export type StorageProviderListResponse = components['schemas']['StorageProviderListResponse']
export type StorageProviderSaveRequest = components['schemas']['StorageProviderSaveRequest']
export type StorageProviderTestResponse = components['schemas']['StorageProviderTestResponse']

export type StorageBucketSummary = components['schemas']['StorageBucketSummary']
export type StorageBucketListResponse = components['schemas']['StorageBucketListResponse']
export type StorageBucketSaveRequest = components['schemas']['StorageBucketSaveRequest']

export type UploadKeySummary = components['schemas']['UploadKeySummary']
export type UploadKeyDetail = components['schemas']['UploadKeyDetail']
export type UploadKeyListResponse = components['schemas']['UploadKeyListResponse']
export type UploadKeySaveRequest = components['schemas']['UploadKeySaveRequest']

export type UploadKeyRuleSummary = components['schemas']['UploadKeyRuleSummary']
export type UploadKeyRuleListResponse = components['schemas']['UploadKeyRuleListResponse']
export type UploadKeyRuleSaveRequest = components['schemas']['UploadKeyRuleSaveRequest']

// ── Provider ────────────────────────────────────────────────────────────────

export async function fetchListStorageProviders(): Promise<StorageProviderListResponse> {
  return unwrap(v5Client.GET('/storage/providers'))
}

export async function fetchGetStorageProvider(id: string): Promise<StorageProviderSummary> {
  return unwrap(v5Client.GET('/storage/providers/{id}', { params: { path: { id } } }))
}

export async function fetchCreateStorageProvider(
  body: StorageProviderSaveRequest
): Promise<StorageProviderSummary> {
  return unwrap(v5Client.POST('/storage/providers', { body }))
}

export async function fetchUpdateStorageProvider(
  id: string,
  body: StorageProviderSaveRequest
): Promise<StorageProviderSummary> {
  return unwrap(
    v5Client.PUT('/storage/providers/{id}', { params: { path: { id } }, body })
  )
}

export async function fetchDeleteStorageProvider(id: string): Promise<void> {
  const { error } = await v5Client.DELETE('/storage/providers/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

export async function fetchTestStorageProvider(
  id: string
): Promise<StorageProviderTestResponse> {
  return unwrap(v5Client.POST('/storage/providers/{id}/test', { params: { path: { id } } }))
}

// ── Bucket ──────────────────────────────────────────────────────────────────

export async function fetchListStorageBuckets(
  providerId?: string
): Promise<StorageBucketListResponse> {
  return unwrap(
    v5Client.GET('/storage/buckets', {
      params: { query: providerId ? { provider_id: providerId } : {} }
    })
  )
}

export async function fetchGetStorageBucket(id: string): Promise<StorageBucketSummary> {
  return unwrap(v5Client.GET('/storage/buckets/{id}', { params: { path: { id } } }))
}

export async function fetchCreateStorageBucket(
  body: StorageBucketSaveRequest
): Promise<StorageBucketSummary> {
  return unwrap(v5Client.POST('/storage/buckets', { body }))
}

export async function fetchUpdateStorageBucket(
  id: string,
  body: StorageBucketSaveRequest
): Promise<StorageBucketSummary> {
  return unwrap(
    v5Client.PUT('/storage/buckets/{id}', { params: { path: { id } }, body })
  )
}

export async function fetchDeleteStorageBucket(id: string): Promise<void> {
  const { error } = await v5Client.DELETE('/storage/buckets/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

// ── UploadKey ───────────────────────────────────────────────────────────────

export async function fetchListUploadKeys(bucketId?: string): Promise<UploadKeyListResponse> {
  return unwrap(
    v5Client.GET('/upload/keys', {
      params: { query: bucketId ? { bucket_id: bucketId } : {} }
    })
  )
}

export async function fetchGetUploadKey(id: string): Promise<UploadKeyDetail> {
  return unwrap(v5Client.GET('/upload/keys/{id}', { params: { path: { id } } }))
}

export async function fetchCreateUploadKey(
  body: UploadKeySaveRequest
): Promise<UploadKeySummary> {
  return unwrap(v5Client.POST('/upload/keys', { body }))
}

export async function fetchUpdateUploadKey(
  id: string,
  body: UploadKeySaveRequest
): Promise<UploadKeySummary> {
  return unwrap(v5Client.PUT('/upload/keys/{id}', { params: { path: { id } }, body }))
}

export async function fetchDeleteUploadKey(id: string): Promise<void> {
  const { error } = await v5Client.DELETE('/upload/keys/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

// ── Rule ────────────────────────────────────────────────────────────────────

export async function fetchListUploadKeyRules(
  uploadKeyId: string
): Promise<UploadKeyRuleListResponse> {
  return unwrap(
    v5Client.GET('/upload/keys/{id}/rules', { params: { path: { id: uploadKeyId } } })
  )
}

export async function fetchCreateUploadKeyRule(
  uploadKeyId: string,
  body: UploadKeyRuleSaveRequest
): Promise<UploadKeyRuleSummary> {
  return unwrap(
    v5Client.POST('/upload/keys/{id}/rules', {
      params: { path: { id: uploadKeyId } },
      body
    })
  )
}

export async function fetchUpdateUploadKeyRule(
  ruleId: string,
  body: UploadKeyRuleSaveRequest
): Promise<UploadKeyRuleSummary> {
  return unwrap(
    v5Client.PUT('/upload/rules/{ruleId}', { params: { path: { ruleId } }, body })
  )
}

export async function fetchDeleteUploadKeyRule(ruleId: string): Promise<void> {
  const { error } = await v5Client.DELETE('/upload/rules/{ruleId}', {
    params: { path: { ruleId } }
  })
  if (error) throw error
}
