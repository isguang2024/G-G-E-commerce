<template>
  <div class="permission-tree-panel">
    <slot name="control" />

    <ElEmpty v-if="!loading && treeData.length === 0" :description="emptyDescription" />

    <div v-else class="tree-wrapper">
      <ElTree
        ref="treeRef"
        :data="treeData"
        node-key="key"
        :props="treeProps"
        :default-expanded-keys="innerExpandedKeys"
        :expand-on-click-node="true"
        :highlight-current="false"
        class="permission-tree"
      >
        <template #default="{ data }">
          <slot name="node" :data="data" />
        </template>
      </ElTree>
    </div>
  </div>
</template>

<script setup lang="ts">
  interface PermissionTreeNode {
    key: string
    children?: PermissionTreeNode[]
  }

  interface Props {
    loading?: boolean
    treeData: PermissionTreeNode[]
    emptyDescription: string
  }

  const props = defineProps<Props>()

  const treeRef = ref()
  const innerExpandedKeys = ref<string[]>([])
  const treeProps = {
    children: 'children',
    label: 'label'
  }

  function collectExpandableKeys(nodes: PermissionTreeNode[]): string[] {
    return nodes.flatMap((node) => {
      if (!node.children?.length) {
        return []
      }
      return [node.key, ...collectExpandableKeys(node.children)]
    })
  }

  function setExpandedKeys(keys: string[]) {
    innerExpandedKeys.value = [...keys]
    nextTick(() => {
      const nodeMap = treeRef.value?.store?.nodesMap || {}
      Object.values(nodeMap).forEach((node: any) => {
        if (node?.level > 0) {
          node.expanded = false
        }
      })
      keys.forEach((key) => {
        const node = nodeMap[key]
        if (node) {
          node.expanded = true
        }
      })
    })
  }

  function expandAll() {
    setExpandedKeys(collectExpandableKeys(props.treeData))
  }

  function collapseAll() {
    setExpandedKeys([])
  }

  defineExpose({
    setExpandedKeys,
    expandAll,
    collapseAll
  })
</script>

<style scoped>
  .tree-wrapper {
    border: 1px solid #e5ebf3;
    border-radius: 16px;
    background: linear-gradient(
      180deg,
      rgba(255, 255, 255, 0.98) 0%,
      rgba(249, 251, 254, 0.96) 100%
    );
    padding: 10px;
    max-height: 60vh;
    overflow: auto;
  }

  .permission-tree {
    max-height: 520px;
    overflow: auto;
    padding-right: 2px;
  }

  :deep(.permission-tree .el-tree-node__content) {
    height: auto;
    min-height: 36px;
    margin: 2px 0;
    padding: 0;
    border-radius: 12px;
  }

  :deep(.permission-tree .el-tree-node__expand-icon) {
    color: #8a94a6;
    font-size: 12px;
  }
</style>
