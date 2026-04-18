export const JAV_PROVIDER_UNKNOWN = 0
export const JAV_PROVIDER_JAVBUS = 1
export const JAV_PROVIDER_JAVDATABASE = 2
export const JAV_PROVIDER_USER = 3

export function normalizeJavSort(sort, fallback = 'recent') {
  if (sort === 'code') return 'code'
  if (sort === 'release') return 'release'
  if (sort === 'play_count') return 'play_count'
  if (sort === 'recent') return 'recent'
  return fallback
}

export function isUserJavTag(tag) {
  return Number(tag?.provider) === JAV_PROVIDER_USER
}
