<template>
  <div>
    <!-- 头部区域 -->
    <div class="flex justify-between page-header">
      <div class="header-container">
        {{ t('env.run.pageDesc.list') }} <span class="desc">{{ t('env.run.pageDesc.desc') }}</span>
      </div>
      <mcp-button :icon="Plus" @click="handleNewEnvDialog({})">{{
        t('env.run.action.add')
      }}</mcp-button>
    </div>

    <!-- 表格区域 -->
    <div v-loading="load.status" :element-loading-text="load.text">
      <TablePlus
        ref="tablePlus"
        search-container="#envSearch"
        :showOperation="true"
        :requestConfig="requestConfig"
        :columns="columns"
        :handlerColumnConfig="{
          fixed: 'right',
          width: '120px',
        }"
        v-model:pageConfig="pageConfig"
      >
        <template #action>
          <div class="flex justify-between mb-4">
            <div class="center">
              <el-image :src="envLogo" style="width: 20px; height: 20px"></el-image>
              <span class="desc">{{ t('env.run.pageDesc.total') }}：{{ pageConfig.total }}</span>
            </div>
            <div id="envSearch"></div>
          </div>
        </template>
        <template #createdAt="{ row }">{{ timestampToDate(row.createdAt) }} </template>
        <template #updatedAt="{ row }">{{ timestampToDate(row.updatedAt) }} </template>
        <template #operation="{ row }">
          <div class="flex align-center">
            <el-button
              type="primary"
              size="small"
              link
              class="base-btn-link"
              @click="handleViewDetail(row)"
            >
              {{ t('env.run.action.detail') }}
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
                  <el-dropdown-item command="handleNewEnvDialog">
                    {{ t('env.run.action.edit') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleConnection">
                    {{ t('env.run.action.connection') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleJumpToPvc">
                    {{ t('env.run.action.pvc') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleJumpToNode">
                    {{ t('env.run.action.node') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="handleDelete">
                    <el-button type="danger" link>
                      {{ t('env.run.action.delete') }}
                    </el-button>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </TablePlus>
    </div>

    <!-- create env model -->
    <NewEnvDialog ref="newEnvDialog"></NewEnvDialog>
    <!-- env detail model -->
    <EnvDetail ref="envDetail"></EnvDetail>
  </div>
</template>

<script setup lang="ts">
import TablePlus from '@/components/TablePlus/index.vue'
import { Plus, More } from '@element-plus/icons-vue'
import { timestampToDate } from '@/utils/system'
import { useEnvTableHooks } from './hooks'
import NewEnvDialog from './modules/new-env-dialog.vue'
import EnvDetail from './modules/env-detail.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import envLogo from '@/assets/logo/env.png'
import McpButton from '@/components/mcp-button/index.vue'

const { t } = useI18n()
const {
  load,
  jumpToPage,
  tablePlus,
  requestConfig,
  columns,
  pageConfig,
  newEnvDialog,
  EnvAPI,
  envDetail,
} = useEnvTableHooks()

/**
 * Handle view env info
 * @param row - env info
 */
const handleViewDetail = (row: any) => {
  envDetail.value.init(row)
}

/**
 * Test of connention
 * @param id - env key
 */
const handleConnection = async (id: string) => {
  try {
    load.value.status = true
    load.value.text = t('env.run.load.connect')
    await EnvAPI.testEnv(id)
    ElMessage({
      type: 'success',
      message: t('env.run.action.connectionSuccess'),
    })
  } finally {
    load.value.status = false
  }
}

/**
 * Handle create or edit a env
 * @param formData - current env data
 */
const handleNewEnvDialog = (formData: any) => {
  newEnvDialog.value.init(formData)
}

/**
 * jump to pvc manage page
 * @param row - current env data
 */
const handleJumpToPvc = (row: any) => {
  jumpToPage({
    url: '/pvc-manage',
    data: {
      environmentId: row.id,
      name: row.name,
    },
  })
}

/**
 * jump to node manage
 * @param row - current env data
 */
const handleJumpToNode = (row: any) => {
  jumpToPage({
    url: '/node-manage',
    data: {
      environmentId: row.id,
      name: row.name,
    },
  })
}

/**
 * delete
 * @param row - item of env
 */
const handleDelete = (row: any) => {
  ElMessageBox.confirm(t('env.run.action.deleteBox'), t('common.warn'), {
    confirmButtonText: t('common.sure'),
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
  }).then(async () => {
    await EnvAPI.delete(row.id)
    ElMessage({
      type: 'success',
      message: t('action.delete'),
    })
    init()
  })
}

/**
 * Handle more opeartion events
 * @param callback - function name
 * @param row - item of env
 */
const handleCommand = (callback: string, row: any) => {
  switch (callback) {
    case 'handleNewEnvDialog':
      handleNewEnvDialog(row)
      break
    case 'handleConnection':
      handleConnection(row.id)
      break
    case 'handleJumpToPvc':
      handleJumpToPvc(row)
      break
    case 'handleJumpToNode':
      handleJumpToNode(row)
      break
    case 'handleDelete':
      handleDelete(row)
      break
    default:
      ElMessage.warning(`未找到 "${callback}" 对应的操作`)
  }
}

/**
 * Handle init table list
 */
const init = () => {
  tablePlus.value.initData()
}

onMounted(init)
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
.danger-btn {
  width: 100px;
  border-radius: 4px;
  &.el-button {
    background-color: rgba(12, 25, 207, 0.08) !important;
  }
}
</style>
