// Test script to verify abort error handling works correctly
// Run this in browser console to test different abort scenarios

import { requestStore } from '@/stores/requestStore'
import { browseFolders } from '@/services/api/files'

// Test 1: Direct abort during fetch
async function testDirectAbort() {
  console.log('=== Test 1: Direct abort during fetch ===')
  const abortController = new AbortController()
  
  // Start request
  const promise = browseFolders([{path: '', include_files: true}], abortController)
  
  // Abort immediately (before response)
  setTimeout(() => {
    console.log('Aborting request...')
    abortController.abort()
  }, 1)
  
  try {
    await promise
    console.log('✅ Direct abort handled correctly')
  } catch (error) {
    console.log('❌ Direct abort failed:', error.message)
  }
}

// Test 2: Abort during SSE streaming
async function testSSEAbort() {
  console.log('=== Test 2: Abort during SSE streaming ===')
  const abortController = new AbortController()
  
  const promise = browseFolders([{path: '', include_files: true, recursive: true}], abortController)
  
  // Abort after some events should have been received
  setTimeout(() => {
    console.log('Aborting SSE stream...')
    abortController.abort()
  }, 100)
  
  try {
    await promise
    console.log('✅ SSE abort handled correctly')
  } catch (error) {
    console.log('❌ SSE abort failed:', error.message)
  }
}

// Test 3: Global request cancellation
async function testGlobalCancel() {
  console.log('=== Test 3: Global request cancellation ===')
  const abortController = new AbortController()
  
  // Add to global store
  requestStore.addRequest('test', abortController)
  
  const promise = browseFolders([{path: '', include_files: true}], abortController)
  
  // Cancel via global store
  setTimeout(() => {
    console.log('Canceling via global store...')
    requestStore.cancelAllRequests()
  }, 50)
  
  try {
    await promise
    console.log('✅ Global cancel handled correctly')
  } catch (error) {
    console.log('❌ Global cancel failed:', error.message)
  }
}

// Test 4: Multiple concurrent requests with abort
async function testConcurrentAbort() {
  console.log('=== Test 4: Multiple concurrent requests with abort ===')
  
  const abortController1 = new AbortController()
  const abortController2 = new AbortController()
  
  const promise1 = browseFolders([{path: 'ENERGY/', include_files: true}], abortController1)
  const promise2 = browseFolders([{path: 'MINING/', include_files: true}], abortController2)
  
  // Abort second request
  setTimeout(() => {
    console.log('Aborting second request only...')
    abortController2.abort()
  }, 50)
  
  try {
    await Promise.all([promise1, promise2])
    console.log('✅ Concurrent abort handled correctly')
  } catch (error) {
    console.log('❌ Concurrent abort failed:', error.message)
  }
}

// Run all tests
export async function runAbortTests() {
  console.log('🧪 Starting comprehensive abort tests...')
  
  await testDirectAbort()
  await new Promise(resolve => setTimeout(resolve, 500))
  
  await testSSEAbort()
  await new Promise(resolve => setTimeout(resolve, 500))
  
  await testGlobalCancel()
  await new Promise(resolve => setTimeout(resolve, 500))
  
  await testConcurrentAbort()
  await new Promise(resolve => setTimeout(resolve, 500))
  
  console.log('🎉 All abort tests completed!')
  console.log('📝 Check console for any error messages - there should be none!')
}

// Auto-run tests
if (typeof window !== 'undefined') {
  // Add to window for manual testing
  window.runAbortTests = runAbortTests
  console.log('🔧 abortTests available - call window.runAbortTests() to run tests')
}