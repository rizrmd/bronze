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
        <Folder class="w-16 h-16 text-blue-500 mb-2" />
        <div>This folder is empty</div>
      </div>
      
      <div v-else class="flex-1 overflow-auto">
        <!-- List View (Table) -->
        <div v-if="viewMode === 'list'" class="bg-white">
          <FileBrowserTable
            :folders="filteredFolders"
            :files="filteredFiles"
            @navigate="navigateToFolder"
            @open-folder="navigateToFolder"
            @open-file="openFile"
          />
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
            @open="openFile"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { RefreshCw, Folder } from 'lucide-vue-next'
import BreadcrumbNavigation from './BreadcrumbNavigation.vue'
import FileBrowserToolbar from './FileBrowserToolbar.vue'
import FileBrowserTable from './FileBrowserTable.vue'
import FolderCard from './FolderCard.vue'
import FileCard from './FileCard.vue'
import { useFileBrowser } from '@/composables/useFileBrowser'
import type { FileInfo } from '@/types'

interface Props {
  initialPath?: string
  initialViewMode?: 'list' | 'grid'
}

interface Emits {
  (e: 'upload'): void
}

const props = withDefaults(defineProps<Props>(), {
  initialViewMode: 'list'
})

const emit = defineEmits<Emits>()

const router = useRouter()

const {
  currentPath,
  searchQuery,
  viewMode,
  loading,
  error,
  breadcrumbPaths,
  filteredFiles,
  filteredFolders,
  navigateToPath,
  navigateToFolder,
  refresh,
  setViewMode,
  setSearchQuery
} = useFileBrowser({
  initialPath: props.initialPath,
  initialViewMode: props.initialViewMode
})

const hasFiles = computed(() => {
  return (filteredFiles.value?.length || 0) > 0 || (filteredFolders.value?.length || 0) > 0
})

// Open file in data browser with full path
const openFile = (file: FileInfo) => {
  // Always navigate to data browser with the full file path
  router.push({
    name: 'DataBrowser',
    query: { file: file.key }
  })
}


</script>