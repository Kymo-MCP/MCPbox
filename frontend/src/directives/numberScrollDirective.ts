import type { Directive } from 'vue'
import { CountUp } from 'countup.js'

interface ICountUpParams {
  el: any
  value: number
  duration?: number
  useEasing?: boolean
  separator?: string
  decimalPlaces?: number
}
const isObject = (value: any) => {
  return Object.prototype.toString.call(value) === '[object Object]'
}

const countToFunc = (params: ICountUpParams) => {
  const { el, value } = params
  const countUp = new CountUp(el, value, {
    duration: params.duration || 0.5,
    useEasing: params.useEasing,
    separator: params.separator || ',',
    decimalPlaces: params.decimalPlaces || 0,
  })
  countUp.start()
}

export default {
  mounted(el, binding) {
    const bindingData: any = binding.value
    let options = { el } as ICountUpParams
    if (isObject(bindingData)) {
      options = {
        ...options,
        ...bindingData,
      }
    } else {
      options = {
        ...options,
        value: Number(bindingData),
      }
    }
    countToFunc(options)
  },
  updated(el, binding) {
    const bindingData: any = binding.value
    let options = { el } as ICountUpParams
    if (isObject(bindingData)) {
      options = {
        ...options,
        ...bindingData,
      }
    } else {
      options = {
        ...options,
        value: Number(bindingData),
      }
    }
    countToFunc(options)
  },
} as Directive
