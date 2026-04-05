import { useEffect, useRef, useState } from 'react'

import {
  DEFAULT_PLAYER_HOTKEYS,
  PLAYER_HOTKEY_ACTIONS,
  formatPlayerHotkeyKey,
  isReservedPlayerHotkeyKey,
  isSupportedPlayerHotkeyKey,
  normalizePlayerHotkey,
  normalizePlayerHotkeyKey,
} from '@/utils/playerHotkeys'

const createRows = (items) =>
  (Array.isArray(items) && items.length ? items : DEFAULT_PLAYER_HOTKEYS).map((item, index) => ({
    id: `${Date.now()}-${index}-${item.key}`,
    key: item.key,
    action: item.action,
    amount: String(item.amount),
  }))

export default function PlayerSettingsModal({ open, onClose, hotkeys, onSave }) {
  const seedRef = useRef(0)
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    seedRef.current += 1
    const source = Array.isArray(hotkeys) ? hotkeys : DEFAULT_PLAYER_HOTKEYS
    setRows(createRows(source))
    setError('')
    setSaving(false)
  }, [open, hotkeys])

  if (!open) return null

  const setRowValue = (id, patch) => {
    setRows((current) => current.map((row) => (row.id === id ? { ...row, ...patch } : row)))
  }

  const handleAdd = () => {
    seedRef.current += 1
    setRows((current) => [
      ...current,
      {
        id: `${seedRef.current}-new`,
        key: '',
        action: PLAYER_HOTKEY_ACTIONS.SEEK,
        amount: '5',
      },
    ])
    setError('')
  }

  const handleReset = () => {
    setRows(createRows(DEFAULT_PLAYER_HOTKEYS))
    setError('')
  }

  const handleSave = async () => {
    setError('')
    const seen = new Set()
    const normalized = []
    for (const row of rows) {
      const key = normalizePlayerHotkeyKey(row.key)
      if (!key) {
        setError('请为每个快捷键设置按键')
        return
      }
      if (!isSupportedPlayerHotkeyKey(key)) {
        setError(`按键 ${formatPlayerHotkeyKey(key)} 不支持`)
        return
      }
      if (isReservedPlayerHotkeyKey(key)) {
        setError(`按键 ${formatPlayerHotkeyKey(key)} 已保留给播放/关闭`)
        return
      }
      if (seen.has(key)) {
        setError(`按键 ${formatPlayerHotkeyKey(key)} 重复了`)
        return
      }
      seen.add(key)
      const item = normalizePlayerHotkey({
        key,
        action: row.action,
        amount: row.amount,
      })
      if (!item) {
        setError(`请检查按键 ${formatPlayerHotkeyKey(key)} 的动作和数值`)
        return
      }
      normalized.push(item)
    }

    setSaving(true)
    try {
      await onSave?.(normalized)
      onClose?.()
    } catch (err) {
      setError(err.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 px-4">
      <div className="w-full max-w-4xl rounded-lg bg-white p-5 shadow-xl">
        <div className="flex items-center justify-between gap-3">
          <div>
            <h2 className="text-base font-semibold">播放器设置</h2>
            <p className="mt-1 text-sm text-gray-500">仅支持视频进度和音量快捷键</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded px-3 py-1 text-sm text-gray-600 hover:bg-gray-100"
          >
            关闭
          </button>
        </div>

        <div className="mt-4 space-y-3">
          <div className="grid grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)_minmax(0,1fr)_auto] gap-3 px-2 text-xs font-medium uppercase tracking-wide text-gray-500">
            <div>按键</div>
            <div>动作</div>
            <div>变化值</div>
            <div>操作</div>
          </div>

          {rows.length === 0 ? (
            <div className="rounded border border-dashed px-4 py-6 text-center text-sm text-gray-500">
              当前没有自定义快捷键
            </div>
          ) : (
            rows.map((row) => (
              <div
                key={row.id}
                className="grid grid-cols-1 gap-3 rounded border p-3 md:grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)_minmax(0,1fr)_auto]"
              >
                <input
                  value={formatPlayerHotkeyKey(row.key)}
                  readOnly
                  onKeyDown={(event) => {
                    event.preventDefault()
                    event.stopPropagation()
                    const nextKey = normalizePlayerHotkeyKey(event.key)
                    if (!nextKey) return
                    setRowValue(row.id, { key: nextKey })
                    setError('')
                  }}
                  onFocus={() => setError('')}
                  placeholder="聚焦后按下按键"
                  className="rounded border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
                <select
                  value={row.action}
                  onChange={(event) => {
                    const nextAction = event.target.value
                    setRowValue(row.id, {
                      action: nextAction,
                      amount: nextAction === PLAYER_HOTKEY_ACTIONS.VOLUME ? '10' : '5',
                    })
                    setError('')
                  }}
                  className="rounded border px-3 py-2 text-sm"
                >
                  <option value={PLAYER_HOTKEY_ACTIONS.SEEK}>调节进度</option>
                  <option value={PLAYER_HOTKEY_ACTIONS.VOLUME}>调节音量</option>
                </select>
                <div className="flex items-center gap-2">
                  <input
                    type="number"
                    step="1"
                    value={row.amount}
                    onChange={(event) => {
                      setRowValue(row.id, { amount: event.target.value })
                      setError('')
                    }}
                    className="w-full rounded border px-3 py-2 text-sm"
                  />
                  <span className="shrink-0 text-sm text-gray-500">
                    {row.action === PLAYER_HOTKEY_ACTIONS.VOLUME ? '%' : '秒'}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() => {
                    setRows((current) => current.filter((item) => item.id !== row.id))
                    setError('')
                  }}
                  className="rounded border px-3 py-2 text-sm text-red-600 hover:bg-red-50"
                >
                  删除
                </button>
              </div>
            ))
          )}
        </div>

        <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={handleAdd}
              className="rounded border px-3 py-1.5 text-sm hover:bg-gray-50"
            >
              新增快捷键
            </button>
            <button
              type="button"
              onClick={handleReset}
              className="rounded border px-3 py-1.5 text-sm hover:bg-gray-50"
            >
              恢复默认
            </button>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded border px-3 py-1.5 text-sm hover:bg-gray-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={saving}
              className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-60"
            >
              {saving ? '保存中…' : '保存'}
            </button>
          </div>
        </div>

        <div className="mt-2 text-xs text-gray-500">
          正数表示增加，负数表示减少。`Space` 和 `Escape` 仍固定用于播放/暂停和关闭播放器。
        </div>
        {error && <div className="mt-2 text-sm text-red-600">{error}</div>}
      </div>
    </div>
  )
}
