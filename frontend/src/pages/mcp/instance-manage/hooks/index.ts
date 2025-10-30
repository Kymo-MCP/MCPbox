import { type InstanceResult, InstanceStatus, ContainerOptions } from '@/types/instance'
import { InstanceAPI } from '@/api/mcp/instance'
import { useMcpStoreHook } from '@/stores'
import { useRouterHooks } from '@/utils/url'
import { instanceTotal, instanceStart, instanceStop, instanceConnect } from '@/utils/logo'
import { AccessType } from '@/types/instance'

export const useInstanceTableHooks = () => {
  const { t } = useI18n()
  const timer = ref<number>(0)
  const tablePlus = ref()
  const instanceDetail = ref()
  const probe = ref()
  const { query } = useRoute()
  const load = ref({
    status: false,
    text: '',
  })
  const selectVisible = ref(false)
  const templateList = ref<any>([])
  const viewConfig = ref()
  const { jumpToPage } = useRouterHooks()
  // const mcpHook = useMcpStoreHook()
  const { accessTypeOptions, mcpProtocolOptions } = useMcpStoreHook()
  const instanceCount = ref<any>({})
  const dataCountList = computed(() => [
    {
      title: t('mcp.instance.count.total'),
      icon: instanceTotal,
      type: 'total',
      count: instanceCount.value.totalInstances,
    },
    {
      title: t('mcp.instance.count.running'),
      icon: instanceStart,
      type: 'active',
      count: instanceCount.value.activeInstances,
    },
    {
      title: t('mcp.instance.count.stoped'),
      icon: instanceStop,
      type: 'inactive',
      count: instanceCount.value.inactiveInstances,
    },
    {
      title: t('mcp.instance.count.connecting'),
      icon: instanceConnect,
      type: 'hosting',
      count: instanceCount.value.hostingInstances,
    },
  ])

  // service status
  const activeOptions = {
    active: {
      label: t('status.' + InstanceStatus.ACTIVE),
      type: 'success',
      value: InstanceStatus.ACTIVE,
    },
    inactive: {
      label: t('status.' + InstanceStatus.INACTIVE),
      type: 'danger',
      value: InstanceStatus.INACTIVE,
    },
  }
  // container status list
  const containerOptions = {
    pending: {
      label: t('status.' + ContainerOptions.PENDING),
      type: 'primary',
      value: ContainerOptions.PENDING,
    },
    running: {
      label: t('status.' + ContainerOptions.RUNNING),
      type: 'success',
      value: ContainerOptions.RUNNING,
    },
    'init-timeout-stop': {
      label: t('status.' + ContainerOptions.INIT_TIMEOUT_STOP),
      type: 'danger',
      value: ContainerOptions.INIT_TIMEOUT_STOP,
    },
    'run-timeout-stop': {
      label: t('status.' + ContainerOptions.RUN_TIMEOUT_STOP),
      type: 'danger',
      value: ContainerOptions.RUN_TIMEOUT_STOP,
    },
    'exception-force-stop': {
      label: t('status.' + ContainerOptions.EXCEPTION_FORCE_STOP),
      type: 'danger',
      value: ContainerOptions.EXCEPTION_FORCE_STOP,
    },
    'manual-stop': {
      label: t('status.' + ContainerOptions.MANUAL_STOP),
      type: 'warning',
      value: ContainerOptions.MANUAL_STOP,
    },
    'create-failed': {
      label: t('status.' + ContainerOptions.CREATE_FAILED),
      type: 'danger',
      value: ContainerOptions.CREATE_FAILED,
    },
    'running-unready': {
      label: t('status.' + ContainerOptions.RUNNING_UNREADY),
      type: 'warning',
      value: ContainerOptions.RUNNING_UNREADY,
    },
  }
  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })
  const columns = ref<any>([
    {
      dataIndex: 'instanceName',
      label: t('mcp.instance.name'),
      searchConfig: {
        component: 'el-input',
        label: t('mcp.instance.form.placeholderName'),
        props: {
          placeholder: t('mcp.instance.form.placeholderName'),
        },
      },
    },
    {
      dataIndex: 'accessType',
      label: t('mcp.instance.form.accessType'),
      searchConfig: {
        component: 'el-select',
        label: t('mcp.instance.form.accessType'),
        props: {
          placeholder: t('mcp.instance.form.accessType'),
          options: accessTypeOptions,
        },
      },
      customRender: ({ row }: { row: InstanceResult }) => {
        return h(
          'span',
          { class: ['text-grey', 'text-primary', 'text-warning', 'text-success'][row.accessType] },
          accessTypeOptions.find((item) => item.value === row.accessType)?.label,
        )
      },
    },
    {
      dataIndex: 'mcpProtocol',
      label: t('mcp.instance.form.mcpProtocol'),
      searchConfig: {
        component: 'el-select',
        label: t('mcp.instance.form.mcpProtocol'),
        props: {
          placeholder: t('mcp.instance.form.mcpProtocol'),
          options: mcpProtocolOptions,
        },
      },
      customRender: ({ row }: { row: InstanceResult }) => {
        return mcpProtocolOptions.find((item) => item.value === row.mcpProtocol)?.label
      },
    },
    { dataIndex: 'publicProxyConfig', label: t('mcp.instance.publicProxyConfig') },
    {
      dataIndex: 'status',
      label: t('mcp.instance.status'),
      searchConfig: {
        component: 'el-select',
        label: t('mcp.instance.form.placeholderStatus'),
        props: {
          placeholder: t('mcp.instance.form.placeholderStatus'),
          options: Object.values(activeOptions),
        },
      },
    },
    {
      dataIndex: 'containerStatus',
      label: t('mcp.instance.packStatus'),
      searchConfig: {
        component: 'el-select',
        label: t('mcp.instance.form.placeholderContainerStatus'),
        props: {
          placeholder: t('mcp.instance.form.placeholderContainerStatus'),
          options: Object.values(containerOptions),
        },
      },
    },
    {
      dataIndex: 'environmentName',
      label: t('mcp.instance.env'),
      customRender: ({ row }: { row: InstanceResult }) => {
        return row.accessType === AccessType.HOSTING ? row.environmentName : '--'
      },
    },
    {
      dataIndex: 'notes',
      label: t('mcp.template.notes'),
      props: {
        'show-overflow-tooltip': true,
        'tooltip-formatter': ({ row }: { row: InstanceResult }) => {
          return h('div', { style: { width: '400px' } }, row.notes)
        },
      },
    },
    { dataIndex: 'createdAt', label: t('mcp.instance.createTime') },
  ])
  const requestConfig = ref<any>({
    api: InstanceAPI.list,
    searchQuery: {
      model: {},
    },
  })
  /**
   *
   * @param form - instance form data
   */
  const handleAddInstance = () => {
    jumpToPage({
      url: '/new-instance',
      data: {},
    })
  }

  return {
    load,
    query,
    columns,
    jumpToPage,
    tablePlus,
    requestConfig,
    pageConfig,
    handleAddInstance,
    activeOptions,
    containerOptions,
    InstanceAPI,
    instanceCount,
    instanceDetail,
    viewConfig,
    accessTypeOptions,
    mcpProtocolOptions,
    dataCountList,
    probe,
    selectVisible,
    templateList,
    timer,
  }
}
