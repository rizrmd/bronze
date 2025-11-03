<script setup lang="ts">
import { inject } from 'vue'

const props = defineProps<{
  value: string
}>()

const select = inject('select') as (value: string) => void
</script>

<template>
  <div
    class="relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground cursor-pointer"
    @click="select(props.value)"
  >
    <slot />
  </div>
</template>
