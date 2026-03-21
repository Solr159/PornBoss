import { Button } from '@mui/material'

export default function SelectionOpsModal({
  open,
  onClose,
  selectedList,
  selectedCount,
  onOpenTags,
}) {
  if (!open) return null

  const list = Array.isArray(selectedList) ? selectedList : []
  const count = Number.isFinite(selectedCount) ? selectedCount : 0

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4">
      <div className="w-full max-w-lg rounded-lg bg-white p-4 shadow-xl">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-base font-semibold">已选择文件</h2>
          <button
            onClick={onClose}
            className="rounded px-2 py-1 text-gray-500 hover:bg-gray-100"
            aria-label="关闭"
          >
            ✕
          </button>
        </div>
        <div className="max-h-64 overflow-y-auto rounded border bg-gray-50 p-2 text-sm">
          {list.length === 0 ? (
            <div className="text-gray-500">暂无选择</div>
          ) : (
            <ul className="space-y-1">
              {list.map((item) => (
                <li key={item.id} className="truncate text-gray-800">
                  {item.label}
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="outlined" size="small" onClick={onOpenTags} disabled={count === 0}>
            添加标签
          </Button>
          <Button variant="contained" size="small" onClick={onClose} color="primary">
            关闭
          </Button>
        </div>
      </div>
    </div>
  )
}
