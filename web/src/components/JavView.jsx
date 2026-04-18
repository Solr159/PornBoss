import { Popover } from '@mui/material'
import { useState } from 'react'
import JavGrid from '@/components/JavGrid'
import Pagination from '@/components/Pagination'
import { zh } from '@/utils/i18n'

export default function JavView({
  javPage,
  javLastPage,
  javHasPrev,
  javHasNext,
  javLoading,
  javRandomMode,
  javPageSort,
  javGlobalSort,
  buildJavUrl,
  setJavPage,
  setJavPageSort,
  javItems,
  onPlay,
  onIdolClick,
  onTagClick,
  onEditTags,
  onOpenFile,
  onRevealFile,
}) {
  const contentClass = javRandomMode ? 'mt-4' : ''
  const [sortAnchorEl, setSortAnchorEl] = useState(null)
  const effectiveSort = javPageSort || javGlobalSort
  const sortOptions = [
    { value: 'recent', label: zh('加入时间', 'Added time') },
    { value: 'code', label: zh('番号', 'Code') },
    { value: 'release', label: zh('发行时间', 'Release date') },
    { value: 'play_count', label: zh('播放次数', 'Play count') },
  ]
  const currentSortLabel =
    sortOptions.find((option) => option.value === effectiveSort)?.label ||
    zh('加入时间', 'Added time')

  const isOptionActive = (value) => {
    return effectiveSort === value
  }

  const openSortMenu = (event) => {
    setSortAnchorEl(event.currentTarget)
  }

  const closeSortMenu = () => {
    setSortAnchorEl(null)
  }

  return (
    <>
      {!javRandomMode && (
        <div className="sticky-pagination mb-4 grid gap-3 md:grid-cols-[1fr_auto_1fr] md:items-center">
          <div className="hidden md:block" />
          <div className="flex justify-center overflow-x-auto">
            <Pagination
              page={javPage}
              lastPage={javLastPage}
              hasPrev={javHasPrev}
              hasNext={javHasNext}
              loading={javLoading}
              buildPageUrl={({ page: targetPage }) => buildJavUrl({ page: targetPage })}
              onFirst={() => setJavPage(1)}
              onPrev={() => {
                if (javHasPrev) setJavPage(javPage - 1)
              }}
              onGoToPage={(p) => setJavPage(p)}
              onNext={() => {
                if (javHasNext) setJavPage(javPage + 1)
              }}
              onLast={() => setJavPage(javLastPage)}
            />
          </div>
          <div className="flex justify-end">
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500">{zh('排序', 'Sort')}</span>
              <button
                type="button"
                onClick={openSortMenu}
                aria-haspopup="dialog"
                aria-expanded={Boolean(sortAnchorEl)}
                aria-label={zh('修改当前 JAV 排序方式', 'Change current JAV sort')}
                className="inline-flex items-center gap-1.5 rounded-md border border-gray-300 bg-white px-2 py-1 text-xs text-gray-700 shadow-sm hover:border-gray-400"
              >
                <span>{currentSortLabel}</span>
                <span
                  aria-hidden="true"
                  className="block h-1.5 w-1.5 rotate-45 border-b border-r border-gray-400"
                />
              </button>
            </div>
            <Popover
              open={Boolean(sortAnchorEl)}
              anchorEl={sortAnchorEl}
              onClose={closeSortMenu}
              disableScrollLock
              anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
              transformOrigin={{ vertical: 'top', horizontal: 'right' }}
            >
              <div className="flex min-w-[120px] flex-col p-1">
                {sortOptions.map((option) => (
                  <button
                    key={option.value || 'default'}
                    type="button"
                    onClick={() => {
                      closeSortMenu()
                      setJavPageSort?.(option.value)
                    }}
                    className={`rounded px-2 py-1 text-left text-xs ${
                      isOptionActive(option.value)
                        ? 'bg-blue-50 text-blue-700'
                        : 'text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    {option.label}
                  </button>
                ))}
              </div>
            </Popover>
          </div>
        </div>
      )}
      {javLoading ? (
        <div
          className={`${contentClass} flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500`}
        >
          {zh('加载中…', 'Loading...')}
        </div>
      ) : (
        <div className={contentClass}>
          <JavGrid
            items={javItems}
            onPlay={onPlay}
            onIdolClick={onIdolClick}
            onTagClick={onTagClick}
            onEditTags={onEditTags}
            onOpenFile={onOpenFile}
            onRevealFile={onRevealFile}
          />
        </div>
      )}
    </>
  )
}
