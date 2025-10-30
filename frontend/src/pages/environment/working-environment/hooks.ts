import { EnvAPI } from '@/api/env'
import { EnvType } from '@/types/env'
import { useRouterHooks } from '@/utils/url'

export const useEnvTableHooks = () => {
  const { t } = useI18n()
  const tablePlus = ref()
  const newEnvDialog = ref()
  const envDetail = ref()
  const { jumpToPage } = useRouterHooks()
  const load = ref({
    status: false,
    text: '',
  })
  const columns = ref([
    {
      dataIndex: 'name',
      label: t('env.run.name'),
      searchConfig: {
        component: 'el-input',
        label: t('env.run.name'),
        props: {
          placeholder: t('env.run.name'),
        },
      },
    },
    {
      dataIndex: 'environment',
      label: t('env.run.environment'),
      searchConfig: {
        component: 'el-select',
        label: t('env.run.environment'),
        props: {
          placeholder: t('env.run.environment'),
          options: [
            { label: 'Kubernetes', value: EnvType.K8S },
            { label: 'Docker', value: EnvType.DOCKER, disabled: true },
          ],
        },
      },
    },
    {
      dataIndex: 'namespace',
      label: t('env.run.namespace'),
      searchConfig: {
        component: 'el-input',
        label: t('env.run.namespace'),
        props: {
          placeholder: t('env.run.namespace'),
        },
      },
    },
    { dataIndex: 'createdAt', label: t('env.run.createdAt') },
    { dataIndex: 'updatedAt', label: t('env.run.updatedAt') },
  ])
  const requestConfig = {
    api: EnvAPI.list,
    searchQuery: {
      model: {},
    },
  }

  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })

  return {
    load,
    jumpToPage,
    tablePlus,
    requestConfig,
    columns,
    pageConfig,
    newEnvDialog,
    EnvAPI,
    envDetail,
  }
}
