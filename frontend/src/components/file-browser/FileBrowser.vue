<template>
  <div class="h-full flex flex-col">
    <!-- Breadcrumb Navigation -->
    <BreadcrumbNavigation
      :current-path="currentPath"
      :breadcrumb-paths="breadcrumbPaths"
      @navigate="navigateToPath"
    />
    
    <!-- Toolbar -->
    <FileBrowserToolbar
      :search-query="searchQuery"
      :view-mode="viewMode"
      :loading="loading"
      @view-mode-change="setViewMode"
      @search="setSearchQuery"
      @refresh="refresh"
      @upload="$emit('upload')"
    />
    
    <!-- File Browser Content -->
    <div class="flex-1 flex flex-col min-h-0">
      <div v-if="loading" class="flex items-center justify-center h-64">
        <RefreshCw class="h-6 w-6 animate-spin text-gray-500" />
        <span class="ml-2 text-gray-500">Loading...</span>
      </div>
      
      <div v-else-if="error" class="flex flex-col items-center justify-center h-64 text-red-500">
        <div class="mb-2">❌</div>
        <div class="mb-4">{{ error }}</div>
        <Button @click="refresh">Retry</Button>
      </div>
      
      <div v-else-if="!hasFiles" class="flex flex-col items-center justify-center h-64 text-gray-400">
        <div class="mb-2 text-6xl">📁</div>
        <div>This folder is empty</div>
      </div>
      
      <div v-else class="flex-1 overflow-auto">
        <!-- List View -->
        <div v-if="viewMode === 'list'" class="bg-white">
          <!-- Folders -->
          <div
            v-for="folder in filteredFolders"
            :key="folder.path"
            class="cursor-pointer hover:bg-gray-50"
            @click="navigateToFolder(folder)"
            @dblclick="navigateToFolder(folder)"
          >
            <FolderListItem :folder="folder" @navigate="navigateToFolder" @open="navigateToFolder" />
          </div>
          
          <!-- Files -->
          <div
            v-for="file in filteredFiles"
            :key="file.key"
            class="cursor-pointer hover:bg-gray-50"
            :class="{ 'bg-blue-50': selectedFiles.has(file.key) }"
            @click="toggleFileSelection(file.key, $event)"
            @dblclick="downloadFile(file)"
          >
            <FileListItem :file="file" @download="downloadFile" @delete="deleteFile" />
          </div>
        </div>
        
        <!-- Grid View -->
        <div v-else class="p-4 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
          <!-- Folders -->
          <FolderCard
            v-for="folder in filteredFolders"
            :key="folder.path"
            :folder="folder"
            @navigate="navigateToFolder"
            @open="navigateToFolder"
          />
          
          <!-- Files -->
          <FileCard
            v-for="file in filteredFiles"
            :key="file.key"
            :file="file"
            @download="downloadFile"
            @delete="deleteFile"
          />
        </div>
      </div>
    </div>
    
    <!-- Actions Bar -->
    <FileBrowserActions
      :has-selection="hasSelection"
      @download="downloadSelected"
      @delete="deleteSelected"
      @select-all="selectAll"
      @clear-selection="clearSelection"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { RefreshCw } from 'lucide-vue-next'
import BreadcrumbNavigation from './BreadcrumbNavigation.vue'
import FileBrowserToolbar from './FileBrowserToolbar.vue'
import FileBrowserActions from './FileBrowserActions.vue'
import FolderCard from './FolderCard.vue'
import FolderListItem from './FolderListItem.vue'
import FileCard from './FileCard.vue'
import FileListItem from './FileListItem.vue'
import { useFileBrowser } from '@/composables/useFileBrowser'

interface Props {
  initialPath?: string
  initialViewMode?: 'list' | 'grid'
  useSSE?: boolean
}

interface Emits {
  (e: 'upload'): void
}

const props = withDefaults(defineProps<Props>(), {
  initialViewMode: 'list',
  useSSE: false
})

const emit = defineEmits<Emits>()

const {
  files,
  folders,
  currentPath,
  parentPath,
  loading,
  error,
  selectedFiles,
  searchQuery,
  viewMode,
  breadcrumbPaths,
  filteredFiles,
  filteredFolders,
  hasSelection,
  navigateToPath,
  navigateToFolder,
  toggleFileSelection,
  clearSelection,
  selectAll,
  refresh,
  setViewMode,
  setSearchQuery
} = useFileBrowser({
  initialPath: props.initialPath,
  initialViewMode: props.initialViewMode,
  useSSE: props.useSSE
})

const hasFiles = computed(() => {
  return (filteredFiles.value?.length || 0) > 0 || (filteredFolders.value?.length || 0) > 0
})

// File operations
const downloadFile = (file: any) => {
  // TODO: Implement download
  console.log('Download file:', file.key)
}

const deleteFile = (file: any) => {
  // TODO: Implement delete
  console.log('Delete file:', file.key)
}

const downloadSelected = () => {
  // TODO: Implement bulk download
  console.log('Download selected:', Array.from(selectedFiles.value))
}

const deleteSelected = () => {
  // TODO: Implement bulk delete
  console.log('Delete selected:', Array.from(selectedFiles.value))
}
</script>