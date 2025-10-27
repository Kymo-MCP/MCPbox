<template>
  <div class="common-layout">
    <el-container class="h-full">
      <el-header class="p-0 flex align-center">
        <div class="title">
          <div>{{ t('code.action.upload') }}</div>
          <div class="flex desc align-end">
            <div class="ml-4">ZIP {{ t('code.desc.compressed') }} (.zip)</div>
            <div class="ml-2">TAR {{ t('code.desc.compressed') }} (.tar, .tar.gz)</div>
          </div>
        </div>
      </el-header>
      <el-main>
        <div class="flex flex-direction h-full">
          <div class="flex-sub center mt-8 link-hover">
            <el-upload
              class="upload-demo"
              drag
              :action="action"
              multiple
              :on-success="handleSuccess"
              :headers="headers"
              accept=".zip, .tar, .tar.gz, application/zip, application/x-tar, application/gzip"
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">
                {{ t('code.desc.suport') }}
              </div>
            </el-upload>
          </div>

          <div class="footer">
            {{ t('code.desc.describe') }}
            <div class="desc">
              <div class="ml-8 mt-2">{{ t('code.desc.text1') }}</div>
              <div class="ml-8 mt-2">{{ t('code.desc.text2') }}</div>
              <div class="ml-8 mt-2">{{ t('code.desc.text3') }}</div>
              <div class="ml-8 mt-2">{{ t('code.desc.text4') }}</div>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { UploadFilled } from '@element-plus/icons-vue'
import { ref } from 'vue'
import baseConfig from '@/config/base_config.ts'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Storage } from '@/utils/storage'

const { t } = useI18n()
const action = ref(baseConfig.SERVER_BASE_URL + baseConfig.baseUrlVersion + '/market/code/upload')
const headers = ref({
  Authorization: `Bearer ${Storage.get('token')}`,
})
const handleSuccess = (response: { code: number; data: { path: string } }) => {
  if (response.code !== 0) {
    return
  }
  ElMessage.success(t('action.upload'))
}
</script>

<style lang="scss" scoped>
.common-layout {
  width: 100vm;
  height: 100%;
  .el-header {
    padding: 0;
  }
  .el-main {
    padding: 0;
  }
  .el-footer {
    padding: 0;
  }
}
.title {
  display: flex;
  font-size: 20px;
  font-weight: 600;
  .desc {
    font-size: 16px;
    color: #999999;
    font-weight: 400;
  }
}
.upload-demo {
  width: 540px;
  color: var(--ep-purple-color);
  :deep(.el-upload-dragger) {
    border: 1px dashed var(--ep-purple-color);
    &:hover {
      border-color: var(--ep-purple-color-hover);
      .el-icon--upload {
        color: var(--ep-purple-color-hover);
      }
      .el-upload__text {
        color: var(--ep-purple-color-hover);
      }
    }
  }
  .el-icon--upload {
    color: var(--ep-purple-color);
  }
  .el-upload__text {
    color: var(--ep-purple-color);
  }
}
.footer {
  font-family:
    PingFangSC,
    PingFang SC;
  font-weight: 600;
  font-size: 20px;
  // color: #cccccc;
  line-height: 28px;
  .desc {
    font-family:
      PingFangSC,
      PingFang SC;
    font-weight: 400;
    font-size: 14px;
    color: #999999;
    line-height: 20px;
    text-align: left;
    font-style: normal;
  }
}
</style>
