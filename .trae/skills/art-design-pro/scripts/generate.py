#!/usr/bin/env python3
"""
Art Design Pro 代码生成器

自动生成标准的 Vue3 页面代码，包括：
- CRUD 列表页（带搜索、弹窗、操作）
- 基础表格页（简单数据展示）
- 仪表板页面（数据统计和图表）

使用示例：
    # 生成 CRUD 列表页
    python3 generate.py crud --name "User" --path "system/user" --fields "username,email,status"

    # 生成基础表格页
    python3 generate.py table --name "Product" --path "product/list" --fields "name,price,stock"

    # 生成仪表板页面
    python3 generate.py dashboard --name "Analytics" --path "dashboard/analytics" --charts "line,bar"
"""

import argparse
import sys
from typing import List, Dict


# ====================================
# 模板定义
# ====================================

CRUD_MAIN_TEMPLATE = """<!-- {name_pascal}管理页面 -->
<!-- 使用 Art Design Pro 组件库快速构建 CRUD 页面 -->
<template>
  <div class="{name_lower}-page art-full-height">
    <!-- 搜索栏 -->
    <{name_pascal}Search v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></{name_pascal}Search>

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" @click="showDialog('add')" v-ripple>新增{name_pascal}</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- {name_pascal}弹窗 -->
      <{name_pascal}Dialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :data="currentData"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import {{ useTable }} from '@/hooks/core/useTable'
  import {name_camel}Search from './modules/{name_lower}-search.vue'
  import {name_pascal}Dialog from './modules/{name_lower}-dialog.vue'
  import {{ DialogType }} from '@/types'

  defineOptions({{ name: '{name_pascal}' }})

  type {name_pascal}Item = Api.YourModule.{name_pascal}Item

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentData = ref<Partial<{name_pascal}Item>>({{}})

  // 搜索表单
  const searchForm = ref({{
{search_fields}
  }})

  const {{
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  }} = useTable({{
    core: {{
      apiFn: yourApiFunction, // TODO: 替换为实际的 API 函数
      apiParams: {{
        current: 1,
        size: 20,
        ...searchForm.value
      }},
      columnsFactory: () => [
{table_columns}
      ]
    }}
  }})

  /**
   * 搜索处理
   */
  const handleSearch = (params: Record<string, any>) => {{
    Object.assign(searchParams, params)
    getData()
  }}

  /**
   * 显示弹窗
   */
  const showDialog = (type: DialogType, row?: {name_pascal}Item): void => {{
    dialogType.value = type
    currentData.value = row || {{}}
    nextTick(() => {{
      dialogVisible.value = true
    }})
  }}

  /**
   * 处理弹窗提交
   */
  const handleDialogSubmit = async () => {{
    try {{
      dialogVisible.value = false
      currentData.value = {{}}
      refreshData()
    }} catch (error) {{
      console.error('提交失败:', error)
    }}
  }}
</script>
"""

CRUD_SEARCH_TEMPLATE = """<!-- {name_pascal}搜索栏 -->
<template>
  <ArtSearchBar @search="handleSearch" @reset="handleReset">
    <ElForm :model="searchForm" label-width="80px">
{form_items}
    </ElForm>
  </ArtSearchBar>
</template>

<script setup lang="ts">
  import ArtSearchBar from '@/components/core/forms/art-search-bar/index.vue'

  defineOptions({{ name: '{name_pascal}Search' }})

  interface SearchForm {{
{interface_fields}
  }}}

  const searchForm = defineModel<SearchForm>({{
    default: () => ({{
{search_field_defaults}
    }})
  }})

  const emit = defineEmits<{{
    search: [params: SearchForm]
    reset: []
  }}>()

  const handleSearch = () => {{
    emit('search', searchForm.value)
  }}

  const handleReset = () => {{
    emit('reset')
  }}
</script>
"""

