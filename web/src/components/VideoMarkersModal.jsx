import { useCallback, useEffect, useMemo, useState } from 'react'
import CloseIcon from '@mui/icons-material/Close'
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline'
import EditOutlinedIcon from '@mui/icons-material/EditOutlined'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import { IconButton, Tooltip } from '@mui/material'
import { createVideoMarker, deleteVideoMarker, fetchVideoMarkers, updateVideoMarker } from '@/api'
import { getVideoDisplayName } from '@/utils/display'
import { zh } from '@/utils/i18n'
import { formatPlaybackTime, parsePlaybackTimeInput } from '@/utils/playbackTime'

function MarkerTimeline({ durationSec, markers, onSeek }) {
  const duration = Number(durationSec)
  const hasDuration = Number.isFinite(duration) && duration > 0

  return (
    <div className="rounded-lg border border-gray-200 bg-gray-50 p-3">
      <div className="mb-2 text-xs font-medium text-gray-600">
        {zh('时间轴', 'Timeline')}
        {hasDuration ? (
          <span className="ml-2 text-gray-400">({formatPlaybackTime(duration)})</span>
        ) : (
          <span className="ml-2 text-gray-400">
            ({zh('时长未知，按比例展示', 'Unknown duration, shown proportionally')})
          </span>
        )}
      </div>
      <div className="relative h-10 rounded-md bg-gray-300/80">
        <div className="absolute inset-y-0 left-0 right-0 rounded-md bg-gradient-to-r from-sky-200/70 to-sky-400/50" />
        {markers.map((marker) => {
          const maxTime = hasDuration
            ? duration
            : Math.max(...markers.map((item) => item.time_sec), marker.time_sec, 1)
          const left = `${Math.min(100, Math.max(0, (marker.time_sec / maxTime) * 100))}%`
          return (
            <Tooltip
              key={marker.id}
              title={`${formatPlaybackTime(marker.time_sec)} — ${marker.note}`}
            >
              <button
                type="button"
                className="absolute top-1/2 z-10 h-3.5 w-3.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white bg-amber-500 shadow hover:scale-110 hover:bg-amber-400"
                style={{ left }}
                aria-label={marker.note}
                onClick={() => onSeek?.(marker.time_sec)}
              />
            </Tooltip>
          )
        })}
      </div>
    </div>
  )
}

