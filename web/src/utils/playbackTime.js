/** Format seconds as m:ss or h:mm:ss. */
export function formatPlaybackTime(seconds) {
  const total = Math.max(0, Number(seconds) || 0)
  const hours = Math.floor(total / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  const secs = Math.floor(total % 60)
  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
  }
  return `${minutes}:${String(secs).padStart(2, '0')}`
}

/** Parse m:ss, mm:ss, or h:mm:ss into seconds. Returns null when invalid. */
export function parsePlaybackTimeInput(value) {
  const raw = String(value || '').trim()
  if (!raw) return null
  const parts = raw.split(':').map((part) => part.trim())
  if (parts.some((part) => part === '' || Number.isNaN(Number(part)))) {
    return null
  }
  if (parts.length === 1) {
    const seconds = Number(parts[0])
    return Number.isFinite(seconds) && seconds >= 0 ? seconds : null
  }
  if (parts.length === 2) {
    const minutes = Number(parts[0])
    const seconds = Number(parts[1])
    if (!Number.isFinite(minutes) || !Number.isFinite(seconds) || minutes < 0 || seconds < 0) {
      return null
    }
    return minutes * 60 + seconds
  }
  if (parts.length === 3) {
    const hours = Number(parts[0])
    const minutes = Number(parts[1])
    const seconds = Number(parts[2])
    if (
      !Number.isFinite(hours) ||
      !Number.isFinite(minutes) ||
      !Number.isFinite(seconds) ||
      hours < 0 ||
      minutes < 0 ||
      seconds < 0
    ) {
      return null
    }
    return hours * 3600 + minutes * 60 + seconds
  }
  return null
}
