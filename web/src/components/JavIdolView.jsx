import JavIdolGrid from '@/components/JavIdolGrid'
import Pagination from '@/components/Pagination'
import WaterfallLoader from '@/components/WaterfallLoader'
import { zh } from '@/utils/i18n'

export default function JavIdolView({
  page,
  lastPage,
  totalItems,
  hasPrev,
  hasNext,
  loading,
  buildPageUrl,
  buildIdolUrl,
  directoryIds = [],
  javMetadataLanguage,
  preferChineseName = false,
  onFirst,
  onPrev,
  onGoToPage,
  onNext,
  onLast,
  items,
  onSelectIdol,
  onOpenFavorites,
  onMerged,
  waterfallMode,
  onWaterfallModeChange,
  onLoadMore,
  loadingMore,
  hasMore,
}) {
  return (
    <>
      <div className="sticky-pagination mb-2.5">
        <div className="relative grid gap-3 md:grid-cols-[1fr_auto_1fr] md:items-center">
          <div />
          <div className="flex justify-center overflow-x-auto">
            <Pagination
              page={page}
              lastPage={lastPage}
              totalItems={totalItems}
              hasPrev={hasPrev}
              hasNext={hasNext}
              loading={loading}
              buildPageUrl={buildPageUrl}
              onFirst={onFirst}
              onPrev={onPrev}
              onGoToPage={onGoToPage}
              onNext={onNext}
              onLast={onLast}
              waterfallMode={waterfallMode}
              onWaterfallModeChange={onWaterfallModeChange}
            />
          </div>
          <div className="hidden md:block" />
        </div>
      </div>
      {loading ? (
        <div className="mt-4 flex min-h-[200px] items-center justify-center rounded border border-dashed border-gray-200 text-gray-500">
          {zh('加载中…', 'Loading...')}
        </div>
      ) : (
        <JavIdolGrid
          items={items}
          onSelectIdol={onSelectIdol}
          onOpenFavorites={onOpenFavorites}
          onMerged={onMerged}
          buildIdolUrl={buildIdolUrl}
          directoryIds={directoryIds}
          javMetadataLanguage={javMetadataLanguage}
          preferChineseName={preferChineseName}
        />
      )}
      <WaterfallLoader
        enabled={waterfallMode && !loading}
        hasMore={hasMore}
        loading={loadingMore}
        onLoadMore={onLoadMore}
      />
    </>
  )
}
