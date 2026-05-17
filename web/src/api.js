import { zh } from '@/utils/i18n'

const jsonHeaders = { 'Content-Type': 'application/json' }
const javIdolResolveInFlight = new Map()

export async function fetchVideos({
  limit = 25,
  offset = 0,
  tags = [],
  search = '',
  sort = '',
  seed = null,
  directoryIds = [],
  hideJav = true,
} = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  if (tags.length) params.set('tags', tags.join(','))
  if (search) params.set('search', search)
  if (sort) params.set('sort', sort)
  if (seed != null) params.set('seed', String(seed))
  if (directoryIds.length) params.set('directory_ids', directoryIds.join(','))
  params.set('hide_jav', hideJav ? '1' : '0')
  const res = await fetch(`/videos?${params.toString()}`)
  if (!res.ok) throw new Error(zh('加载视频失败', 'Failed to load videos'))
  const data = await res.json()
  // Support both new shape {items,total} and legacy array for backward compatibility
  if (Array.isArray(data)) {
    return { items: data, total: data.length }
  }
  return data
}

export async function fetchTags({ directoryIds = [], hideJav = true } = {}) {
  const params = new URLSearchParams()
  if (directoryIds.length) params.set('directory_ids', directoryIds.join(','))
  params.set('hide_jav', hideJav ? '1' : '0')
  const query = params.toString()
  const res = await fetch(`/tags${query ? `?${query}` : ''}`)
  if (!res.ok) throw new Error(zh('加载标签失败', 'Failed to load tags'))
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
    throw new Error(err.error || zh('创建标签失败', 'Failed to create tag'))
  }
  return res.json()
}

export async function fetchConfig() {
  const res = await fetch('/config')
  if (!res.ok) throw new Error(zh('加载配置失败', 'Failed to load config'))
  return res.json()
}

export async function updateConfig(payload) {
  let res
  try {
    res = await fetch('/config', {
      method: 'PATCH',
      headers: jsonHeaders,
      body: JSON.stringify(payload),
    })
  } catch (err) {
    const detail = err?.message ? String(err.message) : String(err)
    throw new Error(
      zh(
        `无法连接服务器（${detail}）。请确认 Go 后端已启动；使用 npm run dev 时代理默认指向 http://127.0.0.1:17654，若端口不同请在 web/.env.local 设置 DEV_API_ORIGIN=你的后端根地址（如 http://127.0.0.1:8080）。`,
        `Cannot reach server (${detail}). Start the Go API; with Vite dev the proxy defaults to http://127.0.0.1:17654. Set DEV_API_ORIGIN in web/.env.local to override (e.g. http://127.0.0.1:8080).`
      )
    )
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('更新配置失败', 'Failed to update config'))
  }
  return res.json()
}

export async function deleteTag(id) {
  const res = await fetch(`/tags/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('删除标签失败', 'Failed to delete tag'))
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
    throw new Error(err.error || zh('批量删除标签失败', 'Failed to delete tags'))
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
    throw new Error(err.error || zh('重命名标签失败', 'Failed to rename tag'))
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
    throw new Error(err.error || zh('添加标签到视频失败', 'Failed to add tag to videos'))
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
    throw new Error(err.error || zh('从视频移除标签失败', 'Failed to remove tag from videos'))
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
    throw new Error(err.error || zh('更新视频标签失败', 'Failed to update video tags'))
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
    throw new Error(err.error || zh('打开文件失败', 'Failed to open file'))
  }
}

export async function playVideoFile({ id, path, dirPath, startTime }) {
  const res = await fetch('/videos/play', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ video_id: id, path, dir_path: dirPath, start_time: startTime }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('播放文件失败', 'Failed to play file'))
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
    throw new Error(err.error || zh('打开所在位置失败', 'Failed to reveal file'))
  }
}

export async function incrementVideoPlayCount(id) {
  const res = await fetch(`/videos/${id}/play`, { method: 'POST' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('增加播放次数失败', 'Failed to increment play count'))
  }
}

export async function fetchPlaybackInfo(id) {
  const res = await fetch(`/videos/${id}/streams`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载播放信息失败', 'Failed to load playback info'))
  }
  return res.json()
}

export async function fetchVideoScreenshots(id) {
  const res = await fetch(`/videos/${id}/screenshots`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载截图失败', 'Failed to load screenshots'))
  }
  const data = await res.json()
  return Array.isArray(data?.items) ? data.items : []
}

export async function deleteVideoScreenshot(videoId, name) {
  const res = await fetch(`/videos/${videoId}/screenshots/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('删除截图失败', 'Failed to delete screenshot'))
  }
}

