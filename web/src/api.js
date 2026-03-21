const jsonHeaders = { 'Content-Type': 'application/json' }

export async function fetchVideos({
  limit = 25,
  offset = 0,
  tags = [],
  search = '',
  sort = '',
  seed = null,
} = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  if (tags.length) params.set('tags', tags.join(','))
  if (search) params.set('search', search)
  if (sort) params.set('sort', sort)
  if (seed != null) params.set('seed', String(seed))
  const res = await fetch(`/videos?${params.toString()}`)
  if (!res.ok) throw new Error('加载视频失败')
  const data = await res.json()
  // Support both new shape {items,total} and legacy array for backward compatibility
  if (Array.isArray(data)) {
    return { items: data, total: data.length }
  }
  return data
}

export async function fetchTags() {
  const res = await fetch('/tags')
  if (!res.ok) throw new Error('加载标签失败')
  return res.json()
}

export async function createTag(name) {
  const res = await fetch('/tags', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ name }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '创建标签失败')
  }
  return res.json()
}

export async function fetchConfig() {
  const res = await fetch('/config')
  if (!res.ok) throw new Error('加载配置失败')
  return res.json()
}

export async function updateConfig(payload) {
  const res = await fetch('/config', {
    method: 'PATCH',
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '更新配置失败')
  }
  return res.json()
}

export async function deleteTag(id) {
  const res = await fetch(`/tags/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '删除标签失败')
  }
}

export async function deleteTagsBatch(tagIds) {
  const res = await fetch('/tags/batch_delete', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ tag_ids: tagIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '批量删除标签失败')
  }
}

export async function renameTag(id, name) {
  const res = await fetch(`/tags/${id}`, {
    method: 'PATCH',
    headers: jsonHeaders,
    body: JSON.stringify({ name }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '重命名标签失败')
  }
}

export async function addTagToVideos(tagId, videoIds) {
  const res = await fetch('/videos/tags/add', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ tag_id: tagId, video_ids: videoIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '添加标签到视频失败')
  }
}

export async function removeTagFromVideos(tagId, videoIds) {
  const res = await fetch('/videos/tags/remove', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ tag_id: tagId, video_ids: videoIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '从视频移除标签失败')
  }
}

export async function replaceTagsForVideos(videoIds, tagIds) {
  const res = await fetch('/videos/tags/replace', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ video_ids: videoIds, tag_ids: tagIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '更新视频标签失败')
  }
}

export async function openVideoFile({ path, dirPath }) {
  const res = await fetch('/videos/open', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ path, dir_path: dirPath }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '打开文件失败')
  }
}

export async function revealVideoLocation({ path, dirPath }) {
  const res = await fetch('/videos/reveal', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ path, dir_path: dirPath }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '打开所在位置失败')
  }
}

export async function incrementVideoPlayCount(id) {
  const res = await fetch(`/videos/${id}/play`, { method: 'POST' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '增加播放次数失败')
  }
}

// Directories
export async function fetchDirectories() {
  const res = await fetch('/directories')
  if (!res.ok) throw new Error('加载目录失败')
  const ct = res.headers.get('content-type') || ''
  if (!ct.includes('application/json')) {
    console.warn('目录接口返回非 JSON，响应类型:', ct)
    return []
  }
  return res.json()
}

export async function createDirectory({ path }) {
  const res = await fetch('/directories', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ path }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '创建目录失败')
  }
  return res.json()
}

export async function pickDirectory() {
  const res = await fetch('/directories/pick', {
    method: 'POST',
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '选择目录失败')
  }
  return res.json()
}

export async function updateDirectory(id, payload) {
  const res = await fetch(`/directories/${id}`, {
    method: 'PATCH',
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '更新目录失败')
  }
  return res.json()
}

export async function deleteDirectory(id) {
  return updateDirectory(id, { is_delete: true })
}

export async function fetchJavs({
  limit = 25,
  offset = 0,
  search = '',
  actors = [],
  tagIds = [],
  sort = '',
  seed = null,
} = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  if (search) params.set('search', search)
  if (actors.length) params.set('actors', actors.join(','))
  if (tagIds.length) params.set('tag_ids', tagIds.join(','))
  if (sort) params.set('sort', sort)
  if (seed != null) params.set('seed', String(seed))
  const res = await fetch(`/jav?${params.toString()}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '加载 JAV 失败')
  }
  return res.json()
}

export async function fetchJavTags() {
  const res = await fetch('/jav/tags')
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '加载 JAV 标签失败')
  }
  return res.json()
}

export async function createJavTag(name) {
  const res = await fetch('/jav/tags', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ name }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '创建 JAV 标签失败')
  }
  return res.json()
}

export async function renameJavTag(id, name) {
  const res = await fetch(`/jav/tags/${id}`, {
    method: 'PATCH',
    headers: jsonHeaders,
    body: JSON.stringify({ name }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '重命名 JAV 标签失败')
  }
}

export async function deleteJavTag(id) {
  const res = await fetch(`/jav/tags/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '删除 JAV 标签失败')
  }
}

export async function deleteJavTagsBatch(tagIds) {
  const res = await fetch('/jav/tags/batch_delete', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ tag_ids: tagIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '批量删除 JAV 标签失败')
  }
}

export async function replaceJavTagsForItems(javIds, tagIds) {
  const res = await fetch('/jav/tags/replace', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ jav_ids: javIds, tag_ids: tagIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '更新 JAV 标签失败')
  }
}

export async function addJavTagToJavs(tagId, javIds) {
  const res = await fetch('/jav/tags/add', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ tag_id: tagId, jav_ids: javIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '添加 JAV 标签失败')
  }
}

export async function removeJavTagFromJavs(tagId, javIds) {
  const res = await fetch('/jav/tags/remove', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ tag_id: tagId, jav_ids: javIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '移除 JAV 标签失败')
  }
}

export async function fetchJavIdols({ limit = 25, offset = 0, search = '', sort = '' } = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  if (search) params.set('search', search)
  if (sort) params.set('sort', sort)
  const res = await fetch(`/jav/idols?${params.toString()}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || '加载女优失败')
  }
  return res.json()
}
