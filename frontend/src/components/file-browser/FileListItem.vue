<template>
  <div class="flex items-center px-4 py-2 border-b hover:bg-gray-50">
    <div class="flex items-center space-x-3 flex-1 mr-2">
      <div class="w-5 h-5 flex items-center justify-center text-gray-400">
        📄
      </div>
      <div class="flex-1">
        <div class="font-medium">{{ fileName }}</div>
        <div class="text-sm text-gray-500">{{ formatFileSize(fileSize) }} • {{ formatDate(lastModified) }}</div>
      </div>
    </div>
    
    <div class="flex items-center space-x-2">
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatFileSize, formatDate } from '@/composables/useFileUtils'
import type { FileInfo } from '@/types'

interface Props {
  file: FileInfo
}

const props = defineProps<Props>()

const fileName = computed(() => {
  return props.file.key?.split('/')?.pop() || 'Unknown'
})

const fileSize = computed(() => props.file.size || 0)
const lastModified = computed(() => props.file.last_modified || '')
</script>