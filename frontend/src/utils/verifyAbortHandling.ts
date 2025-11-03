// Verification script for 500 error handling
// This specifically tests the net::ERR_ABORTED 500 scenario

import { requestStore } from '@/stores/requestStore'

// Simulate the exact error pattern from browser console
const simulatedAbort500Error = {
  message: 'HTTP error! status: 500',
  stack: 'Error: HTTP error! status: 500\n' +
          '    at browseFolders (http://localhost:5173/src/services/api/files.ts:62:13)\n' +
          '    at async fetchCurrentDirectory (http://localhost:5173/src/composables/useFileBrowser.ts:140:7)\n' +
          '    at async watch.immediate (http://localhost:5173/src/composables/useFileBrowser.ts:228:7)'
}

// Test isAbortError function directly
import { isAbortError } from '@/utils/abortUtils'

export function verifyAbortErrorHandling() {
  console.log('🔬 Verifying abort error handling...')
  
  // Test 1: Error without abort signal should NOT be ignored
  const result1 = isAbortError(simulatedAbort500Error, false)
  console.log('❓ 500 error without abort signal:', result1 ? 'IGNORED (bad)' : 'PROPERLY SHOWN (good)')
  
  // Test 2: Error with abort signal should be ignored
  const result2 = isAbortError(simulatedAbort500Error, true)
  console.log('❓ 500 error with abort signal:', result2 ? 'IGNORED (good)' : 'PROPERLY SHOWN (bad)')
  
  // Test 3: Real abort error should be ignored
  const realAbortError = {
    name: 'AbortError',
    message: 'Request aborted'
  }
  const result3 = isAbortError(realAbortError)
  console.log('❓ Real abort error:', result3 ? 'IGNORED (good)' : 'PROPERLY SHOWN (bad)')
  
  // Test 4: Net abort error should be ignored
  const netAbortError = {
    message: 'net::ERR_ABORTED 500'
  }
  const result4 = isAbortError(netAbortError)
  console.log('❓ Net abort error:', result4 ? 'IGNORED (good)' : 'PROPERLY SHOWN (bad)')
  
  console.log('✅ Verification complete - check results above')
}

// Test request tracking
export function verifyRequestTracking() {
  console.log('📊 Verifying request tracking...')
  
  const abortController = new AbortController()
  
  // Test adding and removing requests
  requestStore.addRequest('test', abortController)
  console.log('📝 Added request:', requestStore.getActiveRequestCount() === 1 ? '✅' : '❌')
  
  requestStore.cancelRequest('test')
  console.log('📝 Cancelled request:', requestStore.getActiveRequestCount() === 0 ? '✅' : '❌')
  
  // Test global cancellation
  const abortController2 = new AbortController()
  const abortController3 = new AbortController()
  
  requestStore.addRequest('test2', abortController2)
  requestStore.addRequest('test3', abortController3)
  console.log('📝 Added 2 requests:', requestStore.getActiveRequestCount() === 2 ? '✅' : '❌')
  
  requestStore.cancelAllRequests()
  console.log('📝 Cancelled all requests:', requestStore.getActiveRequestCount() === 0 ? '✅' : '❌')
}

// Auto-run verification
if (typeof window !== 'undefined') {
  window.verifyAbortErrorHandling = verifyAbortErrorHandling
  window.verifyRequestTracking = verifyRequestTracking
  
  console.log('🔧 Verification functions available:')
  console.log('  - window.verifyAbortErrorHandling()')
  console.log('  - window.verifyRequestTracking()')
  
  // Auto-run
  setTimeout(() => {
    verifyAbortErrorHandling()
    verifyRequestTracking()
    console.log('🎯 Verification complete - check console results!')
  }, 100)
}