<template>
  <div v-loading="loading" class="container-earth" id="el-main-home">
    <Global></Global>

    <div class="flex justify-between flex-sub">
      <AnimatedContent
        :distance="100"
        direction="horizontal"
        :reverse="true"
        :duration="0.8"
        ease="power3.out"
        :initial-opacity="0"
        :animate-opacity="true"
        :scale="1"
        :threshold="0.1"
        :delay="0"
      >
        <div class="falls flex flex-direction">
          <div
            class="flex align-center justify-center flex-sub flex-direction cursor-pointer link-hover"
            @click="handleJumpTopage(count.type)"
            v-for="(count, index) in countList"
            :key="index"
          >
            <div class="text-40" v-count-to="count.data"></div>
            <div class="text-16">{{ count.title }}</div>
          </div>
        </div>
      </AnimatedContent>
      <div class="text-40">MCP BOX</div>
      <AnimatedContent
        :distance="100"
        direction="horizontal"
        :reverse="false"
        :duration="0.8"
        ease="power3.out"
        :initial-opacity="0"
        :animate-opacity="true"
        :scale="1"
        :threshold="0.1"
        :delay="0"
      >
        <div class="falls flex flex-direction">
          <div class="case-title center">{{ t('home.case') }}</div>
          <div class="flex-sub flex flex-direction">
            <div
              class="flex align-center flex-sub flex-direction case-card cursor-pointer link-hover"
              v-for="(market, index) in marketList"
              :key="index"
              @click="handleJumpToMcp(market.link)"
            >
              <div class="flex align-center w-full">
                <el-image
                  :src="market.icon"
                  fit="cover"
                  style="width: 40px; height: 40px; border-radius: 4px"
                >
                </el-image>
                <span class="title ellipsis-one">{{ market.name }}</span>
              </div>
              <div class="w-full ellipsis-two desc">
                {{ market.desc }}
              </div>
            </div>
          </div>
        </div>
      </AnimatedContent>
    </div>
    <AnimatedContent
      :distance="100"
      direction="vertical"
      :reverse="false"
      :duration="0.8"
      ease="power3.out"
      :initial-opacity="0"
      :animate-opacity="true"
      :scale="1"
      :threshold="0.1"
      :delay="0"
    >
      <div class="footer-case flex flex-direction">
        <div class="flex justify-between">
          <div></div>
          <div>{{ t('home.usedCase') }}</div>
          <el-button link @click="handleToTemplate">
            {{ t('home.more') }} <el-icon class="el-icon--right"><Right /></el-icon>
          </el-button>
        </div>
        <div class="flex-sub flex justify-between">
          <el-row :gutter="40" class="w-full">
            <el-col :span="6" v-for="itemCase in caseList" :key="itemCase.id">
              <div
                class="center flex-sub flex-direction case-card link-hover cursor-pointer"
                @click="handleToTemplateForm(itemCase.templateId)"
              >
                <el-tooltip class="box-item" effect="dark" placement="top">
                  <div class="flex align-center w-full">
                    <McpImage
                      :src="itemCase.iconPath"
                      width="40"
                      height="40"
                      border-radius="4"
                    ></McpImage>

                    <span class="title ellipsis-one flex-sub">{{ itemCase.name }}</span>
                  </div>
                  <template #content>
                    {{ itemCase.name }}
                  </template>
                </el-tooltip>
                <el-tooltip class="box-item" effect="dark" placement="top">
                  <div class="w-full ellipsis-three desc">
                    {{ itemCase.description }}
                  </div>
                  <template #content>
                    <div class="w-full ellipsis-three desc" style="width: 400px">
                      {{ itemCase.description }}
                    </div>
                  </template>
                </el-tooltip>
              </div>
            </el-col>
          </el-row>
        </div>
      </div>
    </AnimatedContent>
  </div>
</template>

<script setup lang="ts">
import { Right } from '@element-plus/icons-vue'
import { InstanceAPI } from '@/api/mcp/instance'
import { TemplateAPI } from '@/api/mcp/template'
import { useRouterHooks } from '@/utils/url'
import { bigmodel, modelscope, higress, mcpso, smithery } from '@/utils/logo'
import AnimatedContent from '@/components/Animation/AnimatedContent.vue'
import Global from '@/components/Animation/global.vue'
import McpImage from '@/components/mcp-image/index.vue'

