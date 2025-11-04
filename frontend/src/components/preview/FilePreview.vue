<template>
  <div class="bg-white rounded-lg shadow-sm border min-h-[600px]">
    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center h-96">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p class="text-gray-500">Loading preview...</p>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="flex items-center justify-center h-96">
      <div class="text-center text-red-600">
        <AlertCircle class="w-12 h-12 mx-auto mb-4" />
        <p>{{ error }}</p>
      </div>
    </div>

    <!-- Image Preview -->
    <div v-else-if="isImage" class="p-4">
      <img 
        :src="presignedUrl" 
        :alt="getFileName()"
        class="max-w-full h-auto mx-auto rounded"
        @load="loading = false"
        @error="handleImageError"
      />
    </div>

    <!-- PDF Preview -->
    <div v-else-if="isPDF" class="p-4 h-96">
      <iframe 
        :src="presignedUrl" 
        class="w-full h-full border rounded"
        @load="loading = false"
        @error="handleError"
      />
    </div>

    <!-- Text File Preview -->
    <div v-else-if="isText" class="p-4">
      <pre class="bg-gray-50 p-4 rounded overflow-auto max-h-96 text-sm font-mono">{{ textContent }}</pre>
    </div>

    <!-- Video Preview -->
    <div v-else-if="isVideo" class="p-4">
      <video 
        :src="presignedUrl" 
        controls 
        class="max-w-full h-auto mx-auto rounded"
        @loadeddata="loading = false"
        @error="handleError"
      />
    </div>

    <!-- Audio Preview -->
    <div v-else-if="isAudio" class="p-8">
      <audio 
        :src="presignedUrl" 
        controls 
        class="w-full max-w-md mx-auto"
        @loadeddata="loading = false"
        @error="handleError"
      />
    </div>

    <!-- Structured Data (Excel/CSV) Preview -->
    <div v-else-if="isData" class="p-4">
      <DataPreview 
        :file="file" 
        :presigned-url="presignedUrl"
        @error="handleError"
      />
    </div>

    <!-- Unsupported File Type -->
    <div v-else class="flex items-center justify-center h-96">
      <div class="text-center text-gray-500">
        <File class="w-12 h-12 mx-auto mb-4" />
        <p class="font-medium">Preview not available</p>
        <p class="text-sm mt-2">Download the file to view its contents</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { AlertCircle, File } from 'lucide-vue-next'
import DataPreview from './DataPreview.vue'
import type { FileInfo } from '@/types'

interface Props {
  file: FileInfo
  presignedUrl: string
}

interface Emits {
  (e: 'error', error: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const loading = ref(true)
const error = ref<string>('')
const textContent = ref<string>('')

const getFileName = () => {
  return props.file.key?.split('/')?.pop() || 'Unknown'
}

const getFileExtension = () => {
  const fileName = getFileName()
  return fileName.split('.').pop()?.toLowerCase() || ''
}

const isImage = computed(() => {
  const ext = getFileExtension()
  return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp'].includes(ext)
})

const isPDF = computed(() => {
  return getFileExtension() === 'pdf'
})

const isText = computed(() => {
  const ext = getFileExtension()
  return ['txt', 'log', 'md', 'json', 'xml', 'yaml', 'yml', 'csv', 'ts', 'js', 'vue', 'go', 'py'].includes(ext)
})

const isVideo = computed(() => {
  const ext = getFileExtension()
  return ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv'].includes(ext)
})

const isAudio = computed(() => {
  const ext = getFileExtension()
  return ['mp3', 'wav', 'ogg', 'flac', 'aac', 'm4a'].includes(ext)
})

const isData = computed(() => {
  const ext = getFileExtension()
  return ['xlsx', 'xls', 'csv', 'mdb'].includes(ext)
})

const loadTextContent = async () => {
  try {
    const response = await fetch(props.presignedUrl)
    const text = await response.text()
    textContent.value = text
  } catch (err) {
    console.error('Failed to load text content:', err)
    emit('error', 'Failed to load text content')
  }
}

const handleImageError = () => {
  emit('error', 'Failed to load image')
}

const handleError = () => {
  emit('error', 'Failed to load preview')
}

onMounted(() => {
  if (isText.value && !isData.value) {
    loadTextContent().then(() => {
      loading.value = false
    })
  } else if (!isImage.value && !isPDF.value && !isVideo.value && !isAudio.value && !isData.value) {
    loading.value = false
  }
})
</script>