<template>
  <div class="mt-2" v-if="!imageUrl">
    <el-upload
      class="avatar-uploader"
      :action="action"
      :show-file-list="false"
      :headers="headers"
      method="POST"
      name="image"
      :on-success="handleAvatarSuccess"
      accept=".png,.jpg,.jpeg"
    >
      <el-icon class="avatar-uploader-icon"><Plus /></el-icon>
    </el-upload>
  </div>
  <div class="mt-2 user-avatar cursor-pointer" v-else>
    <McpImage :src="imageUrl" width="90" height="90"></McpImage>
    <div class="avatar-overlay center" @click="handleDelAvatar">
      <el-icon class="delete-icon" color="#F56C6C">
        <Delete />
      </el-icon>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
import { ref, watch, withDefaults } from 'vue'
import baseConfig from '@/config/base_config.ts'
import { Storage } from '@/utils/storage'
import { Delete } from '@element-plus/icons-vue'
import McpImage from '../mcp-image/index.vue'

const action = ref(baseConfig.SERVER_BASE_URL + baseConfig.baseUrlVersion + '/market/storage/image')
const headers = ref({
  Authorization: `Bearer ${Storage.get('token')}`,
})
const imageUrl = ref('')
const props = withDefaults(
  defineProps<{
    modelValue?: string

    method?: 'POST' | 'PUT' | 'PATCH'

    name?: string
  }>(),
  {
    method: 'POST',
    name: 'file',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'success', response: any): void
  (e: 'error', error: any): void
}>()
imageUrl.value = props.modelValue || ''

const handleAvatarSuccess = (
  response: { code: number; data: { path: string } },
  uploadFile: { raw: Blob | MediaSource },
) => {
  if (response.code !== 0) {
    return
  }
  imageUrl.value = URL.createObjectURL(uploadFile.raw!)
  if (imageUrl.value) {
    emit('update:modelValue', response.data.path)
  }
  emit('success', response)
}

const handleDelAvatar = () => {
  imageUrl.value = ''
}

watch(
  () => props.modelValue,
  (newVal) => {
    imageUrl.value = newVal || ''
  },
)
</script>

<style lang="scss" scoped>
.avatar-uploader {
  width: 90px;
  height: 90px;
  border-radius: 12px;
  border: 1px dashed var(--ep-purple-color);
  .el-icon.avatar-uploader-icon {
    font-size: 28px;
    color: var(--ep-purple-color);
    width: 90px;
    height: 90px;
    text-align: center;
  }
}
.user-avatar {
  position: relative;
  width: 90px;
  height: 90px;
  border-radius: 12px;
  .avatar-overlay {
    display: none;
  }
}
.user-avatar:hover {
  background: var(--ep-bg-stripe-light);
  .avatar-overlay {
    display: block;
    position: absolute;
    top: calc(50% - 7px);
    left: calc(50% - 7px);
    width: 100%;
    height: 100%;
    transition: opacity 0.3s ease; // 平滑过渡效果
    .delete-icon {
      cursor: pointer;
    }
  }
}
</style>
