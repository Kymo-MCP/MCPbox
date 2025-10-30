/**
 * JSON 格式化工具类
 * 支持格式化（带缩进）、压缩（无空格）、验证JSON有效性
 */
export const JsonFormatter = {
  /**
   * 格式化JSON字符串（带缩进和换行）
   * @param jsonString 原始JSON字符串
   * @param indent 缩进空格数，默认2
   * @returns 格式化后的JSON字符串或错误信息
   */
  format: (jsonString: string, indent: number = 2): string => {
    try {
      const processedStr = jsonString.replace(/\\"/g, '"').replace(/'/g, '"')
      // 先解析验证JSON有效性
      const parsed = JSON.parse(processedStr)
      // 格式化并返回
      return JSON.stringify(parsed, null, indent)
    } catch (error) {
      return jsonString
    }
  },

  /**
   * 压缩JSON字符串（移除所有空格和换行）
   * @param jsonString 原始JSON字符串
   * @returns 压缩后的JSON字符串或错误信息
   */
  minify: (jsonString: string): string => {
    try {
      const parsed = JSON.parse(jsonString)
      return JSON.stringify(parsed)
    } catch (error) {
      return `压缩失败：${(error as Error).message}`
    }
  },

  /**
   * 验证JSON字符串是否有效
   * @param jsonString 待验证的JSON字符串
   * @returns 验证结果对象 { valid: boolean, error?: string }
   */
  validate: (jsonString: string): { valid: boolean; error?: string } => {
    try {
      JSON.parse(jsonString)
      return { valid: true }
    } catch (error) {
      return { valid: false, error: (error as Error).message }
    }
  },

  /**
   * 将JavaScript对象转换为格式化的JSON字符串
   * @param obj 任意JavaScript对象
   * @param indent 缩进空格数，默认2
   * @returns 格式化后的JSON字符串
   */
  stringify: (obj: unknown, indent: number = 2): string => {
    try {
      return JSON.stringify(obj, null, indent)
    } catch (error) {
      return `序列化失败：${(error as Error).message}`
    }
  },
}
