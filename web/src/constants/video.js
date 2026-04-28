export const VIDEO_SORT_OPTIONS = [
  {
    base: 'recent',
    defaultValue: 'recent',
    ascValue: 'recent_asc',
    descValue: 'recent',
    label: ['加入时间', 'Added time'],
    asc: ['旧→新', 'old→new'],
    desc: ['新→旧', 'new→old'],
  },
  {
    base: 'filename',
    defaultValue: 'filename',
    ascValue: 'filename',
    descValue: 'filename_desc',
    label: ['文件名', 'Filename'],
    asc: ['A→Z', 'A→Z'],
    desc: ['Z→A', 'Z→A'],
  },
  {
    base: 'duration',
    defaultValue: 'duration',
    ascValue: 'duration_asc',
    descValue: 'duration',
    label: ['时长', 'Duration'],
    asc: ['短→长', 'short→long'],
    desc: ['长→短', 'long→short'],
  },
  {
    base: 'play_count',
    defaultValue: 'play_count',
    ascValue: 'play_count_asc',
    descValue: 'play_count',
    label: ['播放次数', 'Play count'],
    asc: ['少→多', 'low→high'],
    desc: ['多→少', 'high→low'],
  },
]

const videoSortValues = new Set(
  VIDEO_SORT_OPTIONS.flatMap((option) => [option.defaultValue, option.ascValue, option.descValue])
)

export function normalizeVideoSort(sort, fallback = 'recent') {
  const key = String(sort || '')
    .trim()
    .toLowerCase()
  if (key === 'recent_desc') return 'recent'
  if (key === 'filename_asc') return 'filename'
  if (key === 'duration_desc') return 'duration'
  if (key === 'play_count_desc') return 'play_count'
  if (videoSortValues.has(key)) return key
  return fallback
}

export function findVideoSortOption(sort) {
  const normalized = String(sort || '')
    .trim()
    .toLowerCase()
  return VIDEO_SORT_OPTIONS.find(
    (option) =>
      option.defaultValue === normalized ||
      option.ascValue === normalized ||
      option.descValue === normalized
  )
}

export function getVideoSortDirection(option, sort) {
  if (!option) return 'asc'
  return String(sort || '')
    .trim()
    .toLowerCase() === option.ascValue
    ? 'asc'
    : 'desc'
}

export function reverseVideoSortValue(sort, fallback = 'recent') {
  const option = findVideoSortOption(sort) || findVideoSortOption(fallback)
  if (!option) return fallback
  return getVideoSortDirection(option, sort) === 'asc' ? option.descValue : option.ascValue
}

export function videoSortLabelParts(option, sort, zh) {
  if (!option) return { label: '', separator: '', direction: '' }
  const dir = getVideoSortDirection(option, sort)
  return {
    label: zh(option.label[0], option.label[1]),
    separator: zh('：', ': '),
    direction: zh(option[dir][0], option[dir][1]),
  }
}
