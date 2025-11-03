import { ref, readonly, inject } from 'vue'
import { 
  uploadFile as uploadFileService,
  listFiles as listFilesService,
  createJob as createJobService,
  getJobs as getJobsService,
  cancelJob as cancelJobService,
  updateJobPriority as updateJobPriorityService,
  getJobStats as getJobStatsService,
  updateWorkerCount as updateWorkerCountService,
  getActiveJobs as getActiveJobsService,
  getUnprocessedEvents,
  getEventHistory,
  markEventProcessed as markEventProcessedService
} from '@/services'
import { deleteFile as deleteFileService, downloadFile as downloadFileService } from '@/services/api/files'
import type { JobStats, FileInfo, Job, FileEvent } from '@/types'

export function useApi() {
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

  return {
    loading: readonly(loading),
    error: readonly(error),
    setLoading,
    setError,
    clearError,
    handleSuccess
  }
}

export function useJobs() {
  const { loading, error, setLoading, setError, clearError, handleSuccess } = useApi()
  const jobs = ref<Job[]>([])
  const jobStats = ref<JobStats | null>(null)
  const activeJobs = ref<Job[]>([])

  const fetchJobs = async (status?: string) => {
    setLoading(true)
    clearError()
    try {
      const response = await getJobsService(status)
      if (response.success) {
        jobs.value = response.jobs || []
      } else {
        setError(response.message || 'Failed to fetch jobs')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const fetchJobStats = async () => {
    setLoading(true)
    clearError()
    try {
      const response = await getJobStatsService()
      if (response.success) {
        jobStats.value = response.data || null
      } else {
        setError(response.message || 'Failed to fetch job stats')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const fetchActiveJobs = async () => {
    setLoading(true)
    clearError()
    try {
      const response = await getActiveJobsService()
      if (response.success) {
        activeJobs.value = response.jobs || []
      } else {
        setError(response.message || 'Failed to fetch active jobs')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const createJob = async (jobData: any) => {
    setLoading(true)
    clearError()
    try {
      const response = await createJobService(jobData)
      if (response.success) {
        await fetchJobs()
        handleSuccess('Job created successfully')
        return response.job
      } else {
        setError(response.message || 'Failed to create job')
        return null
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  const cancelJob = async (jobId: string) => {
    setLoading(true)
    clearError()
    try {
      const response = await cancelJobService(jobId)
      if (response.success) {
        await fetchJobs()
        handleSuccess('Job cancelled successfully')
        return true
      } else {
        setError(response.message || 'Failed to cancel job')
        return false
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return false
    } finally {
      setLoading(false)
    }
  }

  const updateJobPriority = async (jobId: string, priority: 'low' | 'medium' | 'high') => {
    setLoading(true)
    clearError()
    try {
      const response = await updateJobPriorityService(jobId, priority)
      if (response.success) {
        await fetchJobs()
        handleSuccess('Job priority updated successfully')
        return true
      } else {
        setError(response.message || 'Failed to update job priority')
        return false
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return false
    } finally {
      setLoading(false)
    }
  }

  const updateWorkerCount = async (count: number) => {
    setLoading(true)
    clearError()
    try {
      const response = await updateWorkerCountService(count)
      if (response.success) {
        handleSuccess('Worker count updated successfully')
        return true
      } else {
        setError(response.message || 'Failed to update worker count')
        return false
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return false
    } finally {
      setLoading(false)
    }
  }

  return {
    loading,
    error,
    jobs,
    jobStats,
    activeJobs,
    fetchJobs,
    fetchJobStats,
    fetchActiveJobs,
    createJob,
    cancelJob,
    updateJobPriority,
    updateWorkerCount
  }
}

export function useFiles() {
  const { loading, error, setLoading, setError, clearError, handleSuccess } = useApi()
  const files = ref<FileInfo[]>([])

  const fetchFiles = async (prefix?: string) => {
    setLoading(true)
    clearError()
    try {
      const response = await listFilesService(prefix)
      if (response.success) {
        files.value = response.files || []
      } else {
        setError(response.message || 'Failed to fetch files')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const uploadFile = async (file: File, objectName?: string) => {
    setLoading(true)
    clearError()
    try {
      const response = await uploadFileService(file, objectName)
      if (response.success) {
        await fetchFiles()
        handleSuccess('File uploaded successfully')
        return response
      } else {
        setError(response.message || 'Failed to upload file')
        return null
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  const deleteFile = async (filename: string) => {
    setLoading(true)
    clearError()
    try {
      const response = await deleteFileService(filename)
      if (response.success) {
        await fetchFiles()
        handleSuccess('File deleted successfully')
        return true
      } else {
        setError(response.message || 'Failed to delete file')
        return false
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return false
    } finally {
      setLoading(false)
    }
  }

  const downloadFile = async (filename: string) => {
    setLoading(true)
    clearError()
    try {
      const blob = await downloadFileService(filename)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
      return true
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return false
    } finally {
      setLoading(false)
    }
  }

  return {
    loading,
    error,
    files,
    fetchFiles,
    uploadFile,
    deleteFile,
    downloadFile
  }
}

export function useWatcher() {
  const { loading, error, setLoading, setError, clearError, handleSuccess } = useApi()
  const events = ref<FileEvent[]>([])
  const unprocessedEvents = ref<FileEvent[]>([])

  const fetchEventHistory = async (limit?: number) => {
    setLoading(true)
    clearError()
    try {
      const response = await getEventHistory(limit)
      if (response.success) {
        events.value = response.events || []
      } else {
        setError(response.message || 'Failed to fetch event history')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const fetchUnprocessedEvents = async (limit?: number) => {
    setLoading(true)
    clearError()
    try {
      const response = await getUnprocessedEvents(limit)
      if (response.success) {
        unprocessedEvents.value = response.events || []
      } else {
        setError(response.message || 'Failed to fetch unprocessed events')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const markEventProcessed = async (eventId: string) => {
    setLoading(true)
    clearError()
    try {
      const response = await markEventProcessedService(eventId)
      if (response.success) {
        await fetchUnprocessedEvents()
        await fetchEventHistory()
        handleSuccess('Event marked as processed successfully')
        return true
      } else {
        setError(response.message || 'Failed to mark event as processed')
        return false
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return false
    } finally {
      setLoading(false)
    }
  }

  return {
    loading,
    error,
    events,
    unprocessedEvents,
    fetchEventHistory,
    fetchUnprocessedEvents,
    markEventProcessed
  }
}