CRUD_DIALOG_TEMPLATE = """<!-- {name_pascal}弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="600px"
    :before-close="handleClose"
  >
    <ElForm :model="formData" :rules="rules" ref="formRef" label-width="100px">
{form_items}
    </ElForm>

    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit" :loading="loading">确定</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import {{ DialogType }} from '@/types'

  defineOptions({{ name: '{name_pascal}Dialog' }})

  interface {name_pascal}Data {{
{interface_fields}
  }}}

  interface Props {{
    visible: boolean
    type: DialogType
    data?: Partial<{name_pascal}Data>
  }}

  const props = withDefaults(defineProps<Props>(), {{
    visible: false,
    type: 'add',
    data: () => ({{}})
  }})

  const emit = defineEmits<{{
    'update:visible': [value: boolean]
    submit: []
  }}>()

  const formRef = ref()
  const loading = ref(false)

  const dialogVisible = computed({{
    get: () => props.visible,
    set: (val) => emit('update:visible', val)
  }})

  const dialogTitle = computed(() => {{
    return props.type === 'add' ? '新增{name_pascal}' : '编辑{name_pascal}'
  }})

  const formData = ref<{name_pascal}Data>({{
{dialog_field_defaults}
  }})

  const rules = {{
    // TODO: 添加校验规则
  }}

  // 监听数据变化
  watch(() => props.data, (newData) => {{
    if (newData && Object.keys(newData).length > 0) {{
      Object.assign(formData.value, newData)
    }}
  }}, {{ immediate: true }})

  const handleClose = () => {{
    formRef.value?.resetFields()
    dialogVisible.value = false
  }}

  const handleSubmit = async () => {{
    try {{
      await formRef.value?.validate()
      loading.value = true
      // TODO: 调用 API 提交数据
      emit('submit')
    }} catch (error) {{
      console.error('验证失败:', error)
    }} finally {{
      loading.value = false
    }}
  }}
</script>
"""

TABLE_TEMPLATE = """<!-- {name_pascal}列表 -->
<template>
  <div class="{name_lower}-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import {{ useTable }} from '@/hooks/core/useTable'

  defineOptions({{ name: '{name_pascal}' }})

  type {name_pascal}Item = Api.YourModule.{name_pascal}Item

  const {{
    data,
    columns,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange
  }} = useTable({{
    core: {{
      apiFn: yourApiFunction, // TODO: 替换为实际的 API 函数
      apiParams: {{
        current: 1,
        size: 20
      }},
      columnsFactory: () => [
{table_columns}
      ]
    }}
  }})
</script>
"""

DASHBOARD_TEMPLATE = """<!-- {name_pascal}仪表板 -->
<template>
  <div class="{name_lower}-dashboard">
    <!-- 统计卡片 -->
    <ElRow :gutter="20">
      <ElCol :xs="24" :sm="12" :md="6" v-for="(stat, index) in stats" :key="index">
        <ArtStatsCard
          :title="stat.title"
          :value="stat.value"
          :icon="stat.icon"
          :trend="stat.trend"
          :trend-value="stat.trendValue"
        />
      </ElCol>
    </ElRow>

    <!-- 图表区域 -->
    <ElRow :gutter="20" style="margin-top: 20px">
{chart_sections}
    </ElRow>
  </div>
</template>

<script setup lang="ts">
  defineOptions({{ name: '{name_pascal}' }})

  // 统计卡片数据
  const stats = ref([
    {{ title: '总用户数', value: 1234, icon: 'user', trend: 'up', trendValue: '12%' }},
    {{ title: '活跃用户', value: 567, icon: 'user-active', trend: 'up', trendValue: '8%' }},
    {{ title: '新增用户', value: 89, icon: 'user-add', trend: 'down', trendValue: '3%' }},
    {{ title: '转化率', value: '23%', icon: 'chart', trend: 'up', trendValue: '5%' }}
  ])

  // TODO: 获取实际数据
  const fetchDashboardData = async () => {{
    // 调用 API 获取数据
  }}

  onMounted(() => {{
    fetchDashboardData()
  }})
</script>

<style scoped lang="scss">
  .{name_lower}-dashboard {{
    padding: 20px;
  }}
</style>
"""


# ====================================
# 生成器类
# ====================================

class CodeGenerator:
    """代码生成器基类"""

    def __init__(self, args):
        self.args = args
        self.name = args.name
        self.path = args.path
        self.fields = self._parse_fields(args.fields) if hasattr(args, 'fields') and args.fields else []

    def _parse_fields(self, fields_str: str) -> List[Dict[str, str]]:
        """解析字段参数"""
        if not fields_str:
            return []

        fields = []
        for field in fields_str.split(','):
            field = field.strip()
            if ':' in field:
                name, ftype = field.split(':', 1)
                fields.append({'name': name.strip(), 'type': ftype.strip()})
            else:
                fields.append({'name': field, 'type': 'string'})
        return fields

    def _to_camel_case(self, snake_str: str) -> str:
        """转换为驼峰命名"""
        components = snake_str.split('_')
        return components[0] + ''.join(x.title() for x in components[1:])

    def _to_pascal_case(self, snake_str: str) -> str:
        """转换为帕斯卡命名（首字母大写）"""
        components = snake_str.split('_')
        return ''.join(x.title() for x in components)

    def _generate_field_label(self, field: Dict[str, str]) -> str:
        """生成字段标签"""
        name = field['name']
        labels = {
            'username': '用户名',
            'email': '邮箱',
            'phone': '手机号',
            'status': '状态',
            'name': '名称',
            'created_at': '创建时间',
            'updated_at': '更新时间',
            'price': '价格',
            'stock': '库存',
            'description': '描述'
        }
        return labels.get(name, self._to_pascal_case(name))

    def generate(self):
        """生成代码（子类实现）"""
        raise NotImplementedError


