// 时间戳转换为标准显示时间

export const timestampToDate = (time: number | string, format: string = 'YYYY-MM-DD HH:mm:ss') => {
  let date: Date
  if (typeof time === 'string') {
    // 若为字符串，先尝试解析为时间戳（如"1761201587000"），失败则直接作为字符串时间解析
    const timestamp = Number(time)
    date = isNaN(timestamp) ? new Date(time.replace(' CST', '')) : new Date(timestamp)
  } else {
    // 若为数字，直接作为时间戳解析
    date = new Date(time)
  }

  // 2. 验证时间有效性（兼容Invalid Date）
  if (isNaN(date.getTime())) {
    console.warn(`无效的时间格式：${time}`)
    return ''
  }

  // 提取时间各部分（月份从0开始，需+1）
  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = date.getHours()
  const minute = date.getMinutes()
  const second = date.getSeconds()
  // 补零工具函数（确保两位数）
  const padZero = (num: number, length: number = 2): string => num.toString().padStart(length, '0')

  // 替换格式字符串
  return format
    .replace('YYYY', year.toString())
    .replace('MM', padZero(month))
    .replace('DD', padZero(day))
    .replace('HH', padZero(hour))
    .replace('mm', padZero(minute)) // 注意：分钟用mm避免与月份MM冲突
    .replace('ss', padZero(second))
}

/**
 * 文件大小转换（字节转 KB/MB/GB/TB）
 * @param bytes 原始文件大小（单位：B，支持数字或字符串类型）
 * @param decimalPlaces 保留的小数位数（默认 2 位，可选 0-10）
 * @returns 格式化后的大小字符串（如 "1.23 MB"），无效输入返回 "0 B"
 */
export function formatFileSize(bytes: number | string, decimalPlaces: number = 2): string {
  // 1. 处理输入：转换为数字并校验有效性
  let byteNum: number
  if (typeof bytes === 'string') {
    byteNum = Number(bytes.trim())
    // 字符串需能转换为有效数字，且非负数
    if (isNaN(byteNum) || byteNum < 0) {
      return '0 B'
    }
  } else {
    // 数字需为非负且有限（排除 Infinity/NaN）
    if (bytes < 0 || !isFinite(bytes)) {
      return '0 B'
    }
    byteNum = bytes
  }

  // 2. 处理 0 字节的特殊情况
  if (byteNum === 0) {
    return '0 B'
  }

  // 3. 定义单位层级（B → KB → MB → GB → TB，1024 进制）
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const unitStep = 1024 // 1 KB = 1024 B

  // 4. 计算适配的单位层级（如 1024 B → 层级 1 → KB）
  const unitIndex = Math.floor(Math.log(byteNum) / Math.log(unitStep))
  // 防止超出最大单位（超过 TB 仍用 TB 显示）
  const safeIndex = Math.min(unitIndex, units.length - 1)

  // 5. 计算转换后的值并保留指定小数位数
  const convertedSize = byteNum / Math.pow(unitStep, safeIndex)
  // 限制小数位数在 0-10 之间（避免无意义的精度）
  const safeDecimal = Math.max(0, Math.min(decimalPlaces, 10))
  const formattedSize = convertedSize.toFixed(safeDecimal)

  // 6. 拼接结果（移除末尾无意义的 ".00"，如 "2.00 MB" → "2 MB"）
  const finalSize = formattedSize.replace(/\.?0*$/, '') || '0'
  return `${finalSize} ${units[safeIndex]}`
}

/**
 * 复制内容
 */
export const setClipboardData = (data: any) => {
  return new Promise((success) => {
    const textarea: any = document.createElement('textarea')
    textarea.value = data
    textarea.readOnly = 'readOnly'
    document.body.appendChild(textarea)
    textarea.select()
    textarea.setSelectionRange(0, data.length)
    document.execCommand('copy')
    textarea.remove()
    success(data)
  })
}
