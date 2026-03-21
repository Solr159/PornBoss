export default function Pagination({
  page,
  lastPage,
  hasPrev,
  hasNext,
  buildPageUrl,
  onFirst,
  onPrev,
  onGoToPage,
  onNext,
  onLast,
}) {
  const isModifiedClick = (e) =>
    e && (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0)
  const windowSize = 11
  const offset = Math.floor(windowSize / 2)
  const totalPages = lastPage && lastPage > 0 ? lastPage : page + offset
  const start = Math.max(1, Math.min(page - offset, totalPages - windowSize + 1))
  const end = Math.min(totalPages, start + windowSize - 1)
  const pages = []
  for (let p = start; p <= end; p++) pages.push(p)

  return (
    <div className="flex flex-col items-center gap-1 py-1 text-sm">
      <div className="flex items-center justify-center gap-1.5">
        <a
          href={buildPageUrl ? buildPageUrl({ page: 1 }) : '#'}
          onClick={(e) => {
            if (isModifiedClick(e) || !hasPrev) return
            e.preventDefault()
            onFirst()
          }}
          className={`rounded border px-2 py-0.5 ${!hasPrev ? 'pointer-events-none opacity-50' : ''}`}
          aria-disabled={!hasPrev}
          aria-label="首页"
        >
          « 首页
        </a>
        <a
          href={buildPageUrl ? buildPageUrl({ page: page - 1 }) : '#'}
          onClick={(e) => {
            if (isModifiedClick(e) || !hasPrev) return
            e.preventDefault()
            onPrev()
          }}
          className={`rounded border px-2 py-0.5 ${!hasPrev ? 'pointer-events-none opacity-50' : ''}`}
          aria-disabled={!hasPrev}
          aria-label="上一页"
        >
          ‹ 上一页
        </a>

        {pages.map((p) => (
          <a
            key={p}
            href={buildPageUrl ? buildPageUrl({ page: p }) : '#'}
            onClick={(e) => {
              if (isModifiedClick(e)) return
              e.preventDefault()
              onGoToPage(p)
            }}
            className={`rounded border px-2.5 py-0.5 leading-tight ${p === page ? 'border-blue-600 bg-blue-600 text-white' : 'bg-white'}`}
            aria-current={p === page ? 'page' : undefined}
          >
            {p}
          </a>
        ))}

        <a
          href={buildPageUrl ? buildPageUrl({ page: page + 1 }) : '#'}
          onClick={(e) => {
            if (isModifiedClick(e) || !hasNext) return
            e.preventDefault()
            onNext()
          }}
          className={`rounded border px-2 py-0.5 ${!hasNext ? 'pointer-events-none opacity-50' : ''}`}
          aria-disabled={!hasNext}
          aria-label="下一页"
        >
          下一页 ›
        </a>
        <a
          href={buildPageUrl ? buildPageUrl({ page: lastPage }) : '#'}
          onClick={(e) => {
            if (isModifiedClick(e) || !hasNext) return
            e.preventDefault()
            onLast()
          }}
          className={`rounded border px-2 py-0.5 ${!hasNext ? 'pointer-events-none opacity-50' : ''}`}
          aria-disabled={!hasNext}
          aria-label="末页"
        >
          末页 »
        </a>
      </div>
    </div>
  )
}
