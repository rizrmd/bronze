<template>
  <div 
    class="flex items-center space-x-2 flex-1 mr-2"
    @click="$emit('navigate', folder)"
    @dblclick="$emit('open', folder)"
  >
    <div class="text-2xl">📁</div>
    <div class="flex-1">
      <div class="font-medium">{{ folder.name || 'Unknown' }}</div>
      <div class="text-sm text-gray-500">Folder</div>
    </div>
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