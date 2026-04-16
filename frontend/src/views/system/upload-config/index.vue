<template>
  <div class="p-4 upload-config-page art-full-height">
    <ElCard class="art-table-card upload-config-main" shadow="never">
      <div class="upload-config-header">
        <div class="upload-config-title">上传配置中心</div>
        <div class="upload-config-tip">
          统一管理文件上传链路的四层配置：存储服务 &rarr; 存储桶 &rarr; 上传配置 &rarr;
          上传规则。所有配置变更会自动失效缓存并广播到运行时上传链路。
        </div>
      </div>

      <ElTabs v-model="activeTab" class="upload-config-tabs" @tab-change="onTabChange">
        <!-- ═══ 存储服务 ═══ -->
        <ElTabPane label="存储服务" name="provider">
          <div class="tab-desc">
            存储服务是最底层的连接配置，对应一个对象存储实例（本地磁盘或云 OSS）。
          </div>
          <ElForm :inline="true" class="upload-config-filters">
            <ElFormItem>
              <ElButton type="primary" @click="loadProviders">刷新</ElButton>
              <ElButton type="success" @click="openProviderCreate">新增存储服务</ElButton>
            </ElFormItem>
          </ElForm>
          <ArtTable
            :loading="provider.loading"
            :data="provider.records"
            :columns="providerColumns"
          />
        </ElTabPane>

        <!-- ═══ 存储桶 ═══ -->
        <ElTabPane label="存储桶" name="bucket">
          <div class="tab-desc">
            存储桶隶属于某个存储服务，代表一个逻辑隔离的文件存放区域，可独立配置公网访问地址和基础路径。
          </div>
          <ElForm :inline="true" class="upload-config-filters">
            <ElFormItem label="所属存储服务">
              <ElSelect
                v-model="bucket.providerFilter"
                clearable
                placeholder="全部"
                style="width: 240px"
                @change="loadBuckets"
              >
                <ElOption
                  v-for="p in provider.records"
                  :key="p.id"
                  :label="`${p.name}（${p.provider_key}）`"
                  :value="p.id"
                />
              </ElSelect>
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="loadBuckets">刷新</ElButton>
              <ElButton type="success" @click="openBucketCreate">新增存储桶</ElButton>
            </ElFormItem>
          </ElForm>
          <ArtTable :loading="bucket.loading" :data="bucket.records" :columns="bucketColumns" />
        </ElTabPane>

        <!-- ═══ 上传配置 ═══ -->
        <ElTabPane label="上传配置" name="upload-key">
          <div class="tab-desc">
            上传配置（UploadKey）对应一个业务上传场景，如头像、附件、编辑器图片等，除了大小、类型和路径模板外，还要明确运行时上传方式、是否开放给前端直传、权限键和可扩展参数。
          </div>
          <ElForm :inline="true" class="upload-config-filters">
            <ElFormItem label="所属存储桶">
              <ElSelect
                v-model="uploadKey.bucketFilter"
                clearable
                placeholder="全部"
                style="width: 240px"
                @change="loadUploadKeys"
              >
                <ElOption
                  v-for="b in bucket.records"
                  :key="b.id"
                  :label="`${b.name}（${b.bucket_key}）`"
                  :value="b.id"
                />
              </ElSelect>
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="loadUploadKeys">刷新</ElButton>
              <ElButton type="success" @click="openUploadKeyCreate">新增上传配置</ElButton>
            </ElFormItem>
          </ElForm>
          <ArtTable
            :loading="uploadKey.loading"
            :data="uploadKey.records"
            :columns="uploadKeyColumns"
          />
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <!-- 存储服务编辑器 -->
    <ElDialog
      v-model="providerEditor.open"
      :title="providerEditor.editingId ? '编辑存储服务' : '新增存储服务'"
      width="640px"
      :close-on-click-modal="false"
      @closed="resetProviderEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="服务标识">
          <ElInput
            v-model="providerEditor.form.provider_key"
            :disabled="!!providerEditor.editingId"
            placeholder="如 local-default、oss-prod"
          />
          <span class="form-tip">唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="providerEditor.form.name" placeholder="便于识别的显示名称" />
        </ElFormItem>
        <ElFormItem label="驱动类型">
          <ElSelect
            v-model="providerEditor.form.driver"
            style="width: 220px"
            @change="onProviderDriverChange"
          >
            <ElOption label="本地存储" value="local" />
            <ElOption label="阿里云 OSS" value="aliyun_oss" />
          </ElSelect>
          <span class="form-tip">选择存储后端类型</span>
        </ElFormItem>
        <ElFormItem label="基础访问地址">
          <ElInput
            v-model="providerEditor.form.base_url"
            placeholder="如 https://cdn.example.com"
          />
          <span class="form-tip">文件公网访问的根地址，通常为 CDN 域名</span>
        </ElFormItem>
        <template v-if="providerEditor.form.driver === 'aliyun_oss'">
          <ElFormItem label="接入点地址">
            <ElInput
              v-model="providerEditor.form.endpoint"
              placeholder="如 oss-cn-hangzhou.aliyuncs.com"
            />
            <span class="form-tip">OSS 服务的接入域名</span>
          </ElFormItem>
          <ElFormItem label="地域">
            <ElInput v-model="providerEditor.form.region" placeholder="如 cn-hangzhou" />
          </ElFormItem>
          <ElFormItem label="访问密钥（AK）">
            <ElInput
              v-model="providerEditor.form.access_key"
              placeholder="留空表示保留原值"
              autocomplete="off"
            />
          </ElFormItem>
          <ElFormItem label="安全密钥（SK）">
            <ElInput
              v-model="providerEditor.form.secret_key"
              type="password"
              show-password
              placeholder="留空表示保留原值"
              autocomplete="new-password"
            />
          </ElFormItem>
        </template>
        <ElFormItem label="设为默认">
          <ElSwitch v-model="providerEditor.form.is_default" />
          <span class="form-tip">开启后，未指定存储服务的场景将使用此服务</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="providerEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
        <div class="driver-guide-card">
          <div class="driver-guide-head">
            <div class="driver-guide-title">
              当前驱动：{{ driverLabel[providerEditor.form.driver] || providerEditor.form.driver }}
            </div>
            <ElButton link type="primary" @click="restoreProviderDriverDefaults"
              >恢复推荐默认值</ElButton
            >
          </div>
          <ul class="driver-guide-list">
            <li v-for="item in providerExtraGuide" :key="item">{{ item }}</li>
          </ul>
        </div>
        <ElCollapse v-model="providerEditor.activePanels" class="driver-collapse">
          <ElCollapseItem
            v-for="section in providerExtraSections"
            :key="section.key"
            :name="section.key"
            :title="section.title"
          >
            <div v-if="section.description" class="driver-section-desc">
              {{ section.description }}
            </div>
            <ElFormItem v-for="field in section.fields" :key="field.key" :label="field.label">
              <ElSwitch
                v-if="field.type === 'boolean'"
                :model-value="readExtraBooleanValue(providerEditor.form.extra, field.key)"
                @update:model-value="setExtraValue(providerEditor.form.extra, field.key, $event)"
              />
              <ElInputNumber
                v-else-if="field.type === 'number'"
                :model-value="readExtraNumberValue(providerEditor.form.extra, field.key)"
                :min="field.min ?? 0"
                :step="field.step ?? 1"
                controls-position="right"
                style="width: 240px"
                @update:model-value="setExtraValue(providerEditor.form.extra, field.key, $event)"
              />
              <ElInput
                v-else-if="field.type === 'object'"
                :model-value="providerEditor.objectExtraText[field.key] || ''"
                type="textarea"
                :autosize="{ minRows: field.rows ?? 4, maxRows: Math.max(field.rows ?? 4, 8) }"
                :placeholder="field.placeholder"
                @update:model-value="
                  setExtraValue(providerEditor.objectExtraText, field.key, $event)
                "
              />
              <ElInput
                v-else
                :model-value="readExtraStringValue(providerEditor.form.extra, field.key)"
                :type="field.multiline ? 'textarea' : 'text'"
                :autosize="
                  field.multiline
                    ? { minRows: field.rows ?? 3, maxRows: Math.max(field.rows ?? 3, 8) }
                    : undefined
                "
                :placeholder="field.placeholder"
                @update:model-value="setExtraValue(providerEditor.form.extra, field.key, $event)"
              />
              <span v-if="formatDriverExtraFieldTip(field)" class="form-tip">
                {{ formatDriverExtraFieldTip(field) }}
              </span>
            </ElFormItem>
          </ElCollapseItem>
          <ElCollapseItem name="custom" title="自定义扩展参数">
            <div class="driver-section-desc">
              用于承载当前页面未结构化覆盖的扩展键，便于未来驱动新增参数或内部定制能力。
            </div>
            <ElFormItem label="附加参数 JSON">
              <ElInput
                v-model="providerEditor.customExtraText"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 10 }"
                placeholder='{&#10;  "custom_endpoint_policy": "internal-only"&#10;}'
              />
              <span class="form-tip">键名不能与上方已结构化字段重复</span>
            </ElFormItem>
          </ElCollapseItem>
        </ElCollapse>
      </ElForm>
      <template #footer>
        <ElButton @click="providerEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="providerEditor.submitting" @click="submitProvider"
          >保存</ElButton
        >
      </template>
    </ElDialog>

    <!-- 存储桶编辑器 -->
    <ElDialog
      v-model="bucketEditor.open"
      :title="bucketEditor.editingId ? '编辑存储桶' : '新增存储桶'"
      width="640px"
      :close-on-click-modal="false"
      @closed="resetBucketEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="所属存储服务">
          <ElSelect
            v-model="bucketEditor.form.provider_id"
            :disabled="!!bucketEditor.editingId"
            style="width: 100%"
            @change="onBucketProviderChange"
          >
            <ElOption
              v-for="p in provider.records"
              :key="p.id"
              :label="`${p.name}（${p.provider_key}）`"
              :value="p.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="存储桶标识">
          <ElInput
            v-model="bucketEditor.form.bucket_key"
            :disabled="!!bucketEditor.editingId"
            placeholder="如 default-bucket"
          />
          <span class="form-tip">唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="bucketEditor.form.name" placeholder="便于识别的显示名称" />
        </ElFormItem>
        <ElFormItem label="存储桶名称">
          <ElInput
            v-model="bucketEditor.form.bucket_name"
            placeholder="对象存储中实际的 Bucket 名称"
          />
          <span class="form-tip">对应云存储服务中的实际 Bucket</span>
        </ElFormItem>
        <ElFormItem label="基础路径">
          <ElInput v-model="bucketEditor.form.base_path" placeholder="可选，文件存储的前缀目录" />
          <span class="form-tip">所有文件都会存储在此目录下</span>
        </ElFormItem>
        <ElFormItem label="公网访问地址">
          <ElInput
            v-model="bucketEditor.form.public_base_url"
            placeholder="访问已上传文件用的公网根地址"
          />
          <span class="form-tip">留空则继承存储服务的基础访问地址</span>
        </ElFormItem>
        <ElFormItem label="公开访问">
          <ElSwitch v-model="bucketEditor.form.is_public" />
          <span class="form-tip">开启后文件可通过公网直接访问，否则需签名</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="bucketEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
        <div class="driver-guide-card">
          <div class="driver-guide-head">
            <div class="driver-guide-title">
              当前驱动：{{ driverLabel[bucketDriver] || bucketDriver || '未选择存储服务' }}
            </div>
            <ElButton
              link
              type="primary"
              :disabled="!bucketDriver"
              @click="restoreBucketDriverDefaults"
              >恢复推荐默认值</ElButton
            >
          </div>
          <ul v-if="bucketExtraGuide.length" class="driver-guide-list">
            <li v-for="item in bucketExtraGuide" :key="item">{{ item }}</li>
          </ul>
          <div v-else class="driver-section-desc"
            >请先选择存储服务，再配置当前桶的驱动扩展参数。</div
          >
        </div>
        <ElCollapse v-model="bucketEditor.activePanels" class="driver-collapse">
          <ElCollapseItem
            v-for="section in bucketExtraSections"
            :key="section.key"
            :name="section.key"
            :title="section.title"
          >
            <div v-if="section.description" class="driver-section-desc">
              {{ section.description }}
            </div>
            <ElFormItem v-for="field in section.fields" :key="field.key" :label="field.label">
              <ElSwitch
                v-if="field.type === 'boolean'"
                :model-value="readExtraBooleanValue(bucketEditor.form.extra, field.key)"
                @update:model-value="setExtraValue(bucketEditor.form.extra, field.key, $event)"
              />
              <ElInputNumber
                v-else-if="field.type === 'number'"
                :model-value="readExtraNumberValue(bucketEditor.form.extra, field.key)"
                :min="field.min ?? 0"
                :step="field.step ?? 1"
                controls-position="right"
                style="width: 240px"
                @update:model-value="setExtraValue(bucketEditor.form.extra, field.key, $event)"
              />
              <ElInput
                v-else-if="field.type === 'object'"
                :model-value="bucketEditor.objectExtraText[field.key] || ''"
                type="textarea"
                :autosize="{ minRows: field.rows ?? 4, maxRows: Math.max(field.rows ?? 4, 8) }"
                :placeholder="field.placeholder"
                @update:model-value="setExtraValue(bucketEditor.objectExtraText, field.key, $event)"
              />
              <ElInput
                v-else
                :model-value="readExtraStringValue(bucketEditor.form.extra, field.key)"
                :type="field.multiline ? 'textarea' : 'text'"
                :autosize="
                  field.multiline
                    ? { minRows: field.rows ?? 3, maxRows: Math.max(field.rows ?? 3, 8) }
                    : undefined
                "
                :placeholder="field.placeholder"
                @update:model-value="setExtraValue(bucketEditor.form.extra, field.key, $event)"
              />
              <span v-if="formatDriverExtraFieldTip(field)" class="form-tip">
                {{ formatDriverExtraFieldTip(field) }}
              </span>
            </ElFormItem>
          </ElCollapseItem>
          <ElCollapseItem name="custom" title="自定义扩展参数">
            <div class="driver-section-desc">
              这里保留给驱动新增键或内部约定参数，避免每次扩展都要等页面改版。
            </div>
            <ElFormItem label="附加参数 JSON">
              <ElInput
                v-model="bucketEditor.customExtraText"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 10 }"
                placeholder='{&#10;  "origin_access_identity": "private"&#10;}'
              />
              <span class="form-tip">键名不能与上方已结构化字段重复</span>
            </ElFormItem>
          </ElCollapseItem>
        </ElCollapse>
      </ElForm>
      <template #footer>
        <ElButton @click="bucketEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="bucketEditor.submitting" @click="submitBucket"
          >保存</ElButton
        >
      </template>
    </ElDialog>

    <!-- 上传配置编辑器 -->
    <ElDialog
      v-model="uploadKeyEditor.open"
      :title="uploadKeyEditor.editingId ? '编辑上传配置' : '新增上传配置'"
      width="680px"
      :close-on-click-modal="false"
      @closed="resetUploadKeyEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="所属存储桶">
          <ElSelect
            v-model="uploadKeyEditor.form.bucket_id"
            :disabled="!!uploadKeyEditor.editingId"
            style="width: 100%"
          >
            <ElOption
              v-for="b in bucket.records"
              :key="b.id"
              :label="`${b.name}（${b.bucket_key}）`"
              :value="b.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="上传标识">
          <ElInput
            v-model="uploadKeyEditor.form.key"
            :disabled="!!uploadKeyEditor.editingId"
            placeholder="如 avatar、attachment、public-asset"
          />
          <span class="form-tip">业务场景的唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="uploadKeyEditor.form.name" placeholder="便于识别的显示名称" />
        </ElFormItem>
        <div class="config-guide-card">
          <div class="config-guide-title">UploadKey 配置顺序</div>
          <ol class="config-guide-list">
            <li>先确定存储路径和默认规则，明确这个业务场景落到哪个 Bucket、默认走哪条规则。</li>
            <li>再决定上传模式、前端是否可见、是否公开访问以及直传阈值。</li>
            <li>最后补权限键、回退 UploadKey 和自定义参数，让业务侧知道哪些附加字段可配置。</li>
          </ol>
        </div>
        <div class="config-section-title">基础路由</div>
        <div class="config-section-tip">决定对象最终写入哪个目录，以及默认使用哪条子规则。</div>
        <ElFormItem label="路径模板">
          <ElInput
            v-model="uploadKeyEditor.form.path_template"
            placeholder="{tenant}/{key}/{date}/{uuid}{ext}"
          />
          <span class="form-tip">支持变量：{tenant}、{key}、{date}、{uuid}、{ext}</span>
        </ElFormItem>
        <ElFormItem label="默认规则标识">
          <ElInput
            v-model="uploadKeyEditor.form.default_rule_key"
            placeholder="可选，留空则使用标记为默认的规则"
          />
        </ElFormItem>
        <ElFormItem label="单文件上限">
          <ElInputNumber
            v-model="uploadKeyEditor.form.max_size_bytes"
            :min="0"
            controls-position="right"
            style="width: 240px"
          />
          <span class="form-tip">单位：字节。0 表示沿用存储桶或全局上限</span>
        </ElFormItem>
        <div class="config-section-title">运行时策略</div>
        <div class="config-section-tip"
          >这里决定前端能否看到该场景，以及实际走直传、后端中转还是自动选择。</div
        >
        <ElFormItem label="前端 accept">
          <ElInput
            :model-value="uploadKeyEditor.clientAcceptText"
            placeholder="逗号分隔，如 image/*,.pdf，仅作为前端选择器提示"
            @update:model-value="onClientAcceptInput"
          />
          <span class="form-tip">只影响前端文件选择器提示，不替代后端 MIME 校验</span>
        </ElFormItem>
        <ElFormItem label="上传方式">
          <ElSelect v-model="uploadKeyEditor.form.upload_mode" style="width: 220px">
            <ElOption label="自动选择" value="auto" />
            <ElOption label="前端直传" value="direct" />
            <ElOption label="后端中转" value="relay" />
          </ElSelect>
          <span class="form-tip">自动模式下会按驱动能力判断；直传要求底层驱动支持直传参数签发</span>
        </ElFormItem>
        <ElFormItem label="直传阈值">
          <ElInputNumber
            v-model="uploadKeyEditor.form.direct_size_threshold_bytes"
            :min="0"
            controls-position="right"
            style="width: 240px"
          />
          <span class="form-tip">单位：字节。大于该值时强制改走后端中转，0 表示不启用阈值</span>
        </ElFormItem>
        <ElFormItem label="可见性">
          <ElSelect v-model="uploadKeyEditor.form.visibility" style="width: 220px">
            <ElOption label="公开" value="public" />
            <ElOption label="私有" value="private" />
          </ElSelect>
          <span class="form-tip">公开文件可直接通过 URL 访问</span>
        </ElFormItem>
        <ElFormItem label="前端可见">
          <ElSwitch v-model="uploadKeyEditor.form.is_frontend_visible" />
          <span class="form-tip">开启后，该 UploadKey 会出现在业务前端可见的上传场景列表中</span>
        </ElFormItem>
        <div class="config-section-title">访问控制与扩展</div>
        <div class="config-section-tip"
          >权限键控制谁能拿到该场景；回退 UploadKey
          用于兜底切换；自定义参数用于补业务侧可填项说明。</div
        >
        <ElFormItem label="权限键">
          <ElInput
            v-model="uploadKeyEditor.form.permission_key"
            placeholder="如 cms.asset.upload，留空表示登录即可使用"
          />
          <span class="form-tip">运行时上传会按权限键过滤；适合区分不同业务场景的上传授权</span>
        </ElFormItem>
        <ElFormItem label="回退 UploadKey">
          <ElInput
            v-model="uploadKeyEditor.form.fallback_key"
            placeholder="可选，主配置不可用时尝试回退到的 UploadKey 标识"
          />
          <span class="form-tip">用于灰度切换或灾备兜底，需保证目标 UploadKey 已存在</span>
        </ElFormItem>
        <ElFormItem label="扩展参数">
          <ExtraSchemaEditor
            ref="uploadKeySchemaEditorRef"
            :model-value="uploadKeyEditor.form.extra_schema"
            title="UploadKey 自定义参数"
          />
          <span class="form-tip">用字段化方式描述扩展参数，便于配置人员直接理解可填项</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="uploadKeyEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="允许文件类型">
          <ElInput
            :model-value="uploadKeyEditor.mimeText"
            placeholder="逗号分隔，如 image/*,video/mp4，留空表示不限"
            @update:model-value="onMimeInput"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="uploadKeyEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="uploadKeyEditor.submitting" @click="submitUploadKey"
          >保存</ElButton
        >
      </template>
    </ElDialog>

    <!-- 上传规则管理抽屉 -->
    <ElDrawer
      v-model="rule.drawerOpen"
      :title="`上传规则管理 — ${rule.parentUploadKeyLabel}`"
      size="860px"
      :close-on-click-modal="false"
    >
      <div class="tab-desc" style="margin-bottom: 8px">
        上传规则是上传配置的子级，同一个 UploadKey 下可按规则覆写上传方式、可见性、前端 accept
        和扩展参数，用来区分图片、附件、海报等细分场景。
      </div>
      <div class="rule-drawer-toolbar">
        <ElButton type="primary" size="small" @click="loadRules">刷新</ElButton>
        <ElButton type="success" size="small" @click="openRuleCreate">新增规则</ElButton>
      </div>
      <ArtTable :loading="rule.loading" :data="rule.records" :columns="ruleColumns" />
    </ElDrawer>

    <!-- 上传规则编辑器 -->
    <ElDialog
      v-model="ruleEditor.open"
      :title="ruleEditor.editingId ? '编辑上传规则' : '新增上传规则'"
      width="620px"
      :close-on-click-modal="false"
      append-to-body
      @closed="resetRuleEditor"
    >
      <ElForm label-width="130px">
        <ElFormItem label="规则标识">
          <ElInput
            v-model="ruleEditor.form.rule_key"
            :disabled="!!ruleEditor.editingId"
            placeholder="如 image、file、poster"
          />
          <span class="form-tip">唯一标识，创建后不可修改</span>
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="ruleEditor.form.name" placeholder="如 图片上传、附件上传" />
        </ElFormItem>
        <div class="config-guide-card">
          <div class="config-guide-title">Rule 配置顺序</div>
          <ol class="config-guide-list">
            <li>先定义子路径、文件名策略和大小限制，明确这条规则与 UploadKey 的差异点。</li>
            <li>再决定是否覆写上传方式、可见性和前端选择提示。</li>
            <li>最后再补扩展参数和默认规则标记，用于更细的业务场景区分。</li>
          </ol>
        </div>
        <div class="config-section-title">基础规则</div>
        <div class="config-section-tip">这些字段主要负责落盘位置、文件名和基础文件限制。</div>
        <ElFormItem label="子路径">
          <ElInput v-model="ruleEditor.form.sub_path" placeholder="可选，追加到上传配置路径之后" />
          <span class="form-tip">文件将存储在上传配置路径 + 子路径下</span>
        </ElFormItem>
        <ElFormItem label="文件名策略">
          <ElSelect v-model="ruleEditor.form.filename_strategy" style="width: 220px">
            <ElOption label="随机生成（UUID）" value="uuid" />
            <ElOption label="保留原文件名" value="original" />
          </ElSelect>
          <span class="form-tip">UUID 可避免文件名冲突</span>
        </ElFormItem>
        <ElFormItem label="单文件上限">
          <ElInputNumber
            v-model="ruleEditor.form.max_size_bytes"
            :min="0"
            controls-position="right"
            style="width: 240px"
          />
          <span class="form-tip">单位：字节。0 表示沿用上传配置的上限</span>
        </ElFormItem>
        <ElFormItem label="允许文件类型">
          <ElInput
            :model-value="ruleEditor.ruleMimeText"
            placeholder="逗号分隔，如 image/*,video/mp4，留空表示不限"
            @update:model-value="onRuleMimeInput"
          />
        </ElFormItem>
        <div class="config-section-title">覆写策略</div>
        <div class="config-section-tip">只有与 UploadKey 默认行为不同的地方才需要在这里覆写。</div>
        <ElFormItem label="前端 accept">
          <ElInput
            :model-value="ruleEditor.clientAcceptText"
            placeholder="逗号分隔，如 image/*,.docx，用于细分规则前端选择提示"
            @update:model-value="onRuleClientAcceptInput"
          />
          <span class="form-tip">规则层优先级高于 UploadKey，可按子场景做更精细的选择器提示</span>
        </ElFormItem>
        <ElFormItem label="上传方式覆写">
          <ElSelect v-model="ruleEditor.form.mode_override" style="width: 220px">
            <ElOption label="继承上传配置" value="inherit" />
            <ElOption label="强制前端直传" value="direct" />
            <ElOption label="强制后端中转" value="relay" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="可见性覆写">
          <ElSelect v-model="ruleEditor.form.visibility_override" style="width: 220px">
            <ElOption label="继承上传配置" value="inherit" />
            <ElOption label="公开" value="public" />
            <ElOption label="私有" value="private" />
          </ElSelect>
        </ElFormItem>
        <div class="config-section-title">扩展与默认</div>
        <div class="config-section-tip"
          >通过扩展参数定义该规则特有的附加字段，同时决定它是否成为默认规则。</div
        >
        <ElFormItem label="扩展参数">
          <ExtraSchemaEditor
            ref="ruleSchemaEditorRef"
            :model-value="ruleEditor.form.extra_schema"
            title="Rule 自定义参数"
          />
          <span class="form-tip">规则层可追加更细的附加字段说明，如变体、回调场景、资源标签等</span>
        </ElFormItem>
        <ElFormItem label="设为默认规则">
          <ElSwitch v-model="ruleEditor.form.is_default" />
          <span class="form-tip">上传时未指定规则将自动使用默认规则</span>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="ruleEditor.form.status" style="width: 220px">
            <ElOption label="启用" value="ready" />
            <ElOption label="停用" value="disabled" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="ruleEditor.open = false">取消</ElButton>
        <ElButton type="primary" :loading="ruleEditor.submitting" @click="submitRule"
          >保存</ElButton
        >
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import {
    ElButton,
    ElCard,
    ElCollapse,
    ElCollapseItem,
    ElDialog,
    ElDrawer,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElPopconfirm,
    ElSelect,
    ElSwitch,
    ElTabPane,
    ElTabs,
    ElTag
  } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import {
    fetchCreateStorageBucket,
    fetchCreateStorageProvider,
    fetchCreateUploadKey,
    fetchCreateUploadKeyRule,
    fetchDeleteStorageBucket,
    fetchDeleteStorageProvider,
    fetchDeleteUploadKey,
    fetchDeleteUploadKeyRule,
    fetchListStorageBuckets,
    fetchListStorageProviders,
    fetchListUploadKeyRules,
    fetchListUploadKeys,
    fetchTestStorageProvider,
    fetchUpdateStorageBucket,
    fetchUpdateStorageProvider,
    fetchUpdateUploadKey,
    fetchUpdateUploadKeyRule,
    type StorageBucketSaveRequest,
    type StorageBucketSummary,
    type StorageProviderSaveRequest,
    type StorageProviderSummary,
    type UploadKeyRuleSaveRequest,
    type UploadKeyRuleSummary,
    type UploadKeySaveRequest,
    type UploadKeySummary
  } from '@/domains/upload-config/api'
  import ExtraSchemaEditor from '@/domains/upload-config/components/ExtraSchemaEditor.vue'
  import {
    getDriverExtraDefaults,
    getDriverExtraSections,
    getDriverGuide,
    type DriverExtraField,
    type DriverExtraSection,
    type DriverExtraScope,
    type StorageDriver
  } from '@/domains/upload-config/driver-extra-registry'

  defineOptions({ name: 'SystemUploadConfig' })

  // ── 显示文本映射 ──────────────────────────────────────────────────────────

  const statusLabel: Record<string, string> = {
    ready: '启用',
    disabled: '停用',
    error: '异常'
  }
  const statusType: Record<string, 'success' | 'info' | 'danger'> = {
    ready: 'success',
    disabled: 'info',
    error: 'danger'
  }
  const visibilityLabel: Record<string, string> = {
    public: '公开',
    private: '私有'
  }
  const uploadModeLabel: Record<string, string> = {
    auto: '自动选择',
    direct: '前端直传',
    relay: '后端中转',
    inherit: '继承上传配置'
  }
  const visibilityOverrideLabel: Record<string, string> = {
    inherit: '继承上传配置',
    public: '公开',
    private: '私有'
  }
  const filenameStrategyLabel: Record<string, string> = {
    uuid: '随机（UUID）',
    original: '保留原名'
  }
  const driverLabel: Record<string, string> = {
    local: '本地存储',
    aliyun_oss: '阿里云 OSS'
  }
  type DriverExtraValueMap = Record<string, any>

  function splitCommaValues(value: string): string[] {
    return value
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
  }

  function normalizeObjectValue<T extends Record<string, unknown>>(value: unknown): T | undefined {
    if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
    return { ...(value as T) }
  }

  function stringifyJsonEditor(value: unknown): string {
    const normalized = normalizeObjectValue<Record<string, unknown>>(value)
    return normalized ? JSON.stringify(normalized, null, 2) : ''
  }

  function parseJsonEditor(
    value: string,
    label: string
  ): Record<string, unknown> | undefined | null {
    const text = value.trim()
    if (!text) return undefined
    try {
      const parsed = JSON.parse(text)
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        ElMessage.warning(`${label}必须是 JSON 对象`)
        return null
      }
      return parsed as Record<string, unknown>
    } catch {
      ElMessage.warning(`${label}不是合法的 JSON`)
      return null
    }
  }

  function formatSchemaConfigured(value: unknown): string {
    return normalizeObjectValue(value) ? '已配置' : '-'
  }

  function flattenDriverExtraFields(sections: DriverExtraSection[]): DriverExtraField[] {
    return sections.flatMap((section) => section.fields)
  }

  function defaultOpenExtraPanels(sections: DriverExtraSection[], customText = ''): string[] {
    const panels = sections.filter((section) => section.defaultOpen).map((section) => section.key)
    if (customText.trim()) panels.push('custom')
    return panels
  }

  function buildDriverExtraDraft(
    driver: StorageDriver | '' | undefined,
    scope: DriverExtraScope,
    value?: unknown
  ) {
    const sections = getDriverExtraSections(driver, scope)
    const fields = flattenDriverExtraFields(sections)
    const defaults = getDriverExtraDefaults(driver, scope)
    const source = normalizeObjectValue<DriverExtraValueMap>(value) || {}
    const knownKeys = new Set(fields.map((field) => field.key))
    const values: DriverExtraValueMap = { ...defaults }
    const custom: Record<string, unknown> = {}
    for (const [key, item] of Object.entries(source)) {
      if (knownKeys.has(key)) {
        values[key] = item
      } else {
        custom[key] = item
      }
    }
    const objectText: Record<string, string> = {}
    for (const field of fields) {
      if (field.type !== 'object') continue
      const current = values[field.key]
      if (current === undefined || current === null || current === '') {
        objectText[field.key] = ''
        continue
      }
      objectText[field.key] =
        typeof current === 'string' ? current : JSON.stringify(current, null, 2)
    }
    const customText = stringifyJsonEditor(custom)
    return {
      values,
      objectText,
      customText,
      activePanels: defaultOpenExtraPanels(sections, customText)
    }
  }

  function buildDriverExtraBody(
    label: string,
    driver: StorageDriver | '' | undefined,
    scope: DriverExtraScope,
    values: DriverExtraValueMap,
    objectText: Record<string, string>,
    customText: string
  ): Record<string, unknown> | undefined | null {
    const sections = getDriverExtraSections(driver, scope)
    const fields = flattenDriverExtraFields(sections)
    const knownKeys = new Set(fields.map((field) => field.key))
    const payload: Record<string, unknown> = {}

    for (const field of fields) {
      if (field.type === 'object') {
        const parsed = parseJsonEditor(objectText[field.key] || '', `${label}${field.label}`)
        if (parsed === null) return null
        if (parsed) payload[field.key] = parsed
        continue
      }
      const current = values[field.key]
      if (field.type === 'boolean') {
        if (typeof current === 'boolean') payload[field.key] = current
        continue
      }
      if (field.type === 'number') {
        if (
          current !== undefined &&
          current !== null &&
          current !== '' &&
          Number.isFinite(Number(current))
        ) {
          payload[field.key] = Number(current)
        }
        continue
      }
      const text = typeof current === 'string' ? current.trim() : String(current ?? '').trim()
      if (text) payload[field.key] = text
    }

    const custom = parseJsonEditor(customText, `${label}自定义扩展参数`)
    if (custom === null) return null
    if (custom) {
      for (const key of Object.keys(custom)) {
        if (knownKeys.has(key)) {
          ElMessage.warning(
            `${label}自定义扩展参数中的 ${key} 已由结构化字段托管，请直接使用上方表单`
          )
          return null
        }
      }
      Object.assign(payload, custom)
    }
    return Object.keys(payload).length ? payload : undefined
  }

  function formatDriverExtraValue(value: unknown): string {
    if (value === undefined || value === null || value === '') return ''
    if (typeof value === 'object') {
      return JSON.stringify(value)
    }
    return String(value)
  }

  function formatDriverExtraFieldTip(field: DriverExtraField): string {
    const parts: string[] = []
    if (field.description) parts.push(field.description)
    if (field.defaultValue !== undefined) {
      parts.push(`推荐默认：${formatDriverExtraValue(field.defaultValue)}`)
    }
    return parts.join(' ')
  }

  function readExtraStringValue(values: DriverExtraValueMap, key: string): string {
    const value = values[key]
    return typeof value === 'string'
      ? value
      : value === undefined || value === null
        ? ''
        : String(value)
  }

  function readExtraNumberValue(values: DriverExtraValueMap, key: string): number | undefined {
    const value = values[key]
    if (typeof value === 'number' && Number.isFinite(value)) return value
    if (typeof value === 'string' && value.trim() && Number.isFinite(Number(value))) {
      return Number(value)
    }
    return undefined
  }

  function readExtraBooleanValue(values: DriverExtraValueMap, key: string): boolean {
    return values[key] === true
  }

  function setExtraValue(values: DriverExtraValueMap, key: string, value: unknown) {
    values[key] = value
  }

  function buildExtraSchemaFromEditor(
    editor: InstanceType<typeof ExtraSchemaEditor> | null,
    label: string
  ): Record<string, unknown> | undefined | null {
    const result = editor?.buildSchema()
    if (!result) return undefined
    if (result.error) {
      ElMessage.warning(`${label}${result.error}`)
      return null
    }
    return result.value
  }

  function renderFeatureTags(labels: string[]) {
    if (!labels.length) return '-'
    return h(
      'div',
      { class: 'config-inline-tags' },
      labels.map((label) => h(ElTag, { type: 'info', effect: 'plain', size: 'small' }, () => label))
    )
  }

  function buildProviderFeatureLabels(row: StorageProviderSummary): string[] {
    const extra = normalizeObjectValue<DriverExtraValueMap>(row.extra)
    if (!extra) return []
    const labels: string[] = []
    if (extra.sts_role_arn) labels.push('STS')
    if (extra.use_cname === true) labels.push('CNAME')
    if (extra.use_path_style === true) labels.push('PathStyle')
    if (extra.disable_ssl === true) labels.push('HTTP')
    if (!labels.length && Object.keys(extra).length) labels.push('已配置')
    return labels
  }

  function buildBucketFeatureLabels(row: StorageBucketSummary): string[] {
    const extra = normalizeObjectValue<DriverExtraValueMap>(row.extra)
    if (!extra) return []
    const labels: string[] = []
    if (extra.success_action_status) labels.push(`状态${extra.success_action_status}`)
    if (extra.content_disposition) labels.push('下载头')
    if (extra.callback || extra.callback_var) labels.push('回调')
    if (!labels.length && Object.keys(extra).length) labels.push('已配置')
    return labels
  }

  type TabKey = 'provider' | 'bucket' | 'upload-key'
  const activeTab = ref<TabKey>('provider')
  const uploadKeySchemaEditorRef = ref<InstanceType<typeof ExtraSchemaEditor> | null>(null)
  const ruleSchemaEditorRef = ref<InstanceType<typeof ExtraSchemaEditor> | null>(null)

  // ── Provider state ────────────────────────────────────────────────────────

  const provider = reactive({
    loading: false,
    records: [] as StorageProviderSummary[]
  })
  const providerEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    objectExtraText: {} as Record<string, string>,
    customExtraText: '',
    activePanels: [] as string[],
    form: {
      provider_key: '',
      name: '',
      driver: 'local' as StorageProviderSaveRequest['driver'],
      endpoint: '',
      region: '',
      base_url: '',
      access_key: '',
      secret_key: '',
      extra: {} as DriverExtraValueMap,
      is_default: false,
      status: 'ready' as Exclude<StorageProviderSaveRequest['status'], undefined>
    }
  })

  const providerExtraGuide = computed(() => getDriverGuide(providerEditor.form.driver, 'provider'))
  const providerExtraSections = computed(() =>
    getDriverExtraSections(providerEditor.form.driver, 'provider')
  )

  function resetProviderDriverExtraState(driver: StorageDriver | '' | undefined, value?: unknown) {
    const draft = buildDriverExtraDraft(driver, 'provider', value)
    providerEditor.form.extra = draft.values
    providerEditor.objectExtraText = draft.objectText
    providerEditor.customExtraText = draft.customText
    providerEditor.activePanels = draft.activePanels
  }

  function resetProviderEditor() {
    providerEditor.submitting = false
    providerEditor.editingId = ''
    providerEditor.objectExtraText = {}
    providerEditor.customExtraText = ''
    providerEditor.activePanels = []
    providerEditor.form = {
      provider_key: '',
      name: '',
      driver: 'local',
      endpoint: '',
      region: '',
      base_url: '',
      access_key: '',
      secret_key: '',
      extra: {},
      is_default: false,
      status: 'ready'
    }
    resetProviderDriverExtraState('local')
  }

  function openProviderCreate() {
    resetProviderEditor()
    providerEditor.open = true
  }

  function openProviderEdit(row: StorageProviderSummary) {
    resetProviderEditor()
    providerEditor.editingId = row.id
    providerEditor.form.provider_key = row.provider_key
    providerEditor.form.name = row.name
    providerEditor.form.driver = row.driver
    providerEditor.form.endpoint = row.endpoint || ''
    providerEditor.form.region = row.region || ''
    providerEditor.form.base_url = row.base_url || ''
    providerEditor.form.access_key = ''
    providerEditor.form.secret_key = ''
    resetProviderDriverExtraState(row.driver, row.extra)
    providerEditor.form.is_default = !!row.is_default
    providerEditor.form.status = row.status === 'error' ? 'ready' : row.status
    providerEditor.open = true
  }

  function onProviderDriverChange(driver: StorageDriver) {
    resetProviderDriverExtraState(driver)
  }

  function restoreProviderDriverDefaults() {
    resetProviderDriverExtraState(providerEditor.form.driver)
  }

  function buildProviderBody(): StorageProviderSaveRequest | null {
    const f = providerEditor.form
    const extra = buildDriverExtraBody(
      '存储服务',
      f.driver,
      'provider',
      f.extra,
      providerEditor.objectExtraText,
      providerEditor.customExtraText
    )
    if (extra === null) return null
    const body: StorageProviderSaveRequest = {
      provider_key: f.provider_key.trim(),
      name: f.name.trim(),
      driver: f.driver,
      is_default: f.is_default,
      status: f.status
    }
    if (f.base_url.trim()) body.base_url = f.base_url.trim()
    if (f.driver === 'aliyun_oss') {
      if (f.endpoint.trim()) body.endpoint = f.endpoint.trim()
      if (f.region.trim()) body.region = f.region.trim()
      if (f.access_key.trim()) body.access_key = f.access_key
      if (f.secret_key.trim()) body.secret_key = f.secret_key
    }
    if (extra) body.extra = extra
    return body
  }

  async function submitProvider() {
    const body = buildProviderBody()
    if (!body) return
    if (!body.provider_key || !body.name) {
      ElMessage.warning('服务标识和名称必填')
      return
    }
    providerEditor.submitting = true
    try {
      if (providerEditor.editingId) {
        await fetchUpdateStorageProvider(providerEditor.editingId, body)
        ElMessage.success('存储服务已更新')
      } else {
        await fetchCreateStorageProvider(body)
        ElMessage.success('存储服务已创建')
      }
      providerEditor.open = false
      await loadProviders()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存存储服务失败')
    } finally {
      providerEditor.submitting = false
    }
  }

  async function removeProvider(row: StorageProviderSummary) {
    try {
      await fetchDeleteStorageProvider(row.id)
      ElMessage.success('已删除')
      await loadProviders()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  async function testProvider(row: StorageProviderSummary) {
    try {
      const result = await fetchTestStorageProvider(row.id)
      const detail = `结果：${result.ok ? '正常' : '异常'}${result.message ? ` / ${result.message}` : ''}${
        typeof result.latency_ms === 'number' ? ` / 延迟 ${result.latency_ms}ms` : ''
      }`
      ElMessageBox.alert(detail, '健康检查结果', { type: result.ok ? 'success' : 'warning' })
    } catch (err: any) {
      ElMessage.error(err?.message || '健康检查失败')
    }
  }

  async function loadProviders() {
    provider.loading = true
    try {
      const res = await fetchListStorageProviders()
      provider.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载存储服务列表失败')
    } finally {
      provider.loading = false
    }
  }

  // ── Bucket state ─────────────────────────────────────────────────────────

  const bucket = reactive({
    loading: false,
    providerFilter: '' as string,
    records: [] as StorageBucketSummary[]
  })
  const bucketEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    objectExtraText: {} as Record<string, string>,
    customExtraText: '',
    activePanels: [] as string[],
    form: {
      provider_id: '',
      bucket_key: '',
      name: '',
      bucket_name: '',
      base_path: '',
      public_base_url: '',
      extra: {} as DriverExtraValueMap,
      is_public: false,
      status: 'ready' as Exclude<StorageBucketSaveRequest['status'], undefined>
    }
  })

  const selectedBucketProvider = computed(
    () => provider.records.find((item) => item.id === bucketEditor.form.provider_id) || null
  )
  const bucketDriver = computed<StorageDriver | ''>(
    () => selectedBucketProvider.value?.driver || ''
  )
  const bucketExtraGuide = computed(() => getDriverGuide(bucketDriver.value, 'bucket'))
  const bucketExtraSections = computed(() => getDriverExtraSections(bucketDriver.value, 'bucket'))

  function getProviderDriverById(providerId: string): StorageDriver | '' {
    return provider.records.find((item) => item.id === providerId)?.driver || ''
  }

  function resetBucketDriverExtraState(driver: StorageDriver | '' | undefined, value?: unknown) {
    const draft = buildDriverExtraDraft(driver, 'bucket', value)
    bucketEditor.form.extra = draft.values
    bucketEditor.objectExtraText = draft.objectText
    bucketEditor.customExtraText = draft.customText
    bucketEditor.activePanels = draft.activePanels
  }

  function resetBucketEditor() {
    bucketEditor.submitting = false
    bucketEditor.editingId = ''
    bucketEditor.objectExtraText = {}
    bucketEditor.customExtraText = ''
    bucketEditor.activePanels = []
    bucketEditor.form = {
      provider_id: '',
      bucket_key: '',
      name: '',
      bucket_name: '',
      base_path: '',
      public_base_url: '',
      extra: {},
      is_public: false,
      status: 'ready'
    }
    resetBucketDriverExtraState('')
  }

  function openBucketCreate() {
    resetBucketEditor()
    if (provider.records[0]) bucketEditor.form.provider_id = provider.records[0].id
    resetBucketDriverExtraState(getProviderDriverById(bucketEditor.form.provider_id))
    bucketEditor.open = true
  }

  function openBucketEdit(row: StorageBucketSummary) {
    resetBucketEditor()
    bucketEditor.editingId = row.id
    bucketEditor.form.provider_id = row.provider_id
    bucketEditor.form.bucket_key = row.bucket_key
    bucketEditor.form.name = row.name
    bucketEditor.form.bucket_name = row.bucket_name
    bucketEditor.form.base_path = row.base_path || ''
    bucketEditor.form.public_base_url = row.public_base_url || ''
    resetBucketDriverExtraState(getProviderDriverById(row.provider_id), row.extra)
    bucketEditor.form.is_public = !!row.is_public
    bucketEditor.form.status = row.status
    bucketEditor.open = true
  }

  function onBucketProviderChange(providerId: string) {
    resetBucketDriverExtraState(getProviderDriverById(providerId))
  }

  function restoreBucketDriverDefaults() {
    resetBucketDriverExtraState(bucketDriver.value)
  }

  function buildBucketBody(): StorageBucketSaveRequest | null {
    const f = bucketEditor.form
    const extra = buildDriverExtraBody(
      '存储桶',
      bucketDriver.value,
      'bucket',
      f.extra,
      bucketEditor.objectExtraText,
      bucketEditor.customExtraText
    )
    if (extra === null) return null
    const body: StorageBucketSaveRequest = {
      provider_id: f.provider_id,
      bucket_key: f.bucket_key.trim(),
      name: f.name.trim(),
      bucket_name: f.bucket_name.trim(),
      is_public: f.is_public,
      status: f.status
    }
    if (f.base_path.trim()) body.base_path = f.base_path.trim()
    if (f.public_base_url.trim()) body.public_base_url = f.public_base_url.trim()
    if (extra) body.extra = extra
    return body
  }

  async function submitBucket() {
    const body = buildBucketBody()
    if (!body) return
    if (!body.provider_id || !body.bucket_key || !body.name || !body.bucket_name) {
      ElMessage.warning('所属存储服务、存储桶标识、名称、存储桶名称均为必填')
      return
    }
    bucketEditor.submitting = true
    try {
      if (bucketEditor.editingId) {
        await fetchUpdateStorageBucket(bucketEditor.editingId, body)
        ElMessage.success('存储桶已更新')
      } else {
        await fetchCreateStorageBucket(body)
        ElMessage.success('存储桶已创建')
      }
      bucketEditor.open = false
      await loadBuckets()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存存储桶失败')
    } finally {
      bucketEditor.submitting = false
    }
  }

  async function removeBucket(row: StorageBucketSummary) {
    try {
      await fetchDeleteStorageBucket(row.id)
      ElMessage.success('已删除')
      await loadBuckets()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  async function loadBuckets() {
    bucket.loading = true
    try {
      const res = await fetchListStorageBuckets(bucket.providerFilter || undefined)
      bucket.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载存储桶列表失败')
    } finally {
      bucket.loading = false
    }
  }

  // ── UploadKey state ──────────────────────────────────────────────────────

  const uploadKey = reactive({
    loading: false,
    bucketFilter: '' as string,
    records: [] as UploadKeySummary[]
  })
  const uploadKeyEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    mimeText: '',
    clientAcceptText: '',
    form: {
      bucket_id: '',
      key: '',
      name: '',
      path_template: '',
      default_rule_key: '',
      max_size_bytes: 0,
      allowed_mime_types: [] as string[],
      upload_mode: 'auto' as Exclude<UploadKeySaveRequest['upload_mode'], undefined>,
      is_frontend_visible: false,
      permission_key: '',
      fallback_key: '',
      client_accept: [] as string[],
      direct_size_threshold_bytes: 0,
      extra_schema: undefined as UploadKeySaveRequest['extra_schema'] | undefined,
      visibility: 'private' as Exclude<UploadKeySaveRequest['visibility'], undefined>,
      status: 'ready' as Exclude<UploadKeySaveRequest['status'], undefined>
    }
  })

  function resetUploadKeyEditor() {
    uploadKeyEditor.submitting = false
    uploadKeyEditor.editingId = ''
    uploadKeyEditor.mimeText = ''
    uploadKeyEditor.clientAcceptText = ''
    uploadKeyEditor.form = {
      bucket_id: '',
      key: '',
      name: '',
      path_template: '',
      default_rule_key: '',
      max_size_bytes: 0,
      allowed_mime_types: [],
      upload_mode: 'auto',
      is_frontend_visible: false,
      permission_key: '',
      fallback_key: '',
      client_accept: [],
      direct_size_threshold_bytes: 0,
      extra_schema: undefined,
      visibility: 'private',
      status: 'ready'
    }
  }

  function openUploadKeyCreate() {
    resetUploadKeyEditor()
    if (bucket.records[0]) uploadKeyEditor.form.bucket_id = bucket.records[0].id
    uploadKeyEditor.open = true
  }

  function openUploadKeyEdit(row: UploadKeySummary) {
    resetUploadKeyEditor()
    uploadKeyEditor.editingId = row.id
    uploadKeyEditor.form.bucket_id = row.bucket_id
    uploadKeyEditor.form.key = row.key
    uploadKeyEditor.form.name = row.name
    uploadKeyEditor.form.path_template = row.path_template || ''
    uploadKeyEditor.form.default_rule_key = row.default_rule_key || ''
    uploadKeyEditor.form.max_size_bytes = Number(row.max_size_bytes ?? 0)
    uploadKeyEditor.form.allowed_mime_types = Array.isArray(row.allowed_mime_types)
      ? row.allowed_mime_types
      : []
    uploadKeyEditor.form.upload_mode = row.upload_mode || 'auto'
    uploadKeyEditor.form.is_frontend_visible = !!row.is_frontend_visible
    uploadKeyEditor.form.permission_key = row.permission_key || ''
    uploadKeyEditor.form.fallback_key = row.fallback_key || ''
    uploadKeyEditor.form.client_accept = Array.isArray(row.client_accept) ? row.client_accept : []
    uploadKeyEditor.form.direct_size_threshold_bytes = Number(row.direct_size_threshold_bytes ?? 0)
    uploadKeyEditor.form.extra_schema = normalizeObjectValue(row.extra_schema)
    uploadKeyEditor.form.visibility = row.visibility
    uploadKeyEditor.form.status = row.status
    uploadKeyEditor.mimeText = uploadKeyEditor.form.allowed_mime_types.join(',')
    uploadKeyEditor.clientAcceptText = uploadKeyEditor.form.client_accept.join(',')
    uploadKeyEditor.open = true
  }

  function onMimeInput(value: string) {
    uploadKeyEditor.mimeText = value
    uploadKeyEditor.form.allowed_mime_types = splitCommaValues(value)
  }

  function onClientAcceptInput(value: string) {
    uploadKeyEditor.clientAcceptText = value
    uploadKeyEditor.form.client_accept = splitCommaValues(value)
  }

  function buildUploadKeyBody(): UploadKeySaveRequest | null {
    const f = uploadKeyEditor.form
    const extraSchema = buildExtraSchemaFromEditor(uploadKeySchemaEditorRef.value, 'UploadKey ')
    if (extraSchema === null) return null
    const body: UploadKeySaveRequest = {
      bucket_id: f.bucket_id,
      key: f.key.trim(),
      name: f.name.trim(),
      upload_mode: f.upload_mode,
      is_frontend_visible: f.is_frontend_visible,
      visibility: f.visibility,
      status: f.status,
      allowed_mime_types: f.allowed_mime_types,
      client_accept: f.client_accept
    }
    if (f.path_template.trim()) body.path_template = f.path_template.trim()
    if (f.default_rule_key.trim()) body.default_rule_key = f.default_rule_key.trim()
    if (f.permission_key.trim()) body.permission_key = f.permission_key.trim()
    if (f.fallback_key.trim()) body.fallback_key = f.fallback_key.trim()
    if (Number.isFinite(f.max_size_bytes) && f.max_size_bytes > 0) {
      body.max_size_bytes = Number(f.max_size_bytes)
    }
    if (Number.isFinite(f.direct_size_threshold_bytes) && f.direct_size_threshold_bytes > 0) {
      body.direct_size_threshold_bytes = Number(f.direct_size_threshold_bytes)
    }
    if (extraSchema) body.extra_schema = extraSchema
    return body
  }

  async function submitUploadKey() {
    const body = buildUploadKeyBody()
    if (!body) return
    if (!body.bucket_id || !body.key || !body.name) {
      ElMessage.warning('所属存储桶、上传标识、名称必填')
      return
    }
    uploadKeyEditor.submitting = true
    try {
      if (uploadKeyEditor.editingId) {
        await fetchUpdateUploadKey(uploadKeyEditor.editingId, body)
        ElMessage.success('上传配置已更新')
      } else {
        await fetchCreateUploadKey(body)
        ElMessage.success('上传配置已创建')
      }
      uploadKeyEditor.open = false
      await loadUploadKeys()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存上传配置失败')
    } finally {
      uploadKeyEditor.submitting = false
    }
  }

  async function removeUploadKey(row: UploadKeySummary) {
    try {
      await fetchDeleteUploadKey(row.id)
      ElMessage.success('已删除')
      await loadUploadKeys()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除失败')
    }
  }

  async function loadUploadKeys() {
    uploadKey.loading = true
    try {
      const res = await fetchListUploadKeys(uploadKey.bucketFilter || undefined)
      uploadKey.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载上传配置列表失败')
    } finally {
      uploadKey.loading = false
    }
  }

  // ── Rule state ───────────────────────────────────────────────────────────

  const rule = reactive({
    drawerOpen: false,
    loading: false,
    parentUploadKeyId: '',
    parentUploadKeyLabel: '',
    records: [] as UploadKeyRuleSummary[]
  })
  const ruleEditor = reactive({
    open: false,
    submitting: false,
    editingId: '',
    ruleMimeText: '',
    clientAcceptText: '',
    form: {
      rule_key: '',
      name: '',
      sub_path: '',
      filename_strategy: 'uuid' as Exclude<
        UploadKeyRuleSaveRequest['filename_strategy'],
        undefined
      >,
      max_size_bytes: 0,
      allowed_mime_types: [] as string[],
      process_pipeline: [] as string[],
      mode_override: 'inherit' as Exclude<UploadKeyRuleSaveRequest['mode_override'], undefined>,
      visibility_override: 'inherit' as Exclude<
        UploadKeyRuleSaveRequest['visibility_override'],
        undefined
      >,
      client_accept: [] as string[],
      extra_schema: undefined as UploadKeyRuleSaveRequest['extra_schema'] | undefined,
      is_default: false,
      status: 'ready' as Exclude<UploadKeyRuleSaveRequest['status'], undefined>
    }
  })

  function resetRuleEditor() {
    ruleEditor.submitting = false
    ruleEditor.editingId = ''
    ruleEditor.ruleMimeText = ''
    ruleEditor.clientAcceptText = ''
    ruleEditor.form = {
      rule_key: '',
      name: '',
      sub_path: '',
      filename_strategy: 'uuid',
      max_size_bytes: 0,
      allowed_mime_types: [],
      process_pipeline: [],
      mode_override: 'inherit',
      visibility_override: 'inherit',
      client_accept: [],
      extra_schema: undefined,
      is_default: false,
      status: 'ready'
    }
  }

  function openRuleDrawer(row: UploadKeySummary) {
    rule.parentUploadKeyId = row.id
    rule.parentUploadKeyLabel = `${row.name}（${row.key}）`
    rule.records = []
    rule.drawerOpen = true
    loadRules()
  }

  function openRuleCreate() {
    resetRuleEditor()
    ruleEditor.open = true
  }

  function openRuleEdit(row: UploadKeyRuleSummary) {
    resetRuleEditor()
    ruleEditor.editingId = row.id
    ruleEditor.form.rule_key = row.rule_key
    ruleEditor.form.name = row.name
    ruleEditor.form.sub_path = row.sub_path || ''
    ruleEditor.form.filename_strategy = row.filename_strategy
    ruleEditor.form.max_size_bytes = Number(row.max_size_bytes ?? 0)
    ruleEditor.form.allowed_mime_types = Array.isArray(row.allowed_mime_types)
      ? row.allowed_mime_types
      : []
    ruleEditor.form.process_pipeline = Array.isArray(row.process_pipeline)
      ? row.process_pipeline
      : []
    ruleEditor.form.mode_override = row.mode_override || 'inherit'
    ruleEditor.form.visibility_override = row.visibility_override || 'inherit'
    ruleEditor.form.client_accept = Array.isArray(row.client_accept) ? row.client_accept : []
    ruleEditor.form.extra_schema = normalizeObjectValue(row.extra_schema)
    ruleEditor.form.is_default = !!row.is_default
    ruleEditor.form.status = row.status
    ruleEditor.ruleMimeText = ruleEditor.form.allowed_mime_types.join(',')
    ruleEditor.clientAcceptText = ruleEditor.form.client_accept.join(',')
    ruleEditor.open = true
  }

  function onRuleMimeInput(value: string) {
    ruleEditor.ruleMimeText = value
    ruleEditor.form.allowed_mime_types = splitCommaValues(value)
  }

  function onRuleClientAcceptInput(value: string) {
    ruleEditor.clientAcceptText = value
    ruleEditor.form.client_accept = splitCommaValues(value)
  }

  function buildRuleBody(): UploadKeyRuleSaveRequest | null {
    const f = ruleEditor.form
    const extraSchema = buildExtraSchemaFromEditor(ruleSchemaEditorRef.value, 'Rule ')
    if (extraSchema === null) return null
    const body: UploadKeyRuleSaveRequest = {
      rule_key: f.rule_key.trim(),
      name: f.name.trim(),
      filename_strategy: f.filename_strategy,
      mode_override: f.mode_override,
      visibility_override: f.visibility_override,
      is_default: f.is_default,
      status: f.status,
      allowed_mime_types: f.allowed_mime_types,
      client_accept: f.client_accept
    }
    if (f.sub_path.trim()) body.sub_path = f.sub_path.trim()
    if (Number.isFinite(f.max_size_bytes) && f.max_size_bytes > 0) {
      body.max_size_bytes = Number(f.max_size_bytes)
    }
    if (f.process_pipeline.length > 0) body.process_pipeline = f.process_pipeline
    if (extraSchema) body.extra_schema = extraSchema
    return body
  }

  async function submitRule() {
    const body = buildRuleBody()
    if (!body) return
    if (!body.rule_key || !body.name) {
      ElMessage.warning('规则标识和名称必填')
      return
    }
    ruleEditor.submitting = true
    try {
      if (ruleEditor.editingId) {
        await fetchUpdateUploadKeyRule(ruleEditor.editingId, body)
        ElMessage.success('上传规则已更新')
      } else {
        await fetchCreateUploadKeyRule(rule.parentUploadKeyId, body)
        ElMessage.success('上传规则已创建')
      }
      ruleEditor.open = false
      await loadRules()
    } catch (err: any) {
      ElMessage.error(err?.message || '保存规则失败')
    } finally {
      ruleEditor.submitting = false
    }
  }

  async function removeRule(row: UploadKeyRuleSummary) {
    try {
      await fetchDeleteUploadKeyRule(row.id)
      ElMessage.success('已删除')
      await loadRules()
    } catch (err: any) {
      ElMessage.error(err?.message || '删除规则失败')
    }
  }

  async function loadRules() {
    if (!rule.parentUploadKeyId) return
    rule.loading = true
    try {
      const res = await fetchListUploadKeyRules(rule.parentUploadKeyId)
      rule.records = res.records || []
    } catch (err: any) {
      ElMessage.error(err?.message || '加载规则列表失败')
    } finally {
      rule.loading = false
    }
  }

  const ruleColumns = computed<ColumnOption[]>(() => [
    { prop: 'rule_key', label: '规则标识', minWidth: 130 },
    { prop: 'name', label: '名称', minWidth: 130 },
    { prop: 'sub_path', label: '子路径', minWidth: 100 },
    {
      prop: 'filename_strategy',
      label: '文件名策略',
      width: 130,
      formatter: (row: UploadKeyRuleSummary) =>
        filenameStrategyLabel[row.filename_strategy] || row.filename_strategy
    },
    {
      prop: 'mode_override',
      label: '上传方式覆写',
      width: 130,
      formatter: (row: UploadKeyRuleSummary) =>
        uploadModeLabel[row.mode_override || 'inherit'] || row.mode_override || 'inherit'
    },
    {
      prop: 'visibility_override',
      label: '可见性覆写',
      width: 130,
      formatter: (row: UploadKeyRuleSummary) =>
        visibilityOverrideLabel[row.visibility_override || 'inherit'] ||
        row.visibility_override ||
        'inherit'
    },
    {
      prop: 'max_size_bytes',
      label: '文件上限',
      width: 110,
      formatter: (row: UploadKeyRuleSummary) => formatBytes(Number(row.max_size_bytes ?? 0))
    },
    {
      prop: 'allowed_mime_types',
      label: '允许类型',
      minWidth: 160,
      showOverflowTooltip: true,
      formatter: (row: UploadKeyRuleSummary) =>
        Array.isArray(row.allowed_mime_types) && row.allowed_mime_types.length
          ? row.allowed_mime_types.join(', ')
          : '不限'
    },
    {
      prop: 'extra_schema',
      label: '扩展参数',
      width: 100,
      formatter: (row: UploadKeyRuleSummary) => formatSchemaConfigured(row.extra_schema)
    },
    {
      prop: 'is_default',
      label: '默认',
      width: 70,
      formatter: (row: UploadKeyRuleSummary) =>
        row.is_default
          ? h(ElTag, { type: 'success', effect: 'plain', size: 'small' }, () => '是')
          : '-'
    },
    {
      prop: 'status',
      label: '状态',
      width: 80,
      formatter: (row: UploadKeyRuleSummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain', size: 'small' },
          () => statusLabel[row.status] || row.status
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 160,
      fixed: 'right',
      formatter: (row: UploadKeyRuleSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openRuleEdit(row) },
            () => '编辑'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该规则？', onConfirm: () => removeRule(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  // ── 列定义 ────────────────────────────────────────────────────────────────

  const providerColumns = computed<ColumnOption[]>(() => [
    { prop: 'provider_key', label: '服务标识', minWidth: 160 },
    { prop: 'name', label: '名称', minWidth: 160 },
    {
      prop: 'driver',
      label: '驱动类型',
      width: 130,
      formatter: (row: StorageProviderSummary) => driverLabel[row.driver] || row.driver
    },
    {
      prop: 'extra',
      label: '扩展能力',
      minWidth: 180,
      formatter: (row: StorageProviderSummary) => renderFeatureTags(buildProviderFeatureLabels(row))
    },
    { prop: 'endpoint', label: '接入点', minWidth: 200, showOverflowTooltip: true },
    { prop: 'access_key_masked', label: '访问密钥', width: 160 },
    {
      prop: 'is_default',
      label: '默认',
      width: 80,
      formatter: (row: StorageProviderSummary) =>
        row.is_default ? h(ElTag, { type: 'success', effect: 'plain' }, () => '默认') : '-'
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row: StorageProviderSummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain' },
          () => statusLabel[row.status] || row.status
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 280,
      fixed: 'right',
      formatter: (row: StorageProviderSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openProviderEdit(row) },
            () => '编辑'
          ),
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => testProvider(row) },
            () => '健康检查'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该存储服务？', onConfirm: () => removeProvider(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  const bucketColumns = computed<ColumnOption[]>(() => [
    { prop: 'bucket_key', label: '存储桶标识', minWidth: 160 },
    { prop: 'name', label: '名称', minWidth: 160 },
    { prop: 'provider_key', label: '所属服务', width: 160 },
    { prop: 'bucket_name', label: '存储桶名称', minWidth: 160 },
    { prop: 'base_path', label: '基础路径', minWidth: 140 },
    {
      prop: 'extra',
      label: '扩展能力',
      minWidth: 180,
      formatter: (row: StorageBucketSummary) => renderFeatureTags(buildBucketFeatureLabels(row))
    },
    {
      prop: 'is_public',
      label: '公开',
      width: 80,
      formatter: (row: StorageBucketSummary) =>
        row.is_public
          ? h(ElTag, { type: 'success', effect: 'plain' }, () => '公开')
          : h(ElTag, { type: 'info', effect: 'plain' }, () => '私有')
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row: StorageBucketSummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain' },
          () => statusLabel[row.status] || row.status
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 200,
      fixed: 'right',
      formatter: (row: StorageBucketSummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openBucketEdit(row) },
            () => '编辑'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该存储桶？', onConfirm: () => removeBucket(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  const uploadKeyColumns = computed<ColumnOption[]>(() => [
    { prop: 'key', label: '上传标识', minWidth: 160 },
    { prop: 'name', label: '名称', minWidth: 160 },
    { prop: 'bucket_key', label: '所属存储桶', width: 160 },
    {
      prop: 'upload_mode',
      label: '上传方式',
      width: 120,
      formatter: (row: UploadKeySummary) =>
        uploadModeLabel[row.upload_mode || 'auto'] || row.upload_mode || 'auto'
    },
    {
      prop: 'is_frontend_visible',
      label: '前端可见',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        row.is_frontend_visible
          ? h(ElTag, { type: 'success', effect: 'plain', size: 'small' }, () => '可见')
          : h(ElTag, { type: 'info', effect: 'plain', size: 'small' }, () => '隐藏')
    },
    {
      prop: 'permission_key',
      label: '权限键',
      minWidth: 160,
      showOverflowTooltip: true,
      formatter: (row: UploadKeySummary) => row.permission_key || '-'
    },
    {
      prop: 'visibility',
      label: '可见性',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        h(
          ElTag,
          { type: row.visibility === 'public' ? 'success' : 'info', effect: 'plain' },
          () => visibilityLabel[row.visibility] || row.visibility
        )
    },
    {
      prop: 'max_size_bytes',
      label: '文件上限',
      width: 140,
      formatter: (row: UploadKeySummary) => formatBytes(Number(row.max_size_bytes ?? 0))
    },
    {
      prop: 'allowed_mime_types',
      label: '允许类型',
      minWidth: 200,
      showOverflowTooltip: true,
      formatter: (row: UploadKeySummary) =>
        Array.isArray(row.allowed_mime_types) && row.allowed_mime_types.length
          ? row.allowed_mime_types.join(', ')
          : '不限'
    },
    {
      prop: 'extra_schema',
      label: '扩展参数',
      width: 100,
      formatter: (row: UploadKeySummary) => formatSchemaConfigured(row.extra_schema)
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row: UploadKeySummary) =>
        h(
          ElTag,
          { type: statusType[row.status] || 'info', effect: 'plain' },
          () => statusLabel[row.status] || row.status
        )
    },
    {
      prop: 'actions',
      label: '操作',
      width: 280,
      fixed: 'right',
      formatter: (row: UploadKeySummary) =>
        h('div', { class: 'config-row-actions' }, [
          h(
            ElButton,
            { type: 'primary', link: true, onClick: () => openUploadKeyEdit(row) },
            () => '编辑'
          ),
          h(
            ElButton,
            { type: 'warning', link: true, onClick: () => openRuleDrawer(row) },
            () => '管理规则'
          ),
          h(
            ElPopconfirm,
            { title: '确认删除该上传配置？', onConfirm: () => removeUploadKey(row) },
            { reference: () => h(ElButton, { type: 'danger', link: true }, () => '删除') }
          )
        ])
    }
  ])

  function formatBytes(value: number): string {
    if (!value || value <= 0) return '不限'
    const units = ['B', 'KB', 'MB', 'GB']
    let size = value
    let unit = 0
    while (size >= 1024 && unit < units.length - 1) {
      size /= 1024
      unit += 1
    }
    return `${size.toFixed(unit === 0 ? 0 : 2)} ${units[unit]}`
  }

  function onTabChange(name: string | number) {
    if (name === 'bucket' && bucket.records.length === 0) {
      loadBuckets()
    } else if (name === 'upload-key' && uploadKey.records.length === 0) {
      loadUploadKeys()
    }
  }

  onMounted(() => {
    loadProviders()
  })
</script>

<style scoped lang="scss">
  .upload-config-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .upload-config-main {
    flex: 1;
    min-height: 0;
  }

  .upload-config-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .upload-config-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0 12px;
  }

  .upload-config-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .upload-config-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .upload-config-tabs {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .upload-config-tabs :deep(.el-tabs__content) {
    flex: 1;
    min-height: 0;
  }

  .upload-config-tabs :deep(.el-tab-pane) {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .upload-config-filters {
    margin-bottom: 4px;
  }

  .config-row-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .form-tip {
    margin-left: 12px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .driver-guide-card {
    margin: 6px 0 14px;
    padding: 12px 14px;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
  }

  .driver-guide-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
  }

  .driver-guide-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .driver-guide-list {
    margin: 0;
    padding-left: 18px;
    color: var(--el-text-color-regular);
    line-height: 1.8;
  }

  .driver-collapse {
    margin-top: 6px;
  }

  .driver-section-desc {
    margin-bottom: 10px;
    font-size: 12px;
    line-height: 1.7;
    color: var(--el-text-color-secondary);
  }

  .config-guide-card {
    margin: 6px 0 14px;
    padding: 12px 14px;
    background: linear-gradient(135deg, var(--el-fill-color-light) 0%, var(--el-fill-color) 100%);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
  }

  .config-guide-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .config-guide-list {
    margin: 8px 0 0;
    padding-left: 18px;
    color: var(--el-text-color-regular);
    line-height: 1.8;
  }

  .config-section-title {
    margin: 12px 0 4px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .config-section-tip {
    margin-bottom: 10px;
    font-size: 12px;
    line-height: 1.7;
    color: var(--el-text-color-secondary);
  }

  .tab-desc {
    margin-bottom: 12px;
    padding: 8px 12px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-lighter);
    border-radius: 4px;
  }

  .rule-drawer-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  .config-inline-tags {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
  }
</style>
