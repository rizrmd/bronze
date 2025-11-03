import { api } from './client'
import type { UploadResponse, FileListResponse } from '@/types'
import { isAbortError } from '@/utils/abortUtils'

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

export async function browseFolders(folders: any[], onEvent: SSEEventCallback, onError?: SSEErrorCallback, abortController?: AbortController): Promise<void> {
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
      signal: abortController?.signal,
    })

    // Early exit if request was aborted during fetch
    if (abortController?.signal.aborted) {
      console.log('browseFolders request cancelled during fetch')
      return
    }

    // Check if response is ok, but first verify it's not due to cancellation
    if (!response.ok) {
      const wasAborted = abortController?.signal.aborted
      const is500Error = response.status === 500
      
      if (wasAborted || is500Error) {
        console.log('browseFolders request cancelled, ignoring response error')
        return
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    if (!response.body) {
      throw new Error('Response body is null')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    
    // Set up abort handler to cancel the reader immediately
    if (abortController) {
      abortController.signal.addEventListener('abort', () => {
        try {
          reader.cancel()
        } catch (e) {
          // Ignore cancel errors
        }
      })
    }

    while (true) {
      // Check if request was aborted before attempting to read
      if (abortController?.signal.aborted) {
        console.log('browseFolders request cancelled before read')
        return
      }
      
      const { done, value } = await reader.read()
      
      // Check if request was aborted during read
      if (abortController?.signal.aborted) {
        console.log('browseFolders request cancelled during read')
        return
      }
      
      if (done) break

      const chunk = decoder.decode(value, { stream: true })
      
      // Check if request was aborted after reading chunk
      if (abortController?.signal.aborted) {
        console.log('browseFolders request cancelled after read')
        return
      }
      
      const lines = chunk.split('\n')

      for (let i = 0; i < lines.length; i++) {
        // Check if request was aborted before processing each line
        if (abortController?.signal.aborted) {
          console.log('browseFolders request cancelled during line processing')
          return
        }
        
        const line = lines[i]?.trim()
        
        if (line?.startsWith('event: ')) {
          // Check if request was aborted before processing event
          if (abortController?.signal.aborted) {
            console.log('browseFolders request cancelled before event processing')
            return
          }
          
          const event = line.substring(7)
          let data = ''
          
          // Look for data lines that follow
          for (let j = i + 1; j < lines.length; j++) {
            // Check if request was aborted during data collection
            if (abortController?.signal.aborted) {
              console.log('browseFolders request cancelled during data collection')
              return
            }
            
            const dataLine = lines[j]?.trim()
            if (dataLine?.startsWith('data: ')) {
              data += dataLine.substring(6) + '\n'
            } else if (!dataLine?.startsWith('data: ')) {
              // End of this event
              i = j - 1 // Skip processed lines
              break
            }
          }
          
          // Final abort check before calling callback
          if (abortController?.signal.aborted) {
            console.log('browseFolders request cancelled before callback')
            return
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
    // Silently ignore abort errors - they're intentional cancellations, not real errors
    // Also check if the HTTP error occurred after abortion (common with 500 errors)
    const wasAborted = abortController?.signal.aborted
    const isAbortErrorType = isAbortError(error, wasAborted)
    
    if (wasAborted || isAbortErrorType) {
      console.log('browseFolders request cancelled')
      return
    }
    
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