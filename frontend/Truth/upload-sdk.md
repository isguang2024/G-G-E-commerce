# 前端上传 Key 配置与使用指南

本文档说明前端如何通过 UploadKey 调用上传接口，覆盖所有预置场景和常见集成方式。

---

## 1. 核心概念

### 上传链路

```
前端 submit(file, target)
  → 解析 target 为 { key, rule }
  → POST /media/prepare（取得上传方式：直传 / 中转）
  → 上传文件
  → POST /media/complete（直传时）
  → 返回 { url, id, ... }
```

### 四层配置继承

```
存储服务（Provider）  → 驱动与凭据
  └─ 存储桶（Bucket）  → 文件存放区域
       └─ 上传配置（UploadKey）  → 业务场景入口
            └─ 上传规则（Rule）   → 细粒度限制
```

后端按 UploadKey 的 `key` 字段查找配置，再结合 Rule 合并出：允许类型、大小上限、路径模板、文件名策略等。

### 当前预置 Key 与 Rule

| UploadKey（key） | 名称 | 大小上限 | 允许类型 | 预置 Rule |
| --- | --- | --- | --- | --- |
| `media.default` | 默认媒体上传 | 10 MB | image/jpeg, image/png, image/webp, image/gif | `image`（默认） |
| `user.avatar` | 用户头像 | 2 MB | image/jpeg, image/png, image/webp | `avatar`（默认） |
| `doc.attachment` | 文档附件 | 50 MB | application/pdf, application/msword 等 | `pdf`（默认）、`office` |
| `editor.inline` | 富文本编辑器图片 | 5 MB | image/jpeg, image/png, image/webp, image/gif | `editor-image`（默认） |

---

## 2. 两种调用方式

### 方式一：对象形式（推荐）

适用于所有 key，语义清晰，支持完整选项。

```ts
import { useUpload } from '@/domains/upload/use-upload'

const { submit, uploading, error } = useUpload()

// 只指定 key，使用该 key 的默认 Rule
const result = await submit(file, { key: 'user.avatar' })

// 指定 key + 明确的 rule
const result = await submit(file, { key: 'doc.attachment', rule: 'office' })

// 完整选项：key + rule + 上传模式 + 进度回调
const result = await submit(file, {
  key: 'media.default',
  rule: 'image',
  mode: 'auto',               // 'auto' | 'direct' | 'relay'
  onProgress(percent) {        // 0-100
    console.log(`上传进度: ${percent}%`)
  }
})
```

### 方式二：字符串简写

> **注意**：字符串按**第一个** `.` 拆分为 key 和 rule。
> 因此只适合 key 本身不含 `.` 的场景（如将来的单段 key `avatar`）。
> 当前预置 key 均为 `xxx.yyy` 格式，**请使用对象形式**。

```ts
// ⚠️ 'user.avatar' 会被拆为 key='user', rule='avatar'，
// 后端找不到 key='user'，会回退到系统默认配置，而不是 user.avatar 的配置。
// 所以不要这样写：
await submit(file, 'user.avatar')  // ❌ 不会匹配 user.avatar 的上传配置

// 正确写法：
await submit(file, { key: 'user.avatar' })  // ✅
```

### 不传 key

不指定任何参数时，后端自动回退到系统默认上传配置（即 `media.default`）。

```ts
// 以下写法等价，均使用默认配置
await submit(file)
await submit(file, {})
await submit(file, { key: '' })
```

---

## 3. 按场景集成示例

### 3.1 用户头像上传

```vue
<template>
  <ElUpload
    :show-file-list="false"
    :http-request="handleUpload"
    accept="image/jpeg,image/png,image/webp"
  >
    <ElAvatar :size="80" :src="avatarUrl" />
  </ElUpload>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useUpload } from '@/domains/upload/use-upload'

const { submit, uploading } = useUpload()
const avatarUrl = ref('')

async function handleUpload({ file }: { file: File }) {
  try {
    const result = await submit(file, { key: 'user.avatar' })
    avatarUrl.value = result.url
    ElMessage.success('头像上传成功')
  } catch {
    ElMessage.error('头像上传失败')
  }
}
</script>
```

**配置效果**：后端按 `user.avatar` 配置校验 ——
- 最大 2 MB，超出直接拒绝
- 只允许 jpeg / png / webp
- 文件名策略 UUID（避免重名）
- 存储子路径 `avatar/`

### 3.2 文档附件上传（带进度条）

```vue
<template>
  <div>
    <ElUpload :http-request="handleUpload" :show-file-list="false">
      <ElButton type="primary">上传文档</ElButton>
    </ElUpload>
    <ElProgress v-if="uploading" :percentage="percent" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useUpload } from '@/domains/upload/use-upload'

const { submit, uploading } = useUpload()
const percent = ref(0)

async function handleUpload({ file }: { file: File }) {
  percent.value = 0
  try {
    // 使用 pdf 规则（默认），最大 50 MB
    const result = await submit(file, {
      key: 'doc.attachment',
      onProgress(p) { percent.value = p }
    })
    ElMessage.success(`文件已上传: ${result.filename}`)
  } catch {
    ElMessage.error('上传失败')
  }
}
</script>
```

