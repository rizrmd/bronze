export const API_ENDPOINTS = {
  HEALTH: '/api/health',
  FILES: '/api/files',
  JOBS: '/api/jobs',
  DATA_BROWSER: '/api/data',
  CONFIG: '/api/config',
  WATCHER: '/api/watcher',
  NESSIE: '/api/nessie'
} as const

export const JOB_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  FAILED: 'failed',
  CANCELLED: 'cancelled'
} as const

export const JOB_PRIORITY = {
  LOW: 'low',
  MEDIUM: 'medium',
  HIGH: 'high'
} as const

export const FILE_TYPES = {
  EXCEL: ['.xlsx', '.xls', '.xlsm'],
  CSV: ['.csv'],
  MDB: ['.mdb'],
  ARCHIVE: ['.zip', '.tar', '.tar.gz', '.tgz']
} as const

export const MAX_FILE_SIZE = 100 * 1024 * 1024 * 1024 // 100GB (following decompression no-limits rule)

export const DEFAULT_PAGE_SIZE = 50
export const MAX_PAGE_SIZE = 1000