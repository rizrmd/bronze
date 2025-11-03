import { api } from './client'
import type { UploadResponse, FileListResponse } from '@/types'

export interface SSEEventCallback {
  (event: string, data: any): void
}

export interface SSEErrorCallback {
  (error: Error): void
}

export async function uploadFile(file: File, objectName?: string): Promise<UploadResponse> {
  const formData = new FormData()
  formData.append('file', file)
  if (objectName) {
    formData.append('object_name', objectName)
  }

  const { data } = await api.post('/api/files', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
  return data
}

export async function listFiles(prefix?: string): Promise<FileListResponse> {
  const params = prefix ? { prefix } : {}
  const { data } = await api.get('/api/files', { params })
  return data
}

export async function browseFolders(folders: any[], onEvent: SSEEventCallback, onError?: SSEErrorCallback): Promise<void> {
  console.log('browseFolders SSE called with:', { folders })
  
  try {
    const response = await fetch('/api/files/browse', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
        'Cache-Control': 'no-cache',
      },
      body: JSON.stringify({ folders }),
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    if (!response.body) {
      throw new Error('Response body is null')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      const chunk = decoder.decode(value, { stream: true })
      const lines = chunk.split('\n')

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i]?.trim()
        
        if (line?.startsWith('event: ')) {
          const event = line.substring(7)
          let data = ''
          
          // Look for data lines that follow
          for (let j = i + 1; j < lines.length; j++) {
            const dataLine = lines[j]?.trim()
            if (dataLine?.startsWith('data: ')) {
              data += dataLine.substring(6) + '\n'
            } else if (!dataLine?.startsWith('data: ')) {
              // End of this event
              i = j - 1 // Skip processed lines
              break
            }
          }
          
          // Parse JSON data
          if (data.trim()) {
            try {
              const parsedData = JSON.parse(data.trim())
              console.log(`SSE Event: ${event}`, parsedData)
              onEvent(event, parsedData)
            } catch (e) {
              console.error('Error parsing SSE data:', data, e)
            }
          }
        }
      }
    }
  } catch (error: any) {
    console.error('browseFolders error:', error)
    if (onError) {
      onError(error)
    }
    throw error
  }
}

export async function getFileInfo(filename: string) {
  const { data } = await api.get(`/api/files/${filename}`)
  return data
}

export async function downloadFile(filename: string): Promise<Blob> {
  const { data } = await api.get(`/api/files/${filename}`, {
    responseType: 'blob'
  })
  return data
}

export async function deleteFile(filename: string) {
  const { data } = await api.delete(`/api/files/${filename}`)
  return data
}

export async function getPresignedUrl(filename: string, expiry?: string) {
  const params = expiry ? { expiry } : {}
  const { data } = await api.get(`/api/files/${filename}/presigned`, { params })
  return data
}