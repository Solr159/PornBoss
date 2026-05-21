import { useEffect, useMemo, useState } from 'react'
import CloseOutlinedIcon from '@mui/icons-material/CloseOutlined'

import { fetchJavSeries, createJavTag } from '@/api'
import { isUserJavTag } from '@/constants/jav'
import { zh } from '@/utils/i18n'

async function fetchAllJavSeries({ directoryIds = [] } = {}) {
  const pageSize = 200
  let offset = 0
  let total = Number.POSITIVE_INFINITY
  const items = []
  while (offset < total) {
    const resp = await fetchJavSeries({ limit: pageSize, offset, directoryIds })
    const batch = Array.isArray(resp?.items) ? resp.items : []
    total = Number.isFinite(resp?.total) ? resp.total : batch.length
    items.push(...batch)
    if (batch.length < pageSize) break
    offset += pageSize
  }
  return items
}

export default function JavMetadataEditModal({
  open,
  item,
  tags = [],
  directoryIds = [],
  metadataLanguage = 'ja',
  onClose,
  onSave,
  saving = false,
}) {
  const isEnglish = metadataLanguage === 'en'
  const [selectedTagIds, setSelectedTagIds] = useState([])
  const [tagSearch, setTagSearch] = useState('')
  const [selectedSeries, setSelectedSeries] = useState(null)
  const [seriesSearch, setSeriesSearch] = useState('')
  const [seriesPickerOpen, setSeriesPickerOpen] = useState(false)
  const [allSeries, setAllSeries] = useState([])
  const [seriesLoading, setSeriesLoading] = useState(false)
  const [seriesError, setSeriesError] = useState('')
  const [newTagName, setNewTagName] = useState('')
  const [creatingTag, setCreatingTag] = useState(false)
  const [createTagError, setCreateTagError] = useState('')
  const [lockEdits, setLockEdits] = useState(true)

  const code = String(item?.code || '').trim()
  const preferredSeries = isEnglish ? item?.series_en : item?.series

  useEffect(() => {
    if (!open || !item) return
    const initialTags = Array.isArray(item.tags) ? item.tags.map((tag) => String(tag.id)) : []
    setSelectedTagIds(initialTags)
    setTagSearch('')
    const series = preferredSeries
    const seriesId = Number(series?.id)
    setSelectedSeries(
      Number.isFinite(seriesId) && seriesId > 0
        ? { id: seriesId, name: String(series?.name || '').trim() || `#${seriesId}` }
        : null
    )
    setSeriesSearch('')
    setSeriesPickerOpen(false)
    setNewTagName('')
    setCreateTagError('')
    setLockEdits(true)
  }, [item, open, preferredSeries])

  useEffect(() => {
    if (!open) return
    let cancelled = false
    setSeriesLoading(true)
    setSeriesError('')
    fetchAllJavSeries({ directoryIds })
      .then((items) => {
        if (!cancelled) setAllSeries(items)
      })
      .catch((err) => {
        if (!cancelled) {
          setAllSeries([])
          setSeriesError(err.message || zh('加载系列失败', 'Failed to load series'))
        }
      })
      .finally(() => {
        if (!cancelled) setSeriesLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [directoryIds, open])

  const tagOptions = useMemo(() => {
    const list = Array.isArray(tags) ? [...tags] : []
    return list.sort((a, b) => String(a?.name || '').localeCompare(String(b?.name || '')))
  }, [tags])

  const filteredTags = useMemo(() => {
    const query = tagSearch.trim().toLowerCase()
    return tagOptions.filter((tag) => {
      if (!query) return true
      return String(tag?.name || '')
        .toLowerCase()
        .includes(query)
    })
  }, [tagOptions, tagSearch])

  const filteredSeries = useMemo(() => {
    const query = seriesSearch.trim().toLowerCase()
    return [...allSeries]
      .filter((series) => {
        if (!query) return true
        return String(series?.name || '')
          .toLowerCase()
          .includes(query)
      })
      .sort((a, b) => {
        const countA = Number.isFinite(a?.work_count) ? a.work_count : 0
        const countB = Number.isFinite(b?.work_count) ? b.work_count : 0
        if (countB !== countA) return countB - countA
        return String(a?.name || '').localeCompare(String(b?.name || ''))
      })
  }, [allSeries, seriesSearch])

  const initialTagIds = useMemo(() => {
    if (!item) return []
    return Array.isArray(item.tags) ? item.tags.map((tag) => String(tag.id)) : []
  }, [item])

  const initialSeriesKey = useMemo(() => {
    const id = Number(preferredSeries?.id)
    if (Number.isFinite(id) && id > 0) return `id:${id}`
    const name = String(preferredSeries?.name || '').trim()
    return name ? `name:${name}` : ''
  }, [preferredSeries])

  const selectedSeriesKey = useMemo(() => {
    if (!selectedSeries) return ''
    const id = Number(selectedSeries.id)
    if (Number.isFinite(id) && id > 0) return `id:${id}`
    const name = String(selectedSeries.name || '').trim()
    return name ? `name:${name}` : ''
  }, [selectedSeries])

  const dirty = useMemo(() => {
    const tagSet = new Set(selectedTagIds)
    const initialSet = new Set(initialTagIds)
    if (tagSet.size !== initialSet.size) return true
    for (const id of tagSet) {
      if (!initialSet.has(id)) return true
    }
    return selectedSeriesKey !== initialSeriesKey
  }, [initialSeriesKey, initialTagIds, selectedSeriesKey, selectedTagIds])

  const toggleTag = (tagId) => {
    const key = String(tagId)
    setSelectedTagIds((prev) => {
      const set = new Set(prev)
      if (set.has(key)) set.delete(key)
      else set.add(key)
      return Array.from(set)
    })
  }

  const handleCreateTag = async () => {
    const name = newTagName.trim()
    if (!name) return
    setCreatingTag(true)
    setCreateTagError('')
    try {
      const created = await createJavTag(name)
      if (created?.id) {
        setSelectedTagIds((prev) => [...new Set([...prev, String(created.id)])])
        setNewTagName('')
      }
    } catch (err) {
      setCreateTagError(err.message || zh('创建标签失败', 'Failed to create tag'))
    } finally {
      setCreatingTag(false)
    }
  }

  const handleSave = () => {
    if (!item?.id) return
    const payload = {
      tag_ids: selectedTagIds.map((id) => Number(id)).filter((id) => id > 0),
      lock: lockEdits,
    }
    const seriesId = Number(selectedSeries?.id)
    if (Number.isFinite(seriesId) && seriesId > 0) {
      payload.series_id = seriesId
    } else if (selectedSeries?.name) {
      payload.series_name = String(selectedSeries.name).trim()
    } else {
      payload.clear_series = true
    }
    onSave?.(item, payload)
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4">
      <div className="flex max-h-[90vh] w-full max-w-lg flex-col rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div>
            <h2 className="text-base font-semibold">{zh('编辑 JAV 信息', 'Edit JAV metadata')}</h2>
            {code ? <p className="text-xs text-gray-500">{code}</p> : null}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-gray-500 hover:bg-gray-100"
            aria-label={zh('关闭', 'Close')}
          >
            <CloseOutlinedIcon fontSize="small" />
          </button>
        </div>

        <div className="space-y-5 overflow-y-auto px-4 py-4">
          <p className="text-xs text-gray-500">
            {zh(
              '修改分类与系列后默认锁定，后续从站点抓取或扫描不会覆盖这些字段。',
              'Edits are locked by default so future scrapes and scans will not overwrite these fields.'
            )}
          </p>

          <label className="flex cursor-pointer items-center gap-2 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={lockEdits}
              onChange={(e) => setLockEdits(e.target.checked)}
              className="rounded border-gray-300"
            />
            {zh('锁定，防止被抓取覆盖', 'Lock to prevent scrape overwrites')}
          </label>

          <section className="space-y-2">
            <div className="text-sm font-semibold text-gray-800">{zh('系列', 'Series')}</div>
            {selectedSeries ? (
              <div className="flex flex-wrap gap-2">
                <span className="inline-flex max-w-full items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-800">
                  <span className="truncate">{selectedSeries.name}</span>
                  <button
                    type="button"
                    onClick={() => setSelectedSeries(null)}
                    className="inline-flex h-4 w-4 items-center justify-center rounded-full hover:bg-emerald-100"
                    aria-label={zh('清除系列', 'Clear series')}
                  >
                    <CloseOutlinedIcon fontSize="inherit" />
                  </button>
                </span>
              </div>
            ) : (
              <p className="text-xs text-gray-400">{zh('未设置系列', 'No series')}</p>
            )}
            <input
              value={seriesSearch}
              onFocus={() => setSeriesPickerOpen(true)}
              onChange={(e) => {
                setSeriesSearch(e.target.value)
                setSeriesPickerOpen(true)
              }}
              className="w-full rounded border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              placeholder={zh('搜索或输入新系列名', 'Search or type a series name')}
            />
            {seriesPickerOpen ? (
              <div className="max-h-40 overflow-y-auto rounded border border-gray-200 bg-white p-1 shadow-sm">
                {seriesSearch.trim() ? (
                  <button
                    type="button"
                    className="flex w-full rounded px-2 py-1.5 text-left text-sm text-blue-700 hover:bg-blue-50"
                    onMouseDown={(e) => e.preventDefault()}
                    onClick={() => {
                      setSelectedSeries({ id: null, name: seriesSearch.trim() })
                      setSeriesPickerOpen(false)
                    }}
                  >
                    {zh(`使用新系列「${seriesSearch.trim()}」`, `Use new series "${seriesSearch.trim()}"`)}
                  </button>
                ) : null}
                {seriesLoading ? (
                  <div className="px-2 py-2 text-sm text-gray-500">{zh('加载中…', 'Loading...')}</div>
                ) : seriesError ? (
                  <div className="px-2 py-2 text-sm text-red-600">{seriesError}</div>
                ) : filteredSeries.length > 0 ? (
                  filteredSeries.slice(0, 40).map((series) => (
                    <button
                      key={series.id}
                      type="button"
                      className="flex w-full rounded px-2 py-1.5 text-left text-sm hover:bg-gray-50"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => {
                        setSelectedSeries({ id: series.id, name: series.name })
                        setSeriesSearch('')
                        setSeriesPickerOpen(false)
                      }}
                    >
                      <span className="truncate">{series.name}</span>
                    </button>
                  ))
                ) : (
                  <div className="px-2 py-2 text-sm text-gray-500">
                    {zh('没有匹配系列', 'No matching series')}
                  </div>
                )}
              </div>
            ) : null}
          </section>

          <section className="space-y-2">
            <div className="text-sm font-semibold text-gray-800">
              {zh('分类 / 标签', 'Categories / tags')}
            </div>
            <input
              value={tagSearch}
              onChange={(e) => setTagSearch(e.target.value)}
              className="w-full rounded border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              placeholder={zh('搜索标签', 'Search tags')}
            />
            <div className="max-h-48 space-y-1 overflow-y-auto rounded border p-2">
              {filteredTags.map((tag) => {
                const checked = selectedTagIds.includes(String(tag.id))
                const isUser = isUserJavTag(tag)
                return (
                  <label
                    key={tag.id}
                    className="flex cursor-pointer items-center gap-2 rounded px-2 py-1 hover:bg-gray-50"
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleTag(tag.id)}
                    />
                    <span className="text-sm text-gray-800">{tag.name}</span>
                    <span
                      className={`ml-auto text-[10px] ${isUser ? 'text-emerald-600' : 'text-orange-600'}`}
                    >
                      {isUser ? zh('我的', 'Mine') : zh('站点', 'Site')}
                    </span>
                  </label>
                )
              })}
            </div>
            <div className="flex gap-2">
              <input
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                className="min-w-0 flex-1 rounded border border-gray-200 px-3 py-1.5 text-sm"
                placeholder={zh('新建标签', 'New tag')}
              />
              <button
                type="button"
                onClick={handleCreateTag}
                disabled={creatingTag || !newTagName.trim()}
                className="shrink-0 rounded border px-3 py-1.5 text-sm hover:bg-gray-50 disabled:opacity-50"
              >
                {zh('添加', 'Add')}
              </button>
            </div>
            {createTagError ? <p className="text-xs text-red-600">{createTagError}</p> : null}
          </section>
        </div>

        <div className="flex justify-end gap-2 border-t px-4 py-3">
          <button
            type="button"
            onClick={onClose}
            className="rounded border px-3 py-1.5 text-sm hover:bg-gray-50"
          >
            {zh('取消', 'Cancel')}
          </button>
          <button
            type="button"
            onClick={handleSave}
            disabled={saving || !dirty}
            className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? zh('保存中…', 'Saving...') : zh('保存', 'Save')}
          </button>
        </div>
      </div>
    </div>
  )
}
