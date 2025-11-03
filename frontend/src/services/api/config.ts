import { api } from './client'
import type { ApiResponse } from '@/types'

export async function getConfig(): Promise<ApiResponse<Record<string, string>>> {
  const { data } = await api.get('/api/config')
  return data
}

export async function updateConfig(config: Record<string, string>): Promise<ApiResponse> {
  const { data } = await api.put('/api/config', config)
  return data
}