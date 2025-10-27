<template>
  <el-form v-if="props.formConfig.length" ref="formRef" :model="formData" v-bind="$attrs">
    <template v-for="item in props.formConfig" :key="item.key">
      <el-form-item :name="item.key" v-bind="item.itemProps">
        <template v-if="item.label" #label>
          <div :style="item.labelStyle">
            {{ item.label }}
          </div>
        </template>
        <slot v-if="item.component === 'slot'" :name="item.slotName" />
        <div v-else-if="item.component === 'innerText'">
          {{
            item.formatter
              ? item.formatter(item.innerText || formData[item.key])
              : item.innerText || formData[item.key]
          }}
        </div>
        <el-switch
          v-else-if="item.component === 'el-switch'"
          v-model="formData[item.key]"
          v-bind="item.props"
        />
        <el-cascader
          v-else-if="item.component === 'el-cascader'"
          v-model="formData[item.key]"
          v-bind="item.props"
        />
        <component
          :is="item.component"
          v-else
          v-model="formData[item.key]"
          clearable
          :show-search="item.component === 'el-select'"
          :filter-option="item.component === 'el-select' ? filterOption : false"
          v-bind="{
            getPopupContainer:
              item.component === 'el-select' ||
              item.component === 'el-date-picker' ||
              item.component === 'el-range-picker'
                ? (trigger: HTMLElement) => {
                    return trigger.parentElement
                  }
                : false,
            ...item.props,
          }"
          class="w-full"
          :teleported="false"
        >
        </component>
        <div v-if="item.tips" class="tips" :style="item.tipsStyle" v-html="item.tips"></div>
      </el-form-item>
    </template>
  </el-form>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue'

const props = defineProps({
  formConfig: {
    type: Array<any>,
    default: () => [],
  },
  formData: {
    type: Object,
    default: () => {},
  },
})
const formData = ref(props.formData)
watch(
  () => props.formData,
  (value) => {
    formData.value = value
  },
  {
    deep: true,
  },
)

const formRef = ref()
const validate = async () => {
  const valid = await formRef.value.validate()
  return valid
}
const resetFields = () => {
  formRef.value.resetFields()
}

/**
 * default
 */
const filterOption = (input: string, option: any) => {
  return option.label.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

defineExpose({ validate, resetFields })
</script>

<style scoped lang="scss">
.tips {
  font-weight: 400;
  font-size: 14px;
  color: #8d8d8d;
  line-height: 2;
  text-align: left;
  font-style: normal;
}

.w-full {
  width: 100%;
}
</style>
