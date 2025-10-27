<template>
  <el-dialog v-model="dialogInfo.visible" width="480px" align-center :show-close="false">
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
    <el-scrollbar
      v-loading="dialogInfo.loading"
      ref="scrollbarRef"
      max-height="75vh"
      always
      class="pr-4"
    >
      <div class="center">
        {{ t('mcp.instance.type') }}：{{
          accessTypeOptions.find((access) => access.value === dialogInfo.instanceInfo.accessType)
            ?.label
        }}
      </div>
      <div>
        <template v-for="(probe, index) in probeStatus">
          <div class="flex justify-between status" :key="index" v-if="!probe.hidden">
            <span>{{ probe.label }}</span>
            <el-text v-if="probe.status" type="success" class="point center">
              {{ t('mcp.instance.probe.ready') }}
            </el-text>
            <el-text v-else style="color: #ffa0a0" class="point center">
              {{ t('mcp.instance.probe.unready') }}
            </el-text>
          </div>
        </template>
      </div>
      <div>
        <div class="flex justify-between status">
          <span>{{ t('mcp.instance.probe.errorInfo') }}</span>
        </div>
        <div class="error-info" v-if="dialogInfo.probeInfo.errorMessage">
          {{ dialogInfo.probeInfo.errorMessage }}
        </div>
        <el-empty class="error-info" v-else></el-empty>
      </div>
      <div
        v-if="
          dialogInfo.instanceInfo.accessType === 3 && dialogInfo.probeInfo.warningEvents?.length
        "
      >
        <el-collapse v-model="isOpenEvent" @change="isOpenEvent = !isOpenEvent">
          <el-collapse-item title="Consistency" name="1">
            <template #title>
              <div class="flex justify-between status">
                <span>
                  {{ t('mcp.instance.probe.warnInfo') }}
                  <span style="color: #ffa0a0">
                    （{{ dialogInfo.probeInfo.warningEvents?.length
                    }}{{ t('mcp.instance.probe.unit') }}）
                  </span>
                </span>
              </div>
            </template>
            <div
              class="warn-event"
              v-for="(warnEvent, index) in dialogInfo.probeInfo.warningEvents"
              :key="index"
            >
              <div class="flex justify-between align-center reason-header">
                <div class="title">{{ warnEvent.type }}</div>
                <span>{{ timestampToDate(warnEvent.lastTimestamp * 1000) }}</span>
              </div>
              {{ warnEvent.message }}
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-scrollbar>
    <template #footer>
      <div class="center">
        <el-button class="w100 mr-2" @click="dialogInfo.visible = false">{{
          t('common.close')
        }}</el-button>

        <mcp-button @click="handleProbe" class="w100">{{ t('common.refresh') }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { useInstanceTableHooks } from '../hooks/index.ts'
import { timestampToDate } from '@/utils/system'
import { useMcpStoreHook } from '@/stores'
import McpButton from '@/components/mcp-button/index.vue'

const { InstanceAPI } = useInstanceTableHooks()
const { t } = useI18n()
const isOpenEvent = ref(false)
const dialogInfo = ref<any>({
  visible: false,
  loading: false,
  title: t('mcp.instance.probe.title'),
  instanceInfo: {
    instanceId: '',
  },
  probeInfo: {},
})

const { accessTypeOptions } = useMcpStoreHook()

/**
 * computed the instance probe status
 */
const probeStatus = computed(() => [
  {
    label: t('mcp.instance.probe.containerStatus'),
    status: dialogInfo.value.probeInfo.containerReady,
    hidden: dialogInfo.value.instanceInfo.accessType !== 3,
  },
  {
    label: t('mcp.instance.probe.serviceReadyStatus'),
    status: dialogInfo.value.probeInfo.serviceReady,
    hidden: dialogInfo.value.instanceInfo.accessType !== 3,
  },
  {
    label: t('mcp.instance.probe.probeHttpStatus'),
    status: dialogInfo.value.probeInfo.probeHttp,
    hidden: false,
  },
])

/**
 * Handle get probe info
 */
const handleProbe = async () => {
  try {
    dialogInfo.value.loading = true
    const data = await InstanceAPI.status(dialogInfo.value.instanceInfo.instanceId)
    dialogInfo.value.probeInfo = data
  } finally {
    dialogInfo.value.loading = false
  }
}

/**
 * Handle init probe model data
 * @param instanceInfo - instance object
 */
const init = (instanceInfo: { [key: string]: string }) => {
  dialogInfo.value.instanceInfo = instanceInfo
  dialogInfo.value.probeInfo = {}
  dialogInfo.value.visible = true
  handleProbe()
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.status {
  margin: 24px 0;
}
.point {
  &::before {
    content: '';
    display: block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 8px;
    background-color: currentColor;
  }
}
.error-info {
  color: #ffa0a0;
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  border-radius: 8px;
  background: #000000;
  border-radius: 8px;
  padding: 16px;
}
.warn-event {
  background: rgb(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 8px;
  .reason-header {
    margin-bottom: 12px;
    > div {
      font-size: 16px;
      line-height: 22px;
    }
    > span {
      font-size: 14px;
    }
  }
}
.open-warns {
  // height: 100px;
  transition: all 0.3s;
}
.close-warns {
  height: 0;
  transition: all 0.3s;
  overflow-y: hidden;
}
</style>
