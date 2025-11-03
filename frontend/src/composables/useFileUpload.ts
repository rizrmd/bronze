import { ref, computed } from 'vue'
import { uploadFile as uploadFileService } from '@/services/api/files'
import { validateFileName } from './useFileUtils'

export interface UploadProgress {
  fileName: string
  progress: number
  status: 'pending' | 'uploading' | 'success' | 'error'
  error?: string
}

export interface UploadOptions {
  onProgress?: (fileName: string, progress: number) => void
  onSuccess?: (fileName: string) => void
  onError?: (fileName: string, error: string) => void
}

export function useFileUpload() {
  const isUploading = ref(false)
  const uploadProgress = ref<UploadProgress[]>([])
  const uploadQueue = ref<File[]>([])

  const addFilesToQueue = (files: FileList | File[]) => {
    const validFiles: File[] = []
    
    for (const file of files) {
      const validation = validateFileName(file.name)
      
      if (!validation.isValid) {
        // Add to progress with error
        uploadProgress.value.push({
          fileName: file.name,
          progress: 0,
          status: 'error',
          error: validation.error
        })
        continue
      }
      
      validFiles.push(file)
      
      // Add to progress queue
      uploadProgress.value.push({
        fileName: file.name,
        progress: 0,
        status: 'pending'
      })
    }
    
    uploadQueue.value.push(...validFiles)
  }

  const clearQueue = () => {
    uploadQueue.value = []
    uploadProgress.value = []
  }

  const startUpload = async (targetPath: string = '', options: UploadOptions = {}) => {
    if (uploadQueue.value.length === 0) return
    
    isUploading.value = true
    
    try {
      for (const file of uploadQueue.value) {
        const progressIndex = uploadProgress.value.findIndex(p => p.fileName === file.name)
        
        if (progressIndex === -1) continue
        
        // Update status to uploading
        uploadProgress.value[progressIndex].status = 'uploading'
        
        try {
          // Create object name with target path
          const objectName = targetPath ? `${targetPath}/${file.name}` : file.name
          
          // Mock progress (in real implementation, you'd track actual upload progress)
          let progress = 0
          const progressInterval = setInterval(() => {
            progress += 10
            if (progress > 90) {
              clearInterval(progressInterval)
              return
            }
            
            uploadProgress.value[progressIndex].progress = progress
            options.onProgress?.(file.name, progress)
          }, 100)
          
          // Upload file
          await uploadFileService(file, objectName)
          
          clearInterval(progressInterval)
          
        // Update to success
        if (uploadProgress.value[progressIndex]) {
          uploadProgress.value[progressIndex].progress = 100
          uploadProgress.value[progressIndex].status = 'success'
        }
          
          options.onSuccess?.(file.name)
          
        } catch (error: any) {
          uploadProgress.value[progressIndex].status = 'error'
          uploadProgress.value[progressIndex].error = error.message || 'Upload failed'
          
          options.onError?.(file.name, error.message || 'Upload failed')
        }
      }
    } finally {
      isUploading.value = false
      uploadQueue.value = []
    }
  }

  const retryUpload = (fileName: string, targetPath: string = '', options: UploadOptions = {}) => {
    const progressIndex = uploadProgress.value.findIndex(p => p.fileName === fileName)
    
    if (progressIndex === -1) return
    
    // Find the original file
    const file = uploadQueue.value.find(f => f.name === fileName)
    if (!file) return
    
    // Reset progress
    uploadProgress.value[progressIndex].progress = 0
    uploadProgress.value[progressIndex].status = 'pending'
    uploadProgress.value[progressIndex].error = undefined
    
    // Re-add to queue
    if (!uploadQueue.value.includes(file)) {
      uploadQueue.value.push(file)
    }
  }

  const removeUpload = (fileName: string) => {
    const progressIndex = uploadProgress.value.findIndex(p => p.fileName === fileName)
    if (progressIndex !== -1) {
      uploadProgress.value.splice(progressIndex, 1)
    }
    
    const queueIndex = uploadQueue.value.findIndex(f => f.name === fileName)
    if (queueIndex !== -1) {
      uploadQueue.value.splice(queueIndex, 1)
    }
  }

  const uploadProgressStats = computed(() => {
    const total = uploadProgress.value.length
    const completed = uploadProgress.value.filter(p => p.status === 'success').length
    const failed = uploadProgress.value.filter(p => p.status === 'error').length
    const uploading = uploadProgress.value.filter(p => p.status === 'uploading').length
    const pending = uploadProgress.value.filter(p => p.status === 'pending').length
    
    return {
      total,
      completed,
      failed,
      uploading,
      pending,
      successRate: total > 0 ? (completed / total) * 100 : 0
    }
  })

  return {
    // State
    isUploading,
    uploadProgress,
    uploadQueue,
    uploadProgressStats,
    
    // Methods
    addFilesToQueue,
    clearQueue,
    startUpload,
    retryUpload,
    removeUpload
  }
}

export type UploadProgressStats = {
  total: number
  completed: number
  failed: number
  uploading: number
  pending: number
  successRate: number
}