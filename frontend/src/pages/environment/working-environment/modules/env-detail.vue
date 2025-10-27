<template>
  <div class="dialog-body">
    <el-dialog v-model="dialogInfo.visible" width="680px" top="10vh" :show-close="false">
      <template #header>
        <div class="center">{{ dialogInfo.title }}</div>
      </template>
      <el-scrollbar ref="scrollbarRef" max-height="75vh" always>
        <el-form :model="dialogInfo.formData" label-width="auto" label-position="top">
          <el-form-item :label="t('env.run.name')">
            <el-input :value="dialogInfo.formData.name" type="text" disabled />
          </el-form-item>
          <el-form-item :label="t('env.run.environment')">
            <el-input :value="dialogInfo.formData.environment" type="text" disabled />
          </el-form-item>
          <el-form-item :label="t('env.run.namespace')">
            <el-input :value="dialogInfo.formData.namespace" type="text" disabled />
          </el-form-item>
          <el-form-item :label="t('env.run.createdAt')">
            <el-input :value="dialogInfo.formData.createdAt" type="text" disabled />
          </el-form-item>
          <el-form-item :label="t('env.run.updatedAt')">
            <el-input :value="dialogInfo.formData.updatedAt" type="text" disabled />
          </el-form-item>
          <el-form-item :label="t('env.run.config')">
            <el-input :value="dialogInfo.formData.config" :rows="8" type="textarea" disabled />
          </el-form-item>
        </el-form>
      </el-scrollbar>
      <template #footer>
        <div class="center">
          <mcp-button @click="dialogInfo.visible = false" class="w100">{{
            t('common.close')
          }}</mcp-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import * as yaml from 'js-yaml'
import { cloneDeep } from 'lodash-es'
import McpButton from '@/components/mcp-button/index.vue'

const { t } = useI18n()
const dialogInfo = ref({
  visible: false,
  title: t('env.run.detail'),
  formData: {
    id: null,
    name: '',
    environment: '',
    config: '',
    namespace: '',
    createdAt: '',
    updatedAt: '',
  },
})

// YAML Format and Validation Functions
const formatYaml = () => {
  try {
    // Parse YAML to validate format
    const parsed = yaml.load(dialogInfo.value.formData.config)
    // Resialize to standard format
    dialogInfo.value.formData.config = yaml.dump(parsed, {
      indent: 2,
      lineWidth: -1,
      noRefs: true,
      quotingType: '"',
      forceQuotes: false,
    })
  } catch (error) {
    // If formatting fails, return the original string
    ElMessage.error(t('env.run.rules.formatFaild'))
  }
}

/**
 * Handle init detail info
 * @param form - form data
 */
const init = (formData: any) => {
  dialogInfo.value.visible = true
  dialogInfo.value.formData = cloneDeep(formData)
  formatYaml()
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.el-descriptions {
  margin-top: 20px;
}
.cell-item {
  display: flex;
  align-items: center;
}
.margin-top {
  margin-top: 20px;
}
.el-text {
  overflow: auto;
  flex-wrap: wrap;
  overflow-wrap: anywhere;
  height: 400px;
}
.dialog-body {
  :deep(.el-dialog) {
    background-color: var(--el-dialog-bg) !important;
    border: 1px solid #999999;
  }
}
</style>
<style>
html.dark {
  --el-disabled-bg-color: rgba(255, 255, 255, 0.08);
}
</style>
