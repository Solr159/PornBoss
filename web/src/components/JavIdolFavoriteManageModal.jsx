import { useEffect, useRef, useState } from 'react'
import CloseRoundedIcon from '@mui/icons-material/CloseRounded'
import DeleteOutlineRoundedIcon from '@mui/icons-material/DeleteOutlineRounded'
import DragIndicatorRoundedIcon from '@mui/icons-material/DragIndicatorRounded'
import EditRoundedIcon from '@mui/icons-material/EditRounded'
import { Button, IconButton } from '@mui/material'
import { zh } from '@/utils/i18n'

export default function JavIdolFavoriteManageModal({
  open,
  groups,
  selectedGroupId,
  loading,
  onClose,
  onReorderGroups,
  onRenameGroup,
  onDeleteGroup,
  onLoadGroupIdols,
  onReorderGroupIdols,
}) {
  const [localGroups, setLocalGroups] = useState([])
  const [editingGroup, setEditingGroup] = useState(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) {
      setLocalGroups([])
      setEditingGroup(null)
      setSaving(false)
      setError('')
      return
    }
    setLocalGroups(normalizeGroups(groups))
  }, [groups, open])

  if (!open) return null

  const saveGroupOrder = async () => {
    setSaving(true)
    setError('')
    try {
      await onReorderGroups?.(localGroups.map((group) => Number(group.id)))
    } catch (err) {
      setError(err.message || zh('保存分组顺序失败', 'Failed to save group order'))
    } finally {
      setSaving(false)
    }
  }

  const handleRename = async (groupId, name) => {
    await onRenameGroup?.(groupId, name)
    setLocalGroups((current) =>
      current.map((group) => (Number(group.id) === Number(groupId) ? { ...group, name } : group))
    )
  }

  const handleDelete = async (groupId) => {
    await onDeleteGroup?.(groupId)
    setLocalGroups((current) => current.filter((group) => Number(group.id) !== Number(groupId)))
    setEditingGroup(null)
  }

  return (
    <>
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/35 px-4">
        <div className="flex max-h-[82vh] w-full max-w-lg flex-col rounded-lg bg-white shadow-xl">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <h2 className="text-base font-semibold text-gray-950">
              {zh('管理女优分组', 'Manage idol groups')}
            </h2>
            <IconButton
              type="button"
              size="small"
              onClick={onClose}
              disabled={saving}
              aria-label={zh('关闭分组管理', 'Close group manager')}
            >
              <CloseRoundedIcon fontSize="small" />
            </IconButton>
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto p-4">
            {error ? (
              <div className="mb-3 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {error}
              </div>
            ) : null}

            <GroupOrderList
              groups={localGroups}
              selectedGroupId={selectedGroupId}
              emptyText={loading ? zh('加载中…', 'Loading...') : zh('暂无分组', 'No groups')}
              onReorder={setLocalGroups}
              onEdit={(group) => setEditingGroup(group)}
            />
          </div>

          <div className="flex justify-end gap-2 border-t px-4 py-3">
            <Button variant="outlined" onClick={onClose} disabled={saving}>
              {zh('关闭', 'Close')}
            </Button>
            <Button
              variant="contained"
              onClick={saveGroupOrder}
              disabled={saving || localGroups.length === 0}
            >
              {saving ? zh('保存中…', 'Saving...') : zh('保存顺序', 'Save order')}
            </Button>
          </div>
        </div>
      </div>

      <FavoriteGroupEditModal
        group={editingGroup}
        onClose={() => setEditingGroup(null)}
        onRename={handleRename}
        onDelete={handleDelete}
        onLoadGroupIdols={onLoadGroupIdols}
        onReorderGroupIdols={onReorderGroupIdols}
      />
    </>
  )
}

function GroupOrderList({ groups, selectedGroupId, emptyText, onReorder, onEdit }) {
  if (!groups.length) {
    return (
      <div className="rounded border border-dashed border-gray-200 px-3 py-8 text-center text-sm text-gray-500">
        {emptyText}
      </div>
    )
  }

  return (
    <SortableList
      items={groups}
      onReorder={onReorder}
      getLabel={(group) => group.name}
      getMeta={(group) => zh(`${group.count || 0} 位`, `${group.count || 0} idols`)}
      isActive={(group) => Number(group.id) === Number(selectedGroupId)}
      renderLeading={(group) => (
        <IconButton
          type="button"
          size="small"
          onClick={() => onEdit(group)}
          aria-label={zh('编辑分组', 'Edit group')}
          sx={{ width: 28, height: 28 }}
        >
          <EditRoundedIcon sx={{ fontSize: 16 }} />
        </IconButton>
      )}
    />
  )
}

