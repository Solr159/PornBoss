import { useEffect, useState } from 'react'
import { Button } from '@mui/material'
import { zh } from '@/utils/i18n'

function membershipForCollection(collectionId, javTargetIds, javItems) {
  const ids = (javTargetIds || [])
    .map((id) => Number(id))
    .filter((id) => Number.isFinite(id) && id > 0)
  if (!ids.length) return 'unknown'
  const idSet = new Set(ids)
  const rows = (javItems || []).filter((j) => idSet.has(Number(j?.id)))
  if (rows.length === 0) return 'unknown'
  let inCount = 0
  for (const row of rows) {
    const cols = Array.isArray(row?.collections) ? row.collections : []
    if (cols.some((x) => Number(x?.id) === Number(collectionId))) inCount++
  }
  if (inCount === rows.length) return 'all'
  if (inCount > 0) return 'partial'
  return 'none'
}

export default function CollectionPickerModal({
  open,
  onClose,
  collections,
  loading,
  onPick,
  onCreateNew,
  javTargetIds = [],
  javItems = [],
}) {
  const [q, setQ] = useState('')
  useEffect(() => {
    if (open) setQ('')
  }, [open])

  const rows = Array.isArray(collections) ? collections : []
  const filtered = q.trim()
    ? rows.filter((c) =>
        String(c.name || '')
          .toLowerCase()
          .includes(q.trim().toLowerCase())
      )
    : rows

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4">
      <div className="max-h-[85vh] w-full max-w-md overflow-hidden rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <h2 className="text-base font-semibold">{zh('选择合集', 'Pick collection')}</h2>
          <button
            type="button"
            className="text-gray-500 hover:text-gray-800"
            onClick={onClose}
            aria-label={zh('关闭', 'Close')}
          >
            ✕
          </button>
        </div>
        <div className="border-b px-4 py-2">
          <input
            className="w-full rounded border px-2 py-1 text-sm"
            placeholder={zh('筛选名称', 'Filter by name')}
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </div>
        <div className="max-h-64 overflow-y-auto">
          {loading ? (
            <div className="p-4 text-sm text-gray-500">{zh('加载中…', 'Loading...')}</div>
          ) : filtered.length === 0 ? (
            <div className="p-4 text-sm text-gray-500">{zh('无合集', 'No collections')}</div>
          ) : (
            <ul>
              {filtered.map((c) => {
                const mem = membershipForCollection(c.id, javTargetIds, javItems)
                const showBadge = javTargetIds?.length > 0 && (mem === 'all' || mem === 'partial')
                return (
                  <li key={c.id}>
                    <button
                      type="button"
                      className={`flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-50 ${
                        mem === 'all'
                          ? 'bg-emerald-50/80'
                          : mem === 'partial'
                            ? 'bg-amber-50/70'
                            : ''
                      }`}
                      onClick={() => onPick?.(c)}
                    >
                      <span className="min-w-0">
                        <span className="font-medium">{c.name}</span>
                        <span className="ml-2 text-xs text-gray-500">({c.count || 0})</span>
                      </span>
                      {showBadge ? (
                        <span
                          className={`shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold ${
                            mem === 'all'
                              ? 'bg-emerald-200 text-emerald-950'
                              : mem === 'partial'
                                ? 'bg-amber-200 text-amber-950'
                                : 'bg-gray-100 text-gray-600'
                          }`}
                        >
                          {mem === 'all'
                            ? zh('已在', 'In')
                            : mem === 'partial'
                              ? zh('部分在', 'Some')
                              : ''}
                        </span>
                      ) : null}
                    </button>
                  </li>
                )
              })}
            </ul>
          )}
        </div>
        <div className="flex justify-end gap-2 border-t px-4 py-3">
          <Button size="small" onClick={onCreateNew}>
            {zh('新建…', 'New…')}
          </Button>
          <Button size="small" variant="outlined" onClick={onClose}>
            {zh('取消', 'Cancel')}
          </Button>
        </div>
      </div>
    </div>
  )
}