export default function VideoMarkersModal({ video, onClose, onPlayAtTime }) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [timeInput, setTimeInput] = useState('0:00')
  const [noteInput, setNoteInput] = useState('')
  const [saving, setSaving] = useState(false)
  const [editingId, setEditingId] = useState(null)
  const [editTimeInput, setEditTimeInput] = useState('')
  const [editNoteInput, setEditNoteInput] = useState('')
  const [deletingId, setDeletingId] = useState(null)

  const open = Boolean(video?.id)
  const title = useMemo(() => getVideoDisplayName(video), [video])
  const durationSec = Number(video?.duration_sec)

  const reload = useCallback(async () => {
    if (!video?.id) return
    setLoading(true)
    setError('')
    try {
      const next = await fetchVideoMarkers(video.id)
      setItems(next)
    } catch (err) {
      setError(err?.message || zh('加载时间点标记失败', 'Failed to load video markers'))
    } finally {
      setLoading(false)
    }
  }, [video?.id])

  useEffect(() => {
    if (!open) return undefined
    setItems([])
    setTimeInput('0:00')
    setNoteInput('')
    setEditingId(null)
    setDeletingId(null)
    reload()
    return undefined
  }, [open, reload])

  const handleCreate = async (event) => {
    event.preventDefault()
    if (!video?.id || saving) return
    const timeSec = parsePlaybackTimeInput(timeInput)
    if (timeSec == null) {
      setError(zh('时间格式无效，请使用 mm:ss 或 h:mm:ss', 'Invalid time, use mm:ss or h:mm:ss'))
      return
    }
    const note = noteInput.trim()
    if (!note) {
      setError(zh('请填写说明', 'Please enter a note'))
      return
    }
    setSaving(true)
    setError('')
    try {
      const created = await createVideoMarker(video.id, { timeSec, note })
      setItems((prev) => [...prev, created].sort((a, b) => a.time_sec - b.time_sec || a.id - b.id))
      setNoteInput('')
    } catch (err) {
      setError(err?.message || zh('添加失败', 'Failed to add marker'))
    } finally {
      setSaving(false)
    }
  }

  const startEdit = (marker) => {
    setEditingId(marker.id)
    setEditTimeInput(formatPlaybackTime(marker.time_sec))
    setEditNoteInput(marker.note)
    setError('')
  }

  const cancelEdit = () => {
    setEditingId(null)
    setEditTimeInput('')
    setEditNoteInput('')
  }

  const saveEdit = async (markerId) => {
    if (!video?.id || saving) return
    const timeSec = parsePlaybackTimeInput(editTimeInput)
    if (timeSec == null) {
      setError(zh('时间格式无效', 'Invalid time format'))
      return
    }
    const note = editNoteInput.trim()
    if (!note) {
      setError(zh('请填写说明', 'Please enter a note'))
      return
    }
    setSaving(true)
    setError('')
    try {
      const updated = await updateVideoMarker(video.id, markerId, { timeSec, note })
      setItems((prev) =>
        prev
          .map((item) => (item.id === markerId ? updated : item))
          .sort((a, b) => a.time_sec - b.time_sec || a.id - b.id)
      )
      cancelEdit()
    } catch (err) {
      setError(err?.message || zh('更新失败', 'Failed to update marker'))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (markerId) => {
    if (!video?.id || deletingId) return
    setDeletingId(markerId)
    setError('')
    try {
      await deleteVideoMarker(video.id, markerId)
      setItems((prev) => prev.filter((item) => item.id !== markerId))
      if (editingId === markerId) cancelEdit()
    } catch (err) {
      setError(err?.message || zh('删除失败', 'Failed to delete marker'))
    } finally {
      setDeletingId(null)
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="flex max-h-[90vh] w-full max-w-2xl flex-col overflow-hidden rounded-xl bg-white shadow-xl">
        <div className="flex items-start justify-between border-b border-gray-200 px-4 py-3">
          <div className="min-w-0 pr-3">
            <h2 className="truncate text-lg font-semibold text-gray-900" title={title}>
              {zh('时间点标记', 'Timeline markers')}
            </h2>
            <p className="truncate text-sm text-gray-500" title={title}>
              {title}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-full p-1 text-gray-500 hover:bg-gray-100"
            aria-label={zh('关闭', 'Close')}
          >
            <CloseIcon fontSize="small" />
          </button>
        </div>

        <div className="flex-1 space-y-4 overflow-y-auto px-4 py-4">
          {items.length > 0 ? (
            <MarkerTimeline
              durationSec={durationSec}
              markers={items}
              onSeek={(timeSec) => onPlayAtTime?.(video, timeSec)}
            />
          ) : null}

          <form onSubmit={handleCreate} className="space-y-3 rounded-lg border border-gray-200 p-3">
            <div className="text-sm font-medium text-gray-800">{zh('添加标记', 'Add marker')}</div>
            <div className="flex flex-wrap gap-3">
              <label className="flex min-w-[8rem] flex-1 flex-col gap-1 text-xs text-gray-600">
                {zh('时间 (mm:ss)', 'Time (mm:ss)')}
                <input
                  type="text"
                  value={timeInput}
                  onChange={(e) => setTimeInput(e.target.value)}
                  placeholder="12:34"
                  className="rounded border border-gray-300 px-2 py-1.5 text-sm text-gray-900"
                />
              </label>
              <label className="flex min-w-[12rem] flex-[2] flex-col gap-1 text-xs text-gray-600">
                {zh('说明', 'Note')}
                <input
                  type="text"
                  value={noteInput}
                  onChange={(e) => setNoteInput(e.target.value)}
                  placeholder={zh('例如：精彩片段', 'e.g. highlight scene')}
                  className="rounded border border-gray-300 px-2 py-1.5 text-sm text-gray-900"
                />
              </label>
            </div>
            <button
              type="submit"
              disabled={saving}
              className="rounded-md bg-sky-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-sky-700 disabled:opacity-60"
            >
              {saving ? zh('保存中…', 'Saving...') : zh('添加', 'Add')}
            </button>
          </form>

          {error ? <p className="text-sm text-red-600">{error}</p> : null}
          {loading ? <p className="text-sm text-gray-500">{zh('加载中…', 'Loading...')}</p> : null}

          <ul className="space-y-2">
            {items.map((marker) => (
              <li
                key={marker.id}
                className="rounded-lg border border-gray-200 bg-white px-3 py-2 shadow-sm"
              >
                {editingId === marker.id ? (
                  <div className="space-y-2">
                    <div className="flex flex-wrap gap-2">
                      <input
                        type="text"
                        value={editTimeInput}
                        onChange={(e) => setEditTimeInput(e.target.value)}
                        className="w-24 rounded border border-gray-300 px-2 py-1 text-sm"
                      />
                      <input
                        type="text"
                        value={editNoteInput}
                        onChange={(e) => setEditNoteInput(e.target.value)}
                        className="min-w-0 flex-1 rounded border border-gray-300 px-2 py-1 text-sm"
                      />
                    </div>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => saveEdit(marker.id)}
                        disabled={saving}
                        className="rounded bg-sky-600 px-2 py-1 text-xs text-white hover:bg-sky-700 disabled:opacity-60"
                      >
                        {zh('保存', 'Save')}
                      </button>
                      <button
                        type="button"
                        onClick={cancelEdit}
                        className="rounded border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                      >
                        {zh('取消', 'Cancel')}
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center gap-2">
                    <span className="shrink-0 rounded bg-gray-100 px-2 py-0.5 font-mono text-xs font-semibold text-gray-800">
                      {formatPlaybackTime(marker.time_sec)}
                    </span>
                    <Tooltip title={marker.note}>
                      <span className="min-w-0 flex-1 truncate text-sm text-gray-800">
                        {marker.note}
                      </span>
                    </Tooltip>
                    <Tooltip title={zh('从此处播放', 'Play from here')}>
                      <IconButton
                        size="small"
                        onClick={() => onPlayAtTime?.(video, marker.time_sec)}
                        aria-label={zh('播放', 'Play')}
                      >
                        <PlayArrowIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={zh('编辑', 'Edit')}>
                      <IconButton
                        size="small"
                        onClick={() => startEdit(marker)}
                        aria-label={zh('编辑', 'Edit')}
                      >
                        <EditOutlinedIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={zh('删除', 'Delete')}>
                      <IconButton
                        size="small"
                        disabled={deletingId === marker.id}
                        onClick={() => handleDelete(marker.id)}
                        aria-label={zh('删除', 'Delete')}
                      >
                        <DeleteOutlineIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </div>
                )}
              </li>
            ))}
            {!loading && items.length === 0 ? (
              <li className="py-6 text-center text-sm text-gray-500">
                {zh('暂无时间点标记', 'No timeline markers yet')}
              </li>
            ) : null}
          </ul>
        </div>
      </div>
    </div>
  )
}
