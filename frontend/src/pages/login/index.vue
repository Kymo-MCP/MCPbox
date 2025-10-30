<template>
  <div class="login-container">
    <div class="action-bar">
      <!-- <el-tooltip :content="t('login.languageToggle')" placement="bottom">
        <LangSelect />
      </el-tooltip> -->
    </div>
    <div class="w-full">
      <el-row justify="center">
        <el-col :span="10" class="text-right">
          <el-image
            style="
              width: 554px;
              height: 553px;
              margin-left: 40%;
              border-radius: 50%;
              background: linear-gradient(180deg, #202021 0%, #000000 100%);
            "
            :src="url"
            fit="fill"
          />
        </el-col>
        <el-col :span="14">
          <div class="flex align-center justify-center h-full login-body">
            <!-- login body -->
            <div class="h-auto sm:w-460px border-rd-10px">
              <div w-full flex flex-col items-center>
                <!-- title -->
                <div class="flex align-center log-login">
                  <el-image style="width: 77px; height: 77px" :src="logo" fit="fill" />
                  <span class="login-title ml-2"> {{ t('login.MCPlogin') }} </span>
                </div>

                <transition name="fade-slide" mode="out-in">
                  <component :is="formComponents[component]" v-model="component" class="w-90%" />
                </transition>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
    <div class="layer1"></div>
    <div class="layer2"></div>
    <div class="layer3"></div>
    <div class="layer4"></div>
  </div>
</template>

<script setup lang="ts">
import logo from '@/assets/logo.png'
import { useUserStore } from '@/stores'
import url from '@/assets/images/global.png'

const userStore = useUserStore()
const { t } = useI18n()

type LayoutMap = 'login' | 'register' | 'resetPwd'

const component = ref<LayoutMap>('login')
const formComponents = {
  login: defineAsyncComponent(() => import('./components/Login.vue')),
  register: defineAsyncComponent(() => import('./components/Register.vue')),
  resetPwd: defineAsyncComponent(() => import('./components/ResetPwd.vue')),
}
const init = async () => {
  await userStore.handleGetEncryptionKey()
}
onMounted(init)
</script>

<style lang="scss" scoped>
@import url('@/styles/star/index.scss');

.login-container {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, #202021 0%, #000000 100%);
  width: 100%;
  height: 100%;
}

.action-bar {
  position: fixed;
  top: 10px;
  right: 10px;
  z-index: 10;
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
}
.login-body {
  color: #fff;
  .log-login {
    margin-bottom: 68px;
    .login-title {
      font-family:
        PingFangSC,
        PingFang SC;
      font-weight: 600;
      font-size: 40px;
      color: #ffffff;
      line-height: 56px;
      text-align: left;
      font-style: normal;
    }
  }
}

/* fade-slide */
.fade-slide-leave-active,
.fade-slide-enter-active {
  transition: all 0.3s;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
