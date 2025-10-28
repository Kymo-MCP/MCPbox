<template>
  <div v-loading="pageInfo.loading">
    <div class="page-title">
      {{
        pageInfo.formData.templateId
          ? t('mcp.template.pageDesc.editTitle')
          : t('mcp.template.pageDesc.createTitle')
      }}
    </div>
    <div class="page-title base-info">{{ t('mcp.instance.formData.baseInfo') }}</div>
    <el-form
      ref="baseInfo"
      :model="pageInfo.formData"
      :rules="pageInfo.rules"
      label-width="auto"
      label-position="top"
    >
      <el-row :gutter="24">
        <el-col :span="12">
          <el-form-item :label="t('mcp.template.formData.name')" prop="name">
            <el-input
              v-model="pageInfo.formData.name"
              :placeholder="t('mcp.template.formData.name')"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12"></el-col>

        <el-col :span="12">
          <el-form-item :label="t('mcp.instance.form.accessType')" prop="accessType">
            <template #label>
              <span class="mr-4">{{ t('mcp.instance.form.accessType') }}</span>
            </template>
            <el-select
              v-model="pageInfo.formData.accessType"
              :placeholder="t('mcp.instance.form.accessType')"
            >
              <el-option
                v-for="(accessType, index) in accessTypeOptions"
                :key="index"
                :label="accessType.label"
                :value="accessType.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('mcp.instance.form.mcpProtocol')" prop="mcpProtocol">
            <el-select
              v-model="pageInfo.formData.mcpProtocol"
              :placeholder="t('mcp.instance.form.mcpProtocol')"
            >
              <el-option
                v-for="(mcpProtocol, index) in mcpProtocolOptions"
                :key="index"
                :label="mcpProtocol.label"
                :value="mcpProtocol.value"
              />
            </el-select>
          </el-form-item>
        </el-col>

        <el-col :span="12">
          <el-form-item
            v-if="showMcpServers"
            :label="t('mcp.instance.formData.mcpServers')"
            prop="mcpServers"
          >
            <el-input
              v-model="pageInfo.formData.mcpServers"
              :rows="11"
              type="textarea"
              :placeholder="placeholderServer"
              @blur="handleFormat"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('mcp.template.formData.notes')" prop="notes">
            <el-input
              v-model="pageInfo.formData.notes"
              :rows="4"
              type="textarea"
              :placeholder="t('mcp.template.formData.notes')"
            />
          </el-form-item>
          <el-form-item :label="t('mcp.template.formData.icon')" prop="iconPath">
            <Upload v-model="pageInfo.formData.iconPath"></Upload>
          </el-form-item>
        </el-col>
        <el-col :span="12"> </el-col>
      </el-row>
    </el-form>
    <div v-if="showImgAddress" class="mt-6">
      <div class="config-info">
        <!-- <span class="text-red">*</span> -->
        {{ t('mcp.instance.formData.configInfo') }}
      </div>
      <el-form
        ref="configInfo"
        :model="pageInfo.formData"
        :rules="pageInfo.rules"
        label-width="auto"
        label-position="top"
      >
        <el-row :gutter="24">
          <el-col :span="12">
            <el-form-item prop="packageId">
              <template #label>
                <span class="mr-2">{{ t('mcp.instance.formData.packageId') }}</span>
              </template>
              <Select
                v-model="selectVisible"
                v-model:selected="pageInfo.formData.packageId"
                red="packageSelect"
                :title="t('mcp.instance.formData.packageId')"
                :options="packageList"
              >
                <template #options="{ option }">
                  <div class="flex justify-between">
                    <div class="flex align-center">
                      <el-image :src="zipLogo" style="width: 32px; height: 32px"></el-image>
                      <span class="ml-2"> {{ option.name }}</span>
                    </div>
                    <div class="flex align-center">
                      {{ t('mcp.template.formData.size') }}：{{ formatFileSize(option.size) }}
                      <span>{{ timestampToDate(option.createdAt) }}</span>
                      <el-button
                        type="primary"
                        size="small"
                        link
                        class="base-btn-link ml-2"
                        @click="handleViewCode(option)"
                      >
                        {{ t('mcp.instance.action.view') }}
                      </el-button>
                    </div>
                  </div>
                </template>
              </Select>
              <div
                v-if="pageInfo.formData.packageId"
                class="select-package flex flex-direction cursor-pointer"
                @click="selectVisible = true"
              >
                <div class="text-center">
                  <el-image :src="zipLogo" style="width: 32px; height: 32px"></el-image>
                  <div>
                    {{ currentPackage.name }}
                  </div>
                </div>
                <div class="center flex-sub">
                  <span class="text-3 text-gray mr-4">
                    {{ t('mcp.template.formData.size') }}：{{ formatFileSize(currentPackage.size) }}
                  </span>
                  <span class="text-3 text-gray mr-4">
                    {{ timestampToDate(currentPackage.createdAt) }}
                  </span>
                </div>
                <div class="text-3 center">
                  {{ t('mcp.template.pageDesc.compress') }} {{ InstanceData.PACKAGE_PATH }}
                </div>
              </div>
              <div
                v-else
                class="select-package flex flex-direction cursor-pointer link-hover center text-gray"
                @click="selectVisible = true"
              >
                <el-icon class="package-logo">
                  <UploadFilled></UploadFilled>
                </el-icon>
                {{ t('mcp.instance.formData.addPackage') }}
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('mcp.instance.formData.initScript')" prop="initScript">
              <el-input
                v-model="pageInfo.formData.initScript"
                :rows="8"
                type="textarea"
                :placeholder="t('mcp.instance.formData.initScript')"
              />
            </el-form-item>
          </el-col>
          <!-- 选择环境 -->
          <el-col :span="12">
            <el-form-item :label="t('mcp.instance.formData.environmentId')" prop="environmentId">
              <el-select
                v-model="pageInfo.formData.environmentId"
                :placeholder="t('mcp.instance.formData.environmentId')"
                @change="handleChangeEnvironmentId"
              >
                <el-option
                  v-for="(env, index) in envList"
                  :key="index"
                  :label="env.name"
                  :value="env.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('mcp.instance.formData.port')" prop="port">
              <el-input
                v-model.number="pageInfo.formData.port"
                type="number"
                :placeholder="t('mcp.instance.formData.port')"
              />
            </el-form-item>
          </el-col>
          <!-- 环境变量 -->
          <el-col :span="12">
            <el-form-item
              :label="t('mcp.instance.formData.environmentVariables')"
              prop="environmentVariables"
            >
              <el-row class="w-full" :gutter="24" justify="space-between">
                <el-col
                  :span="12"
                  v-for="(env, index) in pageInfo.formData.environmentVariables"
                  :key="index"
                >
                  <div class="flex align-center mb-2">
                    <el-input
                      v-model="env.key"
                      :placeholder="t('mcp.instance.formData.key')"
                      class="mr-2"
                    />
                    <el-input
                      v-model="env.value"
                      :placeholder="t('mcp.instance.formData.value')"
                      class="mr-2"
                    />
                    <el-icon
                      class="cursor-pointer"
                      color="#F56C6C"
                      @click="handleDeleteEnvVariable(index)"
                    >
                      <Remove />
                    </el-icon>
                  </div>
                </el-col>
              </el-row>
              <el-button class="add-env" :icon="Plus" plain @click="handleAddEnvVariable">
                {{ t('mcp.instance.formData.add-env') }}
              </el-button>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('mcp.instance.formData.command')" prop="command">
              <template #label>
                <span class="mr-2">{{ t('mcp.instance.formData.command') }}</span>
                <el-tooltip effect="light" placement="top" class="ml-6" :raw-content="true">
                  <el-icon><i class="icon iconfont MCP-tishi1"></i></el-icon>
                  <template #content>
                    <div v-html="pageInfo.tooltip.imgAddress"></div>
                  </template>
                </el-tooltip>
              </template>
              <el-input
                v-model="pageInfo.formData.command"
                :placeholder="
                  showCommand ? InstanceData.COMMAND_TIP : t('mcp.instance.formData.command')
                "
                :disabled="showCommand"
              />
            </el-form-item>
            <el-form-item
              v-if="showServicePath"
              :label="t('mcp.template.formData.servicePath')"
              prop="servicePath"
            >
              <el-input
                v-model="pageInfo.formData.servicePath"
                :placeholder="t('mcp.template.formData.servicePath')"
              />
            </el-form-item>
          </el-col>

          <el-col :span="24">
            <el-form-item
              v-if="pageInfo.formData.environmentId"
              :label="t('mcp.instance.formData.volumeConfig')"
              prop="volumeConfig"
            >
              <el-row class="w-full" :gutter="20">
                <el-col
                  v-for="(volume, index) in pageInfo.formData.volumeMounts"
                  :key="index"
                  :span="12"
                >
                  <el-card class="w-full mb-2">
                    <div>
                      <div class="flex align-center justify-between">
                        <span>
                          <span class="text-red">*</span>{{ t('mcp.instance.formData.volumeType') }}
                        </span>
                        <el-icon
                          color="#F56C6C"
                          class="cursor-pointer"
                          @click="handleDeleteVolume(index)"
                        >
                          <CircleClose />
                        </el-icon>
                      </div>
                      <el-select
                        v-model="volume.type"
                        :placeholder="t('mcp.instance.formData.volumeType')"
                        style="width: 200px"
                      >
                        <el-option label="HostPath" value="HostPath" />
                        <el-option label="PVC" value="PVC" />
                      </el-select>
                    </div>
                    <template v-if="volume.type === 'HostPath'">
                      <div class="flex align-end">
                        <div>
                          <div>
                            <span class="text-red">*</span>{{ t('mcp.instance.formData.nodeName') }}
                          </div>
                          <el-select
                            v-model="volume.nodeName"
                            :placeholder="t('mcp.instance.formData.nodeName')"
                            style="width: 200px"
                          >
                            <el-option
                              v-for="(node, index) in nodeList"
                              :key="index"
                              :label="node.name"
                              :value="node.name"
                            />
                          </el-select>
                        </div>
                        <el-button class="mr-2 ml-2" :icon="RefreshRight" />
                        <div>
                          <div>
                            <span class="text-red">*</span>{{ t('mcp.instance.formData.hostPath') }}
                          </div>
                          <el-input
                            v-model="volume.hostPath"
                            :placeholder="t('mcp.instance.formData.hostPath')"
                          />
                        </div>
                      </div>
                      <div class="flex align-end">
                        <div class="mr-2">
                          <div>
                            <span class="text-red">*</span
                            >{{ t('mcp.instance.formData.mountPath') }}
                          </div>
                          <el-input
                            v-model="volume.mountPath"
                            :placeholder="t('mcp.instance.formData.mountPath')"
                            style="width: 200px"
                          />
                        </div>
                        <div>
                          <div>{{ t('mcp.instance.formData.readOnly') }}</div>
                          <el-switch v-model="volume.readOnly" />
                        </div>
                      </div>
                    </template>

                    <template v-if="volume.type === 'PVC'">
                      <div class="flex align-end">
                        <div>
                          <div>
                            <span class="text-red">*</span>{{ t('mcp.instance.formData.pvcName') }}
                          </div>
                          <el-select
                            v-model="volume.pvcName"
                            placeholder="PVC"
                            style="width: 200px"
                            @change="handlePvcChange($event, volume)"
                          >
                            <el-option
                              v-for="(pvc, index) in pvcList"
                              :key="index"
                              :value="pvc.name"
                              :disabled="disabledPvcNode(pvc)"
                            >
                              <span>{{ pvc.name }}</span>
                              <el-tag v-if="disabledPvcNode(pvc)" color="orange">
                                {{ t('mcp.template.formData.isBind') }}
                              </el-tag>
                            </el-option>
                          </el-select>
                        </div>
                        <el-button class="mr-2 ml-2" :icon="RefreshRight" />
                        <div class="mr-2">
                          <div>
                            <span class="text-red">*</span
                            >{{ t('mcp.instance.formData.mountPath') }}
                          </div>
                          <el-input
                            v-model="volume.mountPath"
                            :placeholder="t('mcp.instance.formData.mountPath')"
                          />
                        </div>
                        <div>
                          <div>{{ t('mcp.instance.formData.readOnly') }}</div>
                          <el-switch
                            v-model="volume.readOnly"
                            :disabled="disabledReadOnly(volume.pvcName)"
                          />
                        </div>
                      </div>
                      <el-space v-if="!selectedPvc(volume.pvcName).length">
                        <el-text type="secondary">
                          {{ t('mcp.template.formData.accessModes') }}:
                          {{ selectedPvc(volume.pvcName).accessModes?.join(', ') }}
                        </el-text>
                        <template v-if="selectedPvc(volume.pvcName).pods?.length > 0">
                          <el-text type="secondary">
                            {{ t('mcp.template.formData.pods') }}:
                            {{ selectedPvc(volume.pvcName).pods.join(', ') }}
                          </el-text>
                        </template>
                      </el-space>
                    </template>
                  </el-card>
                </el-col>
              </el-row>
              <el-button class="add-env" :icon="Plus" plain @click="handleAddVolume">
                {{ t('mcp.instance.formData.addVolume') }}
              </el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>
    <div class="action-footer mt-4 flex">
      <el-button @click="handleCancel" class="mr-4">{{ t('common.cancel') }}</el-button>
      <mcp-button @click="handleConfirm">{{ t('mcp.template.action.save') }}</mcp-button>
      <span class="ml-4">
        <mcp-button @click="handleSaveAndInstance">{{
          t('mcp.template.action.saveAndInstance')
        }}</mcp-button>
      </span>
    </div>
  </div>
