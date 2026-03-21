import { useState } from 'react'

export default function BulkTagActions({ tags, disabled, onAdd, onRemove, onClear }) {
  const [tagId, setTagId] = useState('')
  const parsed = Number(tagId) || 0

  return (
    <div className="flex items-center gap-2">
      <select
        value={tagId}
        onChange={(e) => setTagId(e.target.value)}
        className="rounded border px-2 py-1"
        aria-label="选择标签"
      >
        <option value="">选择标签…</option>
        {tags.map((t) => (
          <option key={t.id} value={t.id}>
            {t.name}
          </option>
        ))}
      </select>
      <button
        disabled={disabled || !parsed}
        onClick={() => parsed && onAdd(parsed)}
        className="rounded bg-emerald-600 px-2 py-1 text-white disabled:opacity-50"
      >
        为选中添加
      </button>
      <button
        disabled={disabled || !parsed}
        onClick={() => parsed && onRemove(parsed)}
        className="rounded bg-amber-600 px-2 py-1 text-white disabled:opacity-50"
      >
        从选中移除
      </button>
      <button
        disabled={disabled}
        onClick={onClear}
        className="rounded border px-2 py-1 disabled:opacity-50"
      >
        清除选择
      </button>
    </div>
  )
}
