<script setup lang="ts">
import { provide, ref } from 'vue'

const modelValue = defineModel<string>()
defineProps<{
  disabled?: boolean
}>()

const isOpen = ref(false)
const selectId = `select-${Math.random().toString(36).substr(2, 9)}`

provide('selectId', selectId)
provide('modelValue', modelValue)
provide('isOpen', isOpen)
provide('close', () => { isOpen.value = false })
provide('select', (value: string) => {
  modelValue.value = value
  isOpen.value = false
})
</script>

<template>
  <div class="relative">
    <slot />
  </div>
</template>