</template>
<script setup lang="ts">
import { Plus, Remove, CircleClose, RefreshRight, UploadFilled } from '@element-plus/icons-vue'
import { TemplateAPI } from '@/api/mcp/template'
import { ElMessage } from 'element-plus'
import { formatFileSize, timestampToDate } from '@/utils/system'
import { useMcpStoreHook } from '@/stores'
import { useTemplateFormHooks } from './hooks/form-template'
import { JsonFormatter } from '@/utils/json'
import { ElLoading } from 'element-plus'
import Upload from '@/components/upload/index.vue'
import Select from '@/components/mcp-select/index.vue'
import zipLogo from '@/assets/logo/zip.png'
import McpButton from '@/components/mcp-button/index.vue'
import { AccessType, McpProtocol, InstanceData, NodeVisible } from '@/types/instance'
import { type VolumeMountsItme, type PvcForm, type Code } from '@/types/index'

const { t } = useI18n()
const {
  jumpToPage,
  router,
  query,
  pageInfo,
  placeholderServer,
  showImgAddress,
  showMcpServers,
  showServicePath,
  showCommand,
  disabledPvcNode,
  selectedPvc,
  disabledReadOnly,
  selectVisible,
} = useTemplateFormHooks()
const { packageList, envList, nodeList, pvcList, accessTypeOptions } = toRefs(useMcpStoreHook())
const { handleGetPackageList, handleGetEnvList, handleGetNodeList, handleGetPvcList } =
  useMcpStoreHook()

