<template>
  <div v-loading="pageInfo.loading" :element-loading-text="pageInfo.loadingText">
    <!-- 头部区域 -->
    <div class="flex justify-between page-header">
      <div class="header-container">
        {{ t('code.pageDesc.list') }} <span class="desc">{{ t('code.pageDesc.desc') }}</span>
      </div>
      <mcp-button :icon="UploadFilled" @click="handleUpdatePackage">{{
        t('code.action.upload')
      }}</mcp-button>
    </div>

    <TablePlus
      :showOperation="true"
      searchContainer="#templateSearch"
      ref="tablePlus"
      :requestConfig="requestConfig"
      :columns="columns"
      v-model:pageConfig="pageConfig"
      :handlerColumnConfig="{
        width: '120px',
        fixed: 'right',
      }"
    >
      <template #action>
        <div class="flex justify-between mb-4">
          <div class="center">
            <el-image :src="codeLogo" style="width: 20px; height: 20px"></el-image>
            <span class="desc">{{ t('code.pageDesc.total') }}：{{ pageConfig.total }}</span>
          </div>
          <div id="templateSearch"></div>
        </div>
      </template>
      <template #name="{ row }">
        <div class="flex align-center">
          <el-image
            :src="zipLogo"
            style="width: 32px; height: 32px"
            fit="cover"
            class="mr-2"
          ></el-image>
          <el-tooltip effect="dark" placement="top" class="flex-sub" :raw-content="true">
            <div class="flex-sub ml-2 ellipsis-two">{{ row.name }}</div>
            <template #content>
              <div style="width: 300px">
                {{ row.name }}
              </div>
            </template>
          </el-tooltip>
        </div>
      </template>
      <template #operation="{ row }">
        <el-button
          type="text"
          size="small"
          link
          @click="handleViewCode(row)"
          class="base-btn-link"
          >{{ t('common.view') }}</el-button
        >

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
              <el-dropdown-item command="handleDownload">
                {{ t('common.download') }}
              </el-dropdown-item>
              <el-dropdown-item command="handleDelete">
                <el-button type="danger" size="small" link>{{ t('common.delete') }}</el-button>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </TablePlus>
  </div>
</template>

<script setup lang="ts">
import { UploadFilled, More } from '@element-plus/icons-vue'
import { CodeAPI } from '@/api/code/index'
import TablePlus from '@/components/TablePlus/index.vue'
import { useCodeTableHooks } from './hooks'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouterHooks } from '@/utils/url'
import codeLogo from '@/assets/logo/code.png'
import zipLogo from '@/assets/logo/zip.png'
import McpButton from '@/components/mcp-button/index.vue'

const { t, tablePlus, columns, requestConfig, pageConfig, pageInfo } = useCodeTableHooks()
const { jumpToPage } = useRouterHooks()

/**
 * Handle jump to the page of update code package
 */
const handleUpdatePackage = () => {
  jumpToPage({
    url: '/update-code-package',
  })
}

/**
 * Handle view code package info
 */
const handleViewCode = (code: any) => {
  jumpToPage({
    url: '/view-code-package',
    data: {
      id: code.id,
      name: code.name,
    },
  })
}

/**
 * Handle more operation events
 * @param callback - function name
 * @param row - item of code package
 */
const handleCommand = (callback: string, row: any) => {
  switch (callback) {
    case 'handleDownload':
      handleDownload(row)
      break
    case 'handleDelete':
      handleDelete(row)
      break
    default:
      ElMessage.warning(`未找到 "${callback}" 对应的操作`)
  }
}

/**
 * Handle download code package
 */
const handleDownload = async (code: any) => {
  try {
    pageInfo.value.loading = true
    const response = await CodeAPI.download(code)
    const blobUrl = URL.createObjectURL(
      new Blob([response.data], { type: response.headers['content-type'] }),
    )
    const link = document.createElement('a')
    link.href = blobUrl
    link.download =
      response.headers['content-disposition']
        ?.split('filename=')[1]
        ?.match(/filename=("?)(.*?)\1/) || code.name
    document.body.appendChild(link)
    link.click()
  } finally {
    pageInfo.value.loading = false
  }
}

/**
 * Handle delete code package
 */
const handleDelete = (code: any) => {
  ElMessageBox.confirm(`${t('code.action.deleteBox')}“${code.name}”?`, t('common.warn'), {
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
  }).then(async () => {
    await CodeAPI.delete(code.id)
    ElMessage({
      type: 'success',
      message: t('action.delete'),
    })
    init()
  })
}

/**
 * Handle init page list
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
</style>
