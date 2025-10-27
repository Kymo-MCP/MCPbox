<template>
  <div class="dialog-body">
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
          <el-form-item :label="t('env.run.formData.name')" prop="name">
            <el-input
              v-model="dialogInfo.formData.name"
              :placeholder="t('env.run.formData.name')"
            />
          </el-form-item>
          <el-form-item :label="t('env.run.formData.environment')" prop="environment">
            <el-select
              v-model="dialogInfo.formData.environment"
              :placeholder="t('env.run.formData.environment')"
              @change="handleGetNamespaceList"
            >
              <el-option
                v-for="(type, index) in environmentOptions"
                :key="index"
                :label="type.label"
                :value="type.value"
                :disabled="type.disabled"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('env.run.formData.namespace')" prop="namespace">
            <el-select
              v-model="dialogInfo.formData.namespace"
              :placeholder="t('env.run.formData.namespace')"
            >
              <el-option
                v-for="(type, index) in namespaceOptions"
                :key="index"
                :label="type.label"
                :value="type.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('env.run.formData.config')" prop="config">
            <template #label>
              {{ t('env.run.formData.config') }}
              <el-button
                type="primary"
                @click="formatYaml"
                size="small"
                :disabled="!dialogInfo.formData.config"
                class="base-btn ml-2"
                >{{ t('env.run.formData.format') }}</el-button
              >
            </template>
            <el-input
              v-model="dialogInfo.formData.config"
              :rows="8"
              type="textarea"
              :placeholder="placeholderConfig"
              @blur="handleGetNamespaceList"
            />
          </el-form-item>
        </el-form>
      </el-scrollbar>

      <template #footer>
        <div class="dialog-footer text-center center">
          <el-button @click="handleCancel" class="mr-4">{{ t('common.cancel') }}</el-button>
          <mcp-button @click="handleConfirm" :loading="dialogInfo.loading">{{
            t('common.save')
          }}</mcp-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { EnvAPI } from '@/api/env'
import { ElMessage } from 'element-plus'
import * as yaml from 'js-yaml'
import { cloneDeep } from 'lodash-es'
import McpButton from '@/components/mcp-button/index.vue'
import { EnvType, EnvFormData } from '@/types/env'

const { t } = useI18n()
const formRef = ref()
const placeholderConfig = computed(() => {
  return t('env.run.formData.placeholderConfig') + EnvFormData.TIP_CONFIG
})
const dialogInfo = ref({
  visible: false,
  loading: false,
  title: t('env.run.formData.title'),
  formData: {
    id: null,
    name: '',
    environment: '',
    config: '',
    namespace: '',
  },
  rules: {
    name: [{ required: true, message: t('env.run.rules.name'), trigger: 'blur' }],
    environment: [{ required: true, message: t('env.run.rules.environment'), trigger: 'change' }],
    config: [{ required: true, message: t('env.run.rules.config'), trigger: 'blur' }],
    namespace: [{ required: true, message: t('env.run.rules.namespace'), trigger: 'change' }],
  },
})
// deployment mode
const environmentOptions = ref([
  { label: 'Kubernetes', value: EnvType.K8S, disabled: false },
  { label: 'Docker', value: EnvType.DOCKER, disabled: true },
])
// Namespace options
const namespaceOptions = ref<any>([])

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
  } catch (error: unknown) {
    // If formatting fails, return the original string
    ElMessage.error(t('env.run.rules.formatFaild'))
  }
}

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
        await (dialogInfo.value.formData.id ? EnvAPI.editEnv : EnvAPI.createEnv)({
          ...dialogInfo.value.formData,
        })
        ElMessage.success(dialogInfo.value.formData.id ? t('action.edit') : t('action.create'))
        dialogInfo.value.visible = false
      } finally {
        dialogInfo.value.loading = false
      }
    }
  })
}

/**
 * Handle get namespace list
 */
const handleGetNamespaceList = async () => {
  formatYaml()
  if (dialogInfo.value.formData.environment && dialogInfo.value.formData.config) {
    const data = await EnvAPI.namespaceList({
      environment: ['kubernetes', 'docker'].indexOf(dialogInfo.value.formData.environment),
      config: dialogInfo.value.formData.config,
    })
    namespaceOptions.value = data.list.map((namespace: string) => {
      return {
        label: namespace,
        value: namespace,
      }
    })
  }
}

/**
 * Handle init form data
 * @param form - form data
 */
const init = (form: any) => {
  if (form.id) {
    Object.assign(dialogInfo.value.formData, cloneDeep(form))
    dialogInfo.value.title = t('env.run.formData.edit')
    handleGetNamespaceList()
  } else {
    formRef.value?.resetFields()
  }
  dialogInfo.value.visible = true
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.dialog-body {
  :deep(.el-dialog) {
    background-color: var(--el-dialog-bg) !important;
    border: 1px solid #999999;
  }
}
</style>