// MCP protocol options
const mcpProtocolOptions = computed(() => {
  return useMcpStoreHook().mcpProtocolOptions.filter((option) => {
    return (
      pageInfo.value.formData.accessType === AccessType.HOSTING ||
      ([AccessType.DIRECT, AccessType.PROXY].includes(pageInfo.value.formData.accessType) &&
        [McpProtocol.SSE, McpProtocol.STEAMABLE_HTTP].includes(option.value))
    )
  })
})

/**
 * The current selected code package
 */
const currentPackage = computed(() => {
  return (
    packageList.value?.find(
      (item: { id: string }) => item.id === pageInfo.value.formData.packageId,
    ) || {name:'',size:"",createdAt:""}
  )
})

const handleFormat = () => {
  pageInfo.value.formData.mcpServers = JsonFormatter.format(pageInfo.value.formData.mcpServers)
}

/**
 * Handle view code package
 */
const handleViewCode = (code: Code) => {
  jumpToPage({
    url: '/view-code-package',
    data: {
      id: code.id,
      name: code.name,
    },
    isOpen: true,
  })
}

/**
 * Handle add an environment variables
 */
const handleAddEnvVariable = () => {
  pageInfo.value.formData.environmentVariables.push({ key: '', value: '' })
}

/**
 * Handle delete environment variables
 * @param index - Index of environment variables
 */
