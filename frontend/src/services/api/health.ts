import { api } from './client'

export async function healthCheck() {
  const { data } = await api.get('/api/health')
  return data
}

export async function getApiInfo() {
  const { data } = await api.get('/api')
  return data
}