<template>
  <table class="w-full">
    <thead class="bg-gray-50 border-b">
      <tr>
        <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
          Name
        </th>
        <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
          Size / Items
        </th>
        <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
          Modified
        </th>
        <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
          Actions
        </th>
      </tr>
    </thead>

    <tbody class="bg-white divide-y divide-gray-200">
      <!-- Folders -->
      <tr v-for="folder in sortedFolders" :key="folder.path" class="hover:bg-gray-50 cursor-pointer"
        @click="$emit('navigate', folder)" @dblclick="$emit('open-folder', folder)">
        <td class="px-4 py-2">
          <div class="flex items-center">
            <Folder class="w-6 h-6 text-blue-500 mr-3" />
            <span class="font-medium">{{ folder.name || 'Unknown' }}</span>
          </div>
        </td>

        <td class="px-4 py-2 text-sm text-gray-500">
          {{ folder.total_count || 0 }} items
          <span v-if="folder.file_count || folder.dir_count" class="text-xs text-gray-400">
            ({{ folder.file_count || 0 }} files, {{ folder.dir_count || 0 }} folders)
          </span>
        </td>

        <td class="px-4 py-2 text-sm text-gray-500">
          Folder
        </td>

      </tr>

      <!-- Files -->
      <tr v-for="file in sortedFiles" :key="file.key" class="hover:bg-gray-50 cursor-pointer" :class="{}"
        @click="$emit('open-file', file)">
        <td class="px-4 py-2">
          <div class="flex items-center gap-3">
            <component :is="getFileIconForFile(file)" />
            <span class="font-medium">
              {{ getFileName(file) }}
            </span>
          </div>
        </td>

        <td class="px-4 py-2 text-sm text-gray-500">
          {{ formatFileSize(file.size) }}
        </td>

        <td class="px-4 py-2 text-sm text-gray-500">
          {{ formatDate(file.last_modified) }}
        </td>

        <td class="px-4 py-2 text-sm text-gray-500">
          <div class="flex items-center gap-2">
            <button @click="$emit('preview', file)" class="p-1.5 hover:bg-gray-100 rounded transition-colors"
              title="Preview">
              <Eye class="w-4 h-4" />
            </button>
          </div>
        </td>

      </tr>
    </tbody>
  </table>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Folder } from 'lucide-vue-next'
import {
  File,
  FileText,
  FileSpreadsheet,
  FileImage,
  FileVideo,
  FileAudio,
  FileArchive,
  Eye
} from 'lucide-vue-next'
import { formatFileSize, formatDate, sortFiles } from '@/composables/useFileUtils'
import type { FileInfo, DirectoryInfo } from '@/types'

interface Props {
  folders: DirectoryInfo[]
  files: FileInfo[]
  sortBy?: 'name' | 'size' | 'date'
  sortDirection?: 'asc' | 'desc'
}

interface Emits {
  (e: 'navigate', folder: DirectoryInfo): void
  (e: 'open-folder', folder: DirectoryInfo): void
  (e: 'open-file', file: FileInfo): void
  (e: 'preview', file: FileInfo): void
}

const props = withDefaults(defineProps<Props>(), {
  sortBy: 'name',
  sortDirection: 'asc'
})

const emit = defineEmits<Emits>()

// Computed sorted items
const sortedFolders = computed(() => {
  return sortFiles(props.folders, props.sortBy, props.sortDirection)
})

const sortedFiles = computed(() => {
  return sortFiles(props.files, props.sortBy, props.sortDirection)
})

// Helper functions
const getFileName = (file: FileInfo) => {
  return file.key?.split('/')?.pop() || 'Unknown'
}

const getFileIconForFile = (file: FileInfo) => {
  const fileName = getFileName(file)
  const ext = fileName.split('.').pop()?.toLowerCase()
  const icons: Record<string, any> = {
    'pdf': FileText,
    'doc': FileText, 'docx': FileText,
    'xls': FileSpreadsheet, 'xlsx': FileSpreadsheet, 'csv': FileSpreadsheet,
    'jpg': FileImage, 'jpeg': FileImage, 'png': FileImage, 'gif': FileImage,
    'mp4': FileVideo, 'avi': FileVideo, 'mov': FileVideo,
    'mp3': FileAudio, 'wav': FileAudio,
    'zip': FileArchive, 'rar': FileArchive, 'tar': FileArchive, 'gz': FileArchive,
    'txt': File, 'log': File
  }
  return icons[ext || ''] || File
}
</script>