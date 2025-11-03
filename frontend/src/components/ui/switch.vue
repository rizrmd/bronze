<script setup lang="ts">
import { computed } from 'vue'

const modelValue = defineModel<boolean>()

const props = defineProps<{
  disabled?: boolean
  id?: string
}>()

const switchClasses = computed(() => [
  'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2',
  modelValue.value ? 'bg-blue-600 focus:ring-blue-500' : 'bg-gray-200 focus:ring-gray-500',
  props.disabled && 'opacity-50 cursor-not-allowed'
])

const thumbClasses = computed(() => [
  'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
  modelValue.value ? 'translate-x-5' : 'translate-x-0'
])

const toggle = () => {
  if (!props.disabled) {
    modelValue.value = !modelValue.value
  }
}
</script>

<template>
  <button
    :id="id"
    type="button"
    :class="switchClasses"
    role="switch"
    :aria-checked="modelValue"
    :disabled="disabled"
    @click="toggle"
  >
    <span :class="thumbClasses" />
  </button>
</template>
