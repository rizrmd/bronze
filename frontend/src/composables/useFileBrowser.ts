import { ref, computed, onMounted } from 'vue'
import type { FileInfo, DirectoryInfo, FolderRequest } from '@/types'
import { browseFolders } from '@/services/api/files'

export interface FileBrowserState {
  files: FileInfo[]
  folders: DirectoryInfo[]
  currentPath: string
  parentPath: string
  loading: boolean
  error: string
  selectedFiles: Set<string>
  searchQuery: string
  viewMode: 'list' | 'grid'
  navigationHistory: string[]
  historyIndex: number
}

export interface FileBrowserOptions {
  initialPath?: string
  initialViewMode?: 'list' | 'grid'
  useSSE?: boolean
}

export function useFileBrowser(options: FileBrowserOptions = {}) {
  // State
  const files = ref<FileInfo[]>([])
  const folders = ref<DirectoryInfo[]>([])
  const currentPath = ref(options.initialPath || '')
  const parentPath = ref('')
  const loading = ref(false)
  const error = ref('')
  const selectedFiles = ref<Set<string>>(new Set())
  const searchQuery = ref('')
  const viewMode = ref<'list' | 'grid'>(options.initialViewMode || 'list')
  const navigationHistory = ref<string[]>([])
  const historyIndex = ref(-1)

  // Computed properties
  const breadcrumbPaths = computed(() => {
    if (!currentPath.value) return [{ name: 'Root', path: '' }]
    
    const parts = currentPath.value.split('/').filter(Boolean)
    const paths = []
    let current = ''
    
    paths.push({ name: 'Root', path: '' })
    
    for (const part of parts) {
      current += (current ? '/' : '') + part
      paths.push({ name: part, path: current })
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

  const hasSelection = computed(() => selectedFiles.value.size > 0)

  const clearSelection = () => {
    selectedFiles.value.clear()
  }

  const selectAll = () => {
    clearSelection()
    filteredFiles.value.forEach(file => {
      selectedFiles.value.add(file.key)
    })
  }

  // Methods
  const fetchCurrentDirectory = async (path: string = '') => {
    loading.value = true
    error.value = ''
    
    const folderRequest: FolderRequest = {
      path,
      include_files: true,
      include_dirs: true,
      recursive: true,
      max_depth: 1,
      include_metadata: true
    }

    try {
      const response = await browseFolders([folderRequest])
      
      if (response.success && response.folders[path]) {
        const folderData = response.folders[path]
        
        // Map backend response to frontend format
        files.value = (folderData.files || []).map((file: any) => ({
          key: file.path || file.name,
          size: file.size,
          last_modified: file.last_modified,
          etag: file.etag,
          content_type: file.content_type
        }))
        
        // Process directories
        const directories = folderData.directories || []
        folders.value = directories.map((dir: any) => ({
          ...dir,
          name: dir.name || dir.path?.split('/').filter(Boolean).pop() || 'Unknown',
          file_count: dir.file_count || 0,
          dir_count: dir.dir_count || 0,
          total_count: (dir.file_count || 0) + (dir.dir_count || 0)
        }))
        
        currentPath.value = path
        
        // Calculate parent path
        if (path) {
          const parts = path.split('/')
          parts.pop()
          parentPath.value = parts.join('/')
        } else {
          parentPath.value = ''
        }
        
        // Update navigation history
        if (historyIndex.value === -1 || navigationHistory.value[historyIndex.value] !== path) {
          navigationHistory.value = navigationHistory.value.slice(0, historyIndex.value + 1)
          navigationHistory.value.push(path)
          historyIndex.value = navigationHistory.value.length - 1
        }
      } else {
        throw new Error(response.message || 'Failed to load directory')
      }
    } catch (err: any) {
      const errorMsg = err.message || 'Failed to load directory'
      error.value = errorMsg
      files.value = []
      folders.value = []
    } finally {
      loading.value = false
    }
  }

  const navigateToPath = (path: string) => {
    fetchCurrentDirectory(path)
  }

  const navigateToFolder = (folder: DirectoryInfo) => {
    navigateToPath(folder.path)
  }

  const refresh = () => {
    selectedFiles.value.clear()
    fetchCurrentDirectory(currentPath.value)
  }

  const toggleFileSelection = (fileKey: string) => {
    if (selectedFiles.value.has(fileKey)) {
      selectedFiles.value.delete(fileKey)
    } else {
      selectedFiles.value.add(fileKey)
    }
  }

  const setViewMode = (mode: 'list' | 'grid') => {
    viewMode.value = mode
  }

  const setSearchQuery = (query: string) => {
    searchQuery.value = query
  }

  // Initialize
  onMounted(() => {
    fetchCurrentDirectory(currentPath.value)
  })

  return {
    // State
    files,
    folders,
    currentPath,
    parentPath,
    loading,
    error,
    selectedFiles,
    searchQuery,
    viewMode,
    navigationHistory,
    historyIndex,
    
    // Computed
    breadcrumbPaths,
    filteredFiles,
    filteredFolders,
    hasFiles,
    hasSelection,
    
    // Methods
    fetchCurrentDirectory,
    navigateToPath,
    navigateToFolder,
    refresh,
    toggleFileSelection,
    setViewMode,
    setSearchQuery,
    clearSelection,
    selectAll
  }
}