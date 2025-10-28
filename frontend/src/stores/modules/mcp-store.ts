import { defineStore } from 'pinia'
import { store } from '@/stores'
import { EnvAPI, PvcAPI, NodeAPI } from '@/api/env'
import { CodeAPI } from '@/api/code'
import { SourceType, AccessType, McpProtocol } from '@/types/instance'
import { type Code, type EnvResult, type nodeResult } from '@/types/index'

export const useMcpStore = defineStore('mcp', () => {
  const { t } = useI18n()
  const packageList = ref<Code[]>([])
  const envList = ref<EnvResult[]>([])
  const nodeList = ref<nodeResult[]>([])
  const pvcList = ref<any[]>([])

  // source of instance list
  const sourceOptions = computed(() => [
    { label: t('mcp.source.unknown'), value: SourceType.UNKONWN }, // unknown
    { label: t('mcp.source.market'), value: SourceType.MARKET }, // market
    { label: t('mcp.source.template'), value: SourceType.TEMPLATE }, // template
    { label: t('mcp.source.custom'), value: SourceType.CUSTOM }, // custom
  ])

  // access type
  const accessTypeOptions = computed(() => [
    { label: t('mcp.type.hosting'), value: AccessType.HOSTING },
    { label: t('mcp.type.direct'), value: AccessType.DIRECT },
    { label: t('mcp.type.proxy'), value: AccessType.PROXY },
  ])
  // mcp protocol
  const mcpProtocolOptions = computed(() => [
    { label: 'STDIO', value: McpProtocol.STDIO },
    { label: 'SSE', value: McpProtocol.SSE },
    { label: 'STEAMABLE_HTTP', value: McpProtocol.STEAMABLE_HTTP },
  ])

  /**
   * Handle get package list
   */
  const handleGetPackageList = async () => {
    const { list } = await CodeAPI.list(null)
    packageList.value = list || []
  }

  /**
   * Handle get env list
   */
  const handleGetEnvList = async () => {
    const data = await EnvAPI.list(null)
    envList.value = data.list || []
  }

  /**
   * Handle get node list
   */
  const handleGetNodeList = async (environmentId: string) => {
    const data = await NodeAPI.list({ environmentId })
    nodeList.value = data.list
  }
  /**
   * Handle get pvc list
   */
  const handleGetPvcList = async (environmentId: string) => {
    const data = await PvcAPI.list({ environmentId })
    pvcList.value = data.list
  }

  return {
    packageList,
    envList,
    nodeList,
    pvcList,
    sourceOptions,
    accessTypeOptions,
    mcpProtocolOptions,
    handleGetPackageList,
    handleGetEnvList,
    handleGetNodeList,
    handleGetPvcList,
  }
})

export function useMcpStoreHook() {
  return useMcpStore(store)
}
