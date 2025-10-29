<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  class?: string
  width?: string | number
  height?: string | number
  circle?: boolean
  lines?: number
}

const props = withDefaults(defineProps<Props>(), {
  lines: 1
})

const skeletonClass = computed(() => {
  const base = 'animate-pulse rounded-md bg-gray-200'
  return [base, props.circle ? 'rounded-full' : 'rounded-md', props.class].filter(Boolean).join(' ')
})

const style = computed(() => {
  const styles: Record<string, string> = {}
  if (props.width) styles.width = typeof props.width === 'number' ? `${props.width}px` : props.width
  if (props.height) styles.height = typeof props.height === 'number' ? `${props.height}px` : props.height
  return styles
})
</script>

<template>
  <div v-if="lines === 1" :class="skeletonClass" :style="style" />
  <div v-else class="space-y-2">
    <div
      v-for="i in lines"
      :key="i"
      :class="skeletonClass"
      :style="{
        ...style,
        width: i === lines ? '70%' : '100%'
      }"
    />
  </div>
</template>