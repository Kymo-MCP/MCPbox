<template>
  <el-dialog v-model="dialogInfo.visible" width="464px" top="10vh" :show-close="false">
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
    <el-scrollbar ref="scrollbarRef" max-height="75vh" always>
      <el-form :model="dialogInfo.formData" label-width="auto" label-position="top">
        <el-form-item :label="t('mcp.instance.name')">
          <el-input :value="dialogInfo.formData.instanceName" type="text" disabled />
        </el-form-item>
        <el-form-item :label="t('mcp.instance.form.accessType')">
          <el-input
            :value="
              accessTypeOptions.find(
                (item) => item.value === Number(dialogInfo.formData.accessType),
              )?.label
            "
            type="text"
            disabled
          />
        </el-form-item>
        <el-form-item :label="t('mcp.instance.form.mcpProtocol')">
          <el-input
            :value="
              mcpProtocolOptions.find(
                (item) => item.value === Number(dialogInfo.formData.mcpProtocol),
              )?.label
            "
            type="text"
            disabled
          />
        </el-form-item>
        <el-form-item :label="t('mcp.instance.status')">
          <el-input
            :value="activeOptions[dialogInfo.formData.status as keyof typeof activeOptions].label"
            type="text"
            disabled
          />
        </el-form-item>
        <el-form-item :label="t('mcp.instance.packStatus')">
          <el-input
            :value="
              containerOptions[dialogInfo.formData.containerStatus as keyof typeof containerOptions]
                .label
            "
            type="text"
            disabled
          />
        </el-form-item>
        <el-form-item :label="t('mcp.instance.env')">
          <el-input :value="dialogInfo.formData.environmentName" type="text" disabled />
        </el-form-item>
        <el-form-item :label="t('mcp.instance.createTime')">
          <el-input :value="timestampToDate(dialogInfo.formData.createdAt)" type="text" disabled />
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
</template>

<script setup lang="ts">
import { cloneDeep } from 'lodash-es'
import { timestampToDate } from '@/utils/system'
import { useInstanceTableHooks } from '../hooks'
import { useMcpStoreHook } from '@/stores'
import McpButton from '@/components/mcp-button/index.vue'

const { t } = useI18n()
const { activeOptions, containerOptions } = useInstanceTableHooks()
const { accessTypeOptions, mcpProtocolOptions } = useMcpStoreHook()

const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.detail'),
  formData: {
    id: null,
    accessType: '',
    instanceName: '',
    mcpProtocol: '',
    status: '',
    containerStatus: '',
    environmentName: '',
    createdAt: '',
  },
})

/**
 * Handle init form data
 * @param form - instance form
 */
const init = (formData: any) => {
  dialogInfo.value.visible = true
  dialogInfo.value.formData = cloneDeep(formData)
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
.w100 {
  width: 100px;
}
</style>
<style>
html.dark {
  --el-disabled-bg-color: rgba(255, 255, 255, 0.08);
}
</style>