**切换为 office 规则**（限 30 MB、保留原文件名）：

```ts
await submit(file, { key: 'doc.attachment', rule: 'office' })
```

### 3.3 富文本编辑器图片

WangEditor 已内置集成，参见 `art-wang-editor/index.vue`。
如需显式指定 key（推荐）：

```ts
// 在 customUpload 中传入 key
async customUpload(file: File, insertFn) {
  const uploaded = await uploadMediaWithPrepare(file, {
    key: 'editor.inline'
  })
  insertFn(uploaded.url, file.name, uploaded.url)
}
```

**配置效果**：
- 最大 5 MB
- 只允许 jpeg / png / webp / gif
- 文件名策略 UUID
- 存储子路径 `img/`

### 3.4 通用媒体上传（默认配置）

```ts
// 不指定 key，回退到 media.default
const { submit } = useUpload()
const result = await submit(file)
```

**配置效果**：
- 最大 10 MB
- 允许常见图片格式
- 默认 Rule 为 `image`

---

## 4. 上传模式

| 模式 | 说明 | 适用场景 |
| --- | --- | --- |
| `auto`（默认） | 调 `/media/prepare`，按服务端返回选择直传或中转 | 生产环境 |
| `relay` | 跳过 prepare，直接走中转 `/api/v1/media/upload` | 调试、或临时禁用直传 |
| `direct` | 调 prepare，但服务端返回中转时抛错 | 验证 STS 直传链路 |

```ts
// 强制走中转（少一次请求往返）
await submit(file, { key: 'user.avatar', mode: 'relay' })

// 强制走直传（排查直传问题时用）
await submit(file, { key: 'user.avatar', mode: 'direct' })
```

---

## 5. 后端 Key 解析规则

理解后端如何解析 key，可以避免踩坑：

```
前端传入 key='doc.attachment', rule='office'
  ① 精确匹配: 查 UploadKey.key = 'doc.attachment' → 找到 ✅
  ② 再查 Rule: rule_key = 'office' → 找到 ✅
  ③ 合并返回: 使用 doc.attachment 配置 + office 规则

前端传入 key='doc.attachment', rule=''
  ① 精确匹配: 查 UploadKey.key = 'doc.attachment' → 找到 ✅
  ② Rule 为空: 使用 default_rule_key 或标记 is_default=true 的 Rule → 'pdf'
  ③ 合并返回: 使用 doc.attachment 配置 + pdf 规则（默认）

前端传入 key='not.exist', rule=''
  ① 精确匹配: 查 UploadKey.key = 'not.exist' → 未找到
  ② 末位拆点: 尝试 key='not', rule='exist' → 查 UploadKey.key = 'not' → 未找到
  ③ 回退: 使用系统默认上传配置（media.default）
```

---

## 6. API 层直接调用

如果不使用 `useUpload` 组合式函数，也可以直接调用底层 API：

```ts
import { uploadMediaWithPrepare } from '@/domains/upload/api'
import type { MediaUploadOptions } from '@/domains/upload/api'

const options: MediaUploadOptions = {
  key: 'user.avatar',
  rule: 'avatar',
  mode: 'auto',
  signal: abortController.signal,  // 支持取消
  onProgress(percent) { /* ... */ },
  onFallback(prepare) {
    // 当 prepare 返回 relay 模式时触发（仅 mode='auto' 时）
    console.log('回退到中转上传')
  }
}

const result = await uploadMediaWithPrepare(file, options)
// result: { id, url, filename, storageKey, mimeType, size, createdAt }
```

---

## 7. 速查表

| 业务场景 | 推荐写法 | 后端匹配的 Key | 生效的 Rule |
| --- | --- | --- | --- |
| 用户头像 | `{ key: 'user.avatar' }` | `user.avatar` | `avatar`（默认） |
| PDF 文档 | `{ key: 'doc.attachment' }` | `doc.attachment` | `pdf`（默认） |
| Office 文档 | `{ key: 'doc.attachment', rule: 'office' }` | `doc.attachment` | `office` |
| 编辑器图片 | `{ key: 'editor.inline' }` | `editor.inline` | `editor-image`（默认） |
| 通用图片 | `{ key: 'media.default' }` 或不传 | `media.default` | `image`（默认） |

---

## 8. 新增业务场景

当现有 Key 无法满足需求时，在「上传配置中心」页面新增：

1. **新增上传配置**（UploadKey）：选择所属存储桶，填写标识（如 `product.gallery`）、大小上限、允许类型
2. **新增规则**（Rule）：在新建的 UploadKey 下添加规则，如 `thumbnail`（缩略图规则）、`original`（原图规则）
3. **前端调用**：

```ts
await submit(file, { key: 'product.gallery', rule: 'thumbnail' })
```

无需修改任何前端基础设施代码，新增 key 即刻可用。