class CrudGenerator(CodeGenerator):
    """CRUD 列表页生成器"""

    def generate(self):
        """生成 CRUD 列表页代码"""
        name_lower = self.name.lower()
        name_camel = self._to_camel_case(name_lower)
        name_pascal = self._to_pascal_case(self.name)

        # 生成各部分代码
        search_fields = self._generate_search_fields()
        table_columns = self._generate_table_columns()

        # 主页面
        main_code = CRUD_MAIN_TEMPLATE.format(
            name_pascal=name_pascal,
            name_lower=name_lower,
            name_camel=name_camel,
            search_fields=search_fields,
            table_columns=table_columns
        )

        # 搜索栏组件
        form_items = self._generate_form_items()
        interface_fields = self._generate_interface_fields()
        search_field_defaults = self._generate_search_fields()

        search_code = CRUD_SEARCH_TEMPLATE.format(
            name_pascal=name_pascal,
            form_items=form_items,
            interface_fields=interface_fields,
            search_field_defaults=search_field_defaults
        )

        # 弹窗组件
        dialog_field_defaults = self._generate_dialog_fields()

        dialog_code = CRUD_DIALOG_TEMPLATE.format(
            name_pascal=name_pascal,
            form_items=form_items,
            interface_fields=interface_fields,
            dialog_field_defaults=dialog_field_defaults
        )

        return {
            'main': main_code,
            'search': search_code,
            'dialog': dialog_code
        }

    def _generate_search_fields(self) -> str:
        """生成搜索表单字段"""
        if not self.fields:
            return '    // 添加搜索字段'

        lines = []
        for field in self.fields[:5]:
            name = field['name']
            lines.append(f"    {name}: undefined,")
        return '\n'.join(lines)

    def _generate_table_columns(self) -> str:
        """生成表格列配置"""
        lines = []
        lines.append("        { type: 'selection' },")
        lines.append("        { type: 'index', width: 60, label: '序号' },")

        for field in self.fields[:8]:
            name = field['name']
            label = self._generate_field_label(field)
            lines.append(f"        {{ prop: '{name}', label: '{label}' }},")

        # 操作列
        lines.append("""        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row) =>
            h('div', [
              h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }),
              h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deleteItem(row)
              })
            ])
        }""")

        return '\n'.join(lines)

    def _generate_form_items(self) -> str:
        """生成表单项"""
        lines = []
        for field in self.fields[:5]:
            name = field['name']
            label = self._generate_field_label(field)
            lines.append(f'        <ElFormItem label="{label}" prop="{name}">')
            lines.append(f'          <ElInput v-model="searchForm.{name}" placeholder="请输入{label}" clearable />')
            lines.append('        </ElFormItem>')

        return '\n'.join(lines)

    def _generate_interface_fields(self) -> str:
        """生成接口字段"""
        if not self.fields:
            return '    // 定义字段类型'

        lines = []
        for field in self.fields[:5]:
            name = field['name']
            ftype = field.get('type', 'string')
            lines.append(f"    {name}?: {ftype}")
        return '\n'.join(lines)

    def _generate_dialog_fields(self) -> str:
        """生成弹窗表单字段"""
        if not self.fields:
            return '    // 定义字段'

        lines = []
        for field in self.fields:
            name = field['name']
            lines.append(f"    {name}: '',")
        return '\n'.join(lines)


class TableGenerator(CodeGenerator):
    """基础表格页生成器"""

    def generate(self):
        """生成基础表格页代码"""
        name_lower = self.name.lower()
        name_pascal = self._to_pascal_case(self.name)

        table_columns = self._generate_columns()

        code = TABLE_TEMPLATE.format(
            name_pascal=name_pascal,
            name_lower=name_lower,
            table_columns=table_columns
        )

        return {'main': code}

    def _generate_columns(self) -> str:
        """生成表格列"""
        if not self.fields:
            return "        { prop: 'id', label: 'ID' }\n        // 添加更多列"

        lines = []
        for field in self.fields[:10]:
            name = field['name']
            label = self._generate_field_label(field)
            lines.append(f"        {{ prop: '{name}', label: '{label}' }},")

        return '\n'.join(lines)


