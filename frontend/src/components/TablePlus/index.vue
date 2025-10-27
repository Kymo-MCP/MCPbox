<template>
  <Teleport v-if="props.searchContainer && isMounted" :to="props.searchContainer">
    <div class="flex">
      <el-input
        v-model="formData[formConfig[0].key]"
        v-bind="{ ...formConfig[0].props }"
        @keyup.enter="handleQuery"
        :suffix-icon="Search"
      ></el-input>
      <el-button style="width: 32px" class="ml-3" @click="initData">
        <el-icon><Refresh /></el-icon>
      </el-button>
      <el-popover
        v-model:visible="showMoreSearch"
        placement="bottom"
        :width="464"
        trigger="click"
        :show-arrow="false"
        append-to-body
        style="padding: 24px"
      >
        <template #reference>
          <el-button style="width: 32px">
            <el-icon><i class="icon iconfont MCP-shaixuan1"></i></el-icon>
          </el-button>
        </template>

        <FormPlus
          ref="searchFromRef"
          :form-config="formConfig"
          :form-data="formData"
          name="formPlus"
          label-width="108"
          label-position="left"
        >
          <template #handler>
            <el-form-item class="search-buttons flex-sub text-right">
              <div class="flex-sub flex justify-end">
                <el-button @click="resetFields" class="mr-2">{{ t('common.reseat') }}</el-button>
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
                  <el-button type="primary" @click="handleQuery" class="base-btn">{{
                    t('common.ok')
                  }}</el-button>
                </GlareHover>
              </div>
            </el-form-item>
          </template>
        </FormPlus>
      </el-popover>
    </div>
  </Teleport>

  <slot name="action"></slot>
  <el-table
    ref="dataTableRef"
    v-loading="loading"
    :data="list"
    highlight-current-row
    stripe
    :header-cell-style="{
      'background-color': 'var(--ep-bg-grey)',
    }"
    class="data-table__content"
  >
    <el-table-column
      v-for="(column, index) in props.columns"
      :key="index || column.dataIndex"
      :label="column.label"
      :prop="column.dataIndex"
      align="left"
      min-width="100"
      v-bind="{ ...column.props }"
    >
      <template #default="scope">
        <!-- 优先使用插槽 -->
        <slot :name="column.dataIndex" :row="scope.row" :index="scope.$index">
          <!-- 其次使用 customRender -->
          <template v-if="typeof column.customRender === 'function'">
            <!-- 用 component 动态渲染 h 函数返回的 VNode -->
            <component
              :is="column.customRender({ row: scope.row, index: scope.$index })"
              v-if="isVNode(column.customRender({ row: scope.row, index: scope.$index }))"
            />
            <!-- 普通值（字符串/数字等）直接显示 -->
            <template v-else>
              {{ column.customRender({ row: scope.row, index: scope.$index }) }}
            </template>
          </template>

          <!-- 最后使用默认值 -->
          <span v-else>
            {{ scope.row[column.dataIndex] || '--' }}
          </span>
        </slot>
      </template>
    </el-table-column>
    <el-table-column
      v-if="props.showOperation"
      :fixed="props.handlerColumnConfig?.fixed || 'right'"
      :width="props.handlerColumnConfig?.width || '240px'"
      :label="t('common.operation')"
    >
      <template #default="scope">
        <slot name="operation" :row="scope.row" :index="scope.$index"></slot>
      </template>
    </el-table-column>
    <template #empty>
      <el-empty :image-size="200" :description="t('status.noData')" />
    </template>
  </el-table>
  <div class="mt-8 flex justify-end" v-if="showPage">
    <el-pagination
      background
      :total="_pagerConfig.total"
      :page="_pagerConfig.page"
      :limit="_pagerConfig.pageSize"
      @current-change="handlePageChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, withDefaults, isVNode } from 'vue'
import { useI18n } from 'vue-i18n'
import { cloneDeep } from 'lodash-es'
import FormPlus from '../FormPlus/index.vue'
import { Search, Refresh } from '@element-plus/icons-vue'
import GlareHover from '../Animation/GlareHover.vue'

