<template>
  <div>
    <el-form
      ref="loginFormRef"
      :model="loginFormData"
      :rules="loginRules"
      size="large"
      :validate-on-rule-change="false"
    >
      <!-- 用户名 -->
      <el-form-item prop="username">
        <el-input
          v-model.trim="loginFormData.username"
          :placeholder="t('login.username')"
          :disabled="loading"
        >
          <template #prefix>
            <el-icon><User /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <!-- 密码 -->
      <el-tooltip :visible="isCapsLock" :content="t('login.capsLock')" placement="right">
        <el-form-item prop="password">
          <el-input
            v-model.trim="loginFormData.password"
            :placeholder="t('login.password')"
            type="password"
            show-password
            :disabled="loading"
            @keyup="checkCapsLock"
            @keyup.enter="handleLoginSubmit"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
      </el-tooltip>

      <!-- 登录按钮 -->
      <el-form-item>
        <el-button
          :loading="loading"
          type="primary"
          class="w-full login-btn"
          @click="handleLoginSubmit"
        >
          {{ loading ? t('login.logging') : t('login.login') }}
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
<script setup lang="ts">
import type { FormInstance } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { type LoginFormData } from '@/api/auth-api'
import router from '@/router'
import { useUserStore } from '@/stores'

const { t } = useI18n()
const userStore = useUserStore()
const route = useRoute()

const loginFormRef = ref<FormInstance>()
const loading = ref(false)
// 是否大写锁定
const isCapsLock = ref(false)

const loginFormData = ref<LoginFormData>({
  username: 'admin',
  password: 'admin123',
  encryptedPassword: '',
})

const loginRules = computed(() => {
  return {
    username: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.message.username.required'),
      },
    ],
    password: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.message.password.required'),
      },
      {
        min: 6,
        message: t('login.message.password.min'),
        trigger: 'blur',
      },
    ],
    captchaCode: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.message.captchaCode.required'),
      },
    ],
  }
})

/**
 * 登录提交
 */
async function handleLoginSubmit() {
  try {
    // 1. 表单验证
    const valid = await loginFormRef.value?.validate()
    if (!valid) return

    loading.value = true

    // 2. 执行登录
    await userStore.login(loginFormData.value)

    // const redirectPath = (route.query.redirect as string) || '/'

    await router.push('/home')
  } catch (error) {
  } finally {
    loading.value = false
  }
}

// 检查输入大小写
function checkCapsLock(event: KeyboardEvent) {
  // 防止浏览器密码自动填充时报错
  if (event instanceof KeyboardEvent) {
    isCapsLock.value = event.getModifierState('CapsLock')
  }
}
</script>

<style lang="scss" scoped>
.el-input {
  --el-input-bg-color: rgba(40, 40, 40);
  --el-input-border-color: #dddddd;
  --el-input-hover-border-color: #a083f7;
  --el-input-focus-border-color: #a083f7;
}
.login-btn {
  background: linear-gradient(180deg, #a083f7 0%, #2a029f 100%);
  border: none;
}
.el-input__inner {
  color: #fff;
}
</style>
