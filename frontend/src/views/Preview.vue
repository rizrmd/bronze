<template>
  <div class="h-screen flex flex-col">
    <header class="bg-white border-b px-6 py-4 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <button 
          @click="goBack" 
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft class="w-5 h-5" />
        </button>
        <h1 class="text-xl font-semibold">
          {{ getFileName(file) }}
        </h1>
      </div>
      <div class="flex items-center gap-3">
        <button 
          @click="downloadFile"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
        >
          <Download class="w-4 h-4" />
          Download
        </button>
      </div>
    </header>

    <main class="flex-1 bg-gray-50 p-6 overflow-auto">
      <div class="max-w-6xl mx-auto">
        <FilePreview 
          v-if="file"
          :file="file" 
          :presigned-url="presignedUrl"
          @error="handleError"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeft, Download } from 'lucide-vue-next'
import { getPresignedUrl } from '@/services/api/files'
import { downloadFile as downloadFileService } from '@/services/api/files'
import FilePreview from '@/components/preview/FilePreview.vue'
import type { FileInfo } from '@/types'

const router = useRouter()
const route = useRoute()

const file = ref<FileInfo | null>(null)
const presignedUrl = ref<string>('')
const loading = ref(true)
const error = ref<string>('')

const fileName = computed(() => route.query.file as string)
const filePath = computed(() => route.query.path as string || '')

const getFileName = (file: FileInfo | null) => {
  if (!file) return 'Preview'
  return file.key?.split('/')?.pop() || 'Unknown'
}

const goBack = () => {
  if (filePath.value) {
    router.push(`/files/${filePath.value}`)
  } else {
    router.push('/files')
  }
}

const loadPreview = async () => {
  if (!fileName.value) {
    error.value = 'No file specified'
    loading.value = false
    return
  }

  try {
    // Get presigned URL for preview
    const fullPath = filePath.value ? `${filePath.value}/${fileName.value}` : fileName.value
    const response = await getPresignedUrl(fullPath)
    presignedUrl.value = response.url

    // Create file info object
    file.value = {
      key: fullPath,
      size: 0, // Would need to fetch file info if needed
      last_modified: new Date().toISOString(),
      etag: ''
    }

    loading.value = false
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load preview'
    loading.value = false
  }
}

const downloadFile = async () => {
  if (!file.value) return
  
  try {
    const blob = await downloadFileService(file.value.key)
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = getFileName(file.value)
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (err) {
    console.error('Download failed:', err)
  }
}

const handleError = (err: string) => {
  error.value = err
}

onMounted(() => {
  loadPreview()
})
</script>