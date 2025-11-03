import { api } from './client'
import type { CreateJobRequest, JobResponse, JobListResponse, ApiResponse, JobStats } from '@/types'

export async function createJob(jobData: CreateJobRequest): Promise<JobResponse> {
  const { data } = await api.post('/api/jobs', jobData)
  return data
}

export async function getJobs(status?: string): Promise<JobListResponse> {
  const params = status ? { status } : {}
  const { data } = await api.get('/api/jobs', { params })
  return data
}

export async function getJob(id: string): Promise<JobResponse> {
  const { data } = await api.get(`/api/jobs/${id}`)
  return data
}

export async function cancelJob(id: string): Promise<ApiResponse> {
  const { data } = await api.delete(`/api/jobs/${id}`)
  return data
}

export async function updateJobPriority(id: string, priority: 'low' | 'medium' | 'high'): Promise<ApiResponse> {
  const { data } = await api.put(`/api/jobs/${id}/priority`, { priority })
  return data
}

export async function getJobStats(): Promise<ApiResponse<JobStats>> {
  const { data } = await api.get('/api/jobs/stats')
  return data
}

export async function updateWorkerCount(count: number): Promise<ApiResponse> {
  const { data } = await api.put('/api/jobs/workers', { count })
  return data
}

export async function getActiveJobs(): Promise<JobListResponse> {
  const { data } = await api.get('/api/jobs/workers/active')
  return data
}