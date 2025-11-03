// Global request management store
class RequestStore {
  private activeRequests = new Map<string, AbortController>()

  // Add a new request
  addRequest(key: string, controller: AbortController): void {
    // Cancel existing request with same key
    this.cancelRequest(key)
    this.activeRequests.set(key, controller)
  }

  // Cancel specific request
  cancelRequest(key: string): boolean {
    const controller = this.activeRequests.get(key)
    if (controller) {
      console.log(`Canceling request: ${key}`)
      controller.abort()
      this.activeRequests.delete(key)
      return true
    }
    return false
  }

  // Cancel all active requests
  cancelAllRequests(): void {
    console.log(`Canceling ${this.activeRequests.size} active requests`)
    this.activeRequests.forEach((controller, key) => {
      console.log(`Canceling request: ${key}`)
      controller.abort()
    })
    this.activeRequests.clear()
  }

  // Get active request count
  getActiveRequestCount(): number {
    return this.activeRequests.size
  }

  // Check if request is active
  isRequestActive(key: string): boolean {
    return this.activeRequests.has(key)
  }

  // Remove completed request (called by component when request completes)
  removeRequest(key: string): void {
    this.activeRequests.delete(key)
  }
}

// Create singleton instance
export const requestStore = new RequestStore()