<template>
  <div>
    <div class="title flex justify-between">
      <div class="font-bold text-5">{{ t('code.detail') }} - {{ query.name }}</div>
      <div>
        <GlareHover
          width="auto"
          height="auto"
          background="transparent"
          border-color="#222222"
          border-radius="4px"
          glare-color="#ffffff"
          :glare-opacity="0.3"
          :glare-size="300"
          :transition-duration="800"
          :play-once="false"
        >
          <el-button type="primary" @click="handleDownload" class="base-btn">
            <el-icon class="mr-2"><i class="icon iconfont MCP-xiazai"></i></el-icon>
            {{ t('code.action.download') }}</el-button
          >
        </GlareHover>
      </div>
    </div>
    <div class="mt-4">
      <el-row :gutter="32" :style="{ height: 'calc(100vh - 200px)' }">
        <el-col :span="4">
          <el-card :style="{ height: '100%' }">
            <template #header>
              <span>{{ t('code.fileTree') }}</span>
            </template>
            <el-tree
              :style="{ height: 'calc(100% - 40px)', overflow: 'auto' }"
              :data="fileTree"
              :props="defaultProps"
              @node-click="handleNodeClick"
              default-expand-all
              :expand-on-click-node="false"
            >
              <!-- 自定义节点内容 -->
              <template #default="{ node, data }">
                <div class="flex items-center">
                  <!-- 根据节点层级显示不同图标 -->
                  <component :is="getNodeIcon(data)" class="ml-2" :size="4" />
                  <span class="ml-2">{{ data.label }}</span>
                </div>
              </template>
            </el-tree>
          </el-card>
        </el-col>
        <el-col :span="20" v-loading="contentLoading">
          <el-card :style="{ height: '100%' }">
            <template #header>
              <el-space>
                <span>
                  {{ t('code.codeEditor') }}<span class="ml-4">{{ currentDir.label }}</span>
                </span>
              </el-space>
            </template>
            <el-scrollbar ref="scrollbarRef" height="calc(100vh - 300px)" always>
              <div v-if="!currentContent">{{ t('code.empty') }}</div>
              <div class="text-conetnt" v-else>
                <!-- 图片显示 -->
                {{ currentContent }}
              </div>
            </el-scrollbar>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Notebook,
  Document,
  Picture,
  VideoPlay,
  Headset,
  Box as Collection,
} from '@element-plus/icons-vue'
import { CodeAPI } from '@/api/code/index'
import Files from '../modules/files.vue'
import GlareHover from '@/components/Animation/GlareHover.vue'

const { t } = useI18n()
const $route = useRoute()
const { query } = $route
const contentLoading = ref(false)
const currentContent = ref('')
const currentDir = ref<any>({})
const fileTree = ref<any>([])

/**
 * file tree config
 */
const defaultProps = {
  children: 'children',
  label: 'label',
  isLeaf: 'isLeaf',
}

/**
 * Handle icon by fileName
 * @param fileName
 * @returns - return file icon
 */
const getNodeIcon = (nodeData: any) => {
  if (!nodeData.isLeaf) return Files
  const fileName = nodeData.label
  const ext = fileName.split('.').pop()?.toLowerCase() || ''
  switch (ext) {
    case 'js':
    case 'ts':
    case 'jsx':
    case 'tsx':
    case 'py':
    case 'txt':
    case 'xml':
    case 'yml':
      return Document
    case 'png':
    case 'jpg':
    case 'jpeg':
    case 'gif':
    case 'svg':
      return Picture
    case 'mp4':
    case 'mov':
    case 'avi':
      return VideoPlay
    case 'mp3':
    case 'wav':
    case 'flac':
      return Headset
    case 'zip':
    case 'rar':
    case 'tar':
    case 'gz':
      return Collection
    case 'txt':
    case 'md':
    case 'markdown':
      return Notebook
    default:
      return Files // default
  }
}

/**
 * interchange tree data
 */
const convertToTreeData = (nodes: any) => {
  return nodes.map((node: any) => ({
    label: node.name,
    isLeaf: !node.isDir,
    originalData: node, // save origin data
    children: node.children ? convertToTreeData(node.children) : undefined,
  }))
}

/**
 * Handle download code package
 */
const handleDownload = async () => {
  const data = await CodeAPI.download({ ...query })
  const blobUrl = URL.createObjectURL(new Blob([data], { type: 'application/zip' }))
  const link = document.createElement('a')
  link.href = blobUrl
  link.download = `code-package-${query.id}.zip`
  document.body.appendChild(link)
  link.click()
}

/**
 * Handle node click
 */
const handleNodeClick = (data: any) => {
  if (data.originalData.isDir) {
    return
  } else {
    currentDir.value = data
  }
  handleGetContent(data.originalData.path)
}

/**
 * Handle get file content by API
 */
const handleGetContent = async (filePath: string) => {
  try {
    contentLoading.value = true
    const data = await CodeAPI.fileContent({
      packageId: query.id,
      filePath,
    })
    currentContent.value = data.content
  } finally {
    contentLoading.value = false
  }
}

/**
 * Handle get code package file tree by API
 */
const handleGetFileTree = async () => {
  const data = await CodeAPI.fileTree({ packageId: query.id })
  fileTree.value = convertToTreeData([data.fileStructure])
}

/**
 * Handle init page data
 */
const init = () => {
  handleGetFileTree()
}

onMounted(init)
</script>

<style lang="scss" scoped>
.text-conetnt {
  height: calc(100vh - 300px);
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  background-color: #000;
  padding: 10px;
  border-radius: 8px;
}
:deep(.el-card__body) {
  height: calc(100% - 100px);
  overflow-y: auto;
}
/* 图标样式：区分文件夹和文件 */
.folder-icon {
  color: #f59e0b; // 文件夹橙色
  margin-right: 4px;
}
.file-icon {
  color: #60a5fa; // 文件蓝色（可根据需要调整）
  margin-right: 4px;
}
/* 修复树节点文字与图标对齐 */
:deep(.el-tree-node__content) {
  align-items: center !important;
  padding: 4px 0 !important;
}
:deep(.el-tree-node.is-expanded.is-current) {
}
</style>
