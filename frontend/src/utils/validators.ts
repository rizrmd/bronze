export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

export function isValidUrl(url: string): boolean {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

export function isValidObjectname(name: string): boolean {
  // Basic validation for object names (no slashes, no empty strings)
  if (!name || name.trim() === '') return false
  if (name.includes('/') || name.includes('\\')) return false
  if (name.includes('..')) return false
  return true
}

export function sanitizeFileName(name: string): string {
  // Remove or replace problematic characters
  return name
    .replace(/[<>:"/\\|?*]/g, '_')
    .replace(/\s+/g, '_')
    .toLowerCase()
}