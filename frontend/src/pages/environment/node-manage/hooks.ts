import { NodeAPI } from '@/api/env'
import type { nodeResult } from '@/types'

export const usePvcTableHooks = () => {
  const { t } = useI18n()
  const tablePlus = ref()
  const $route = useRoute()
  const { query } = $route
  const statusOptions = {
    Ready: 'success',
    NotReady: 'danger',
    Unknown: 'info',
  }
  const columns = ref([
    {
      dataIndex: 'name',
      label: t('env.node.name'),
      searchConfig: {
        component: 'el-input',
        label: t('env.node.name'),
        props: {
          placeholder: t('env.node.name'),
        },
      },
    },
    {
      dataIndex: 'status',
      label: t('env.node.status'),
      searchConfig: {
        component: 'el-select',
        label: t('env.node.status'),
        props: {
          placeholder: t('env.node.status'),
          options: [
            { label: 'Ready', value: 'Ready' },
            { label: 'NotReady', value: 'NotReady' },
            { label: 'Unknown', value: 'Unknown' },
          ],
        },
      },
    },
    {
      dataIndex: 'roles',
      label: t('env.pvc.roles'),
    },
    {
      dataIndex: 'internalIp',
      label: t('env.node.internalIp'),
      customRender: ({ row }: { row: nodeResult }) => {
        return row.internalIp || '--'
      },
    },
    {
      dataIndex: 'externalIp',
      label: t('env.node.externalIp'),
    },
    { dataIndex: 'creationTime', label: t('env.node.creationTime') },
    {
      dataIndex: 'operatingSystem',
      label: t('env.node.operatingSystem'),
    },
  ])
  const requestConfig = {
    api: NodeAPI.list,
    searchQuery: {
      model: {
        environmentId: query.environmentId,
      },
    },
  }
  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })

  return { tablePlus, columns, requestConfig, pageConfig, statusOptions, query }
}
