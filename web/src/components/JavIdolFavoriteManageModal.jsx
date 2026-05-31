import { useEffect, useState } from 'react'
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
    <div className="rounded border border-gray-200 p-1">
      {groups.map((group) => (
        <ReorderRow
          key={group.id}
          item={group}
          items={groups}
          active={Number(group.id) === Number(selectedGroupId)}
          label={group.name}
          meta={zh(`${group.count || 0} 位`, `${group.count || 0} idols`)}
          onReorder={onReorder}
          leading={
            <IconButton
              type="button"
              size="small"
              onClick={() => onEdit(group)}
              aria-label={zh('编辑分组', 'Edit group')}
              sx={{ width: 28, height: 28 }}
            >
              <EditRoundedIcon sx={{ fontSize: 16 }} />
            </IconButton>
          }
        />
      ))}
    </div>
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
            {zh('点击拖动按钮后上下拖动，调整组内女优顺序', 'Drag the handle to reorder idols')}
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
    <div className="rounded border border-gray-200 p-1">
      {idols.map((idol) => (
        <ReorderRow
          key={idol.id}
          item={idol}
          items={idols}
          label={idol.name || zh('未知女优', 'Unknown idol')}
          meta={zh(`${Number(idol?.work_count) || 0} 部`, `${Number(idol?.work_count) || 0} works`)}
          onReorder={onReorder}
        />
      ))}
    </div>
  )
}

function ReorderRow({ item, items, active = false, label, meta, leading = null, onReorder }) {
  const [dragId, setDragId] = useState(null)
  const id = String(item.id)

  const moveTo = (sourceId, targetId) => {
    if (!sourceId || !targetId || sourceId === targetId) return
    onReorder?.(moveItem(items, sourceId, targetId))
  }

  return (
    <div
      onDragOver={(event) => {
        event.preventDefault()
        event.dataTransfer.dropEffect = 'move'
      }}
      onDrop={(event) => {
        event.preventDefault()
        moveTo(event.dataTransfer.getData('text/plain') || dragId, id)
        setDragId(null)
      }}
      className={`mb-1 flex items-center gap-2 rounded border px-2 py-1.5 last:mb-0 ${
        active ? 'border-blue-200 bg-blue-50' : 'border-transparent bg-gray-50'
      }`}
    >
      {leading}
      <span className="min-w-0 flex-1 truncate text-sm text-gray-900">{label}</span>
      <span className="shrink-0 text-xs text-gray-500">{meta}</span>
      <button
        type="button"
        draggable
        onDragStart={(event) => {
          setDragId(id)
          event.dataTransfer.effectAllowed = 'move'
          event.dataTransfer.setData('text/plain', id)
        }}
        onDragEnd={() => setDragId(null)}
        className={`inline-flex h-7 w-7 shrink-0 cursor-move items-center justify-center rounded border ${
          dragId === id ? 'border-blue-300 bg-blue-100' : 'border-gray-200 bg-white text-gray-500'
        }`}
        aria-label={zh('拖动排序', 'Drag to reorder')}
        title={zh('拖动排序', 'Drag to reorder')}
      >
        <DragIndicatorRoundedIcon sx={{ fontSize: 17 }} />
      </button>
    </div>
  )
}

function moveItem(items, sourceId, targetId) {
  const next = [...items]
  const from = next.findIndex((item) => String(item.id) === String(sourceId))
  const to = next.findIndex((item) => String(item.id) === String(targetId))
  if (from < 0 || to < 0 || from === to) return next
  const [item] = next.splice(from, 1)
  next.splice(to, 0, item)
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
