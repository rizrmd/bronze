<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  value?: number
  max?: number
  class?: string
  color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple'
  size?: 'sm' | 'md' | 'lg'
  showLabel?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  max: 100,
  color: 'blue',
  size: 'md',
  showLabel: false
})

const percentage = computed(() => {
  if (props.max === 0) return 0
  return Math.min((props.value || 0) / props.max * 100, 100)
})

const containerClass = computed(() => {
  const sizes = {
    sm: 'h-1',
    md: 'h-2',
    lg: 'h-3'
  }
  
  return [
    'w-full bg-gray-200 rounded-full overflow-hidden',
    sizes[props.size],
    props.class
  ].filter(Boolean).join(' ')
})

const barClass = computed(() => {
  const colors = {
    blue: 'bg-blue-600',
    green: 'bg-green-600',
    yellow: 'bg-yellow-600',
    red: 'bg-red-600',
    purple: 'bg-purple-600'
  }
  
  return [
    'h-full transition-all duration-300 ease-out',
    colors[props.color]
  ].join(' ')
})

// const labelColor = computed(() => {
//   const colors = {
//     blue: 'text-blue-600',
//     green: 'text-green-600',
//     yellow: 'text-yellow-600',
//     red: 'text-red-600',
//     purple: 'text-purple-600'
//   }
//   return colors[props.color]
// })
</script>

<template>
  <div class="w-full">
    <div v-if="showLabel" class="flex justify-between mb-1">
      <span class="text-sm font-medium text-gray-700">
        <slot name="label">{{ Math.round(percentage) }}%</slot>
      </span>
    </div>
    <div :class="containerClass">
      <div
        :class="barClass"
        :style="{ width: `${percentage}%` }"
      />
    </div>
    <div v-if="showLabel" class="mt-1 text-xs text-gray-500">
      <slot name="details">{{ value || 0 }} / {{ max }}</slot>
    </div>
  </div>
</template>