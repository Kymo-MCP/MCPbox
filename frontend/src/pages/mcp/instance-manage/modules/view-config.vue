<template>
  <el-dialog v-model="dialogInfo.visible" width="1000px" top="10vh">
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
    <el-scrollbar ref="scrollbarRef" max-height="590px" always>
      <div class="config-info">{{ dialogInfo.config }}</div>
      <el-icon class="base-btn-link copy-icon" @click="handleCopy"><CopyDocument /></el-icon>
    </el-scrollbar>
    <template #footer>
      <div class="center">
        <mcp-button @click="handleCopy" class="w100">{{
          t('mcp.instance.action.copy')
        }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { setClipboardData } from '@/utils/system'
import { JsonFormatter } from '@/utils/json.ts'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'

const { t } = useI18n()
const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.config'),
  config: '',
})

/**
 * Handle copy config info
 */
const handleCopy = async () => {
  await setClipboardData(dialogInfo.value.config)
  ElMessage.success(t('action.copy'))
}

/**
 * Handle init model data
 * @param config - public proxy config
 */
const init = (config: any) => {
  dialogInfo.value.visible = true
  dialogInfo.value.config = JsonFormatter.format(config)
}
defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.config-info {
  min-height: 590px;
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  border-radius: 8px;
  background: #000000;
  border-radius: 8px;
  padding: 24px;
}
.copy-icon {
  position: absolute;
  top: 12px;
  right: 12px;
  cursor: pointer;
}
</style>
