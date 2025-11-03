import { ref, readonly, inject } from 'vue'
import { 
  browseData as browseDataService,
  listDataFiles as listDataFilesService,
  createExportJob as createExportJobService,
  exportSingleFile as exportSingleFileService,
  exportMultipleFiles as exportMultipleFilesService
} from '@/services'
import type { BrowseRequest, ExportRequest } from '@/types'

export function useDataBrowser() {
  const loading = ref(false)
  const error = ref<string | null>(null)
  const toast = inject('toast') as any

  const setLoading = (state: boolean) => {
    loading.value = state
  }

  const setError = (err: string | null) => {
    error.value = err
    if (err) {
      toast?.error('Error', err)
    }
  }

  const clearError = () => {
    error.value = null
  }

  const handleSuccess = (message: string) => {
    toast?.success('Success', message)
  }

  const browseData = async (request: BrowseRequest) => {
    setLoading(true)
    clearError()
    try {
      const response = await browseDataService(request)
      if (response.success) {
        return response
      } else {
        setError(response.message || 'Failed to browse data')
        return null
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  const listDataFiles = async () => {
    setLoading(true)
    clearError()
    try {
      const response = await listDataFilesService()
      if (response.success) {
        return response
      } else {
        setError(response.message || 'Failed to list data files')
        return null
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  const createExportJob = async (request: ExportRequest) => {
    setLoading(true)
    clearError()
    try {
      const response = await createExportJobService(request)
      if (response.success) {
        const jobMessage = response.job_id 
          ? `Export job created successfully. Job ID: ${response.job_id}`
          : 'File exported successfully'
        handleSuccess(jobMessage)
        return response
      } else {
        setError(response.message || 'Failed to create export job')
        return null
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  const exportSingleFile = async (request: ExportRequest) => {
    setLoading(true)
    clearError()
    try {
      const response = await exportSingleFileService(request)
      if (response.success) {
        handleSuccess('File exported successfully')
        return response
      } else {
        setError(response.message || 'Failed to export file')
        return null
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  const exportMultipleFiles = async (request: ExportRequest) => {
    setLoading(true)
    clearError()
    try {
      const response = await exportMultipleFilesService(request)
      if (response.success) {
        handleSuccess('Files exported successfully')
        return response
      } else {
        setError(response.message || 'Failed to export files')
        return null
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  return {
    loading: readonly(loading),
    error: readonly(error),
    browseData,
    listDataFiles,
    createExportJob,
    exportSingleFile,
    exportMultipleFiles,
    clearError,
    setError
  }
}
