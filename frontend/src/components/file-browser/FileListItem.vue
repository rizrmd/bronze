<template>
  <div class="flex items-center px-4 py-2 border-b hover:bg-gray-50">
    <div class="flex items-center space-x-3 flex-1 mr-2">
      <div class="text-xl">{{ getFileIcon(fileName) }}</div>
      <div class="flex-1">
        <div class="font-medium">{{ fileName }}</div>
        <div class="text-sm text-gray-500">{{ formatFileSize(fileSize) }} • {{ formatDate(lastModified) }}</div>
      </div>
    </div>
    
    <div class="flex items-center space-x-2">
      <Button
        @click="$emit('download', file)"
        variant="ghost"
        size="sm"
        title="Download"
      >
        <Download class="h-4 w-4" />
      </Button>
      
      <Button
        @click="$emit('delete', file)"
        variant="ghost"
        size="sm"
        title="Delete"
        class="text-red-600 hover:text-red-700"
      >
        <Trash2 class="h-4 w-4" />
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Download, Trash2 } from 'lucide-vue-next'
import { formatFileSize, formatDate, getFileIcon } from '@/composables/useFileUtils'
import type { FileInfo } from '@/types'

interface Props {
  file: FileInfo
}

interface Emits {
  (e: 'download', file: FileInfo): void
  (e: 'delete', file: FileInfo): void
}

const props = defineProps<Props>()
defineEmits<Emits>()

const fileName = computed(() => {
  return props.file.key?.split('/')?.pop() || 'Unknown'
})

const fileSize = computed(() => props.file.size || 0)
const lastModified = computed(() => props.file.last_modified || '')
</script>