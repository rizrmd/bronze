<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useFiles } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { 
  Upload, 
  Download, 
  Trash2, 
  Search, 
  File,
  FolderOpen,
  RefreshCw
} from 'lucide-vue-next'

const { files, fetchFiles, uploadFile, deleteFile, downloadFile, loading, error } = useFiles()

const searchQuery = ref('')
const selectedFiles = ref<string[]>([])
const uploadDialogOpen = ref(false)
const dragOver = ref(false)

const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  return files.value.filter(file => 
    file.key.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    if (file) {
      await uploadFile(file)
      target.value = ''
    }
  }
}

const handleDrop = async (event: DragEvent) => {
  event.preventDefault()
  dragOver.value = false
  
  if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
    const file = event.dataTransfer.files[0]
    if (file) {
      await uploadFile(file)
    }
  }
}

const handleDragOver = (event: DragEvent) => {
  event.preventDefault()
  dragOver.value = true
}

const handleDragLeave = () => {
  dragOver.value = false
}

const handleDeleteFile = async (fileKey: string) => {
  if (confirm(`Are you sure you want to delete ${fileKey}?`)) {
    await deleteFile(fileKey)
  }
}

const handleDownloadFile = async (fileKey: string) => {
  await downloadFile(fileKey)
}

const toggleFileSelection = (fileKey: string) => {
  const index = selectedFiles.value.indexOf(fileKey)
  if (index > -1) {
    selectedFiles.value.splice(index, 1)
  } else {
    selectedFiles.value.push(fileKey)
  }
}

const refreshFiles = () => {
  fetchFiles()
}

onMounted(() => {
  fetchFiles()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold text-gray-900">Files</h2>
        <p class="mt-2 text-gray-600">Manage your files in MinIO storage</p>
      </div>
      <div class="flex items-center space-x-3">
        <Button @click="refreshFiles" variant="outline" size="sm">
          <RefreshCw class="w-4 h-4 mr-2" />
          Refresh
        </Button>
        <Button @click="uploadDialogOpen = true">
          <Upload class="w-4 h-4 mr-2" />
          Upload File
        </Button>
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="error" class="bg-red-50 border border-red-200 rounded-md p-4">
      <div class="flex">
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error</h3>
          <div class="mt-2 text-sm text-red-700">
            {{ error }}
          </div>
        </div>
      </div>
    </div>

    <!-- Upload Area -->
    <Card>
      <CardHeader>
        <CardTitle>Upload Files</CardTitle>
        <CardDescription>
          Drag and drop files here or click to browse
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div
          @drop="handleDrop"
          @dragover="handleDragOver"
          @dragleave="handleDragLeave"
          :class="[
            'border-2 border-dashed rounded-lg p-8 text-center transition-colors',
            dragOver ? 'border-blue-400 bg-blue-50' : 'border-gray-300'
          ]"
        >
          <Upload class="mx-auto h-12 w-12 text-gray-400" />
          <div class="mt-4">
            <label for="file-upload" class="cursor-pointer">
              <span class="mt-2 block text-sm font-medium text-gray-900">
                Drop files here or 
                <span class="text-blue-600 hover:text-blue-500">browse</span>
              </span>
              <input
                id="file-upload"
                type="file"
                class="sr-only"
                @change="handleFileUpload"
              />
            </label>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Search and Filter -->
    <Card>
      <CardContent class="pt-6">
        <div class="flex items-center space-x-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              v-model="searchQuery"
              placeholder="Search files..."
              class="pl-10"
            />
          </div>
          <div v-if="selectedFiles.length > 0" class="flex items-center space-x-2">
            <Badge variant="secondary">
              {{ selectedFiles.length }} selected
            </Badge>
            <Button variant="outline" size="sm">
              Delete Selected
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Files List -->
    <Card>
      <CardHeader>
        <CardTitle>Files ({{ filteredFiles.length }})</CardTitle>
        <CardDescription>
          All files in your MinIO bucket
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-center py-8">
          <RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
          <p class="mt-2 text-gray-500">Loading files...</p>
        </div>
        
        <div v-else-if="filteredFiles.length === 0" class="text-center py-8 text-gray-500">
          <FolderOpen class="mx-auto h-12 w-12 text-gray-400" />
          <p class="mt-2">No files found</p>
        </div>

        <div v-else class="space-y-2">
            <div
              v-for="file in filteredFiles"
              :key="file.key"
              class="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors"
            >
              <div class="flex items-center space-x-3">
                <input
                  type="checkbox"
                  :checked="selectedFiles.includes(file.key)"
                  @change="toggleFileSelection(file.key)"
                  class="rounded border-gray-300"
                />
                <File class="h-5 w-5 text-gray-400" />
                <div>
                  <p class="font-medium text-gray-900">{{ file.key }}</p>
                  <p class="text-sm text-gray-500">
                    {{ formatFileSize(file.size) }} • {{ formatDate(file.last_modified) }}
                  </p>
                </div>
              </div>
              
              <div class="flex items-center space-x-2">
                <Badge variant="outline">
                  {{ file.content_type || 'Unknown' }}
                </Badge>
                <Button
                  variant="ghost"
                  size="sm"
                  @click="handleDownloadFile(file.key)"
                >
                  <Download class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  @click="handleDeleteFile(file.key)"
                  class="text-red-600 hover:text-red-700"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>