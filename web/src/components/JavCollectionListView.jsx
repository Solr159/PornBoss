import { Button } from '@mui/material'

import { zh } from '@/utils/i18n'

export default function JavCollectionListView({
  items,
  loading,
  error,
  onOpenCollection,
  onCreateClick,
  onRefresh,
  buildCollectionUrl,
}) {
  const list = Array.isArray(items) ? items : []

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
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
        <div className="flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500">
          {zh('加载中…', 'Loading...')}
        </div>
      ) : list.length === 0 ? (
        <div className="flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500">
          {zh('暂无合集', 'No collections yet')}
        </div>
      ) : (
        <div className="grid gap-4 bg-white sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
          {list.map((item) => (
            <CollectionCard
              key={item.id || item.name}
              item={item}
              href={buildCollectionUrl?.(item)}
              onOpenCollection={onOpenCollection}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function CollectionCard({ item, href, onOpenCollection }) {
  const cover = item?.sample_code ? `/jav/${encodeURIComponent(item.sample_code)}/cover` : null
  const name = item?.name || zh('未命名合集', 'Untitled collection')
  const workCount = item?.count || 0
  const description = String(item?.description || '').trim()

  const handleClick = (e) => {
    const selection = window.getSelection?.()
    if (selection && String(selection).trim() !== '') {
      e.preventDefault()
      return
    }
    const isModified = e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0
    if (isModified) {
      return
    }
    e.preventDefault()
    onOpenCollection?.(item)
  }

  return (
    <a
      href={href || '#'}
      className="group flex cursor-pointer flex-col overflow-hidden rounded-lg border bg-white shadow-sm transition hover:shadow-lg"
      onClick={handleClick}
      onKeyDown={(e) => {
        if (e.key === ' ') {
          e.preventDefault()
          onOpenCollection?.(item)
        }
      }}
    >
      <div className="relative aspect-[800/538] w-full overflow-hidden bg-gray-100">
        {cover ? (
          <img
            src={cover}
            alt={name}
            className="h-full w-full object-cover transition duration-200 group-hover:scale-[1.03]"
            loading="lazy"
          />
        ) : (
          <div className="absolute inset-0 flex h-full w-full items-center justify-center bg-gradient-to-br from-indigo-50 to-violet-100 p-4 text-center text-lg font-semibold text-indigo-900">
            {name}
          </div>
        )}
      </div>
      <div className="flex flex-1 flex-col gap-1 p-3">
        <div className="line-clamp-2 text-sm font-semibold leading-tight">{name}</div>
        {description ? (
          <div className="line-clamp-2 text-xs text-gray-600">{description}</div>
        ) : null}
        <div className="text-xs text-gray-500">
          {zh(`${workCount} 部作品`, `${workCount} works`)}
        </div>
      </div>
    </a>
  )
}