const { jumpToPage } = useRouterHooks()
const { t } = useI18n()
const loading = ref(false)
const instanceCount = ref<any>({
  totalEnvironments: 0,
})
// market list
const marketList = ref([
  {
    name: 'BigModel',
    icon: bigmodel,
    desc: t('home.market.bigModel'),
    link: 'https://bigmodel.cn/',
  },
  {
    name: 'Modelscope',
    icon: modelscope,
    desc: t('home.market.modelScope'),
    link: 'https://www.modelscope.cn/',
  },
  {
    name: 'Higressai',
    icon: higress,
    desc: t('home.market.higress'),
    link: 'https://mcp.higress.ai/',
  },
  { name: 'Mcpso', icon: mcpso, desc: t('home.market.mcpso'), link: 'https://mcp.so/' },
  {
    name: 'Smithery',
    icon: smithery,
    desc: t('home.market.smithery'),
    link: 'https://smithery.ai/',
  },
])

// count instance data
const countList = computed(() => {
  return [
    { title: t('home.total'), data: instanceCount.value.totalInstances || 0, type: 'total' },
    { title: t('home.start'), data: instanceCount.value.activeInstances || 0, type: 'active' },
    { title: t('home.stop'), data: instanceCount.value.inactiveInstances || 0, type: 'inactive' },
    { title: t('home.hosting'), data: instanceCount.value.hostingInstances || 0, type: 'hosting' },
  ]
})

// case list
const caseList = ref<any>([])

/**
 * Jump to instance list with search data
 * @param type
 */
const handleJumpTopage = (type: string) => {
  jumpToPage({
    url: '/instance-manage',
    data: { type },
  })
}
/**
 * jump to template list page
 */
const handleToTemplate = () => {
  jumpToPage({
    url: '/template-manage',
    data: {},
  })
}
/**
 * jump to create instance page
 */
const handleToTemplateForm = (templateId: string) => {
  jumpToPage({
    url: '/new-instance',
    data: { templateId },
  })
}

/**
 * jump to others MCP
 */
const handleJumpToMcp = (url: string) => {
  jumpToPage({
    url,
    isOpen: true,
  })
}

/**
 * Handle get count Data
 */
const handleGetCount = async () => {
  const data = await InstanceAPI.count()
  instanceCount.value = data
}

/**
 * Handle get instance list
 */
const handleGetCases = async () => {
  const data = await TemplateAPI.cases()
  caseList.value = data.cases || []
}

/**
 * Handle init page data
 */
const init = () => {
  handleGetCount()
  handleGetCases()
}

onMounted(init)
</script>

<style scoped lang="scss">
.container-earth {
  width: 100%;
  height: 100%;
  padding: 40px;
  box-sizing: border-box;
  overflow: hidden;
  // background: url('@/assets/images/global-plain.png') no-repeat center 20vh;
  // background-size: 665px 644px;
  position: relative;
  display: flex;
  flex-direction: column;
  .falls {
    background: var(--ep-home-glass);
    border-radius: 8px;
    border: 1px solid var(--ep-border-color);
    width: 262px;
    height: 100%;
    padding: 16px;
    .case-title {
      height: 60px;
    }
    .case-card {
      padding: 8px 0;
      .title {
        font-size: 18px;
        line-height: 28px;
        margin-left: 12px;
      }
      .desc {
        font-family:
          PingFangSC,
          PingFang SC;
        font-weight: 400;
        font-size: 14px;
        color: #666666;
        line-height: 20px;
        text-align: left;
        font-style: normal;
      }
    }
  }
  .footer-case {
    height: 268px;
    background: var(--ep-home-glass);
    border-radius: 8px;
    border: 1px solid var(--ep-border-color);
    margin-top: 40px;
    padding: 22px 40px 40px;
    .case-card {
      padding: 8px 0;
      height: 100%;
      margin-top: 10px;
      .title {
        font-size: 20px;
        line-height: 28px;
        margin-left: 12px;
      }
      .desc {
        font-family:
          PingFangSC,
          PingFang SC;
        font-weight: 400;
        font-size: 14px;
        height: 60px;
        color: #666666;
        line-height: 20px;
        text-align: left;
        font-style: normal;
        margin-top: 10px;
      }
    }
  }
  .text-40 {
    font-size: 40px;
    font-family:
      PingFangSC,
      PingFang SC;
    font-weight: 500;
  }
  .text-16 {
    font-family:
      PingFangSC,
      PingFang SC;
    font-weight: 400;
    font-size: 16px;
    color: #999999;
    background-color: transparent;
  }
}
</style>
