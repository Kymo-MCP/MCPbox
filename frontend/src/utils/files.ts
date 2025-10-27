/**
 * 文件数据下载至本地
 * @param fileInfo
 * @param suffix
 */
export const downloadData = (
  fileInfo: {
    fileName: string
    data: string
  },
  suffix: string = 'txt',
) => {
  const blob = new Blob([fileInfo.data], { type: 'text/plain;charset=utf-8' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${fileInfo.fileName}.${suffix}`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}
