<template>
  <el-dropdown
    ref="addDropdown"
    placement="bottom-start"
    trigger="click"
    :show-arrow="false"
    :hide-on-click="false"
    :virtual-trigger="false"
    :popper-options="props.popperOptions"
    append-to-body
    @command="handleJumpToPage"
    @click.stop
  >
    <slot></slot>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item command="/new-template">
          <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
          {{ t('mcp.template.formData.title') }}
        </el-dropdown-item>
        <el-dropdown-item command="/new-instance">
          <el-dropdown
            trigger="click"
            :popperOptions="{
              modifiers: [
                {
                  name: 'offset',
                  options: {
                    offset: [128, 0], // 向右偏移100px，避免紧贴按钮
                  },
                },
              ],
            }"
            :show-arrow="false"
            @command="handleCommand"
          >
            <div>
              <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
              {{ t('mcp.instance.action.create') }}
            </div>

            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="/new-instance">
                  <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
                  {{ t('mcp.instance.action.customize') }}
                </el-dropdown-item>
                <el-dropdown-item :command="`/new-instance?templateId=${template.templateId}`">
                  <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
                  {{ t('mcp.instance.action.byTemplate') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-dropdown-item>
        <el-dropdown-item command="env">
          <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
          {{ t('env.run.formData.title') }}
        </el-dropdown-item>
        <el-dropdown-item command="/update-code-package">
          <el-icon><i class="icon iconfont MCP-shangchuan"></i></el-icon>
          {{ t('code.action.upload') }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
  <!-- 创建环境 -->
  <NewEnvDialog ref="newEnvDialog"></NewEnvDialog>
  <!-- 选择模板 -->
  <Select
    v-model="template.visible"
    v-model:selected="template.templateId"
    ref="packageSelect"
    :title="t('mcp.instance.action.selectTempalte')"
    :options="template.templateList"
    @confirm="handleConfirmTemplate"
  >
    <template #options="{ option }">
      <div class="flex justify-between">
        <div class="flex align-center w-full mt-2 mb-2">
          <el-image :src="option.iconPath" style="width: 32px; height: 32px"></el-image>
          <el-tooltip effect="dark" placement="top" class="ml-2" :raw-content="true">
            <div class="flex-sub ml-2 ellipsis-one">{{ option.name }}</div>
            <template #content>
              <div style="width: 300px">
                {{ option.name }}
              </div>
            </template>
          </el-tooltip>
        </div>
      </div>
    </template>
  </Select>
</template>

<script lang="ts" setup>
import { useRouterHooks } from '@/utils/url'
import NewEnvDialog from '@/pages/environment/working-environment/modules/new-env-dialog.vue'
import { TemplateAPI } from '@/api/mcp/template'
import Select from '@/components/mcp-select/index.vue'

const newEnvDialog = ref()
const { jumpToPage } = useRouterHooks()
const { t } = useI18n()
const template = ref<any>({
  visible: false,
  templateId: '',
  templateList: [],
})
const addDropdown = ref()

const props = defineProps({
  popperOptions: {
    type: Object,
    default: () => {
      return {
        modifiers: [
          {
            name: 'offset',
            options: {
              offset: [0, 0], // 向右偏移100px，避免紧贴按钮
            },
          },
        ],
      }
    },
  },
})

// jump to sub page
const handleJumpToPage = (url: string) => {
  if (url === '/new-instance') {
    return
  }
  addDropdown.value.handleClose()
  if (url === 'env') {
    newEnvDialog.value.init({})
    return
  }
  jumpToPage({
    url: url,
    data: {},
  })
}

const handleCommand = (cmd: string) => {
  if (cmd === '/new-instance') {
    jumpToPage({
      url: cmd,
      data: {},
    })
    return
  }
  handleAddByTemplate()
}

/**
 * Handle select a template by list
 */
const handleAddByTemplate = async () => {
  template.value.visible = true
  const data = await TemplateAPI.list({ page: '1', pageSize: '999' })
  template.value.templateList = data.list.map((template: any) => ({
    id: template.templateId,
    name: template.name,
    ...template,
  }))
}

/**
 * handle confirm selected template
 */
const handleConfirmTemplate = (templateId: string) => {
  jumpToPage({
    url: '/new-instance',
    data: { templateId },
  })
}
</script>

<style lang="scss" scoped>
.el-menu-item.is-disabled {
  opacity: 1 !important;
  cursor: default !important;
  .el-dropdown {
    width: 100%;
  }
}
</style>
