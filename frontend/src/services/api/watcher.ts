import { api } from './client'
import type { FileEventsResponse, ApiResponse } from '@/types'

export async function getUnprocessedEvents(limit?: number): Promise<FileEventsResponse> {
  const params = limit ? { limit } : {}
  const { data } = await api.get('/api/watcher/events/unprocessed', { params })
  return data
}

export async function getEventHistory(limit?: number): Promise<FileEventsResponse> {
  const params = limit ? { limit } : {}
  const { data } = await api.get('/api/watcher/events/history', { params })
  return data
}

export async function markEventProcessed(eventId: string): Promise<ApiResponse> {
  const { data } = await api.post('/api/watcher/events/mark-processed', { event_id: eventId })
  return data
}