const handleDeleteEnvVariable = (index: number) => {
  pageInfo.value.formData.environmentVariables.splice(index, 1)
}

/**
 * Handle add a volume
 */
const handleAddVolume = () => {
  pageInfo.value.formData.volumeMounts.push({
    type: '',
    nodeName: '', // HostPath name
    hostPath: '',
    mountPath: '',
    pvcName: '',
    readOnly: false,
  })
}

/**
 * Handle delete Volume mounting
 * @param index - Index of need Delete Volume
 */
const handleDeleteVolume = (index: number) => {
  pageInfo.value.formData.volumeMounts.splice(index, 1)
}

/**
 *  Update readOnly status
 * @param key - $event
 * @param volume - item of pvc
 */
const handlePvcChange = (key: any, volume: VolumeMountsItme) => {
  const pvc = pvcList.value.find((pvc: PvcForm) => pvc.name === volume.pvcName)
  if (pvc) {
    const accessModes = pvc.accessModes || []
    let readOnlyValue = false
    if (accessModes.includes(NodeVisible.ROM)) {
      readOnlyValue = true
    } else if (accessModes.includes(NodeVisible.RWM)) {
      readOnlyValue = false
    } else if (accessModes.includes(NodeVisible.RWO)) {
      readOnlyValue = false
    }
    // Update readOnly status
    volume.readOnly = readOnlyValue
  }
}

