import type { StorageProviderSaveRequest } from '@/domains/upload-config/api'

export type StorageDriver = StorageProviderSaveRequest['driver']
export type DriverExtraScope = 'provider' | 'bucket'
export type DriverExtraFieldType = 'string' | 'number' | 'boolean' | 'object'

export interface DriverExtraField {
  key: string
  label: string
  type: DriverExtraFieldType
  description?: string
  placeholder?: string
  defaultValue?: unknown
  multiline?: boolean
  rows?: number
  min?: number
  step?: number
}

export interface DriverExtraSection {
  key: string
  title: string
  description?: string
  defaultOpen?: boolean
  fields: DriverExtraField[]
}

export interface DriverExtraProfile {
  driver: StorageDriver
  providerGuide: string[]
  bucketGuide: string[]
  providerSections: DriverExtraSection[]
  bucketSections: DriverExtraSection[]
}

const registry: Record<StorageDriver, DriverExtraProfile> = {
  local: {
    driver: 'local',
    providerGuide: [
      '本地存储只需要基础连接信息，通常只配置名称、默认状态和访问根地址。',
      '若文件经由 Nginx 或静态资源站暴露，可在基础访问地址里填写统一公网域名。',
      '驱动当前没有预置高级参数，如需特殊能力，可在自定义扩展参数中补充。'
    ],
    bucketGuide: [
      '本地桶主要负责目录隔离和公网访问策略，基础路径建议按业务模块分层。',
      '公开访问关闭时，业务侧应通过后端签名或代理方式访问文件。',
      '驱动当前没有预置桶级高级参数，可直接使用自定义扩展参数。'
    ],
    providerSections: [],
    bucketSections: []
  },
  aliyun_oss: {
    driver: 'aliyun_oss',
    providerGuide: [
      '最小可用配置：Endpoint、地域、AccessKey、SecretKey。',
      '如果业务要走前端直传，优先补齐 CNAME / STS，并确认 Bucket 侧开启成功状态码与回调参数。',
      '网络兼容和超时重试集中在高级参数里，日常只需要动基础连接与 STS。'
    ],
    bucketGuide: [
      'Bucket 名称对应阿里云 OSS 实际 Bucket；基础路径用于统一加业务前缀。',
      '前端直传时推荐保留成功状态码，并按需补充下载头或 OSS 回调配置。',
      '若使用回调，callback / callback_var 通常填写 OSS 约定的 base64 JSON。'
    ],
    providerSections: [
      {
        key: 'compatibility',
        title: '域名与兼容性',
        description: '控制 CNAME、自定义网关和协议兼容选项。',
        defaultOpen: true,
        fields: [
          {
            key: 'use_cname',
            label: '启用 CNAME',
            type: 'boolean',
            description: '通过自定义域名访问 OSS，对接 CDN 时通常开启。',
            defaultValue: false
          },
          {
            key: 'use_path_style',
            label: '启用 Path-Style',
            type: 'boolean',
            description: '兼容部分代理网关或对象存储兼容层。',
            defaultValue: false
          }
        ]
      },
      {
        key: 'sts',
        title: 'STS 临时凭证',
        description: '前端直传或最小权限访问时建议启用。',
        defaultOpen: true,
        fields: [
          {
            key: 'sts_role_arn',
            label: 'STS Role ARN',
            type: 'string',
            placeholder: 'acs:ram::123456789012:role/upload-direct',
            description: '启用 STS 时必填。'
          },
          {
            key: 'sts_external_id',
            label: 'STS External ID',
            type: 'string',
            placeholder: '可选的 External ID',
            description: '跨账号或更严格信任关系时按需配置。'
          },
          {
            key: 'sts_session_name',
            label: 'STS Session Name',
            type: 'string',
            placeholder: 'gge-upload',
            description: 'AssumeRole 会话名。',
            defaultValue: 'gge-upload'
          },
          {
            key: 'sts_duration_seconds',
            label: 'STS 凭证时长(秒)',
            type: 'number',
            min: 900,
            step: 60,
            description: '默认 3600 秒，越短越安全。',
            defaultValue: 3600
          },
          {
            key: 'sts_endpoint',
            label: 'STS Endpoint',
            type: 'string',
            placeholder: 'sts.cn-hangzhou.aliyuncs.com',
            description: '仅在需要自定义 STS 接入点时填写。'
          },
          {
            key: 'sts_policy',
            label: 'STS Policy',
            type: 'object',
            multiline: true,
            rows: 5,
            placeholder: '{\n  "Statement": []\n}',
            description: 'AssumeRole 附加策略 JSON。'
          }
        ]
      },
      {
        key: 'advanced',
        title: '高级网络参数',
        description: '测试、代理、超时和重试参数，通常仅在复杂网络环境下调整。',
        fields: [
          {
            key: 'insecure_skip_verify',
            label: '跳过 TLS 校验',
            type: 'boolean',
            description: '仅用于内网测试或自签证书场景。',
            defaultValue: false
          },
          {
            key: 'disable_ssl',
            label: '禁用 HTTPS',
            type: 'boolean',
            description: '改走 HTTP 访问 OSS，仅限可信内网。',
            defaultValue: false
          },
          {
            key: 'connect_timeout_ms',
            label: '连接超时(ms)',
            type: 'number',
            min: 0,
            step: 100,
            description: 'SDK 建连超时时间。'
          },
          {
            key: 'read_write_timeout_ms',
            label: '读写超时(ms)',
            type: 'number',
            min: 0,
            step: 100,
            description: '上传下载的读写超时时间。'
          },
          {
            key: 'retry_max_attempts',
            label: '最大重试次数',
            type: 'number',
            min: 0,
            step: 1,
            description: 'SDK 请求失败后的重试上限。'
          }
        ]
      }
    ],
    bucketSections: [
      {
        key: 'direct-upload',
        title: '直传响应与对象头',
        description: '控制浏览器直传返回值和对象默认下载行为。',
        defaultOpen: true,
        fields: [
          {
            key: 'success_action_status',
            label: '直传成功状态码',
            type: 'string',
            placeholder: '200',
            description: '浏览器表单直传成功后的 HTTP 状态码。',
            defaultValue: '200'
          },
          {
            key: 'content_disposition',
            label: 'Content-Disposition',
            type: 'string',
            placeholder: 'inline; filename="demo.png"',
            description: '对象默认下载头，控制 inline / attachment。'
          }
        ]
      },
      {
        key: 'callback',
        title: 'OSS 回调',
        description: '上传完成后由 OSS 回调业务服务，常用于媒体入库或异步处理。',
        fields: [
          {
            key: 'callback',
            label: '上传回调配置',
            type: 'string',
            multiline: true,
            rows: 4,
            placeholder: 'base64 callback body',
            description: 'OSS 回调配置字符串，通常是 base64 JSON。'
          },
          {
            key: 'callback_var',
            label: '上传回调变量',
            type: 'string',
            multiline: true,
            rows: 4,
            placeholder: 'base64 callback vars',
            description: 'OSS 回调变量配置，通常是 base64 JSON。'
          }
        ]
      }
    ]
  }
}

export function getDriverExtraProfile(
  driver?: StorageDriver | '' | null
): DriverExtraProfile | undefined {
  if (!driver) return undefined
  return registry[driver]
}

export function getDriverGuide(
  driver: StorageDriver | '' | undefined | null,
  scope: DriverExtraScope
): string[] {
  const profile = getDriverExtraProfile(driver)
  if (!profile) return []
  return scope === 'provider' ? profile.providerGuide : profile.bucketGuide
}

export function getDriverExtraSections(
  driver: StorageDriver | '' | undefined | null,
  scope: DriverExtraScope
): DriverExtraSection[] {
  const profile = getDriverExtraProfile(driver)
  if (!profile) return []
  return scope === 'provider' ? profile.providerSections : profile.bucketSections
}

export function getDriverExtraDefaults(
  driver: StorageDriver | '' | undefined | null,
  scope: DriverExtraScope
): Record<string, unknown> {
  const sections = getDriverExtraSections(driver, scope)
  const defaults: Record<string, unknown> = {}
  for (const section of sections) {
    for (const field of section.fields) {
      if (field.defaultValue === undefined) continue
      defaults[field.key] = field.defaultValue
    }
  }
  return defaults
}