// Directories
export async function fetchDirectories() {
  const res = await fetch('/directories')
  if (!res.ok) throw new Error(zh('加载目录失败', 'Failed to load directories'))
  const ct = res.headers.get('content-type') || ''
  if (!ct.includes('application/json')) {
    console.warn(
      zh('目录接口返回非 JSON，响应类型:', 'Directory API returned non-JSON content type:'),
      ct
    )
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
    throw new Error(err.error || zh('创建目录失败', 'Failed to create directory'))
  }
  return res.json()
}

export async function pickDirectory() {
  const res = await fetch('/directories/pick', {
    method: 'POST',
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('选择目录失败', 'Failed to choose directory'))
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
    throw new Error(err.error || zh('更新目录失败', 'Failed to update directory'))
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
  idolIds = [],
  tagIds = [],
  studioId = null,
  collectionId = null,
  sort = '',
  seed = null,
  directoryIds = [],
} = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  if (search) params.set('search', search)
  if (idolIds.length) params.set('idol_ids', idolIds.join(','))
  if (tagIds.length) params.set('tag_ids', tagIds.join(','))
  if (studioId) params.set('studio_id', String(studioId))
  if (collectionId) params.set('collection_id', String(collectionId))
  if (sort) params.set('sort', sort)
  if (seed != null) params.set('seed', String(seed))
  if (directoryIds.length) params.set('directory_ids', directoryIds.join(','))
  const res = await fetch(`/jav?${params.toString()}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载 JAV 失败', 'Failed to load JAV'))
  }
  return res.json()
}

export async function fetchJavTags({ directoryIds = [] } = {}) {
  const params = new URLSearchParams()
  if (directoryIds.length) params.set('directory_ids', directoryIds.join(','))
  const query = params.toString()
  const res = await fetch(`/jav/tags${query ? `?${query}` : ''}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载 JAV 标签失败', 'Failed to load JAV tags'))
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
    throw new Error(err.error || zh('创建 JAV 标签失败', 'Failed to create JAV tag'))
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
    throw new Error(err.error || zh('重命名 JAV 标签失败', 'Failed to rename JAV tag'))
  }
}

export async function deleteJavTag(id) {
  const res = await fetch(`/jav/tags/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('删除 JAV 标签失败', 'Failed to delete JAV tag'))
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
    throw new Error(err.error || zh('批量删除 JAV 标签失败', 'Failed to delete JAV tags'))
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
    throw new Error(err.error || zh('更新 JAV 标签失败', 'Failed to update JAV tags'))
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
    throw new Error(err.error || zh('添加 JAV 标签失败', 'Failed to add JAV tag'))
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
    throw new Error(err.error || zh('移除 JAV 标签失败', 'Failed to remove JAV tag'))
  }
}

export async function fetchJavIdols({
  limit = 25,
  offset = 0,
  search = '',
  sort = '',
  directoryIds = [],
} = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  if (search) params.set('search', search)
  if (sort) params.set('sort', sort)
  if (directoryIds.length) params.set('directory_ids', directoryIds.join(','))
  const res = await fetch(`/jav/idols?${params.toString()}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载女优失败', 'Failed to load idols'))
  }
  return res.json()
}

