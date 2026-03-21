import { Button } from '@mui/material'
import CheckBoxOutlineBlankIcon from '@mui/icons-material/CheckBoxOutlineBlank'
import Pagination from './Pagination'
import VideoGrid from './VideoGrid'

export default function VideoView({
  selectMode,
  setSelectMode,
  selectedCount,
  setSelectionOpsOpen,
  page,
  lastPage,
  canPrev,
  canNext,
  loading,
  randomMode,
  buildVideoUrl,
  setPage,
  goToLastPage,
  videos,
  selectedVideoIds,
  toggleSelectVideo,
  onToggleSelectPage,
  openPlayer,
  setTagPickerFor,
  onTagClick,
}) {
  const pageIds = videos.map((video) => video?.id).filter(Boolean)
  const pageSelectable = pageIds.length > 0
  const pageAllSelected = pageSelectable && pageIds.every((id) => selectedVideoIds.has(id))

  return (
    <>
      <div className="sticky-pagination mb-4 grid grid-cols-[auto_1fr_auto] items-start gap-3">
        <div className="flex items-center gap-2 pt-1">
          <Button
            startIcon={<CheckBoxOutlineBlankIcon fontSize="small" />}
            variant={selectMode ? 'contained' : 'outlined'}
            onClick={() => setSelectMode(!selectMode)}
            size="small"
          >
            {selectMode ? '退出选择' : '选择模式'}
          </Button>
          {selectMode && (
            <>
              <Button
                variant="outlined"
                size="small"
                onClick={onToggleSelectPage}
                disabled={!pageSelectable}
              >
                {pageAllSelected ? '取消本页' : '全选本页'}
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={() => setSelectionOpsOpen(true)}
                disabled={selectedCount === 0}
              >
                操作
              </Button>
            </>
          )}
        </div>
        <div className="flex justify-center">
          {!randomMode && (
            <Pagination
              page={page}
              lastPage={lastPage}
              hasPrev={canPrev}
              hasNext={canNext}
              loading={loading}
              buildPageUrl={({ page: targetPage }) =>
                buildVideoUrl({ page: targetPage, random: false })
              }
              onFirst={() => setPage(1)}
              onPrev={() => {
                if (canPrev) setPage(page - 1)
              }}
              onGoToPage={(p) => setPage(p)}
              onNext={() => {
                if (canNext) setPage(page + 1)
              }}
              onLast={() => {
                goToLastPage()
              }}
            />
          )}
        </div>
        <div className="pointer-events-none flex items-center gap-2 opacity-0">
          <Button
            startIcon={<CheckBoxOutlineBlankIcon fontSize="small" />}
            variant={selectMode ? 'contained' : 'outlined'}
            size="small"
          >
            {selectMode ? '退出选择' : '选择模式'}
          </Button>
          {selectMode && (
            <>
              <Button variant="outlined" size="small">
                {pageAllSelected ? '取消本页' : '全选本页'}
              </Button>
              <Button variant="outlined" size="small">
                操作
              </Button>
            </>
          )}
        </div>
      </div>
      {loading ? (
        <div className="mt-4 flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500">
          加载中…
        </div>
      ) : (
        <VideoGrid
          videos={videos}
          selectedIds={selectedVideoIds}
          onToggleSelect={toggleSelectVideo}
          onPlay={(video) => openPlayer(video)}
          selectMode={selectMode}
          onOpenTagPicker={(vid) => setTagPickerFor(vid)}
          onTagClick={onTagClick}
        />
      )}
    </>
  )
}
