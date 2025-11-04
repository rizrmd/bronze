import { ref, computed, onUnmounted, nextTick, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { FileInfo, DirectoryInfo, FolderRequest } from '@/types'
import { browseFolders } from '@/services/api/files'
import { requestStore } from '@/stores/requestStore'
import { isAbortError } from '@/utils/abortUtils'

export interface FileBrowserState {
  files: FileInfo[]
  folders: DirectoryInfo[]
  currentPath: string
  parentPath: string
  loading: boolean
  error: string

  searchQuery: string
  viewMode: 'list' | 'grid'
}

export interface FileBrowserOptions {
  initialPath?: string
  initialViewMode?: 'list' | 'grid'
}

export function useFileBrowser(options: FileBrowserOptions = {}) {
  const router = useRouter()
  const route = useRoute()
  
  // State - no currentPath, derived from URL
  const files = ref<FileInfo[]>([])
  const folders = ref<DirectoryInfo[]>([])
  const loading = ref(false)
  const error = ref('')
  
  // Track seen items to avoid duplicates during streaming
  const seenFiles = ref(new Set<string>())
  const seenFolders = ref(new Set<string>())

  const searchQuery = ref('')
  const viewMode = ref<'list' | 'grid'>(options.initialViewMode || 'list')
  const navigationHistory = ref<string[]>([])
  
  // Cancel all active browse requests
  const cancelAllBrowseRequests = () => {
    requestStore.cancelAllRequests()
  }

  // Cancel specific browse request
  const cancelBrowseRequest = (path: string) => {
    requestStore.cancelRequest(path)
  }
  
  // Single source of truth: derive path from URL
  const currentPath = computed(() => route.params.path as string || '')
  const parentPath = computed(() => {
    const path = currentPath.value
    if (!path) return ''
    const parts = path.split('/')
    parts.pop()
    return parts.join('/')
  })

  // Computed properties
  const breadcrumbPaths = computed(() => {
    const path = currentPath.value
    if (!path) return [{ name: 'Root', path: '' }]
    
    const parts = path.split('/').filter(Boolean)
    const paths = [{ name: 'Root', path: '' }]
    let currentPathSegment = ''
    
    for (const part of parts) {
      currentPathSegment += (currentPathSegment ? '/' : '') + part
      paths.push({ name: part, path: currentPathSegment })
    }
    
    return paths
  })

  const filteredFiles = computed(() => {
    if (!searchQuery.value || !files.value) return files.value || []
    return files.value.filter(file => 
      file.key && file.key.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  })

  const filteredFolders = computed(() => {
    if (!searchQuery.value || !folders.value) return folders.value || []
    return folders.value.filter(folder => 
      folder.name && folder.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  })

  const hasFiles = computed(() => {
    return (filteredFiles.value?.length || 0) > 0 || (filteredFolders.value?.length || 0) > 0
  })



  // Navigation methods - just router operations
  const navigateToPath = (path: string) => {
    if (path) {
      router.push({ 
        name: 'FilesWithFolder', 
        params: { path } 
      }).catch(() => {}) // Ignore duplicate navigation
    } else {
      router.push({ 
        name: 'Files' 
      }).catch(() => {})
    }
  }

  const navigateToFolder = (folder: DirectoryInfo) => {
    navigateToPath(folder.path)
  }

  // Data fetching
  const fetchCurrentDirectory = async (path: string = '') => {
    loading.value = true
    error.value = ''
    
    // Create a unique key for this request (use empty string for root)
    const requestKey = path || 'root'
    
    // Cancel any existing request for this path
    cancelBrowseRequest(requestKey)
    
    // Create new abort controller for this request - ensure it's fresh
    let abortController = new AbortController()
    
    // Double-check the new controller isn't already aborted (edge case)
    if (abortController.signal.aborted) {
      console.log('⚠️ New AbortController was already aborted, creating fresh one')
      abortController = new AbortController()
    }
    
    requestStore.addRequest(requestKey, abortController, `/files/${path || ''}`)
    
    const folderRequest: FolderRequest = {
      path,
      include_files: true,
      include_dirs: true,
      recursive: true,
      max_depth: 1,
      include_metadata: true
    }

    try {
      // Use SSE streaming for all requests
      await browseFolders(
        [folderRequest],
        (event, data) => {
          // If request was aborted, exit early and reset loading
          if (abortController?.signal.aborted) {
            console.log('📡 Request aborted during SSE processing, resetting loading state')
            loading.value = false
            return
          }
          
          switch (event) {
            case 'folder_start':
              // Clear current data and tracking sets when starting new folder browse
              files.value = []
              folders.value = []
              seenFiles.value.clear()
              seenFolders.value.clear()
              loading.value = true
              
              // Pre-populate with directory listing if available
              if (data.items && Array.isArray(data.items)) {
                data.items.forEach((item: any) => {
                  if (item.type === 'dir') {
                    folders.value.push({
                      name: item.name,
                      path: data.path,
                      last_modified: new Date().toISOString(),
                      file_count: 0,
                      dir_count: 0
                    })
                    seenFolders.value.add(item.name)
                  } else {
                    files.value.push({
                      key: item.name,
                      size: 0,
                      last_modified: new Date().toISOString(),
                      etag: '',
                      content_type: ''
                    })
                    seenFiles.value.add(item.name)
                  }
                })
                
                // Stop loading since we have items to display
                loading.value = false
              }
              
              // Trigger immediate UI update
              nextTick()
              break
              
            case 'item':
              // Stop loading after first item arrives for instant UI
              if (loading.value) {
                loading.value = false
              }
              
              if (data.type === 'file') {
                const fileName = data.name || data.path
                
                if (!seenFiles.value.has(fileName)) {
                  // New file - add it with full metadata
                  seenFiles.value.add(fileName)
                  files.value.push({
                    key: data.path,
                    size: data.size,
                    last_modified: data.last_modified,
                    etag: data.etag,
                    content_type: data.contentType
                  })
                } else {
                  // Existing file - update its metadata
                  const fileIndex = files.value.findIndex(file => 
                    file.key === data.path || file.key === fileName
                  )
                  if (fileIndex !== -1) {
                    files.value[fileIndex] = {
                      ...files.value[fileIndex],
                      key: data.path,
                      size: data.size,
                      last_modified: data.last_modified,
                      etag: data.etag,
                      content_type: data.contentType
                    }
                  }
                }
                
                // Trigger immediate UI update
                nextTick()
              } else if (data.type === 'directory') {
                const folderName = data.name || data.path
                
                if (!seenFolders.value.has(folderName)) {
                  // New folder - add it with full metadata
                  seenFolders.value.add(folderName)
                  folders.value.push({
                    ...data,
                    name: data.name,
                    path: data.path,
                    file_count: 0,
                    dir_count: 0,
                    total_count: data.size || 0, // For folders, size is item count
                    size: data.size || 0 // Store original size (item count for folders)
                  })
                } else {
                  // Existing folder - update its metadata
                  const folderIndex = folders.value.findIndex(folder => 
                    folder.path === data.path || folder.name === folderName
                  )
                  if (folderIndex !== -1) {
                    folders.value[folderIndex] = {
                      ...folders.value[folderIndex],
                      ...data,
                      name: data.name,
                      path: data.path,
                      file_count: 0,
                      dir_count: 0,
                      total_count: data.size || 0,
                      size: data.size || 0
                    }
                  }
                }
                
                // Trigger immediate UI update
                nextTick()
              }
              break
              
            case 'folder_complete':
            case 'complete':
              loading.value = false
              break
              
            case 'error':
              error.value = data.error || 'Unknown error occurred'
              loading.value = false
              break
          }
        },
        (err: any) => {
          // Don't show error for aborted requests - they're intentional cancellations
          if (!isAbortError(err, abortController?.signal.aborted)) {
            error.value = err.message || 'Unknown error'
            loading.value = false
          }
        },
        abortController
      )
    } catch (err: any) {
      // Don't show error for aborted requests - they're intentional cancellations
      if (isAbortError(err, abortController?.signal.aborted)) {
        console.log('✅ Request aborted, resetting loading state')
        loading.value = false  // Always reset loading on abort!
      } else {
        error.value = err.message || 'Failed to fetch directory'
        loading.value = false
      }
    } finally {
      // Clean up controller when done
      requestStore.removeRequest(requestKey)
    }
  }

  // Watch route changes to fetch data
  const stopWatch = watch(currentPath, (newPath, oldPath) => {
    // Cancel the previous path's request before fetching new path, unless it's back navigation
    if (oldPath && oldPath !== newPath) {
      // Check if this is back navigation to parent
      const isBackToParent = oldPath && newPath && 
                           oldPath.startsWith(newPath) &&
                           (newPath === '' || oldPath.includes(newPath + '/'))
      
      const oldRequestKey = oldPath || 'root'
      
      if (isBackToParent) {
        console.log(`Path changed from ${oldPath} to ${newPath} (back to parent), NOT canceling previous request`)
      } else {
        console.log(`Path changed from ${oldPath} to ${newPath} (forward navigation), canceling previous request: ${oldRequestKey}`)
        cancelBrowseRequest(oldRequestKey)
      }
    }
    fetchCurrentDirectory(newPath)
  }, { immediate: true })

  onUnmounted(() => {
    stopWatch()
    // Cancel all pending requests when component is unmounted
    cancelAllBrowseRequests()
  })

  return {
    // State
    files,
    folders,
    currentPath,
    parentPath,
    loading,
    error,

    searchQuery,
    viewMode,
    navigationHistory,

    // Computed
    breadcrumbPaths,
    filteredFiles,
    filteredFolders,
    hasFiles,

    // Methods
    navigateToPath,
    navigateToFolder,
    fetchCurrentDirectory,
    
    refresh: () => fetchCurrentDirectory(currentPath.value),
    setViewMode: (mode: 'list' | 'grid') => { viewMode.value = mode },
    setSearchQuery: (query: string) => { searchQuery.value = query },
    cancelAllBrowseRequests,
    cancelBrowseRequest
  }
}
