import axios from 'axios'
import type { 
  ApiResponse, 
  JobStats, 
  FileInfo, 
  UploadResponse, 
  CreateJobRequest,
  JobListResponse,
  JobResponse,
  FileListResponse,
  FileEventsResponse
} from '@/types'

const API_BASE_URL = import.meta.env.VITE_API_URL || ''

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

export const apiClient = {
  // Health & Info
  async healthCheck(): Promise<ApiResponse> {
    const { data } = await api.get('/')
    return data
  },

  async getApiInfo(): Promise<ApiResponse> {
    const { data } = await api.get('/api')
    return data
  },

  // Files
  async uploadFile(file: File, objectName?: string): Promise<UploadResponse> {
    const formData = new FormData()
    formData.append('file', file)
    if (objectName) {
      formData.append('object_name', objectName)
    }

    const { data } = await api.post('/files', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    return data
  },

  async listFiles(prefix?: string): Promise<FileListResponse> {
    const params = prefix ? { prefix } : {}
    const { data } = await api.get('/files', { params })
    return data
  },

  // Multi-folder browsing
  async browseFolders(folders: any[], limit?: number): Promise<any> {
    console.log('API browseFolders called with:', { folders, limit })
    try {
      const response = await api.post('/api/files/browse', { folders, limit })
      console.log('API browseFolders response:', response)
      console.log('API browseFolders response data:', response.data)
      return response.data
    } catch (error: any) {
      console.error('API browseFolders error:', error)
      console.error('Error response:', error.response?.data)
      throw error
    }
  },

  async getFileInfo(filename: string): Promise<ApiResponse<FileInfo>> {
    const { data } = await api.get(`/files/${filename}`)
    return data
  },

  async downloadFile(filename: string): Promise<Blob> {
    const { data } = await api.get(`/files/${filename}`, {
      responseType: 'blob'
    })
    return data
  },

  async deleteFile(filename: string): Promise<ApiResponse> {
    const { data } = await api.delete(`/files/${filename}`)
    return data
  },

  async getPresignedUrl(filename: string, expiry?: string): Promise<ApiResponse<{ url: string }>> {
    const params = expiry ? { expiry } : {}
    const { data } = await api.get(`/files/${filename}/presigned`, { params })
    return data
  },

  // Jobs
  async createJob(jobData: CreateJobRequest): Promise<JobResponse> {
    const { data } = await api.post('/jobs', jobData)
    return data
  },

  async getJobs(status?: string): Promise<JobListResponse> {
    const params = status ? { status } : {}
    const { data } = await api.get('/jobs', { params })
    return data
  },

  async getJob(id: string): Promise<JobResponse> {
    const { data } = await api.get(`/jobs/${id}`)
    return data
  },

  async cancelJob(id: string): Promise<ApiResponse> {
    const { data } = await api.delete(`/jobs/${id}`)
    return data
  },

  async updateJobPriority(id: string, priority: 'low' | 'medium' | 'high'): Promise<ApiResponse> {
    const { data } = await api.put(`/jobs/${id}/priority`, { priority })
    return data
  },

  async getJobStats(): Promise<ApiResponse<JobStats>> {
    const { data } = await api.get('/jobs/stats')
    return data
  },

  async updateWorkerCount(count: number): Promise<ApiResponse> {
    const { data } = await api.put('/jobs/workers', { count })
    return data
  },

  async getActiveJobs(): Promise<JobListResponse> {
    const { data } = await api.get('/jobs/workers/active')
    return data
  },

  // Watcher Events
  async getUnprocessedEvents(limit?: number): Promise<FileEventsResponse> {
    const params = limit ? { limit } : {}
    const { data } = await api.get('/watcher/events/unprocessed', { params })
    return data
  },

  async getEventHistory(limit?: number): Promise<FileEventsResponse> {
    const params = limit ? { limit } : {}
    const { data } = await api.get('/watcher/events/history', { params })
    return data
  },

  async markEventProcessed(eventId: string): Promise<ApiResponse> {
    const { data } = await api.post('/watcher/events/mark-processed', { event_id: eventId })
    return data
  },

  // Configuration
  async getConfig(): Promise<ApiResponse<Record<string, string>>> {
    const { data } = await api.get('/config')
    return data
  },

  async updateConfig(config: Record<string, string>): Promise<ApiResponse> {
    const { data } = await api.put('/config', config)
    return data
  }
}

export { api }
export default apiClient