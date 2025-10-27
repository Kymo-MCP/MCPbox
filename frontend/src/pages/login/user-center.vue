<template>
  <div class="center">
    <div class="user-setting">
      <div>{{ t('login.userSetting') }}</div>
      <el-divider />
      <div class="flex">
        <div class="flex-sub">
          <div class="mb-6">{{ t('login.changePassword') }}</div>
          <el-form ref="formRef" :model="formData" :rules="rules" label-position="top">
            <!-- 用户名 -->
            <el-form-item prop="oldPassword" :label="t('login.currentPw')">
              <el-input
                v-model.trim="formData.oldPassword"
                :placeholder="t('login.password')"
                type="password"
                show-password
                disabled
              >
              </el-input>
            </el-form-item>
            <el-form-item prop="newPassword" :label="t('login.newPassword')">
              <el-input
                v-model.trim="formData.newPassword"
                :placeholder="t('login.message.password.required')"
                type="password"
                show-password
              >
              </el-input>
            </el-form-item>
            <el-form-item prop="confirmPassword" :label="t('login.confirmPassword')">
              <el-input
                v-model.trim="formData.confirmPassword"
                :placeholder="t('login.message.password.confirm')"
                type="password"
                show-password
              >
              </el-input>
            </el-form-item>
          </el-form>
          <div class="mt-24">
            <el-button plain @click="handleCancel">{{ t('common.cancel') }}</el-button>
            <el-button type="primary" class="base-btn" @click="handleSavePassword">
              {{ t('common.save') }}</el-button
            >
          </div>
        </div>
        <div class="flex-sub avatar">
          <div class="mb-6">{{ t('login.avatar') }}</div>
          <div class="mt-12" v-if="!imageUrl">
            <el-upload
              class="avatar-uploader"
              :action="action"
              :show-file-list="false"
              :headers="headers"
              method="PUT"
              name="image"
              :on-success="handleAvatarSuccess"
              accept=".png,.jpg,.jpeg"
            >
              <el-icon class="avatar-uploader-icon"><Plus /></el-icon>
            </el-upload>
          </div>
          <div class="mt-12 user-avatar cursor-pointer" v-else>
            <McpImage :src="imageUrl"> </McpImage>
            <div class="avatar-overlay center" @click="handleDelAvatar">
              <el-icon class="delete-icon">
                <Delete />
              </el-icon>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useUserStoreHook, useUserStore } from '@/stores'
import baseConfig from '@/config/base_config.ts'
import { Plus, Delete } from '@element-plus/icons-vue'
import { Storage } from '@/utils/storage'
import AuthAPI from '@/api/auth-api.ts'
import { ElMessageBox } from 'element-plus'
import { useRouterHooks } from '@/utils/url'
import McpImage from '@/components/mcp-image/index.vue'

const action = ref(
  baseConfig.SERVER_BASE_URL + baseConfig.baseUrlVersion + '/authz/users/update-avatar',
)
const { t } = useI18n()
const headers = ref({
  Authorization: `Bearer ${Storage.get('token')}`,
})
const { loginFormData } = toRefs(useUserStoreHook())
const formData = ref<any>({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})
const rules = ref({
  newPassword: [{ required: true, message: t('mcp.instance.rules.name'), trigger: 'blur' }],
  confirmPassword: [{ required: true, message: t('mcp.instance.rules.name'), trigger: 'blur' }],
})
const imageUrl = ref('')
const { logout } = useUserStore()
const { jumpBack } = useRouterHooks()

// 上传成功
const handleAvatarSuccess = (
  response: { code: number },
  uploadFile: { raw: Blob | MediaSource },
) => {
  if (response.code !== 0) {
    return
  }
  imageUrl.value = URL.createObjectURL(uploadFile.raw!)
}

/**
 * 取消
 */
const handleCancel = () => {
  jumpBack()
}

/**
 * 保存修改密码
 */
const handleSavePassword = async () => {
  await AuthAPI.changePassword({
    ...formData.value,
  })
  ElMessageBox.alert('密码修改成功，需要重新登录', t('common.warn'), {
    confirmButtonText: t('common.ok'),
    type: 'warning',
    customClass: 'tips-box',
    center: true,
    showClose: false,
    confirmButtonClass: 'base-btn',
    customStyle: {
      width: '517px',
      height: '247px',
    },
  }).then(() => {
    logout()
  })
}

// 删除头像
const handleDelAvatar = () => {
  imageUrl.value = ''
}

/**
 * 初始化
 */
const init = () => {
  formData.value.oldPassword = loginFormData.value.password
}

onMounted(init)
</script>

<style lang="scss" scoped>
.user-setting {
  width: 1000px;
  margin-top: 80px;
  .avatar {
    margin-left: 70px;
    .avatar-uploader {
      width: 134px;
      height: 134px;
      border-radius: 50%;
      border: 1px dashed var(--ep-purple-color);
      .el-icon.avatar-uploader-icon {
        font-size: 28px;
        color: var(--ep-purple-color);
        width: 134px;
        height: 134px;
        text-align: center;
      }
    }
  }
  .user-avatar {
    position: relative;
    width: 134px;
    height: 134px;
    border-radius: 50%;
    .avatar-overlay {
      display: none;
    }
  }
  .user-avatar:hover {
    background: rgba(0, 0, 0, 0.5);
    .avatar-overlay {
      display: block;
      position: absolute;
      top: calc(50% - 7px);
      left: calc(50% - 7px);
      width: 100%;
      height: 100%;
      transition: opacity 0.3s ease; // 平滑过渡效果
      border-radius: 50%; // 与头像保持一致的圆角
      .delete-icon {
        cursor: pointer;
        transition: color 0.2s ease; // 图标颜色过渡
      }
    }
  }
}
</style>
