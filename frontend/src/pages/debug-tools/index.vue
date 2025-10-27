<!-- 目标页面组件（如 InspectorPage.vue） -->
<template>
  <div class="inspector-page">
    <h1>集成 ModelContext Inspector 的页面</h1>
    <!-- 渲染 Inspector 组件 -->
    <!-- <Inspector :config="inspectorConfig" style="width: 100%; height: 600px" /> -->
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
// import { Inspector } from '@modelcontextprotocol/inspector'
// import '@modelcontextprotocol/inspector/dist/style.css' // 引入样式

// 初始化配置
// const inspectorConfig = {
//   apiKey: 'your-api-key',
//   defaultContext: {},
//   onEvent: (event) => {
//   },
// }

onMounted(async () => {
  // 可选：页面加载后初始化
  // await Inspector.init(inspectorConfig);
})
</script>
