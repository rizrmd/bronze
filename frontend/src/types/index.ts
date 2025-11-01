export interface Job {
  id: string
  type: string
  priority: 'low' | 'medium' | 'high'
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
  file_path: string
  bucket: string
  object_name: string
  created_at: string
  started_at?: string
  completed_at?: string
  error?: string
  result?: any
  progress: number
  metadata: Record<string, any>
}

export interface FileInfo {
  key: string
  size: number
  last_modified: string
  etag: string
  content_type?: string
}

export interface QueueStats {
  total: number
  pending: number
  processing: number
  completed: number
  failed: number
  cancelled: number
}

export interface WorkerPoolStats {
  total_workers: number
  active_jobs: number
  is_running: boolean
}

export interface JobStats {
  queue: QueueStats
  workers: WorkerPoolStats
}

export interface FileEvent {
  id: string
  event_type: 's3:ObjectCreated:*' | 's3:ObjectRemoved:*' | 's3:ObjectMetadata:*'
  key: string
  bucket: string
  size?: number
  etag?: string
  event_time: string
  metadata?: Record<string, string>
  processed: boolean
  processed_at?: string
}

export interface ApiResponse<T = any> {
  success: boolean
  message: string
  data?: T
  error?: string
}

// Specific API response types matching backend
export interface JobListResponse {
  success: boolean
  message: string
  jobs: Job[]
  count: number
}

export interface JobResponse {
  success: boolean
  message: string
  job?: Job
}

export interface FileListResponse {
  success: boolean
  message: string
  files: FileInfo[]
  count: number
}

export interface FileEventsResponse {
  success: boolean
  message: string
  events: FileEvent[]
  count: number
}

export interface UploadResponse {
  success: boolean
  message: string
  object_name: string
  size: number
  etag?: string
}

export interface CreateJobRequest {
  type: string
  file_path?: string
  bucket: string
  object_name: string
  priority: 'low' | 'medium' | 'high'
}

// Multi-folder browsing types
export interface FolderRequest {
  path: string
  include_files: boolean
  include_dirs: boolean
  recursive: boolean
  max_depth?: number
}

export interface DirectoryInfo {
  name: string
  path: string
  last_modified: string
  file_count?: number
  size?: number
}

export interface FolderResult {
  path: string
  directories?: DirectoryInfo[]
  files?: FileInfo[]
  total_count: number
  file_count: number
  dir_count: number
  total_size_bytes: number
  last_modified: string
  subfolders?: Record<string, FolderResult>
}

export interface MultiFolderResponse {
  success: boolean
  message: string
  folders: Record<string, FolderResult>
}