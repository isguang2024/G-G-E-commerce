import { v5Client, unwrap, type V5Query, type V5RequestBody } from './_shared'

export async function fetchListRegisterEntries() {
  return unwrap(v5Client.GET('/system/register-entries'))
}

export async function fetchCreateRegisterEntry(
  body: V5RequestBody<'/system/register-entries', 'post'>
) {
  return unwrap(v5Client.POST('/system/register-entries', { body }))
}

export async function fetchUpdateRegisterEntry(
  id: string,
  body: V5RequestBody<'/system/register-entries/{id}', 'put'>
) {
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
  return unwrap(v5Client.GET('/system/register-policies'))
}

export async function fetchCreateRegisterPolicy(
  body: V5RequestBody<'/system/register-policies', 'post'>
) {
  return unwrap(v5Client.POST('/system/register-policies', { body }))
}

export async function fetchUpdateRegisterPolicy(
  code: string,
  body: V5RequestBody<'/system/register-policies/{code}', 'put'>
) {
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
  const query: V5Query<'/system/users/register-logs', 'get'> = params
  return unwrap(
    v5Client.GET('/system/users/register-logs', { params: { query } })
  )
}

export async function fetchListLoginPageTemplates() {
  return unwrap(v5Client.GET('/system/login-page-templates'))
}

export async function fetchCreateLoginPageTemplate(
  body: V5RequestBody<'/system/login-page-templates', 'post'>
) {
  return unwrap(v5Client.POST('/system/login-page-templates', { body }))
}

export async function fetchUpdateLoginPageTemplate(
  templateKey: string,
  body: V5RequestBody<'/system/login-page-templates/{templateKey}', 'put'>
) {
  return unwrap(
    v5Client.PUT('/system/login-page-templates/{templateKey}', {
      params: { path: { templateKey } },
      body
    })
  )
}

export async function fetchDeleteLoginPageTemplate(templateKey: string) {
  const { error } = await v5Client.DELETE('/system/login-page-templates/{templateKey}', {
    params: { path: { templateKey } }
  })
  if (error) throw new Error('delete failed')
}
