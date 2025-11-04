import { ref } from 'vue'

// Global store for managing active requests and abort controllers
interface RequestInfo {
  controller: AbortController
  url: string // Track which URL/path this request is for
  createdAt: number // Track when request was created
}

export const requestStore = {
  // Map of request keys to AbortControllers
  activeRequests: ref<Map<string, RequestInfo>>(new Map()),

  // Add a request to the store
  addRequest(key: string, controller: AbortController, url?: string): void {
    // Cancel existing request with same key
    this.cancelRequest(key)
    
    const requestInfo: RequestInfo = {
      controller,
      url: url || key,
      createdAt: Date.now()
    }
    this.activeRequests.value.set(key, requestInfo)
  },

  // Remove a request from the store
  removeRequest(key: string): void {
    this.activeRequests.value.delete(key)
  },

  // Get request info by key
  getRequest(key: string): RequestInfo | undefined {
    return this.activeRequests.value.get(key)
  },

  // Cancel a specific request
  cancelRequest(key: string): boolean {
    const requestInfo = this.activeRequests.value.get(key)
    if (requestInfo) {
      console.log(`Canceling request: ${key} for URL: ${requestInfo.url}`)
      requestInfo.controller.abort()
      this.activeRequests.value.delete(key)
      return true
    }
    return false
  },

  // Cancel requests for specific URL (most important method)
  cancelRequestsForUrl(url: string): void {
    console.log(`Canceling all requests for URL: ${url}`)
    const requestsToCancel: string[] = []
    
    this.activeRequests.value.forEach((requestInfo, key) => {
      // Cancel if request is for specified URL
      if (requestInfo.url === url) {
        requestsToCancel.push(key)
      }
    })
    
    requestsToCancel.forEach(key => {
      this.cancelRequest(key)
    })
  },

  // Cancel all requests except those for specific URL (for navigation)
  cancelAllExceptForUrl(keepUrl: string): void {
    console.log(`Canceling all requests except for URL: ${keepUrl}`)
    const requestsToCancel: string[] = []
    
    this.activeRequests.value.forEach((requestInfo, key) => {
      // Cancel if request is NOT for the target URL
      if (requestInfo.url !== keepUrl) {
        requestsToCancel.push(key)
      }
    })
    
    requestsToCancel.forEach(key => {
      this.cancelRequest(key)
    })
  },

  // Cancel all active requests (fallback method)
  cancelAllRequests(): void {
    console.log(`Canceling ${this.activeRequests.value.size} active requests`)
    this.activeRequests.value.forEach((requestInfo, key) => {
      console.log(`Canceling request: ${key} for URL: ${requestInfo.url}`)
      requestInfo.controller.abort()
    })
    this.activeRequests.value.clear()
  },

  // Get active request count
  getActiveRequestCount(): number {
    return this.activeRequests.value.size
  },

  // Get all active URLs
  getActiveUrls(): string[] {
    return Array.from(this.activeRequests.value.values()).map(info => info.url)
  },

  // Check if request is active
  isRequestActive(key: string): boolean {
    return this.activeRequests.value.has(key)
  },

  // Clean up old requests (fallback safety net)
  cleanupOldRequests(maxAge: number = 30000): void {
    const now = Date.now()
    const requestsToCancel: string[] = []
    
    this.activeRequests.value.forEach((requestInfo, key) => {
      if (now - requestInfo.createdAt > maxAge) {
        console.log(`Cleaning up old request: ${key} (${now - requestInfo.createdAt}ms old)`)
        requestsToCancel.push(key)
      }
    })
    
    requestsToCancel.forEach(key => {
      this.cancelRequest(key)
    })
  }
}