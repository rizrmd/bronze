import { ref, computed } from 'vue'
import { getConfig as getConfigService, updateConfig as updateConfigService } from '@/services'

export interface ConfigData {
  SERVER_HOST: string
  SERVER_PORT: string
  MINIO_ENDPOINT: string
  MINIO_ACCESS_KEY: string
  MINIO_SECRET_KEY: string
  MINIO_BUCKET: string
  MINIO_REGION: string
  MAX_WORKERS: string
  QUEUE_SIZE: string
  WATCH_INTERVAL: string
  TEMP_DIR: string
  DECOMPRESSION_ENABLED: string
  MAX_EXTRACT_SIZE: string
  MAX_FILES_PER_ARCHIVE: string
  NESTED_ARCHIVE_DEPTH: string
  PASSWORD_PROTECTED: string
  EXTRACT_TO_SUBFOLDER: string
}

export function useConfig() {
  const config = ref<ConfigData>({
    SERVER_HOST: 'localhost',
    SERVER_PORT: '8060',
      MINIO_ENDPOINT: 'localhost:9000',
    MINIO_ACCESS_KEY: 'minioadmin',
    MINIO_SECRET_KEY: 'minioadmin',
    MINIO_BUCKET: 'files',
    MINIO_REGION: 'us-east-1',
    MAX_WORKERS: '3',
    QUEUE_SIZE: '100',
    WATCH_INTERVAL: '5s',
    TEMP_DIR: '/tmp/bronze',
    DECOMPRESSION_ENABLED: 'true',
    MAX_EXTRACT_SIZE: '1GB',
    MAX_FILES_PER_ARCHIVE: '1000',
    NESTED_ARCHIVE_DEPTH: '3',
    PASSWORD_PROTECTED: 'true',
    EXTRACT_TO_SUBFOLDER: 'true'
  })

  const loading = ref(false)
  const error = ref<string | null>(null)
  let fetched = false

  const serverConfig = computed(() => ({
    host: config.value.SERVER_HOST,
    port: config.value.SERVER_PORT
  }))

  const minioConfig = computed(() => ({
    endpoint: config.value.MINIO_ENDPOINT,
    accessKey: config.value.MINIO_ACCESS_KEY,
    secretKey: config.value.MINIO_SECRET_KEY,
    useSSL: config.value.MINIO_ENDPOINT.startsWith('https://'),
    bucket: config.value.MINIO_BUCKET,
    region: config.value.MINIO_REGION
  }))

  const processingConfig = computed(() => ({
    maxWorkers: config.value.MAX_WORKERS,
    queueSize: config.value.QUEUE_SIZE,
    watchInterval: config.value.WATCH_INTERVAL,
    tempDir: config.value.TEMP_DIR,
    decompressionEnabled: config.value.DECOMPRESSION_ENABLED === 'true',
    maxExtractSize: config.value.MAX_EXTRACT_SIZE,
    maxFilesPerArchive: config.value.MAX_FILES_PER_ARCHIVE,
    nestedArchiveDepth: config.value.NESTED_ARCHIVE_DEPTH,
    passwordProtected: config.value.PASSWORD_PROTECTED === 'true',
    extractToSubfolder: config.value.EXTRACT_TO_SUBFOLDER === 'true'
  }))

  const fetchConfig = async () => {
    loading.value = true
    error.value = null
    
    try {
      const response = await getConfigService()
      if (response.success && response.data) {
        config.value = { ...config.value, ...response.data }
        fetched = true
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch configuration'
    } finally {
      loading.value = false
    }
  }

  // Auto-fetch config on first use
  if (!fetched && !loading.value) {
    fetchConfig()
  }

  const updateConfig = async (updates: Partial<ConfigData>) => {
    loading.value = true
    error.value = null
    
    try {
      const response = await updateConfigService(updates as Record<string, string>)
      if (response.success) {
        // Update local config with the changes
        config.value = { ...config.value, ...updates }
        return true
      }
      return false
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update configuration'
      return false
    } finally {
      loading.value = false
    }
  }

  const resetConfig = () => {
    config.value = {
      SERVER_HOST: 'localhost',
      SERVER_PORT: '8060',
    MINIO_ENDPOINT: 'localhost:9000',
      MINIO_ACCESS_KEY: 'minioadmin',
      MINIO_SECRET_KEY: 'minioadmin',
      MINIO_BUCKET: 'files',
      MINIO_REGION: 'us-east-1',
      MAX_WORKERS: '3',
      QUEUE_SIZE: '100',
      WATCH_INTERVAL: '5s',
      TEMP_DIR: '/tmp/bronze',
      DECOMPRESSION_ENABLED: 'true',
      MAX_EXTRACT_SIZE: '1GB',
      MAX_FILES_PER_ARCHIVE: '1000',
      NESTED_ARCHIVE_DEPTH: '3',
      PASSWORD_PROTECTED: 'true',
      EXTRACT_TO_SUBFOLDER: 'true'
    }
  }

  return {
    config,
    loading,
    error,
    serverConfig,
    minioConfig,
    processingConfig,
    fetchConfig,
    updateConfig,
    resetConfig
  }
}