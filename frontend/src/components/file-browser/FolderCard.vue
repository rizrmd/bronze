<template>
  <div 
    class="flex flex-col items-center p-3 rounded-lg hover:bg-gray-100 cursor-pointer transition-colors"
    @click="$emit('navigate', folder)"
    @dblclick="$emit('open', folder)"
  >
    <Folder class="w-10 h-10 text-blue-500 mb-2" />
    <span class="text-xs text-center truncate w-full">{{ folder.name || 'Unknown' }}</span>
    <span v-if="folder.total_count !== undefined" class="text-xs text-gray-500">
      {{ folder.total_count }} items
    </span>
  </div>
</template>

<script setup lang="ts">
import { Folder } from 'lucide-vue-next'
import type { DirectoryInfo } from '@/types'

interface Props {
  folder: DirectoryInfo
}

interface Emits {
  (e: 'navigate', folder: DirectoryInfo): void
  (e: 'open', folder: DirectoryInfo): void
}

defineProps<Props>()
defineEmits<Emits>()
</script>