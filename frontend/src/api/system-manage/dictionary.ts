import { v5Client, unwrap, type V5Query, type V5RequestBody } from '@/domains/governance/api/_shared'
import type { components } from '@/api/v5/schema'

// ─── Types ───────────────────────────────────────────────────────────────────

export type DictTypeSummary = components['schemas']['DictTypeSummary']
export type DictTypeDetail = components['schemas']['DictTypeDetail']
export type DictItemSummary = components['schemas']['DictItemSummary']
export type DictTypeSaveRequest = components['schemas']['DictTypeSaveRequest']
export type DictItemSaveRequest = components['schemas']['DictItemSaveRequest']
export type DictsByCodesResponse = components['schemas']['DictsByCodesResponse']

// ─── Dict Type CRUD ──────────────────────────────────────────────────────────

export async function fetchDictTypeList(params?: {
  current?: number
  size?: number
  keyword?: string
  status?: string
}) {
  const query: V5Query<'/dictionaries', 'get'> = params ?? {}
  return unwrap(v5Client.GET('/dictionaries', { params: { query } }))
}

export async function fetchCreateDictType(body: V5RequestBody<'/dictionaries', 'post'>) {
  return unwrap(v5Client.POST('/dictionaries', { body }))
}

export async function fetchGetDictType(id: string) {
  return unwrap(
    v5Client.GET('/dictionaries/{id}', { params: { path: { id } } })
  )
}

export async function fetchUpdateDictType(
  id: string,
  body: V5RequestBody<'/dictionaries/{id}', 'put'>
) {
  return unwrap(
    v5Client.PUT('/dictionaries/{id}', { params: { path: { id } }, body })
  )
}

export async function fetchDeleteDictType(id: string) {
  const { error } = await v5Client.DELETE('/dictionaries/{id}', {
    params: { path: { id } }
  })
  if (error) throw new Error('delete failed')
}

// ─── Dict Item ───────────────────────────────────────────────────────────────

export async function fetchDictItems(dictTypeId: string): Promise<DictItemSummary[]> {
  return unwrap(
    v5Client.GET('/dictionaries/{id}/items', { params: { path: { id: dictTypeId } } })
  )
}

export async function fetchSaveDictItems(
  dictTypeId: string,
  body: V5RequestBody<'/dictionaries/{id}/items', 'put'>
) {
  return unwrap(
    v5Client.PUT('/dictionaries/{id}/items', {
      params: { path: { id: dictTypeId } },
      body
    })
  )
}

export async function fetchCreateDictItem(
  dictTypeId: string,
  body: V5RequestBody<'/dictionaries/{id}/items', 'post'>
): Promise<DictItemSummary> {
  return unwrap(
    v5Client.POST('/dictionaries/{id}/items', {
      params: { path: { id: dictTypeId } },
      body
    })
  )
}

export async function fetchUpdateDictItem(
  dictTypeId: string,
  itemId: string,
  body: V5RequestBody<'/dictionaries/{id}/items/{itemId}', 'put'>
): Promise<DictItemSummary> {
  return unwrap(
    v5Client.PUT('/dictionaries/{id}/items/{itemId}', {
      params: { path: { id: dictTypeId, itemId } },
      body
    })
  )
}

export async function fetchDeleteDictItem(dictTypeId: string, itemId: string): Promise<void> {
  await unwrap(
    v5Client.DELETE('/dictionaries/{id}/items/{itemId}', {
      params: { path: { id: dictTypeId, itemId } }
    })
  )
}

// ─── Consumer (batch query by codes) ─────────────────────────────────────────

export async function fetchDictsByCodes(codes: string[]) {
  const query: V5Query<'/dictionaries/by-codes', 'get'> = {
    codes: codes.join(',')
  }
  return unwrap(
    v5Client.GET('/dictionaries/by-codes', { params: { query } })
  )
}
