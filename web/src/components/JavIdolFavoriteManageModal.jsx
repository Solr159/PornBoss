import { useEffect, useMemo, useState } from 'react'
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
  const [mode, setMode] = useState('groups')
  const [localGroups, setLocalGroups] = useState([])
  const [activeGroupId, setActiveGroupId] = useState(null)
  const [groupName, setGroupName] = useState('')
  const [groupIdols, setGroupIdols] = useState([])
  const [idolsLoading, setIdolsLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) {
      setMode('groups')
      setLocalGroups([])
      setActiveGroupId(null)
      setGroupName('')
      setGroupIdols([])
      setIdolsLoading(false)
      setSaving(false)
      setError('')
      return
    }
    const nextGroups = normalizeGroups(groups)
    setLocalGroups(nextGroups)
    const selected = Number(selectedGroupId)
    const initial = nextGroups.some((group) => Number(group.id) === selected)
      ? selected
      : Number(nextGroups[0]?.id) || null
    setActiveGroupId(initial)
    setGroupName(String(nextGroups.find((group) => Number(group.id) === initial)?.name || ''))
  }, [groups, open, selectedGroupId])

  useEffect(() => {
    if (!open || mode !== 'edit' || !activeGroupId) return undefined
    let cancelled = false
    setIdolsLoading(true)
    setError('')
    onLoadGroupIdols?.(activeGroupId)
      .then((items) => {
        if (!cancelled) setGroupIdols(Array.isArray(items) ? items : [])
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.message || zh('加载收藏夹女优失败', 'Failed to load favorite group idols'))
        }
      })
      .finally(() => {
        if (!cancelled) setIdolsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [activeGroupId, mode, onLoadGroupIdols, open])

  const activeGroup = useMemo(
    () => localGroups.find((group) => Number(group.id) === Number(activeGroupId)) || null,
    [activeGroupId, localGroups]
  )

  useEffect(() => {
    setGroupName(String(activeGroup?.name || ''))
  }, [activeGroup])

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

  const saveRename = async () => {
    const name = groupName.trim()
    if (!activeGroupId || !name) return
    setSaving(true)
    setError('')
    try {
      await onRenameGroup?.(activeGroupId, name)
      setLocalGroups((current) =>
        current.map((group) =>
          Number(group.id) === Number(activeGroupId) ? { ...group, name } : group
        )
      )
    } catch (err) {
      setError(err.message || zh('重命名分组失败', 'Failed to rename group'))
    } finally {
      setSaving(false)
    }
  }

  const deleteGroup = async () => {
    if (!activeGroupId || !activeGroup) return
    if (
      !window.confirm(zh(`删除分组“${activeGroup.name}”？`, `Delete group "${activeGroup.name}"?`))
    ) {
      return
    }
    setSaving(true)
    setError('')
    try {
      await onDeleteGroup?.(activeGroupId)
      const next = localGroups.filter((group) => Number(group.id) !== Number(activeGroupId))
      setLocalGroups(next)
      setActiveGroupId(Number(next[0]?.id) || null)
      setGroupIdols([])
    } catch (err) {
      setError(err.message || zh('删除分组失败', 'Failed to delete group'))
    } finally {
      setSaving(false)
    }
  }

  const saveIdolOrder = async () => {
    if (!activeGroupId) return
    setSaving(true)
    setError('')
    try {
      await onReorderGroupIdols?.(
        activeGroupId,
        groupIdols.map((idol) => Number(idol.id))
      )
    } catch (err) {
      setError(err.message || zh('保存女优顺序失败', 'Failed to save idol order'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/35 px-4">
      <div className="flex max-h-[86vh] w-full max-w-2xl flex-col rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-gray-950">
              {zh('管理女优分组', 'Manage idol groups')}
            </h2>
          </div>
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

        <div className="flex border-b px-4 pt-3">
          <ModeButton active={mode === 'groups'} onClick={() => setMode('groups')}>
            {zh('分组顺序', 'Group order')}
          </ModeButton>
          <ModeButton active={mode === 'edit'} onClick={() => setMode('edit')}>
            {zh('编辑分组', 'Edit group')}
          </ModeButton>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          {error ? (
            <div className="mb-3 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          ) : null}

          {mode === 'groups' ? (
            <div className="grid gap-3">
              <ReorderList
                items={localGroups}
                emptyText={loading ? zh('加载中…', 'Loading...') : zh('暂无分组', 'No groups')}
                getLabel={(group) => group.name}
                getMeta={(group) => zh(`${group.count || 0} 位`, `${group.count || 0} idols`)}
                onReorder={setLocalGroups}
              />
              <div className="flex justify-end">
                <Button
                  variant="contained"
                  onClick={saveGroupOrder}
                  disabled={saving || localGroups.length === 0}
                >
                  {saving ? zh('保存中…', 'Saving...') : zh('保存顺序', 'Save order')}
                </Button>
              </div>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-[14rem_1fr]">
              <div className="min-w-0 rounded border border-gray-200 p-1">
                {localGroups.length === 0 ? (
                  <div className="px-3 py-8 text-center text-sm text-gray-500">
                    {zh('暂无分组', 'No groups')}
                  </div>
                ) : (
                  localGroups.map((group) => {
                    const active = Number(group.id) === Number(activeGroupId)
                    return (
                      <button
                        key={group.id}
                        type="button"
                        onClick={() => setActiveGroupId(Number(group.id))}
                        className={`flex w-full items-center gap-2 rounded px-2 py-1.5 text-left text-sm ${
                          active ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-50'
                        }`}
                      >
                        <span className="min-w-0 flex-1 truncate">{group.name}</span>
                        <span className="shrink-0 text-xs text-gray-400">{group.count || 0}</span>
                      </button>
                    )
                  })
                )}
              </div>

              <div className="min-w-0">
                {activeGroup ? (
                  <div className="grid gap-4">
                    <div className="flex flex-wrap items-center gap-2">
                      <EditRoundedIcon className="text-gray-400" fontSize="small" />
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

                    <div className="grid gap-3">
                      <div className="text-xs text-gray-500">
                        {zh('拖动女优调整组内顺序', 'Drag idols to adjust group order')}
                      </div>
                      <ReorderList
                        items={groupIdols}
                        emptyText={
                          idolsLoading
                            ? zh('加载中…', 'Loading...')
                            : zh('该分组暂无女优', 'No idols in this group')
                        }
                        getLabel={(idol) => idol.name || zh('未知女优', 'Unknown idol')}
                        getMeta={(idol) => {
                          const count = Number(idol?.work_count) || 0
                          return zh(`${count} 部`, `${count} works`)
                        }}
                        onReorder={setGroupIdols}
                      />
                      <div className="flex justify-end">
                        <Button
                          variant="contained"
                          onClick={saveIdolOrder}
                          disabled={saving || idolsLoading || groupIdols.length === 0}
                        >
                          {saving
                            ? zh('保存中…', 'Saving...')
                            : zh('保存女优顺序', 'Save idol order')}
                        </Button>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="rounded border border-dashed border-gray-200 px-3 py-10 text-center text-sm text-gray-500">
                    {zh('请选择分组', 'Select a group')}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function ModeButton({ active, onClick, children }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`border-b-2 px-3 pb-2 text-sm ${
        active ? 'border-blue-600 text-blue-700' : 'border-transparent text-gray-500'
      }`}
    >
      {children}
    </button>
  )
}

function ReorderList({ items, emptyText, getLabel, getMeta, onReorder }) {
  const [dragId, setDragId] = useState(null)
  const list = Array.isArray(items) ? items : []

  if (list.length === 0) {
    return (
      <div className="rounded border border-dashed border-gray-200 px-3 py-8 text-center text-sm text-gray-500">
        {emptyText}
      </div>
    )
  }

  const moveTo = (sourceId, targetId) => {
    if (!sourceId || !targetId || sourceId === targetId) return
    onReorder?.(moveItem(list, sourceId, targetId))
  }

  return (
    <div className="rounded border border-gray-200 p-1">
      {list.map((item) => {
        const id = String(item.id)
        return (
          <div
            key={id}
            draggable
            onDragStart={(event) => {
              setDragId(id)
              event.dataTransfer.effectAllowed = 'move'
              event.dataTransfer.setData('text/plain', id)
            }}
            onDragOver={(event) => {
              event.preventDefault()
              event.dataTransfer.dropEffect = 'move'
            }}
            onDrop={(event) => {
              event.preventDefault()
              moveTo(event.dataTransfer.getData('text/plain') || dragId, id)
              setDragId(null)
            }}
            onDragEnd={() => setDragId(null)}
            className={`mb-1 flex cursor-move items-center gap-2 rounded border px-2 py-1.5 last:mb-0 ${
              dragId === id ? 'border-blue-300 bg-blue-50' : 'border-transparent bg-gray-50'
            }`}
          >
            <DragIndicatorRoundedIcon className="shrink-0 text-gray-400" fontSize="small" />
            <span className="min-w-0 flex-1 truncate text-sm text-gray-900">{getLabel(item)}</span>
            <span className="shrink-0 text-xs text-gray-500">{getMeta?.(item)}</span>
          </div>
        )
      })}
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
