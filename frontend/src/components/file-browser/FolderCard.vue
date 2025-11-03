<template>
  <div 
    class="flex flex-col items-center p-3 rounded-lg hover:bg-gray-100 cursor-pointer transition-colors"
    @click="$emit('navigate', folder)"
    @dblclick="$emit('open', folder)"
  >
    <div class="text-4xl mb-2">📁</div>
    <span class="text-xs text-center truncate w-full">{{ folder.name || 'Unknown' }}</span>
  </div>
</template>

<script setup lang="ts">
import type { DirectoryInfo } from '@/types'

interface Props {
  folder: DirectoryInfo
}

interface Emits {
  (e: 'navigate', folder: DirectoryInfo): void
  (e: 'open', folder: DirectoryInfo): void
}

const props = defineProps<Props>()
defineEmits<Emits>()
</script>