/**
 * Handle cancel
 */
const handleCancel = () => {
  router.push('/template-manage')
}
/**
 * Handle confirm
 */
const baseInfo = ref()
const configInfo = ref()
const handleConfirm = async () => {
  await handleSaveTemplate()
  jumpToPage({
    url: '/template-manage',
    data: {},
  })
}

/**
 * Handle save template and create a instance
 */
const handleSaveAndInstance = async () => {
  const result = await handleSaveTemplate()
  jumpToPage({
    url: '/new-instance',
    data: { templateId: result.templateId || query.templateId },
  })
}

/**
 * Handle confirm save
 */
const handleSaveTemplate = async () => {
  try {
    pageInfo.value.loading = true
    // valid baseInfo
    const validateBase = new Promise<boolean>((resolve) => {
      baseInfo.value.validate((valid: boolean) => resolve(valid))
    })
    // valid configInfo
    const validateConfig = new Promise<boolean>((resolve) => {
      if (configInfo.value) {
        configInfo.value.validate((valid: boolean) => resolve(valid))
      } else {
        resolve(true)
      }
    })
    const [isBaseValid, isConfigValid] = await Promise.all([validateBase, validateConfig])

    // pass and handle create
    if (isBaseValid && isConfigValid) {
      try {
        pageInfo.value.loading = true
        const data = await (query.templateId ? TemplateAPI.edit : TemplateAPI.create)({
          ...pageInfo.value.formData,
          environmentVariables: pageInfo.value.formData.environmentVariables?.reduce(
            (obj: any, item: any) => ({ ...obj, [item.key]: item.value }),
            {},
          ),
        })
        pageInfo.value.visible = false
        ElMessage.success(
          pageInfo.value.formData.templateId ? t('action.edit') : t('action.success'),
        )
        return data
        // router.push('/template-manage')
      } finally {
        pageInfo.value.loading = false
      }
    } else {
      // One form failed validation
      ElMessage.warning(t('mcp.template.rules.validForm'))
    }
  } finally {
    pageInfo.value.loading = false
  }
}

/**
 * Handle change environment event
 */
const handleChangeEnvironmentId = async () => {
  handleGetNodeList(pageInfo.value.formData.environmentId)
  handleGetPvcList(pageInfo.value.formData.environmentId)
}

/**
 * handle get template detail info
 */
const handleGetTemplateDetail = async () => {
  const data = await TemplateAPI.detail({
    id: query.templateId,
  })
  pageInfo.value.formData = data
  pageInfo.value.formData.mcpServers = JsonFormatter.format(data.mcpServers)
  pageInfo.value.formData.environmentVariables = data.environmentVariables
    ? Object.keys(data.environmentVariables)?.map((key) => ({
        key,
        value: data.environmentVariables[key],
      }))
    : []
  pageInfo.value.formData.volumeMounts = data.volumeMounts || []
}

/**
 * init data
 * @param form - instance form data
 */
const init = () => {
  let loadingInstance
  try {
    loadingInstance = ElLoading.service({ fullscreen: true, text: t('status.loading') + '...' })
    handleGetEnvList()
    handleGetPackageList()
    if (query.templateId) {
      handleGetTemplateDetail()
    }
  } finally {
    loadingInstance?.close()
  }
}

onMounted(init)
</script>

<style lang="scss" scoped>
.add-env {
  width: 100%;
  border: 1px dashed var(--el-border-color);
}
.w-full {
  width: 100%;
}
.page-title {
  font-family:
    PingFangSC,
    PingFang SC;
  font-weight: 600;
  font-size: 20px;
  line-height: 28px;
  &.base-info {
    margin-top: 40px;
    margin-bottom: 16px;
  }
}
.config-info {
  margin-top: 40px;
  margin-bottom: 24px;
  font-weight: 600;
  font-size: 20px;
  line-height: 28px;
}
.select-package {
  width: 100%;
  height: 176px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
  padding: 24px;
  &:hover {
    border-color: var(--ep-purple-color);
  }
  .package-logo {
    font-size: 65px;
  }
}
:deep(.el-input__wrapper) {
  background: var(--ep-bg-form);
}
:deep(.el-select__wrapper) {
  background: var(--ep-bg-form);
}
:deep(.el-textarea__inner) {
  background: var(--ep-bg-form);
}
:deep(.el-button) {
  background-color: transparent;
}
</style>