function FavoriteGroupEditModal({
  group,
  onClose,
  onRename,
  onDelete,
  onLoadGroupIdols,
  onReorderGroupIdols,
}) {
  const [groupName, setGroupName] = useState('')
  const [idols, setIdols] = useState([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const groupId = Number(group?.id) || null

  useEffect(() => {
    if (!groupId) {
      setGroupName('')
      setIdols([])
      setLoading(false)
      setSaving(false)
      setError('')
      return
    }
    setGroupName(String(group?.name || ''))
    setIdols([])
    setLoading(true)
    setError('')
    let cancelled = false
    onLoadGroupIdols?.(groupId)
      .then((items) => {
        if (!cancelled) setIdols(Array.isArray(items) ? items : [])
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.message || zh('加载收藏夹女优失败', 'Failed to load favorite group idols'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [group, groupId, onLoadGroupIdols])

  if (!groupId) return null

  const saveRename = async () => {
    const name = groupName.trim()
    if (!name) return
    setSaving(true)
    setError('')
    try {
      await onRename?.(groupId, name)
    } catch (err) {
      setError(err.message || zh('重命名分组失败', 'Failed to rename group'))
    } finally {
      setSaving(false)
    }
  }

  const deleteGroup = async () => {
    if (!window.confirm(zh(`删除分组“${group.name}”？`, `Delete group "${group.name}"?`))) {
      return
    }
    setSaving(true)
    setError('')
    try {
      await onDelete?.(groupId)
    } catch (err) {
      setError(err.message || zh('删除分组失败', 'Failed to delete group'))
    } finally {
      setSaving(false)
    }
  }

  const saveIdolOrder = async () => {
    setSaving(true)
    setError('')
    try {
      await onReorderGroupIdols?.(
        groupId,
        idols.map((idol) => Number(idol.id))
      )
    } catch (err) {
      setError(err.message || zh('保存女优顺序失败', 'Failed to save idol order'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 px-4">
      <div className="flex max-h-[86vh] w-full max-w-xl flex-col rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <h2 className="min-w-0 truncate text-base font-semibold text-gray-950">
            {zh('编辑分组', 'Edit group')}
          </h2>
          <IconButton
            type="button"
            size="small"
            onClick={onClose}
            disabled={saving}
            aria-label={zh('关闭编辑分组', 'Close group editor')}
          >
            <CloseRoundedIcon fontSize="small" />
          </IconButton>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          {error ? (
            <div className="mb-3 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          ) : null}

          <div className="mb-4 flex flex-wrap items-center gap-2">
            <input
              value={groupName}
              onChange={(event) => setGroupName(event.target.value)}
              className="h-8 min-w-0 flex-1 rounded border border-gray-200 px-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              disabled={saving}
            />
            <Button
              variant="outlined"
              size="small"
              onClick={saveRename}
              disabled={saving || !groupName.trim()}
            >
              {zh('重命名', 'Rename')}
            </Button>
            <IconButton
              type="button"
              size="small"
              onClick={deleteGroup}
              disabled={saving}
              aria-label={zh('删除分组', 'Delete group')}
            >
              <DeleteOutlineRoundedIcon fontSize="small" />
            </IconButton>
          </div>

          <div className="mb-2 text-xs text-gray-500">
            {zh('按住拖动按钮，整行会跟随鼠标移动并挤压其它项', 'Drag the handle to reorder idols')}
          </div>
          <IdolOrderList idols={idols} loading={loading} onReorder={setIdols} />
        </div>

        <div className="flex justify-end gap-2 border-t px-4 py-3">
          <Button variant="outlined" onClick={onClose} disabled={saving}>
            {zh('关闭', 'Close')}
          </Button>
          <Button
            variant="contained"
            onClick={saveIdolOrder}
            disabled={saving || loading || idols.length === 0}
          >
            {saving ? zh('保存中…', 'Saving...') : zh('保存女优顺序', 'Save idol order')}
          </Button>
        </div>
      </div>
    </div>
  )
}

function IdolOrderList({ idols, loading, onReorder }) {
  if (!idols.length) {
    return (
      <div className="rounded border border-dashed border-gray-200 px-3 py-8 text-center text-sm text-gray-500">
        {loading ? zh('加载中…', 'Loading...') : zh('该分组暂无女优', 'No idols in this group')}
      </div>
    )
  }
  return (
    <SortableList
      items={idols}
      onReorder={onReorder}
      getLabel={(idol) => idol.name || zh('未知女优', 'Unknown idol')}
      getMeta={(idol) =>
        zh(`${Number(idol?.work_count) || 0} 部`, `${Number(idol?.work_count) || 0} works`)
      }
    />
  )
}

function SortableList({
  items,
  onReorder,
  getLabel,
  getMeta,
  isActive = () => false,
  renderLeading = null,
}) {
  const containerRef = useRef(null)
  const rowRefs = useRef(new Map())
  const [drag, setDrag] = useState(null)

  useEffect(() => {
    if (!drag) return undefined

    const handlePointerMove = (event) => {
      event.preventDefault()
      setDrag((current) =>
        current
          ? {
              ...current,
              pointerX: event.clientX,
              pointerY: event.clientY,
            }
          : current
      )

      const nextIndex = calculateDropIndex(items, rowRefs.current, drag.id, event.clientY)
      const currentIndex = items.findIndex((item) => String(item.id) === drag.id)
      if (nextIndex >= 0 && currentIndex >= 0 && nextIndex !== currentIndex) {
        onReorder?.(moveItemToIndex(items, drag.id, nextIndex))
      }
    }

    const handlePointerUp = () => {
      setDrag(null)
    }

    window.addEventListener('pointermove', handlePointerMove, { passive: false })
    window.addEventListener('pointerup', handlePointerUp)
    window.addEventListener('pointercancel', handlePointerUp)
    return () => {
      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('pointerup', handlePointerUp)
      window.removeEventListener('pointercancel', handlePointerUp)
    }
  }, [drag, items, onReorder])

  const startDrag = (event, item) => {
    if (event.button !== 0) return
    const id = String(item.id)
    const row = rowRefs.current.get(id)
    if (!row) return
    const rect = row.getBoundingClientRect()
    event.preventDefault()
    setDrag({
      id,
      pointerX: event.clientX,
      pointerY: event.clientY,
      offsetX: event.clientX - rect.left,
      offsetY: event.clientY - rect.top,
      left: rect.left,
      top: rect.top,
      width: rect.width,
      height: rect.height,
    })
  }

  return (
    <div
      ref={containerRef}
      className={`rounded border border-gray-200 p-1 ${drag ? 'select-none' : ''}`}
    >
      {items.map((item) => {
        const id = String(item.id)
        const isDragging = drag?.id === id
        return (
          <SortableRow
            key={id}
            refCallback={(node) => {
              if (node) rowRefs.current.set(id, node)
              else rowRefs.current.delete(id)
            }}
            item={item}
            label={getLabel(item)}
            meta={getMeta?.(item)}
            leading={renderLeading?.(item)}
            active={isActive(item)}
            dragging={isDragging}
            onHandlePointerDown={(event) => startDrag(event, item)}
          />
        )
      })}
      {drag ? (
        <div
          className="pointer-events-none fixed z-[80]"
          style={{
            left: drag.pointerX - drag.offsetX,
            top: drag.pointerY - drag.offsetY,
            width: drag.width,
            height: drag.height,
          }}
        >
          {(() => {
            const item = items.find((candidate) => String(candidate.id) === drag.id)
            if (!item) return null
            return (
              <SortableRow
                item={item}
                label={getLabel(item)}
                meta={getMeta?.(item)}
                leading={renderLeading?.(item)}
                active={isActive(item)}
                dragging={false}
                floating
              />
            )
          })()}
        </div>
      ) : null}
    </div>
  )
}

function SortableRow({
  item,
  label,
  meta,
  leading = null,
  active = false,
  dragging = false,
  floating = false,
  refCallback,
  onHandlePointerDown,
}) {
  return (
    <div
      ref={refCallback}
      className={`mb-1 flex items-center gap-2 rounded border px-2 py-1.5 last:mb-0 ${
        floating ? 'shadow-lg ring-1 ring-blue-200' : 'transition-[background-color,opacity]'
      } ${dragging ? 'opacity-0' : 'opacity-100'} ${
        active ? 'border-blue-200 bg-blue-50' : 'border-transparent bg-gray-50'
      }`}
      data-sortable-id={item.id}
    >
      {leading}
      <span className="min-w-0 flex-1 truncate text-sm text-gray-900">{label}</span>
      <span className="shrink-0 text-xs text-gray-500">{meta}</span>
      <button
        type="button"
        onPointerDown={onHandlePointerDown}
        className="inline-flex h-7 w-7 shrink-0 cursor-grab touch-none items-center justify-center rounded border border-gray-200 bg-white text-gray-500 active:cursor-grabbing"
        aria-label={zh('拖动排序', 'Drag to reorder')}
        title={zh('拖动排序', 'Drag to reorder')}
      >
        <DragIndicatorRoundedIcon sx={{ fontSize: 17 }} />
      </button>
    </div>
  )
}

function calculateDropIndex(items, refs, draggingId, pointerY) {
  let nextIndex = 0
  for (const item of items) {
    const id = String(item.id)
    if (id === draggingId) continue
    const row = refs.get(id)
    if (!row) continue
    const rect = row.getBoundingClientRect()
    if (pointerY > rect.top + rect.height / 2) {
      nextIndex += 1
    }
  }
  return Math.min(nextIndex, items.length - 1)
}

function moveItemToIndex(items, sourceId, targetIndex) {
  const next = [...items]
  const from = next.findIndex((item) => String(item.id) === String(sourceId))
  if (from < 0) return next
  const [item] = next.splice(from, 1)
  const safeIndex = Math.max(0, Math.min(targetIndex, next.length))
  next.splice(safeIndex, 0, item)
  return next
}

function normalizeGroups(groups) {
  return [...(Array.isArray(groups) ? groups : [])].sort((a, b) => {
    const orderA = Number(a?.sort_order) || 0
    const orderB = Number(b?.sort_order) || 0
    if (orderA !== orderB) return orderA - orderB
    return String(a?.name || '').localeCompare(String(b?.name || ''))
  })
}
