<template>
  <div class="h-full">
    <FileBrowser
      :initial-path="initialPath"
      :initial-view-mode="initialViewMode"
      :use-s-s-e="useSSE"
    />
    
    <!-- Upload Dialog -->
    <UploadDialog
      v-model:is-open="uploadDialogOpen"
      :is-uploading="isUploading"
      :upload-progress="uploadProgress"
      :upload-progress-stats="uploadProgressStats"
      @files-selected="handleFileSelection"
      @start-upload="handleUploadStart"
      @cancel-upload="handleUploadCancel"
      @clear-queue="handleUploadClearQueue"
      @retry="handleUploadRetry"
      @remove="handleUploadRemove"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import FileBrowser from '@/components/file-browser/FileBrowser.vue'
import UploadDialog from '@/components/file-upload/UploadDialog.vue'
import { useFileUpload } from '@/composables/useFileUpload'

interface Props {
  initialPath?: string
  initialViewMode?: 'list' | 'grid'
  useSSE?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  initialViewMode: 'list',
  useSSE: false
})

// File upload
const {
  isUploading,
  uploadProgress,
  uploadProgressStats,
  addFilesToQueue,
  startUpload,
  clearQueue
} = useFileUpload()

// Dialog states
const uploadDialogOpen = ref(false)

// File upload handlers
const handleFileSelection = (files: File[]) => {
  addFilesToQueue(files)
  // Auto-open upload dialog when files are selected
  if (!uploadDialogOpen.value) {
    uploadDialogOpen.value = true
  }
}

const handleUploadStart = async () => {
  await startUpload(props.initialPath || '', {
    onProgress: (fileName, progress) => {
      console.log(`Upload progress for ${fileName}: ${progress}%`)
    },
    onSuccess: (fileName) => {
      console.log(`Upload complete for ${fileName}`)
    },
    onError: (fileName, error) => {
      console.error(`Upload failed for ${fileName}: ${error}`)
    }
  })
}

const handleUploadCancel = () => {
  // TODO: Implement upload cancellation
  console.log('Upload cancelled')
}

const handleUploadClearQueue = () => {
  clearQueue()
}

const handleUploadRetry = (fileName: string) => {
  // TODO: Implement upload retry
  console.log('Retrying upload for:', fileName)
}

const handleUploadRemove = (fileName: string) => {
  // The useFileUpload composable handles this internally
  console.log('Removed from queue:', fileName)
}

// Expose upload dialog control to parent components
defineExpose({
  openUploadDialog: () => { uploadDialogOpen.value = true },
  closeUploadDialog: () => { uploadDialogOpen.value = false }
})
</script>