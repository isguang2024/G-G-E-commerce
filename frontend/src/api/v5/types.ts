import type { paths } from './schema'

export type V5Method = 'get' | 'post' | 'put' | 'delete' | 'patch'
export type V5Path = keyof paths

type V5Operation<Path extends V5Path, Method extends V5Method> = NonNullable<paths[Path][Method]>
type V5Parameters<Path extends V5Path, Method extends V5Method> =
  V5Operation<Path, Method> extends { parameters: infer Parameters } ? Parameters : never

export type V5Query<Path extends V5Path, Method extends V5Method> =
  V5Parameters<Path, Method> extends { query?: infer Query } ? NonNullable<Query> : never

export type V5PathParams<Path extends V5Path, Method extends V5Method> =
  V5Parameters<Path, Method> extends { path?: infer Params } ? NonNullable<Params> : never

export type V5RequestBody<Path extends V5Path, Method extends V5Method> =
  V5Operation<Path, Method> extends {
    requestBody?: {
      content: {
        'application/json': infer Body
      }
    }
  }
    ? Body
    : never
