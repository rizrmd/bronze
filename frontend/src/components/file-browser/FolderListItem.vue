<template>
  <div 
    class="flex items-center space-x-2 flex-1 mr-2"
    @click="$emit('navigate', folder)"
    @dblclick="$emit('open', folder)"
  >
    <Folder class="w-6 h-6 text-blue-500" />
    <div class="flex-1">
      <div class="font-medium">{{ folder.name || 'Unknown' }}</div>
      <div class="text-sm text-gray-500">
        Folder <span v-if="folder.total_count !== undefined">({{ folder.total_count }} items)</span>
      </div>
    </div>
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