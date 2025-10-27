import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { CodeAPI } from '@/api/code/index'
import { timestampToDate, formatFileSize } from '@/utils/system'

export const useCodeTableHooks = () => {
  const { t } = useI18n()

  const tablePlus = ref()
  const pageInfo = ref({
    loading: false,
    loadingText: t('code.action.loadingText'),
  })
  const columns = ref([
    {
      label: t('code.name'),
      dataIndex: 'name',
      searchConfig: {
        component: 'el-input',
        label: t('code.name'),
        props: {
          placeholder: t('code.name'),
        },
      },
    },
    {
      label: t('code.size'),
      dataIndex: 'size',

      customRender: ({ row }: any) => {
        return formatFileSize(row.size)
      },
    },
    {
      label: t('code.columns.type'),
      dataIndex: 'type',
      searchConfig: {
        component: 'el-select',
        label: t('code.columns.type'),
        props: {
          placeholder: t('code.columns.type'),
          options: [
            { label: t('code.columns.tar'), value: 1 },
            { label: t('code.columns.zip'), value: 2 },
          ],
        },
      },
      customRender: ({ row }: any) => {
        return [t('code.columns.unspecified'), t('code.columns.tar'), t('code.columns.zip')][
          row.type
        ]
      },
    },
    {
      dataIndex: 'createdAt',
      label: t('code.columns.createdAt'),
      customRender: ({ row }: any) => {
        return timestampToDate(row.createdAt)
      },
    },
    {
      dataIndex: 'updatedAt',
      label: t('code.columns.updatedAt'),
      customRender: ({ row }: any) => {
        return timestampToDate(row.updatedAt)
      },
    },
  ])
  const requestConfig = {
    api: CodeAPI.list,
    searchQuery: {
      model: {},
    },
  }
  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })

  return { t, columns, requestConfig, tablePlus, pageConfig, pageInfo }
}
