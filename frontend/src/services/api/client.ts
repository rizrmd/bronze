import axios from 'axios'
import { isAbortError } from '@/utils/abortUtils'

const API_BASE_URL = import.meta.env.VITE_API_URL || ''

export const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Silently ignore abort errors - they're intentional cancellations
    if (isAbortError(error)) {
      return Promise.reject(error) // Don't log, but still reject to allow proper handling
    }
    
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

export default api