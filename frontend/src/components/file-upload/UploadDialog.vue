<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-4xl max-h-[80vh] overflow-auto">
      <DialogHeader>
        <DialogTitle>Upload Files</DialogTitle>
      </DialogHeader>
      
      <div class="space-y-6">
        <!-- Drop Zone -->
        <div
          class="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-gray-400 transition-colors"
          :class="{ 'border-blue-400 bg-blue-50': dragOver }"
          @drop="handleDrop"
          @dragover="handleDragOver"
          @dragleave="handleDragLeave"
        >
          <Upload class="mx-auto h-12 w-12 text-gray-400 mb-4" />
          <p class="text-lg text-gray-600 mb-2">Drop files here to upload</p>
          <p class="text-sm text-gray-500 mb-4">or</p>
          <Button @click="fileInput?.click()">
            Choose Files
          </Button>
          <input
            ref="fileInput"
            type="file"
            multiple
            class="hidden"
            @change="handleFileSelect"
          />
        </div>
        
        <!-- Upload Queue -->
        <div v-if="uploadProgress.length > 0" class="space-y-4">
          <h3 class="text-lg font-semibold">Upload Queue</h3>
          
          <div class="space-y-2 max-h-60 overflow-y-auto">
            <div
              v-for="progress in uploadProgress"
              :key="progress.fileName"
              class="flex items-center justify-between p-3 border rounded-lg"
              :class="{
                'bg-green-50 border-green-200': progress.status === 'success',
                'bg-red-50 border-red-200': progress.status === 'error',
                'bg-blue-50 border-blue-200': progress.status === 'uploading',
                'bg-gray-50 border-gray-200': progress.status === 'pending'
              }"
            >
              <div class="flex items-center space-x-3 flex-1">
                <File class="w-5 h-5 text-gray-500" />
                <div class="flex-1">
                  <div class="font-medium">{{ progress.fileName }}</div>
                  <div class="text-sm text-gray-500">
                    {{ getStatusText(progress.status) }}
                    <span v-if="progress.progress > 0 && progress.status !== 'success'">
                      ({{ progress.progress }}%)
                    </span>
                  </div>
                  <div v-if="progress.error" class="text-sm text-red-500">
                    {{ progress.error }}
                  </div>
                </div>
              </div>
              
              <div class="flex items-center space-x-2">
                <Button
                  v-if="progress.status === 'error'"
                  @click="$emit('retry', progress.fileName)"
                  variant="outline"
                  size="sm"
                >
                  Retry
                </Button>
                
                <Button
                  @click="$emit('remove', progress.fileName)"
                  variant="ghost"
                  size="sm"
                  :disabled="progress.status === 'uploading'"
                >
                  ✕
                </Button>
              </div>
            </div>
            
            <!-- Progress Bar for Active Upload -->
            <div v-if="uploadProgressStats.uploading > 0" class="w-full">
              <div class="flex items-center justify-between text-sm text-gray-500 mb-1">
                <span>Overall Progress</span>
                <span>{{ overallProgress }}%</span>
              </div>
              <div class="w-full bg-gray-200 rounded-full h-2">
                <div
                  class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                  :style="{ width: overallProgress + '%' }"
                ></div>
              </div>
            </div>
          </div>
          
          <!-- Upload Stats -->
          <div class="flex items-center justify-between text-sm text-gray-600 pt-2 border-t">
            <span>
              {{ uploadProgressStats.total }} files • 
              {{ uploadProgressStats.completed }} completed • 
              {{ uploadProgressStats.failed }} failed
            </span>
            
            <div class="space-x-2">
              <Button
                v-if="!isUploading && uploadProgressStats.pending > 0"
                @click="$emit('start-upload')"
                variant="default"
                size="sm"
              >
                Start Upload ({{ uploadProgressStats.pending }})
              </Button>
              
              <Button
                v-if="isUploading"
                @click="$emit('cancel-upload')"
                variant="outline"
                size="sm"
              >
                Cancel
              </Button>
              
              <Button
                @click="$emit('clear-queue')"
                variant="outline"
                size="sm"
                :disabled="isUploading"
              >
                Clear Queue
              </Button>
            </div>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Upload, File } from 'lucide-vue-next'
import type { UploadProgress, UploadProgressStats } from '@/composables/useFileUpload'

interface Props {
  isOpen: boolean
  isUploading: boolean
  uploadProgress: UploadProgress[]
  uploadProgressStats: UploadProgressStats
}

interface Emits {
  (e: 'update:isOpen', value: boolean): void
  (e: 'files-selected', files: File[]): void
  (e: 'start-upload'): void
  (e: 'cancel-upload'): void
  (e: 'clear-queue'): void
  (e: 'retry', fileName: string): void
  (e: 'remove', fileName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const fileInput = ref<HTMLInputElement>()
const dragOver = ref(false)

const overallProgress = computed(() => {
  if (props.uploadProgressStats.total === 0) return 0
  return Math.round(
    ((props.uploadProgressStats.completed + props.uploadProgressStats.uploading) / 
     props.uploadProgressStats.total) * 100
  )
})

const getStatusText = (status: string) => {
  const statusTexts: Record<string, string> = {
    pending: 'Waiting to upload',
    uploading: 'Uploading...',
    success: 'Upload complete',
    error: 'Upload failed'
  }
  return statusTexts[status] || status
}

const handleDrop = (event: DragEvent) => {
  event.preventDefault()
  dragOver.value = false
  
  const files = Array.from(event.dataTransfer?.files || [])
  if (files.length > 0) {
    emit('files-selected', files)
  }
}

const handleDragOver = (event: DragEvent) => {
  event.preventDefault()
  dragOver.value = true
}

const handleDragLeave = () => {
  dragOver.value = false
}

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = Array.from(target.files || [])
  if (files.length > 0) {
    emit('files-selected', files)
  }
}

// Handle v-model
const isOpen = computed({
  get: () => props.isOpen,
  set: (value) => emit('update:isOpen', value)
})
</script>