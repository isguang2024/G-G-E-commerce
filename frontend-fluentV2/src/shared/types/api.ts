export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export interface ApiErrorPayload {
  code?: number | string
  message?: string
  data?: unknown
}

export interface ApiErrorShape {
  message: string
  status?: number
  code?: string
  businessCode?: number | string
  details?: unknown
}