export async function fetchJavStudios({
  limit = 25,
  offset = 0,
  search = '',
  directoryIds = [],
} = {}) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  if (search) params.set('search', search)
  if (directoryIds.length) params.set('directory_ids', directoryIds.join(','))
  const res = await fetch(`/jav/studios?${params.toString()}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载片商失败', 'Failed to load studios'))
  }
  return res.json()
}

export async function fetchJavIdolPreview(id, { directoryIds = [] } = {}) {
  const params = new URLSearchParams()
  if (directoryIds.length) params.set('directory_ids', directoryIds.join(','))
  const query = params.toString()
  const res = await fetch(`/jav/idols/${encodeURIComponent(id)}${query ? `?${query}` : ''}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载女优预览失败', 'Failed to load idol preview'))
  }
  return res.json()
}

export async function resolveJavIdols(ids = []) {
  const clean = Array.from(
    new Set(
      (ids || [])
        .map((id) => Number.parseInt(String(id), 10))
        .filter((id) => Number.isFinite(id) && id > 0)
    )
  ).sort((a, b) => a - b)
  if (!clean.length) return []
  const key = clean.join(',')
  if (javIdolResolveInFlight.has(key)) {
    return javIdolResolveInFlight.get(key)
  }
  const params = new URLSearchParams()
  params.set('ids', clean.join(','))
  const request = fetch(`/jav/idols/resolve?${params.toString()}`)
    .then(async (res) => {
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || zh('加载女优名称失败', 'Failed to load idol names'))
      }
      const data = await res.json()
      return Array.isArray(data?.items) ? data.items : []
    })
    .finally(() => {
      javIdolResolveInFlight.delete(key)
    })
  javIdolResolveInFlight.set(key, request)
  return request
}

export async function fetchCollections() {
  const res = await fetch('/collections')
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加载合集失败', 'Failed to load collections'))
  }
  return res.json()
}

export async function createCollection({ name, description = '' }) {
  const res = await fetch('/collections', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ name, description }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('创建合集失败', 'Failed to create collection'))
  }
  return res.json()
}

export async function updateCollection(id, payload) {
  const res = await fetch(`/collections/${id}`, {
    method: 'PATCH',
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('更新合集失败', 'Failed to update collection'))
  }
  return res.json()
}

export async function deleteCollection(id) {
  const res = await fetch(`/collections/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('删除合集失败', 'Failed to delete collection'))
  }
}

export async function addJavsToCollection(collectionId, javIds) {
  const res = await fetch('/collections/javs/add', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ collection_id: collectionId, jav_ids: javIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('加入合集失败', 'Failed to add to collection'))
  }
}

export async function removeJavsFromCollection(collectionId, javIds) {
  const res = await fetch('/collections/javs/remove', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ collection_id: collectionId, jav_ids: javIds }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('从合集移除失败', 'Failed to remove from collection'))
  }
}

export async function analyzeCollection(id) {
  const res = await fetch(`/collections/${id}/analyze`, { method: 'POST' })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('AI 分析失败', 'AI analysis failed'))
  }
  return res.json()
}

export async function postJavNlQuery({
  query,
  directoryIds = [],
  collectionId = null,
  limit = 24,
  offset = 0,
  sort = 'recent',
} = {}) {
  const res = await fetch('/jav/nl_query', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({
      query,
      directory_ids: directoryIds,
      collection_id: collectionId || 0,
      limit,
      offset,
      sort,
    }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('自然语言检索失败', 'Natural language search failed'))
  }
  return res.json()
}

export async function postJavLibraryChat({
  message,
  history = [],
  directoryIds = [],
  collectionId = null,
  mode = 'library',
  focusCode = '',
} = {}) {
  const res = await fetch('/jav/library_chat', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({
      message,
      history,
      directory_ids: directoryIds,
      collection_id: collectionId || 0,
      mode: mode || 'library',
      focus_code: String(focusCode || '').trim(),
    }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('对话请求失败', 'Chat request failed'))
  }
  return res.json()
}

export async function postJavCodeInsight(code) {
  const res = await fetch('/jav/code_insight', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ code: String(code || '').trim() }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || zh('AI 番号解读失败', 'Code insight request failed'))
  }
  return res.json()
}
