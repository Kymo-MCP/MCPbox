import { TemplateAPI } from '@/api/mcp/template'
import { useMcpStoreHook } from '@/stores'
import { timestampToDate } from '@/utils/system'
import { AccessType } from '@/types/instance'

export const useTemplateTableHooks = () => {
  const { t } = useI18n()
  const tablePlus = ref()
  const { accessTypeOptions, mcpProtocolOptions } = useMcpStoreHook()
  const columns = ref([
    {
      label: t('mcp.template.name'),
      dataIndex: 'name',
      props: {
        width: 260,
      },
      searchConfig: {
        component: 'el-input',
        label: t('mcp.template.name'),
        props: {
          placeholder: t('mcp.template.name'),
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
      customRender: ({ row }: any) => {
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
      customRender: ({ row }: any) => {
        return mcpProtocolOptions.find((item) => item.value === row.mcpProtocol)?.label
      },
    },
    {
      dataIndex: 'environmentName',
      label: t('mcp.template.env'),
      customRender: ({ row }: any) => {
        return row.accessType === AccessType.HOSTING ? row.environmentName : '--'
      },
    },
    {
      dataIndex: 'notes',
      label: t('mcp.template.notes'),
      props: {
        'show-overflow-tooltip': true,
        'tooltip-formatter': ({ row }: any) => {
          return h('div', { style: { width: '400px' } }, row.notes)
        },
      },
    },
    {
      dataIndex: 'createdAt',
      label: t('mcp.template.createTime'),
      customRender: ({ row }: any) => {
        return timestampToDate(row.createdAt)
      },
    },
  ])
  const requestConfig = {
    api: TemplateAPI.list,
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
    tablePlus,
    columns,
    requestConfig,
    pageConfig,
  }
}
