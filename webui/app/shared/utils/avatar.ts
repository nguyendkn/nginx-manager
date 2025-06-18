/**
 * Utility functions for avatar handling
 */

import { STORAGE_PATHS } from '~/shared/constants/api-paths'

/**
 * Generate avatar URL from file ID
 * @param fileId File ID from the database
 * @returns Avatar URL for streaming
 */
export const getAvatarUrl = (fileId?: string): string | null => {
  if (!fileId) return null
  return STORAGE_PATHS.FILE_PREVIEW(fileId)
}

/**
 * Generate avatar URL with fallback
 * @param user User object with avatar_id or face_id
 * @returns Avatar URL or null
 */
export const getUserAvatarUrl = (user?: { avatar_id?: string; face_id?: string }): string | null => {
  if (!user) return null

  // Prefer avatar_id (new file-based system)
  if (user.avatar_id) {
    return getAvatarUrl(user.avatar_id)
  }

  // Fallback to face_id (legacy URL-based system)
  if (user.face_id) {
    // If face_id is already a URL, return it
    if (user.face_id.startsWith('http') || user.face_id.startsWith('/')) {
      return user.face_id
    }
    // If face_id is a file ID, convert to stream URL
    return getAvatarUrl(user.face_id)
  }

  return null
}

/**
 * Get initials from user name for avatar fallback
 * @param name User name
 * @returns Initials (max 2 characters)
 */
export const getInitials = (name?: string): string => {
  if (!name) return '??'

  const words = name.trim().split(' ').filter(word => word.length > 0)
  if (words.length === 0) return '??'
  if (words.length === 1) {
    return (words[0]?.charAt(0) || '?').toUpperCase()
  }

  const firstChar = words[0]?.charAt(0) || '?'
  const lastChar = words[words.length - 1]?.charAt(0) || '?'
  return (firstChar + lastChar).toUpperCase()
}

/**
 * Generate a placeholder avatar URL with initials
 * @param name User name
 * @param size Avatar size (default: 40)
 * @param bgColor Background color (default: random)
 * @returns Data URL for placeholder avatar
 */
export const getPlaceholderAvatar = (
  name?: string,
  size: number = 40,
  bgColor?: string
): string => {
  const initials = getInitials(name)

  // Generate a consistent color based on name
  const colors = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7',
    '#DDA0DD', '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E9'
  ]

  const colorIndex = name ? name.charCodeAt(0) % colors.length : 0
  const backgroundColor = bgColor || colors[colorIndex]

  // Create SVG
  const svg = `
    <svg width="${size}" height="${size}" xmlns="http://www.w3.org/2000/svg">
      <rect width="${size}" height="${size}" fill="${backgroundColor}" rx="${size / 2}"/>
      <text x="50%" y="50%" text-anchor="middle" dy="0.35em"
            font-family="Arial, sans-serif" font-size="${size * 0.4}"
            font-weight="bold" fill="white">
        ${initials}
      </text>
    </svg>
  `

  return `data:image/svg+xml;base64,${btoa(svg)}`
}