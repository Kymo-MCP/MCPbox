import { PvcAPI } from '@/api/env'
import { NodeVisible } from '@/types/instance'

export const usePvcTableHooks = () => {
  const { t } = useI18n()
  const { query } = useRoute()
  const storageClassOptions = ref<any>([])
  const tablePlus = ref()
  const columns = computed(() => {
    return [
      {
        dataIndex: 'name',
        label: t('env.pvc.name'),
        searchConfig: {
          component: 'el-input',
          label: t('env.pvc.name'),
          props: {
            placeholder: t('env.pvc.name'),
          },
        },
      },
      {
        dataIndex: 'namespace',
        label: t('env.pvc.namespace'),
        searchConfig: {
          component: 'el-input',
          label: t('env.pvc.namespace'),
          props: {
            placeholder: t('env.pvc.namespace'),
          },
        },
      },
      {
        dataIndex: 'storageClass',
        label: t('env.pvc.storageClass'),
        searchConfig: {
          component: 'el-select',
          label: t('env.pvc.storageClass'),
          props: {
            placeholder: t('env.pvc.storageClass'),
            options: storageClassOptions.value,
          },
        },
      },
      {
        dataIndex: 'accessModes',
        label: t('env.pvc.accessModes'),
        searchConfig: {
          component: 'el-select',
          label: t('env.pvc.accessModes'),
          props: {
            placeholder: t('env.pvc.accessModes'),
            options: [
              { label: NodeVisible.RWO, value: NodeVisible.RWO },
              { label: NodeVisible.ROM, value: NodeVisible.ROM },
              { label: NodeVisible.RWM, value: NodeVisible.RWM },
            ],
          },
        },
      },
      {
        dataIndex: 'capacity',
        label: t('env.pvc.capacity'),
        customRender: ({ row }: any) => {
          return row.capacity || '--'
        },
      },
      {
        dataIndex: 'status',
        label: t('env.pvc.status'),
      },
      {
        dataIndex: 'pods',
        label: t('env.pvc.pods'),
      },
      { dataIndex: 'creationTime', label: t('env.pvc.creationTime') },
    ]
  })
  // ref()
  const requestConfig = {
    api: PvcAPI.list,
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
  const newPvcDialog = ref()

  return {
    PvcAPI,
    tablePlus,
    columns,
    storageClassOptions,
    requestConfig,
    pageConfig,
    newPvcDialog,
    query,
  }
}