const { t } = useI18n()
interface RequestConfig {
  // eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
  api: Function
  searchQuery: {
    model: Record<string, any>
  }
}

interface TableColumn {
  label: string
  dataIndex: string
  customRender?: (params: { row: Record<string, any>; index: number }) => any
  [key: string]: any
}

interface PageConfig {
  page: number
  pageSize: number
  total: number
}
interface HandlerColumnConfig {
  width: string | null
  fixed: string | null
}

const props = withDefaults(
  defineProps<{
    requestConfig: RequestConfig
    columns: TableColumn[]
    pageConfig: PageConfig
    showOperation: boolean
    searchContainer: string
    handlerColumnConfig: HandlerColumnConfig | null | undefined
    // eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
    queryFormatter?: Function
    showPage?: boolean
  }>(),
  {
    showPage: () => true,
  },
)

const showMoreSearch = ref(false)
const loading = ref(false)
const list = ref<unknown[]>([])
const dataTableRef =
  ref<InstanceType<(typeof import('element-plus/lib/components/table/src/table.vue'))['default']>>()
const _pagerConfig = ref(Object.assign({}, props.pageConfig))

const emit = defineEmits<{
  (e: 'update:pageConfig', value: PageConfig): void
  (e: 'resetFields', value: any): void
}>()
const handlePageChange = (newPage: number) => {
  _pagerConfig.value.page = newPage
  initData()
}

/**
 * init search data config
 */
const formConfig = computed(() => {
  const config = cloneDeep(props.columns)
    .filter((f: any) => f.searchConfig)
    .map((o: any) => {
      return {
        key: o.dataIndex,
        ...o.searchConfig,
        span: o.searchConfig.span || 4,
      }
    })
  config.push({
    span: 5,
    component: 'slot',
    slotName: 'handler',
  })
  return config
})

/**
 * init form data
 */
const formData = ref<any>({})
const initFormData = () => {
  cloneDeep(props.columns)
    .filter((f: any) => f.searchConfig)
    .forEach((o: any) => {
      if (!formData.value.hasOwnProperty.call(o.dataIndex)) formData.value[o.dataIndex] = undefined
    })
}

/**
 * Handle search event
 */
const handleQuery = () => {
  showMoreSearch.value = false
  initData()
}

/**
 * reset form data
 */
const searchFromRef = ref()
const resetFields = () => {
  showMoreSearch.value = false
  initFormData()
  searchFromRef.value?.resetFields()
  emit('resetFields', null)
  initData()
}

//Init search data
const initSearchQuery = () => {
  if (props.queryFormatter) {
    props.queryFormatter(formData.value)
  }
}

const customize = (searchData: object) => {
  formData.value = Object.assign({ ...formData.value }, searchData)
}

/**
 * Handle init table data
 */
const initData = async () => {
  await initSearchQuery()
  try {
    loading.value = true
    const data = await props.requestConfig.api({
      page: props.pageConfig.page,
      pageSize: props.pageConfig.pageSize,
      ...formData.value,
      ...props.requestConfig.searchQuery.model,
    })

    list.value = data.list
    _pagerConfig.value.total = Number(data.total) || data.list?.length || 0
    emit('update:pageConfig', _pagerConfig.value)
  } catch (error) {
    console.error('数据加载失败', error)
  } finally {
    loading.value = false
  }
}

const isMounted = ref(false)
onMounted(() => {
  isMounted.value = true
  initFormData()
})
defineExpose({
  initData,
  resetFields,
  customize,
})
</script>

<style lang="scss">
.el-popover.el-popper {
  padding: 24px 24px 12px !important;
}
.el-table__body tr.current-row > td.el-table__cell {
  background-color: var(--ep-bg-purple-color);
}
.el-table--striped .el-table__body tr.el-table__row--striped.current-row td.el-table__cell {
  background-color: var(--ep-bg-purple-color);
}
</style>
