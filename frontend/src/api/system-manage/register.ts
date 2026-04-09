import { v5Client, unwrap } from './_shared'

export async function fetchListRegisterEntries() {
  return unwrap(v5Client.GET('/system/register-entries', {} as any))
}

export async function fetchCreateRegisterEntry(body: any) {
  return unwrap(v5Client.POST('/system/register-entries', { body }))
}

export async function fetchUpdateRegisterEntry(id: string, body: any) {
  return unwrap(
    v5Client.PUT('/system/register-entries/{id}', { params: { path: { id } }, body })
  )
}

export async function fetchDeleteRegisterEntry(id: string) {
  const { error } = await v5Client.DELETE('/system/register-entries/{id}', {
    params: { path: { id } }
  })
  if (error) throw new Error('delete failed')
}

export async function fetchListRegisterPolicies() {
  return unwrap(v5Client.GET('/system/register-policies', {} as any))
}

export async function fetchCreateRegisterPolicy(body: any) {
  return unwrap(v5Client.POST('/system/register-policies', { body }))
}

export async function fetchUpdateRegisterPolicy(code: string, body: any) {
  return unwrap(
    v5Client.PUT('/system/register-policies/{code}', { params: { path: { code } }, body })
  )
}

export async function fetchDeleteRegisterPolicy(code: string) {
  const { error } = await v5Client.DELETE('/system/register-policies/{code}', {
    params: { path: { code } }
  })
  if (error) throw new Error('delete failed')
}

export async function fetchListRegisterLogs(params: {
  source?: string
  entry_code?: string
  policy_code?: string
  page?: number
  page_size?: number
}) {
  return unwrap(
    v5Client.GET('/system/users/register-logs', { params: { query: params as any } })
  )
}
