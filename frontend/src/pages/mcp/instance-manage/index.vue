<template>
  <div>
    <!-- head model -->
    <div class="flex justify-between page-header">
      <div class="header-container">
        {{ t('mcp.instance.pageDesc.list') }}
        <span class="desc">{{ t('mcp.instance.pageDesc.desc') }}</span>
      </div>
      <el-dropdown
        trigger="click"
        class="ml-4"
        @click.stop
        :show-arrow="false"
        @command="(cmd: string) => handleCommand(cmd, null)"
      >
        <mcp-button :icon="Plus">{{ t('mcp.instance.action.add') }}</mcp-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="handleAddInstance">
              <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
              {{ t('mcp.instance.action.customize') }}
            </el-dropdown-item>
            <el-dropdown-item command="handleAddByTemplate">
              <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
              {{ t('mcp.instance.action.byTemplate') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- count data model -->
    <el-row justify="space-around">
      <el-col :span="6" class="center" v-for="(dataCount, index) in dataCountList" :key="index">
        <GlareHover
          width="338px"
          height="124px"
          background="transparent"
          border-color="#222222"
          border-radius="6px"
          glare-color="#ffffff"
          :glare-opacity="0.3"
          :glare-size="300"
          :transition-duration="800"
          :play-once="false"
          class="data-card"
        >
          <div
            class="data-card flex justify-around align-center"
            @click="handleSearchByCount(dataCount.type)"
          >
            <div class="center">
              <el-image
                :src="dataCount.icon"
                style="width: 20px; height: 20px"
                class="mr-2"
              ></el-image>
              {{ dataCount.title }}
            </div>
            <span class="count" v-count-to="dataCount.count"></span>
          </div>
        </GlareHover>
      </el-col>
    </el-row>

    <!-- table model -->
    <div v-loading="load.status" :element-loading-text="load.text">
      <TablePlus
        ref="tablePlus"
        searchContainer="#instanceSearch"
        :showOperation="true"
        :requestConfig="requestConfig"
        :columns="columns"
        v-model:pageConfig="pageConfig"
        :handlerColumnConfig="{
          fixed: 'right',
          width: '120px',
        }"
      >
        <template #action>
          <div class="flex justify-between mb-4">
            <div class="center">
              <el-image :src="instanceLogo" style="width: 20px; height: 20px"></el-image>
              <span class="desc"
                >{{ t('mcp.instance.pageDesc.total') }}：{{ pageConfig.total }}</span
              >
            </div>
            <div id="instanceSearch"></div>
          </div>
        </template>
        <template #instanceName="{ row }">
          <div class="flex align-center">
            <mcp-image :src="row.iconPath" width="32" height="32"></mcp-image>
            <el-tooltip effect="dark" placement="top" class="ml-6" :raw-content="true">
              <div class="flex-sub ml-2 ellipsis-two">{{ row.instanceName }}</div>
              <template #content>
                <div style="width: 300px">{{ row.instanceName }}</div>
                <div style="width: 300px">{{ row.instanceId }}</div>
              </template>
            </el-tooltip>
          </div>
        </template>
        <template #status="{ row }">
          <el-text :type="activeOptions[row.status as keyof typeof activeOptions].type" link>
            {{ activeOptions[row.status as keyof typeof activeOptions].label }}
          </el-text>
        </template>
        <template #containerStatus="{ row }">
          <div v-if="row.accessType === AccessType.HOSTING">
            <el-text
              :type="containerOptions[row.containerStatus as keyof typeof containerOptions]?.type"
              link
            >
              <span class="mr-2">
                {{ containerOptions[row.containerStatus as keyof typeof containerOptions]?.label }}
              </span>
            </el-text>
            <el-popover v-if="!row.containerIsReady" placement="right" width="250">
              <div>{{ t('mcp.instance.pageDesc.packStatusTips') }}</div>
              <template #reference>
                <el-icon class="cursor-pointer"><Warning /></el-icon>
              </template>
            </el-popover>
          </div>

          <span v-else>--</span>
        </template>
        <template #createdAt="{ row }">
          <span>{{ timestampToDate(row.createdAt) }}</span>
        </template>
        <template #publicProxyConfig="{ row }">
          <el-link
            type="success"
            class="base-btn-link"
            :underline="false"
            @click="handleViewConfig(row.publicProxyConfig)"
          >
            {{ t('mcp.instance.viewConfig') }}
          </el-link>
        </template>
        <template #operation="{ row }">
          <div class="flex align-center">
            <el-button
              type="primary"
              size="small"
              link
              class="base-btn-link"
              @click="handleViewDetail(row)"
            >
              {{ t('mcp.instance.action.view') }}
            </el-button>

            <el-dropdown
              trigger="click"
              class="ml-4"
              @click.stop
              :show-arrow="false"
              @command="(cmd: string) => handleCommand(cmd, row)"
            >
              <el-icon class="link-hover cursor-pointer"><More /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item
                    v-if="row.accessType === AccessType.HOSTING"
                    command="handleViewLog"
                  >
                    {{ t('mcp.instance.action.log') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleEditInstance">
                    {{ t('env.run.action.edit') }}
                  </el-dropdown-item>
                  <el-dropdown-item
                    v-if="row.accessType !== AccessType.DIRECT"
                    :command="
                      row.status === InstanceStatus.INACTIVE
                        ? 'handleRestartInstance'
                        : 'handleStopInstance'
                    "
                  >
                    {{
                      row.status === InstanceStatus.INACTIVE
                        ? t('mcp.instance.action.start')
                        : t('mcp.instance.action.stop')
                    }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleRestartInstance">
                    {{ t('mcp.instance.action.reStart') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleViewStatus">
                    {{ t('mcp.instance.action.probe') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleDeleteInstance">
                    <el-button type="danger" link>
                      {{ t('mcp.instance.action.delete') }}
                    </el-button>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </TablePlus>
    </div>

    <!-- view detail model -->
    <InstanceDetail ref="instanceDetail"></InstanceDetail>
    <!-- view config model -->
    <ViewConfig ref="viewConfig"></ViewConfig>
    <!-- probe instance dialog model -->
    <ProbeStatus ref="probe"></ProbeStatus>
    <!-- select template -->
    <Select
      v-model="selectVisible"
      ref="packageSelect"
      :title="t('mcp.instance.action.selectTempalte')"
      :options="templateList"
      @confirm="handleConfirmSelect"
    >
      <template #options="{ option }">
        <div class="flex justify-between">
          <div class="flex align-center w-full">
            <el-image :src="option.iconPath" style="width: 32px; height: 32px"></el-image>
            <el-tooltip effect="dark" placement="top" class="ml-2" :raw-content="true">
              <div class="flex-sub ml-2 ellipsis-one">{{ option.name }}</div>
              <template #content>
                <div style="width: 300px">
                  {{ option.name }}
                </div>
              </template>
            </el-tooltip>
          </div>
        </div>
      </template>
    </Select>
  </div>
</template>

<script setup lang="ts">
import { Plus, More, Warning } from '@element-plus/icons-vue'
import { timestampToDate } from '@/utils/system'
import TablePlus from '@/components/TablePlus/index.vue'
import { useInstanceTableHooks } from './hooks/index.ts'
import { ElMessage, ElMessageBox } from 'element-plus'
import InstanceDetail from './modules/instance-detail.vue'
import instanceLogo from '@/assets/logo/instance.png'
import GlareHover from '@/components/Animation/GlareHover.vue'
import McpButton from '@/components/mcp-button/index.vue'
import ViewConfig from './modules/view-config.vue'
import ProbeStatus from './modules/probe-dialog.vue'
import Select from '@/components/mcp-select/index.vue'
import { TemplateAPI } from '@/api/mcp/template'
import McpImage from '@/components/mcp-image/index.vue'
import { AccessType, InstanceStatus } from '@/types/instance'

const { t } = useI18n()
const {
  load,
  query,
  columns,
  jumpToPage,
  tablePlus,
  requestConfig,
  pageConfig,
  handleAddInstance,
  activeOptions,
  containerOptions,
  InstanceAPI,
  instanceCount,
  instanceDetail,
  viewConfig,
  dataCountList,
  probe,
  selectVisible,
  templateList,
  timer,
} = useInstanceTableHooks()

/**
 * Handle create a instance by template list
 */
const handleAddByTemplate = async () => {
  selectVisible.value = true
  const data = await TemplateAPI.list({ page: '1', pageSize: '999' })
  templateList.value = data.list.map((template: any) => ({
    id: template.templateId,
    name: template.name,
    ...template,
  }))
}

/**
 * handle confirm selected template
 * @param templateId - selected of templateId
 */
const handleConfirmSelect = (templateId: string) => {
  jumpToPage({
    url: '/new-instance',
    data: { templateId },
  })
}

/**
 * Handle search instance list by count data
 * @param type - dimension of count
 */
const handleSearchByCount = (type: string) => {
  switch (type) {
    case 'total':
      tablePlus.value.resetFields()
      return
    case 'active':
      tablePlus.value.customize({
        status: InstanceStatus.ACTIVE,
        accessType: null,
        mcpProtocol: null,
      })
      break
    case 'inactive':
      tablePlus.value.customize({
        status: InstanceStatus.INACTIVE,
        accessType: null,
        mcpProtocol: null,
      })
      break
    case 'hosting':
      tablePlus.value.customize({
        status: null,
        accessType: AccessType.HOSTING,
      })
      break
    default:
      tablePlus.value.resetFields()
  }
  tablePlus.value.initData()
}

/**
 * Handle view instance detail info
 * @param row - item of instance
 */
const handleViewDetail = (row: any) => {
  instanceDetail.value.init(row)
}

/**
 * handle view instance server log
 * @param row - item of instance
 */
const handleViewLog = (row: any) => {
  jumpToPage({
    url: '/instance-log',
    data: {
      instanceId: row.instanceId,
    },
    isOpen: true,
  })
}
/**
 * Handle eidt the instance form
 * @param row - instance form data
 */
const handleEditInstance = (row: any) => {
  jumpToPage({
    url: '/new-instance',
    data: {
      instanceId: row.instanceId,
    },
  })
}
/**
 * handle stop instance server
 * @param instanceId - 实例ID
 */
const handleStopInstance = async (instanceId: string) => {
  try {
    load.value.status = true
    load.value.text = t('mcp.instance.loading.stop')
    await InstanceAPI.stop({
      instanceId,
    })
    ElMessage.success(t('mcp.instance.action.stopInstance'))
    init()
  } finally {
    load.value.status = false
  }
}

/**
 * Handle start instance
 * @param instanceId - instance id
 */
const handleRestartInstance = async (instanceId: string) => {
  try {
    load.value.status = true
    load.value.text = t('mcp.instance.loading.restart')
    await InstanceAPI.restart({
      instanceId,
    })
    ElMessage.success(t('mcp.instance.action.restart'))
    init()
  } finally {
    load.value.status = false
  }
}

/**
 * Handle view public proxy config
 * @param config - JSON of config info
 */
const handleViewConfig = async (config: string) => {
  viewConfig.value.init(config)
}

/**
 * Handle probe insatnce status
 * @param instanceInfo - instance form data
 */
const handleViewStatus = async (instanceInfo: { [key: string]: string }) => {
  probe.value.init(instanceInfo)
}

/**
 * Handle delete instance
 * @param instanceId - instance id
 */
const handleDeleteInstance = (instanceId: string) => {
  ElMessageBox.confirm(t('mcp.instance.action.deleteBox'), t('common.warn'), {
    confirmButtonText: t('common.ok'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
    customClass: 'tips-box',
    center: true,
    showClose: false,
    confirmButtonClass: 'is-plain el-button--danger danger-btn',
    customStyle: {
      width: '517px',
      height: '247px',
    },
  })
    .then(async () => {
      await InstanceAPI.delete(instanceId)
      ElMessage({
        type: 'success',
        message: t('action.delete'),
      })
      init()
    })
    .catch(() => {})
}

/**
 * Collect more opearation
 * @param callback - function name
 * @param row - item of instance data
 */
const handleCommand = (callback: string, row: any) => {
  switch (callback) {
    case 'handleViewLog':
      handleViewLog(row)
      break
    case 'handleEditInstance':
      handleEditInstance(row)
      break
    case 'handleStopInstance':
      handleStopInstance(row.instanceId)
      break
    case 'handleRestartInstance':
      handleRestartInstance(row.instanceId)
      break
    case 'handleViewStatus':
      handleViewStatus(row)
      break
    case 'handleDeleteInstance':
      handleDeleteInstance(row.instanceId)
      break
    case 'handleAddInstance':
      handleAddInstance(null)
      break
    case 'handleAddByTemplate':
      handleAddByTemplate()
      break
    default:
      ElMessage.warning(`未找到 "${callback}" 对应的操作`)
  }
}

/**
 * Handle get count data
 */
const handleGetCount = async () => {
  const data = await InstanceAPI.count()
  instanceCount.value = data
}

/**
 * Init data for page
 */
const init = () => {
  if (query.type) {
    handleSearchByCount(query.type as string)
  } else {
    tablePlus.value.initData()
  }
  handleGetCount()
}

onBeforeUnmount(() => {
  if (timer.value) {
    clearInterval(timer.value)
  }
})

onMounted(() => {
  init()
  if (timer.value) {
    return
  } else {
    timer.value = setInterval(init, 30000)
  }
})
</script>
<style lang="scss" scoped>
.page-header {
  margin-bottom: 24px;
  .header-container {
    font-size: 20px;
  }
}
.desc {
  font-size: 16px;
  color: #999999;
  margin-left: 16px;
}
.data-card {
  width: 338px;
  height: 124px;
  background: url('@/assets/images/data-card.png') no-repeat center center;
  background-size: 100% 100%;
  margin-bottom: 24px;
  cursor: pointer;
  color: white;
  .count {
    font-size: 40px;
  }
}

.danger-btn {
  width: 100px;
  border-radius: 4px;
  &.el-button {
    background-color: rgba(12, 25, 207, 0.08) !important;
  }
}
.option-item {
  width: 580px;
}
</style>
