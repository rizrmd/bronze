// Helper utilities for handling abort errors consistently across the application

/**
 * Checks if an error is an abort error (intentional cancellation)
 * @param error - The error to check
 * @param wasAborted - Whether the request was aborted (optional)
 * @returns true if error is an abort error, false otherwise
 */
export function isAbortError(error: any, wasAborted?: boolean): boolean {
  const has500Error = error.message?.includes('HTTP error! status: 500')
  
  return (
    error.name === 'AbortError' ||
    error.message?.includes('Request aborted') ||
    error.message?.includes('The user aborted a request') ||
    error.message?.includes('net::ERR_ABORTED') ||
    error.code === 'ERR_CANCELED' ||
    // 500 errors after abort (only if wasAborted is true)
    (has500Error && wasAborted) ||
    // Axios cancellation
    (typeof error.__CANCEL__ === 'boolean' && error.__CANCEL__) ||
    // DOMException for abort
    (error instanceof DOMException && error.name === 'AbortError')
  )
}

/**
 * Handles abort errors by silently suppressing them and rethrowing other errors
 * @param error - The error to handle
 * @param onError - Optional callback to call for non-abort errors
 */
export function handleAbortError(error: any, onError?: (error: any) => void): void {
  if (isAbortError(error)) {
    console.log('Request cancelled by user or navigation')
    return
  }
  
  // It's not an abort error, handle it normally
  if (onError) {
    onError(error)
  }
}