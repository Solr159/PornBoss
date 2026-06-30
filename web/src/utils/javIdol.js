import { zh } from '@/utils/i18n'

export function getIdolDisplayName(item, javMetadataLanguage = 'zh', preferChineseName = false) {
  return getIdolDisplayNames(item, javMetadataLanguage, preferChineseName).primaryName
}

export function getIdolDisplayNames(item, javMetadataLanguage = 'zh', preferChineseName = false) {
  return buildDisplayNames({
    name: item?.name || '',
    japaneseName: item?.japanese_name || '',
    chineseName: item?.chinese_name || '',
    javMetadataLanguage,
    preferChineseName,
  })
}

function buildDisplayNames({
  name,
  japaneseName,
  chineseName,
  javMetadataLanguage,
  preferChineseName,
}) {
  if (preferChineseName && String(chineseName || '').trim()) {
    return {
      primaryName: String(chineseName || '').trim(),
      secondaryName: joinUniqueDisplayParts([name], [chineseName]),
    }
  }

  if (javMetadataLanguage === 'en') {
    const primaryName = firstNonEmpty(
      name,
      japaneseName,
      chineseName,
      zh('Unknown idol', 'Unknown idol')
    )
    return {
      primaryName,
      secondaryName: joinUniqueDisplayParts([japaneseName, chineseName], [primaryName]),
    }
  }

  const primaryName = firstNonEmpty(name, chineseName, zh('未知女优', 'Unknown idol'))
  return {
    primaryName,
    secondaryName: joinUniqueDisplayParts([chineseName], [primaryName]),
  }
}

function firstNonEmpty(...values) {
  for (const value of values) {
    const trimmed = String(value || '').trim()
    if (trimmed) return trimmed
  }
  return ''
}

function joinUniqueDisplayParts(values, exclude = []) {
  const excluded = new Set(exclude.map((value) => String(value || '').trim()).filter(Boolean))
  const seen = new Set()
  const parts = []
  for (const value of values) {
    const trimmed = String(value || '').trim()
    if (!trimmed || excluded.has(trimmed) || seen.has(trimmed)) {
      continue
    }
    seen.add(trimmed)
    parts.push(trimmed)
  }
  return parts.join(' · ')
}
