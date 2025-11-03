<script setup lang="ts">
import { inject, watch, type Ref } from 'vue'

const isOpen = inject('isOpen') as Ref<boolean>
const selectId = inject('selectId') as string

watch(isOpen, (newValue) => {
  if (newValue) {
    setTimeout(() => {
      document.getElementById(selectId)?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    }, 0)
  }
})
</script>

<template>
  <div v-show="isOpen" class="absolute top-full left-0 z-50 mt-1 w-full rounded-md border bg-popover p-1 text-popover-foreground shadow-md">
    <slot />
  </div>
</template>
