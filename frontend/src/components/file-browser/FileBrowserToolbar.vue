<template>
  <div class="flex items-center justify-between p-2 border-b bg-gray-50">
    <div class="flex items-center space-x-2 flex-1 mr-2">
      <Input
        :model-value="searchQuery"
        placeholder="Filter files and folders..."
        class="flex-1"
        @input="handleSearchInput"
      />
    </div>
    
    <div class="flex items-center space-x-2">
      <div class="flex items-center space-x-1 border rounded-lg">
        <button
          @click="$emit('view-mode-change', 'list')"
          class="px-3 py-1 rounded-l-md hover:bg-gray-100 transition-colors"
          :class="{ 'bg-white shadow-sm': viewMode === 'list' }"
          title="List view"
        >
          <List class="h-4 w-4" />
        </button>
        <button
          @click="$emit('view-mode-change', 'grid')"
          class="px-3 py-1 rounded-r-md hover:bg-gray-100 transition-colors"
          :class="{ 'bg-white shadow-sm': viewMode === 'grid' }"
          title="Grid view"
        >
          <Grid3X3 class="h-4 w-4" />
        </button>
      </div>
      
      <Button
        @click="$emit('refresh')"
        variant="outline"
        size="sm"
        :disabled="loading"
        title="Refresh"
      >
        <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" />
      </Button>
      
      <Button
        @click="$emit('create-folder')"
        variant="outline"
        size="sm"
        title="New folder"
      >
        <FolderOpen class="h-4 w-4" />
      </Button>
      
      <Button
        @click="$emit('upload')"
        variant="default"
        size="sm"
        title="Upload files"
      >
        <Upload class="h-4 w-4" />
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { List, Grid3X3, RefreshCw, Upload, FolderOpen } from 'lucide-vue-next'
import { debounce } from '@/composables/useFileUtils'

interface Props {
  searchQuery: string
  viewMode: 'list' | 'grid'
  loading: boolean
}

interface Emits {
  (e: 'search', query: string): void
  (e: 'view-mode-change', mode: 'list' | 'grid'): void
  (e: 'refresh'): void
  (e: 'create-folder'): void
  (e: 'upload'): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

const debouncedSearch = debounce((query: string) => {
  emit('search', query)
}, 300)

// Handle search input
const handleSearchInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  debouncedSearch(target.value)
}
</script>