import { api } from './client'
import type { BrowseRequest, BrowseResponse, DataFileListResponse, ExportRequest, ExportResponse } from '@/types'

export async function browseData(request: BrowseRequest): Promise<BrowseResponse> {
  const { data } = await api.post('/api/data/browse', request)
  return data
}

export async function listDataFiles(): Promise<DataFileListResponse> {
  const { data } = await api.get('/api/data/files')
  return data
}

export async function createExportJob(request: ExportRequest): Promise<ExportResponse> {
  const { data } = await api.post('/api/data/export-job', request)
  return data
}

export async function exportSingleFile(request: ExportRequest): Promise<ExportResponse> {
  const { data } = await api.post('/api/data/export-single', request)
  return data
}

export async function exportMultipleFiles(request: ExportRequest): Promise<ExportResponse> {
  const { data } = await api.post('/api/data/export-multiple', request)
  return data
}