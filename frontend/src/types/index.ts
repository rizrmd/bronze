export interface Job {
  id: string
  job_type?: string
  type: string
  priority: 'low' | 'medium' | 'high'
  status: 'pending' | 'processing' | 'running' | 'completed' | 'failed' | 'cancelled'
  file_path: string
  bucket: string
  object_name: string
  created_at: string
  started_at?: string
  completed_at?: string
  error?: string
  error_message?: string
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
  include_metadata?: boolean
}

export interface DirectoryInfo {
  name: string
  path: string
  last_modified: string
  file_count?: number
  dir_count?: number
  total_count?: number | string
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

// Data Browser types
export interface BrowseRequest {
  file_name: string
  sheet_name?: string
  max_rows?: number
  offset?: number
  has_headers?: boolean
  treat_as_csv?: boolean
  auto_detect_headers?: boolean
  stream_mode?: boolean
  chunk_size?: number
}

export interface BrowseResponse {
  success: boolean
  message: string
  data_type: string
  file_name: string
  sheet_name?: string
  columns: string[]
  rows: string[][]
  total_rows: number
  row_count: number
  offset: number
  has_headers: boolean
  sheets?: string[]
}

export interface DataFileInfo {
  name: string
  size: number
  last_modified: string
  data_type: string
  sheets?: string[]
  columns?: string[]
  row_count?: number
}

export interface DataFileListResponse {
  success: boolean
  message: string
  files: DataFileInfo[]
  count: number
}

export interface ExportRequest {
  files: FileExportInfo[]
  table_name: string
  operation: 'create' | 'append'
  database?: string
  max_errors?: number
  stop_on_error?: boolean
  collect_errors?: boolean
  schema_resolution?: string
  max_concurrent_files?: number
  batch_size?: number
  auto_type_conversion?: boolean
}

export interface FileExportInfo {
  file_name: string
  sheet_name?: string
  treat_as_csv?: boolean
}

export interface ExportResponse {
  success: boolean
  message: string
  job_id?: string
  job_type?: string
  status?: string
  table_name: string
  files_processed: number
  rows_exported: number
  rows_failed: number
  processing_time: number
  column_mismatches?: any[]
  row_errors?: any[]
  error_summary?: Record<string, number>
  database?: string
}