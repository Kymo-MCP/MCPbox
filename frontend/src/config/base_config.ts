export default {
  SERVER_BASE_URL: '', // 当使用vite代理的时候基础IP设置为空
  // SERVER_BASE_URL: 'https://134.175.7.229:443', // 当使用vite代理的时候基础IP设置为空；当axios 配置了baseURL之后；会导致请求直接跳过vite代理
  baseUrlVersion: '/api',
}