class DashboardGenerator(CodeGenerator):
    """仪表板页面生成器"""

    def generate(self):
        """生成仪表板页代码"""
        name_lower = self.name.lower()
        name_pascal = self._to_pascal_case(self.name)

        charts = self.args.charts.split(',') if hasattr(self.args, 'charts') and self.args.charts else ['line']
        chart_sections = self._generate_chart_sections(charts)

        code = DASHBOARD_TEMPLATE.format(
            name_pascal=name_pascal,
            name_lower=name_lower,
            chart_sections=chart_sections
        )

        return {'main': code}

    def _generate_chart_sections(self, charts: List[str]) -> str:
        """生成图表区块"""
        sections = []

        chart_types = {
            'line': ('ArtLineChart', '折线图'),
            'bar': ('ArtBarChart', '柱状图'),
            'pie': ('ArtRingChart', '饼图'),
            'radar': ('ArtRadarChart', '雷达图')
        }

        for chart in charts[:4]:
            chart = chart.strip()
            if chart in chart_types:
                component, title = chart_types[chart]
                sections.append(f'      <ElCol :xs="24" :sm="24" :md="12" :lg="12" :xl="12">')
                sections.append('        <ElCard shadow="never">')
                sections.append('          <template #header>')
                sections.append(f'            <span>{title}</span>')
                sections.append('          </template>')
                sections.append(f'          <{component} :data="chartData" height="300px" />')
                sections.append('        </ElCard>')
                sections.append('      </ElCol>')

        return '\n'.join(sections) if sections else '      <!-- 添加图表 -->'


# ====================================
# 主函数
# ====================================

def main():
    """主函数"""
    parser = argparse.ArgumentParser(
        description='Art Design Pro 代码生成器',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog='''
使用示例：
  # 生成 CRUD 列表页
  python3 generate.py crud --name "User" --path "system/user" --fields "username,email,status"

  # 生成基础表格页
  python3 generate.py table --name "Product" --path "product/list" --fields "name,price,stock"

  # 生成仪表板页
  python3 generate.py dashboard --name "Analytics" --path "dashboard/analytics" --charts "line,bar"
        '''
    )

    subparsers = parser.add_subparsers(dest='type', help='页面类型')

    # CRUD 页面生成器
    crud_parser = subparsers.add_parser('crud', help='生成 CRUD 列表页')
    crud_parser.add_argument('--name', required=True, help='实体名称（如 User、Product）')
    crud_parser.add_argument('--path', required=True, help='页面路径（如 system/user）')
    crud_parser.add_argument('--fields', required=True, help='字段列表（逗号分隔，如 username,email,status）')

    # 表格页面生成器
    table_parser = subparsers.add_parser('table', help='生成基础表格页')
    table_parser.add_argument('--name', required=True, help='实体名称')
    table_parser.add_argument('--path', required=True, help='页面路径')
    table_parser.add_argument('--fields', required=True, help='字段列表')

    # 仪表板页面生成器
    dashboard_parser = subparsers.add_parser('dashboard', help='生成仪表板页')
    dashboard_parser.add_argument('--name', required=True, help='仪表板名称')
    dashboard_parser.add_argument('--path', required=True, help='页面路径')
    dashboard_parser.add_argument('--charts', help='图表类型（逗号分隔：line,bar,pie,radar）')

    args = parser.parse_args()

    if not args.type:
        parser.print_help()
        sys.exit(1)

    # 创建生成器
    generators = {
        'crud': CrudGenerator,
        'table': TableGenerator,
        'dashboard': DashboardGenerator
    }

    generator_class = generators.get(args.type)
    if not generator_class:
        print(f'❌ 错误：不支持的页面类型 "{args.type}"')
        sys.exit(1)

    generator = generator_class(args)
    result = generator.generate()

    # 输出生成的代码
    if args.type == 'crud':
        print('=' * 80)
        print('📄 主页面 (index.vue):')
        print('=' * 80)
        print(result['main'])
        print('\n')
        print('=' * 80)
        print('📄 搜索栏组件 (modules/{}-search.vue):'.format(generator.name.lower()))
        print('=' * 80)
        print(result['search'])
        print('\n')
        print('=' * 80)
        print('📄 弹窗组件 (modules/{}-dialog.vue):'.format(generator.name.lower()))
        print('=' * 80)
        print(result['dialog'])
    else:
        print('=' * 80)
        print(f'📄 生成的 {args.type} 页面:')
        print('=' * 80)
        print(result['main'])

    print('\n✅ 代码生成完成！')
    print(f'💡 提示：请将生成的代码复制到对应目录，并根据实际需求调整')
    print(f'📂 目标目录: frontend/src/views/{args.path}/')


if __name__ == '__main__':
    main()
