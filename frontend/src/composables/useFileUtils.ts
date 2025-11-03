export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

export function formatDate(dateString: string): string {
  if (!dateString) return 'Unknown'
  
  const date = new Date(dateString)
  return date.toLocaleString()
}

import { 
  File, 
  FileText, 
  FileSpreadsheet, 
  FileImage, 
  FileVideo, 
  FileAudio, 
  FileArchive,
  Folder
} from 'lucide-vue-next'

export function getFileIcon(fileName: string): any {
  if (!fileName || typeof fileName !== 'string') return File
  const ext = fileName.split('.').pop()?.toLowerCase()
  const icons: Record<string, any> = {
    'pdf': FileText,
    'doc': FileText, 'docx': FileText,
    'xls': FileSpreadsheet, 'xlsx': FileSpreadsheet, 'csv': FileSpreadsheet,
    'jpg': FileImage, 'jpeg': FileImage, 'png': FileImage, 'gif': FileImage,
    'mp4': FileVideo, 'avi': FileVideo, 'mov': FileVideo,
    'mp3': FileAudio, 'wav': FileAudio,
    'zip': FileArchive, 'rar': FileArchive, 'tar': FileArchive, 'gz': FileArchive,
    'txt': File, 'log': File
  }
  return icons[ext || ''] || File
}

export function getFolderIcon(): any {
  return Folder
}

export function getRelativePath(basePath: string, fullPath: string): string {
  if (!basePath || !fullPath) return fullPath
  
  // Remove trailing slashes
  const cleanBase = basePath.replace(/\/+$/, '')
  const cleanFull = fullPath.replace(/\/+$/, '')
  
  if (cleanFull.startsWith(cleanBase + '/')) {
    return cleanFull.substring(cleanBase.length + 1)
  }
  
  return cleanFull
}

export function validateFileName(fileName: string): { isValid: boolean; error?: string } {
  if (!fileName || fileName.trim() === '') {
    return { isValid: false, error: 'File name cannot be empty' }
  }
  
  // Check for invalid characters
  const invalidChars = /[<>:"/\\|?*]/
  if (invalidChars.test(fileName)) {
    return { isValid: false, error: 'File name contains invalid characters' }
  }
  
  // Check for reserved names (Windows)
  const reservedNames = ['CON', 'PRN', 'AUX', 'NUL', 'COM1', 'COM2', 'COM3', 'COM4', 'COM5', 'COM6', 'COM7', 'COM8', 'COM9', 'LPT1', 'LPT2', 'LPT3', 'LPT4', 'LPT5', 'LPT6', 'LPT7', 'LPT8', 'LPT9']
  const nameParts = fileName.split('.')
  const nameWithoutExt = nameParts[0]?.toUpperCase() || fileName.toUpperCase()
  if (reservedNames.includes(nameWithoutExt)) {
    return { isValid: false, error: 'File name is reserved' }
  }
  
  return { isValid: true }
}

export function sortFiles<T extends { key?: string; name?: string; last_modified?: string; size?: number }>(
  files: T[],
  sortBy: 'name' | 'date' | 'size',
  direction: 'asc' | 'desc' = 'asc'
): T[] {
  return [...files].sort((a, b) => {
    let comparison = 0
    
    switch (sortBy) {
      case 'name':
        const nameA = a.key || a.name || ''
        const nameB = b.key || b.name || ''
        comparison = nameA.localeCompare(nameB)
        break
        
      case 'date':
        const dateA = a.last_modified || ''
        const dateB = b.last_modified || ''
        comparison = new Date(dateA).getTime() - new Date(dateB).getTime()
        break
        
      case 'size':
        const sizeA = a.size || 0
        const sizeB = b.size || 0
        comparison = sizeA - sizeB
        break
    }
    
    return direction === 'asc' ? comparison : -comparison
  })
}

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}