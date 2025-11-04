<template>
  <div 
    class="p-4 border rounded-lg hover:bg-gray-50 transition-colors cursor-pointer"
    @dblclick="$emit('open', file)"
  >
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-3">
        <div class="w-5 h-5 flex items-center justify-center text-gray-400">
          📄
        </div>
        <div>
          <div class="font-medium">{{ fileName || 'Unknown' }}</div>
          <div class="text-sm text-gray-500">{{ formatFileSize(fileSize) }} • {{ formatDate(lastModified) }}</div>
        </div>
      </div>
      
      <div class="flex items-center space-x-2">
        <button 
          @click.stop="$emit('preview', file)"
          class="p-1.5 hover:bg-gray-100 rounded transition-colors"
          title="Preview"
        >
          <Eye class="w-4 h-4" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Eye } from 'lucide-vue-next'
import { formatFileSize, formatDate } from '@/composables/useFileUtils'
import type { FileInfo } from '@/types'

interface Props {
  file: FileInfo
}

interface Emits {
  (e: 'open', file: FileInfo): void
  (e: 'preview', file: FileInfo): void
}

const props = defineProps<Props>()
defineEmits<Emits>()

const fileName = computed(() => {
  return props.file.key?.split('/')?.pop() || 'Unknown'
})

const fileSize = computed(() => props.file.size || 0)
const lastModified = computed(() => props.file.last_modified || '')
</script>