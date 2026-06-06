import { useEffect, useState } from 'react'
import { Button, TextField } from '@mui/material'
import { zh } from '@/utils/i18n'

export default function CreateCollectionModal({ open, onClose, onCreate }) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    setName('')
    setDescription('')
    setSaving(false)
    setError('')
  }, [open])

  if (!open) return null

  const handleSubmit = async (event) => {
    event.preventDefault()
    const trimmedName = name.trim()
    if (!trimmedName) {
      setError(zh('请输入合集名称', 'Please enter a collection name'))
      return
    }
    setSaving(true)
    setError('')
    try {
      await onCreate?.({
        name: trimmedName,
        description: description.trim(),
      })
      onClose?.()
    } catch (err) {
      setError(err?.message || zh('创建失败', 'Create failed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/30 px-4">
      <div className="w-full max-w-md overflow-hidden rounded-lg bg-white shadow-xl">
        <form onSubmit={handleSubmit}>
          <div className="flex items-center justify-between border-b px-4 py-3">
            <h2 className="text-base font-semibold">{zh('新建合集', 'New collection')}</h2>
            <button
              type="button"
              className="text-gray-500 hover:text-gray-800"
              onClick={onClose}
              disabled={saving}
              aria-label={zh('关闭', 'Close')}
            >
              ✕
            </button>
          </div>
          <div className="space-y-3 px-4 py-4">
            <TextField
              label={zh('名称', 'Name')}
              value={name}
              onChange={(event) => {
                setName(event.target.value)
                setError('')
              }}
              required
              fullWidth
              size="small"
              autoFocus
              disabled={saving}
            />
            <TextField
              label={zh('简介', 'Description')}
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              fullWidth
              size="small"
              multiline
              minRows={3}
              disabled={saving}
              placeholder={zh('可选', 'Optional')}
            />
            {error ? <div className="text-sm text-red-600">{error}</div> : null}
          </div>
          <div className="flex justify-end gap-2 border-t px-4 py-3">
            <Button size="small" variant="outlined" onClick={onClose} disabled={saving}>
              {zh('取消', 'Cancel')}
            </Button>
            <Button size="small" variant="contained" type="submit" disabled={saving}>
              {saving ? zh('创建中…', 'Creating...') : zh('创建', 'Create')}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
