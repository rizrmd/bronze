<script setup lang="ts">
import { inject, type Ref } from 'vue'

const selectId = inject('selectId') as string
const isOpen = inject('isOpen') as Ref<boolean>
const toggle = () => {
  isOpen.value = !isOpen.value
}
</script>

<template>
  <button
    :id="selectId"
    type="button"
    class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
    @click="toggle"
  >
    <slot />
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4 opacity-50">
      <path d="m6 9 6 6 6-6"/>
    </svg>
  </button>
</template>
