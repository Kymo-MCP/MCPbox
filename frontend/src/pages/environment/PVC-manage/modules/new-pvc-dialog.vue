<template>
  <el-dialog v-model="dialogInfo.visible" :show-close="false" width="680px" top="10vh">
    <template #header>
      <div class="center mt-4 mb-4">{{ dialogInfo.title }}</div>
    </template>
    <el-scrollbar ref="scrollbarRef" max-height="75vh" always>
      <el-form
        ref="formRef"
        :model="dialogInfo.formData"
        :rules="dialogInfo.rules"
        label-width="auto"
        label-position="top"
        class="mr-2 ml-2"
      >
        <el-form-item :label="t('env.pvc.formData.name')" prop="name">
          <el-input v-model="dialogInfo.formData.name" :placeholder="t('env.pvc.formData.name')" />
        </el-form-item>
        <el-form-item :label="t('env.pvc.formData.storageClass')" prop="storageClass">
          <el-select
            v-model="dialogInfo.formData.storageClass"
            :placeholder="t('env.pvc.formData.storageClass')"
          >
            <el-option
              v-for="(type, index) in storageClassOptions"
              :key="index"
              :label="type.label"
              :value="type.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('env.pvc.formData.accessMode')" prop="accessMode">
          <el-select
            v-model="dialogInfo.formData.accessMode"
            :placeholder="t('env.pvc.formData.accessMode')"
          >
            <el-option
              v-for="(type, index) in accessModeOptions"
              :key="index"
              :label="type.label"
              :value="type.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('env.pvc.formData.storageSize')" prop="storageSize">
          <el-input
            v-model.number="dialogInfo.formData.storageSize"
            type="number"
            :placeholder="t('env.pvc.placeholder.storageSize')"
          />
        </el-form-item>

        <el-form-item :label="t('env.pvc.formData.nodeName')" prop="nodeName">
          <el-select
            v-model="dialogInfo.formData.nodeName"
            :placeholder="t('env.pvc.formData.nodeName')"
          >
            <el-option
              v-for="(type, index) in nodeNameOptions"
              :key="index"
              :label="type.label"
              :value="type.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('env.pvc.formData.hostPath')" prop="hostPath">
          <el-input
            v-model="dialogInfo.formData.hostPath"
            :placeholder="t('env.pvc.formData.hostPath')"
          />
        </el-form-item>
      </el-form>
    </el-scrollbar>

    <template #footer>
      <div class="center text-center">
        <el-button @click="handleCancel" class="mr-2">{{ t('common.cancel') }}</el-button>

        <mcp-button @click="handleConfirm" :loading="dialogInfo.loading">{{
          t('common.save')
        }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { PvcAPI, NodeAPI } from '@/api/env'
import { ElMessage } from 'element-plus'
import McpButton from '@/components/mcp-button/index.vue'
import { type PvcForm, NodeVisible } from '@/types/index'

const { t } = useI18n()
const $route = useRoute()
const { query } = $route
const formRef = ref()
const emit = defineEmits(['onRefresh'])
const dialogInfo = ref({
  visible: false,
  loading: false,
  title: t('env.pvc.formData.title'),
  formData: {
    name: '',
    storageClass: '',
    accessMode: '',
    storageSize: 0,
    nodeName: '',
    hostPath: '',
  },
  rules: {
    name: [{ required: true, message: t('env.pvc.rules.name'), trigger: 'blur' }],
    accessMode: [{ required: true, message: t('env.pvc.rules.accessMode'), trigger: 'change' }],
    storageSize: [
      { required: true, type: 'number', message: t('env.pvc.rules.storageSize'), trigger: 'blur' },
    ],
  },
})
const storageClassOptions = ref<{ label: string; value: string }[]>([])
// access model list
const accessModeOptions = ref<{ label: NodeVisible; value: NodeVisible }[]>([
  { label: NodeVisible.RWO, value: NodeVisible.RWO },
  { label: NodeVisible.ROM, value: NodeVisible.ROM },
  { label: NodeVisible.RWM, value: NodeVisible.RWM },
])
// node name options list
const nodeNameOptions = ref<{ label: string; value: string }[]>([])

/**
 * Handle cancel
 */
const handleCancel = () => {
  dialogInfo.value.visible = false
}
/**
 * Handle confirm
 */
const handleConfirm = () => {
  formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      try {
        dialogInfo.value.loading = true
        await PvcAPI.createPvc({
          ...dialogInfo.value.formData,
          environmentId: Number(query.environmentId),
        })
        ElMessage.success(t('action.create'))
        dialogInfo.value.visible = false
        emit('onRefresh')
      } finally {
        dialogInfo.value.loading = false
      }
    }
  })
}

/**
 * Handle get storageClass list
 */
const handleGetStorageClassList = async () => {
  const data = await PvcAPI.storageList({ environmentId: query.environmentId })
  storageClassOptions.value = data.list.map((storage) => {
    return { label: storage.name, value: storage.name }
  })
}

/**
 * Handle get node list
 */
const handleGetNodeList = async () => {
  const data = await NodeAPI.list({ environmentId: query.environmentId })
  nodeNameOptions.value = data.list.map((node) => {
    return { label: `${node.name}(${node.status})`, value: node.name }
  })
}

/**
 * Handle init form data
 * @param form - form data
 */
const init = () => {
  handleGetNodeList()
  handleGetStorageClassList()
  dialogInfo.value.visible = true
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.add-anv {
  width: 100%;
  border: 1px dashed var(--el-border-color);
}
</style>
