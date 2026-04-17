<template>
  <ElDrawer
    v-model="visible"
    :title="dialogTitle"
    size="980px"
    direction="rtl"
    class="page-entry-drawer config-drawer"
    destroy-on-close
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <div class="dialog-intro">
        <div class="dialog-intro__line">
          <span class="dialog-intro__desc">{{ configHintDescription }}</span>
          <ElButton text type="primary" size="small" @click="showExamples = !showExamples">
            {{ showExamples ? '收起示例' : '查看示例' }}
          </ElButton>
        </div>
        <div v-if="isUnregisteredCandidate" class="dialog-intro__meta">
          <ElTag size="small" effect="plain" type="warning">未注册来源，组件路径固定</ElTag>
        </div>
        <div v-if="showExamples" class="dialog-intro__examples">
          <div v-for="item in pageExamples" :key="item" class="dialog-intro__example">{{
            item
          }}</div>
        </div>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">基础信息</div>
          </div>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="页面名称" prop="name">
              <template #label>
                <PageFieldLabel
                  label="页面名称"
                  help="给人看的名称，显示在页面管理、面包屑预览和关联选择里。"
                />
              </template>
              <ElInput v-model="form.name" placeholder="请输入页面名称" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="页面标识" prop="pageKey">
              <template #label>
                <PageFieldLabel
                  label="页面标识"
                  help="页面的稳定业务标识，用于父子页面关联、同步识别和配置引用。上线后尽量不要改。"
                />
              </template>
              <ElInput v-model="form.pageKey" placeholder="例如 store.management.detail" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="页面类型" prop="pageType">
              <template #label>
                <PageFieldLabel
                  label="页面类型"
                  help="内页必须继承菜单或上级页面；独立页可选择当前 App 全局可见，或只对指定空间开放。"
                />
              </template>
              <ElSelect v-model="form.pageType" style="width: 100%">
                <ElOption label="内页" value="inner" />
                <ElOption label="独立页" value="standalone" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="排序" prop="sortOrder">
              <template #label>
                <PageFieldLabel label="排序" help="同级页面或分组的排序值，数字越小越靠前。" />
              </template>
              <ElInputNumber v-model="form.sortOrder" :min="0" :step="1" style="width: 100%" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="模块标识" prop="moduleKey">
              <template #label>
                <PageFieldLabel
                  label="模块标识"
                  help="页面所属业务模块，用于筛选、归类和后续批量管理，例如 system、dashboard、order。"
                />
              </template>
              <ElInput v-model="form.moduleKey" placeholder="例如 system / order" />
            </ElFormItem>
          </ElCol>
          <ElCol v-if="showVisibilityScopeField" :span="12">
            <ElFormItem label="可见范围" prop="visibilityScope">
              <template #label>
                <PageFieldLabel
                  label="可见范围"
                  help="全局页与独立页可在当前 App 下全局可见，或只在指定菜单空间开放；内页不单独配置这里。"
                />
              </template>
              <ElSelect v-model="form.visibilityScope" style="width: 100%">
                <ElOption label="当前 App 全局可见（所有空间都能访问）" value="app" />
                <ElOption label="仅指定空间可见（只在选中的空间暴露）" value="spaces" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="showSpaceBindingField" :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="开放空间" prop="menuSpaceKeys">
              <template #label>
                <PageFieldLabel
                  label="开放空间"
                  help="仅在选中的菜单空间暴露该独立页；不再使用“全空间可见”占位值。"
                />
              </template>
              <ElSelect
                v-model="form.menuSpaceKeys"
                multiple
                collapse-tags
                collapse-tags-tooltip
                clearable
                filterable
                style="width: 100%"
              >
                <ElOption
                  v-for="item in menuSpaceOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="路由名称" prop="routeName">
              <template #label>
                <PageFieldLabel
                  label="路由名称"
                  help="Vue Router 内部路由名。可留空，留空时默认使用页面标识。"
                />
              </template>
              <ElInput
                v-model="form.routeName"
                placeholder="例如 StoreManagementDetail；留空时默认使用页面标识"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">路由与渲染</div>
          </div>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="路由路径" prop="routePath">
              <template #label>
                <PageFieldLabel
                  label="路由路径"
                  help="单段路径会按上级菜单或上级页面自动拼接；多段绝对路径会按完整路径注册。"
                />
              </template>
              <ElInput v-model="form.routePath" :placeholder="routePathPlaceholder" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="组件路径" prop="component">
              <template #label>
                <PageFieldLabel
                  label="组件路径"
                  help="实际渲染的前端页面组件路径。外链模式下会自动改为 /outside/Iframe。"
                />
              </template>
              <ElInput
                v-model="form.component"
                :disabled="form.isIframe || isComponentLocked"
                :placeholder="getComponentPlaceholder()"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="是否外链" prop="isIframe">
              <template #label>
                <PageFieldLabel
                  label="是否外链"
                  help="开启后页面将通过 iframe 加载外部地址，组件路径自动切为 /outside/Iframe。"
                />
              </template>
              <ElSwitch v-model="form.isIframe" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="状态" prop="status">
              <template #label>
                <PageFieldLabel
                  label="状态"
                  help="正常状态才会参与运行时注册；停用后页面保留数据，但不会被动态加载。"
                />
              </template>
              <div class="inline-flex items-center gap-2">
                <ElSwitch v-model="form.status" active-value="normal" inactive-value="suspended" />
                <ElTag :type="form.status === 'normal' ? 'success' : 'danger'" effect="plain">
                  {{ form.status === 'normal' ? '正常' : '停用' }}
                </ElTag>
              </div>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.isIframe" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="外链地址" prop="link">
              <template #label>
                <PageFieldLabel
                  label="外链地址"
                  help="外链模式下必填，填写要加载的 http:// 或 https:// 地址。"
                />
              </template>
              <ElInput v-model="form.link" placeholder="例如 https://example.com" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElFormItem label="最终路径" class="final-path-item">
          <template #label>
            <PageFieldLabel
              label="最终路径"
              help="系统按规则推导真实访问路径：不挂载时=路由路径本身；挂到菜单时，单段路径会拼到菜单路径后；挂到页面时，单段路径会拼到上级页面路径后；填写多段绝对路径则按原样注册。"
            />
          </template>
          <div class="route-preview-box">
            <code>{{ resolvedRoutePreview || '-' }}</code>
          </div>
        </ElFormItem>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">挂载与归属</div>
          </div>
        </div>

        <ElRow v-if="showMountSection" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="挂载方式" prop="mountMode">
              <template #label>
                <PageFieldLabel
                  label="挂载方式"
                  help="决定当前页面是独立存在，还是归属某个菜单，或归属到另一个页面/分组下面。"
                />
              </template>
              <ElRadioGroup v-model="mountMode" class="mount-mode-group">
                <ElRadioButton label="none">不挂载 · 独立路径</ElRadioButton>
                <ElRadioButton label="menu">挂到菜单 · 出现在左侧菜单</ElRadioButton>
                <ElRadioButton label="page">挂到页面 · 作为父页面子路由</ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>
            <div v-if="mountOwnershipSummary" class="mount-hint">{{ mountOwnershipSummary }}</div>
          </ElCol>
        </ElRow>

        <ElRow v-if="showMountSection && mountMode === 'menu'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="菜单空间">
              <template #label>
                <PageFieldLabel
                  label="菜单空间"
                  help="按菜单空间过滤下方上级菜单候选。留空表示不限空间（可能跨空间显示重复项）。"
                />
              </template>
              <ElSelect
                v-model="mountSpaceKey"
                clearable
                filterable
                style="width: 100%"
                placeholder="可选，按空间过滤菜单"
              >
                <ElOption
                  v-for="item in menuSpaceOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="上级菜单" prop="parentMenuId">
              <template #label>
                <PageFieldLabel
                  label="上级菜单"
                  help="页面直接归属的菜单。单段路由会自动拼到该菜单路径后，并继承菜单高亮与菜单准入。若页面再单独配置权限，则最终按菜单权限与页面权限交集放行。"
                />
              </template>
              <ElCascader
                v-model="form.parentMenuId"
                :options="menuTreeOptions"
                :props="menuCascaderProps"
                filterable
                clearable
                show-all-levels
                style="width: 100%"
                placeholder="请选择上级菜单"
              />
            </ElFormItem>
            <div v-if="mountMenuSummary" class="mount-hint">{{ mountMenuSummary }}</div>
          </ElCol>
        </ElRow>

        <ElRow v-if="showMountSection && mountMode === 'page'" :gutter="14">
          <ElCol :span="24">
            <ElFormItem label="上级页面" prop="parentPageKey">
              <template #label>
                <PageFieldLabel
                  label="上级页面"
                  help="页面直接归属的父页面或逻辑分组。选择后会优先继承其访问路径、菜单链和默认面包屑。"
                />
              </template>
              <ElSelect
                v-model="form.parentPageKey"
                clearable
                filterable
                style="width: 100%"
                placeholder="请选择上级页面或逻辑分组"
              >
                <ElOption
                  v-for="item in parentPageOptions"
                  :key="item.pageKey"
                  :label="`${item.name} (${item.pageKey})`"
                  :value="item.pageKey"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow
          v-if="
            form.pageType !== 'standalone' && mountMode !== 'page'
          "
          :gutter="14"
        >
          <ElCol :span="24">
            <ElFormItem label="普通分组" prop="displayGroupKey">
              <template #label>
                <PageFieldLabel
                  label="普通分组"
                  help="仅用于页面管理列表归类，不影响页面的菜单挂载、路径、权限和面包屑继承。"
                />
              </template>
              <ElSelect
                v-model="form.displayGroupKey"
                clearable
                filterable
                style="width: 100%"
                placeholder="可选，选择普通分组"
              >
                <ElOption
                  v-for="item in displayGroupOptions"
                  :key="item.pageKey"
                  :label="item.name"
                  :value="item.pageKey"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>
      </div>

      <div class="form-section">
        <div class="form-section__header">
          <div>
            <div class="form-section__title">访问与行为</div>
          </div>
          <ElButton
            text
            type="primary"
            v-if="showMountSection"
            @click="showAdvanced = !showAdvanced"
          >
            {{ showAdvanced ? '收起高级配置' : '展开高级配置' }}
          </ElButton>
        </div>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="访问模式" prop="accessMode">
              <template #label>
                <PageFieldLabel
                  label="访问模式"
                  help="继承表示跟随上级菜单或页面；登录表示只验登录；权限表示还需校验权限键。挂到菜单时，继承即默认跟菜单权限走；若改成权限模式，则在菜单准入基础上再校验页面权限。"
                />
              </template>
              <ElSelect v-model="form.accessMode" style="width: 100%">
                <ElOption
                  v-for="item in accessModeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="权限键" prop="permissionKey">
              <template #label>
                <PageFieldLabel
                  label="权限键"
                  help="仅在访问模式为权限时生效。挂到菜单时，这里不是覆盖菜单权限，而是在菜单准入基础上追加页面权限校验。"
                />
              </template>
              <ElInput
                v-model="form.permissionKey"
                :disabled="form.accessMode !== 'permission'"
                placeholder="accessMode=permission 时必填"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="14">
          <ElCol :span="12">
            <ElFormItem label="缓存页面" prop="keepAlive">
              <template #label>
                <PageFieldLabel
                  label="缓存页面"
                  help="开启后页面会进入 keep-alive 缓存，适合表单或列表类页面；内嵌页通常不缓存。"
                />
              </template>
              <ElSwitch v-model="form.keepAlive" :disabled="form.isIframe" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="全屏页面" prop="isFullPage">
              <template #label>
                <PageFieldLabel
                  label="全屏页面"
                  help="开启后页面按全屏模式展示，通常用于沉浸式页面或不依赖常规布局的场景。"
                />
              </template>
              <ElSwitch v-model="form.isFullPage" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="showAdvanced && showMountSection" :gutter="14" class="advanced-grid">
          <ElCol :span="12">
            <ElFormItem label="面包屑模式" prop="breadcrumbMode">
              <template #label>
                <PageFieldLabel
                  label="面包屑模式"
                  help="继承菜单表示按菜单链展示；继承页面表示把父页面链也带上；自定义用于高级覆盖。"
                />
              </template>
              <ElSelect v-model="form.breadcrumbMode" style="width: 100%">
                <ElOption label="继承菜单" value="inherit_menu" />
                <ElOption label="继承页面" value="inherit_page" />
                <ElOption label="自定义" value="custom" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem label="高亮菜单路径" prop="activeMenuPath">
              <template #label>
                <PageFieldLabel
                  label="高亮菜单路径"
                  help="仅在自动推导不满足时手工覆盖菜单高亮路径。大多数页面可留空。"
                />
              </template>
              <ElInput v-model="form.activeMenuPath" placeholder="可选，例如 /system/page" />
            </ElFormItem>
          </ElCol>
        </ElRow>
      </div>
    </ElForm>

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="handleClose">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
      </div>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  // 视图脚本：所有 reactive state、computed、watch、handler 集中在 usePageEntryDialog。
  // 这里只做：1) 引入子组件；2) 调用 composable；3) 把返回值拉到 setup 作用域供模板访问。
  import PageFieldLabel from './page-field-label.vue'
  import { usePageEntryDialog } from './use-page-entry-dialog'

  type PageItem = Api.SystemManage.PageItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit' | 'copy'
    pageData?: Partial<PageItem>
    appKey?: string
    menuSpaces?: Api.SystemManage.MenuSpaceItem[]
    initialParentPageKey?: string
    initialParentMenuId?: string
    initialPageType?: PageItem['pageType']
    defaultData?: Partial<PageItem>
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    pageData: undefined,
    menuSpaces: () => [],
    initialParentPageKey: '',
    initialParentMenuId: '',
    initialPageType: 'standalone',
    defaultData: undefined
  })

  const emit = defineEmits<Emits>()

  const {
    formRef,
    submitting,
    mountMode,
    mountSpaceKey,
    showAdvanced,
    showExamples,
    visible,
    dialogTitle,
    form,
    rules,
    menuTreeOptions,
    menuSpaceOptions,
    showMountSection,
    showVisibilityScopeField,
    showSpaceBindingField,
    menuCascaderProps,
    parentPageOptions,
    displayGroupOptions,
    isUnregisteredCandidate,
    isComponentLocked,
    configHintDescription,
    mountOwnershipSummary,
    accessModeOptions,
    routePathPlaceholder,
    mountMenuSummary,
    resolvedRoutePreview,
    pageExamples,
    getComponentPlaceholder,
    handleClose,
    handleSubmit
  } = usePageEntryDialog(props, emit)
