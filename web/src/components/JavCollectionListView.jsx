import { Button } from '@mui/material'
import { zh } from '@/utils/i18n'

export default function JavCollectionListView({
  items,
  loading,
  error,
  onOpenCollection,
  onCreateClick,
  onRefresh,
}) {
  const list = Array.isArray(items) ? items : []

  return (
    <div className="mx-auto max-w-screen-2xl px-6 py-4">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-lg font-semibold text-gray-900">{zh('JAV 合集', 'JAV Collections')}</h1>
        <div className="flex gap-2">
          <Button variant="outlined" size="small" onClick={onRefresh} disabled={loading}>
            {zh('刷新', 'Refresh')}
          </Button>
          <Button variant="contained" size="small" onClick={onCreateClick}>
            {zh('新建合集', 'New collection')}
          </Button>
        </div>
      </div>
      {error ? (
        <div className="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      ) : null}
      {loading ? (
        <div className="text-sm text-gray-500">{zh('加载中…', 'Loading...')}</div>
      ) : list.length === 0 ? (
        <div className="text-sm text-gray-500">{zh('暂无合集', 'No collections yet')}</div>
      ) : (
        <ul className="divide-y rounded-lg border border-gray-200 bg-white">
          {list.map((row) => (
            <li key={row.id}>
              <button
                type="button"
                className="flex w-full items-start gap-3 px-4 py-3 text-left hover:bg-gray-50"
                onClick={() => onOpenCollection?.(row)}
              >
                <div className="min-w-0 flex-1">
                  <div className="font-medium text-gray-900">{row.name}</div>
                  {row.description ? (
                    <div className="mt-0.5 line-clamp-2 text-xs text-gray-600">
                      {row.description}
                    </div>
                  ) : null}
                </div>
                <div className="shrink-0 text-xs text-gray-500">
                  {zh(`${row.count || 0} 部`, `${row.count || 0} items`)}
                </div>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