</script>

<style scoped lang="scss">
  .dialog-intro {
    background: linear-gradient(
      180deg,
      var(--el-fill-color-light) 0%,
      color-mix(in srgb, var(--el-fill-color-light) 72%, white) 100%
    );
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 14px;
    margin-bottom: 18px;
    padding: 14px 16px;
  }

  .dialog-intro__line {
    display: flex;
    align-items: center;
    gap: 12px;
    justify-content: space-between;
  }

  .dialog-intro__desc {
    color: var(--el-text-color-secondary);
    flex: 1;
    font-size: 12px;
    line-height: 1.6;
  }

  .dialog-intro__meta {
    margin-top: 8px;
  }

  .mount-hint {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.6;
    margin: -4px 0 12px;
  }

  .dialog-intro__examples {
    border-top: 1px dashed var(--el-border-color-lighter);
    margin-top: 8px;
    padding-top: 12px;
  }

  .dialog-intro__example {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.7;
  }

  .form-section {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 14px;
    margin-bottom: 16px;
    padding: 16px 16px 8px;
  }

  .form-section__header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  .form-section__title {
    color: var(--el-text-color-primary);
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .mount-mode-group {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .route-preview-box {
    align-items: center;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    color: var(--el-text-color-primary);
    display: flex;
    min-height: 40px;
    padding: 0 12px;
    width: 100%;
  }

  .route-preview-box code {
    color: inherit;
    font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
    font-size: 12px;
    word-break: break-all;
  }

  .advanced-grid {
    border-top: 1px dashed var(--el-border-color-lighter);
    margin-top: 4px;
    padding-top: 12px;
  }

  .drawer-footer {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  :deep(.el-drawer__body) {
    max-height: calc(100vh - 126px);
    overflow-y: auto;
    padding: 14px 20px 12px;
  }

  :deep(.el-drawer__footer) {
    border-top: 1px solid var(--el-border-color-lighter);
    padding: 14px 20px 18px;
  }

  :deep(.final-path-item .el-form-item__content) {
    align-items: stretch;
  }

  :deep(.mount-mode-group .el-radio-button__inner) {
    min-width: 96px;
    justify-content: flex-end;
  }
</style>
