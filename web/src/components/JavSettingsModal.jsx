export default function JavSettingsModal({
  open,
  onClose,
  javPageSizeInput,
  onJavPageSizeChange,
  idolPageSizeInput,
  onIdolPageSizeChange,
  javSortInput,
  onJavSortChange,
  idolSortInput,
  onIdolSortChange,
  onSave,
}) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4">
      <div className="w-full max-w-sm rounded-lg bg-white p-3 shadow-xl">
        <div className="mb-2 flex items-center justify-between">
          <h2 className="text-base font-semibold">JAV 设置</h2>
          <button
            onClick={onClose}
            className="rounded px-2 py-1 text-gray-500 hover:bg-gray-100"
            aria-label="关闭设置"
          >
            ✕
          </button>
        </div>
        <div className="space-y-2">
          <label className="flex items-center justify-between gap-3 text-sm font-medium text-gray-700">
            <span>每页 JAV 数量</span>
            <input
              type="number"
              min="1"
              value={javPageSizeInput}
              onChange={(e) => onJavPageSizeChange?.(e.target.value)}
              className="w-24 rounded border px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
          <label className="flex items-center justify-between gap-3 text-sm font-medium text-gray-700">
            <span>每页 女优 数量</span>
            <input
              type="number"
              min="1"
              value={idolPageSizeInput}
              onChange={(e) => onIdolPageSizeChange?.(e.target.value)}
              className="w-24 rounded border px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
          <div className="text-sm font-medium text-gray-700">作品排序</div>
          <label
            htmlFor="jav-sort-recent"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按加入时间排序"
          >
            <input
              id="jav-sort-recent"
              type="radio"
              name="jav-sort"
              value="recent"
              checked={javSortInput === 'recent'}
              onChange={() => onJavSortChange?.('recent')}
            />
            <div className="text-sm font-semibold">
              加入时间（默认）
              <span className="ml-2 text-xs font-normal text-gray-500">最近入库的 JAV 优先</span>
            </div>
          </label>
          <label
            htmlFor="jav-sort-code"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按番号排序"
          >
            <input
              id="jav-sort-code"
              type="radio"
              name="jav-sort"
              value="code"
              checked={javSortInput === 'code'}
              onChange={() => onJavSortChange?.('code')}
            />
            <div className="text-sm font-semibold">
              番号
              <span className="ml-2 text-xs font-normal text-gray-500">按番号字母顺序</span>
            </div>
          </label>
          <label
            htmlFor="jav-sort-release"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按发行时间排序"
          >
            <input
              id="jav-sort-release"
              type="radio"
              name="jav-sort"
              value="release"
              checked={javSortInput === 'release'}
              onChange={() => onJavSortChange?.('release')}
            />
            <div className="text-sm font-semibold">
              发行时间
              <span className="ml-2 text-xs font-normal text-gray-500">最新发行优先</span>
            </div>
          </label>
          <label
            htmlFor="jav-sort-play-count"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按播放次数排序"
          >
            <input
              id="jav-sort-play-count"
              type="radio"
              name="jav-sort"
              value="play_count"
              checked={javSortInput === 'play_count'}
              onChange={() => onJavSortChange?.('play_count')}
            />
            <div className="text-sm font-semibold">
              播放次数（多→少）
              <span className="ml-2 text-xs font-normal text-gray-500">按作品累计播放次数</span>
            </div>
          </label>
          <div className="text-sm font-medium text-gray-700">女优排序</div>
          <label
            htmlFor="idol-sort-work"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按作品数量排序"
          >
            <input
              id="idol-sort-work"
              type="radio"
              name="idol-sort"
              value="work"
              checked={idolSortInput === 'work'}
              onChange={() => onIdolSortChange?.('work')}
            />
            <div className="text-sm font-semibold">
              作品数量（默认）
              <span className="ml-2 text-xs font-normal text-gray-500">作品越多越靠前</span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-birth"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按出生日期排序"
          >
            <input
              id="idol-sort-birth"
              type="radio"
              name="idol-sort"
              value="birth"
              checked={idolSortInput === 'birth'}
              onChange={() => onIdolSortChange?.('birth')}
            />
            <div className="text-sm font-semibold">
              年龄
              <span className="ml-2 text-xs font-normal text-gray-500">由小到大（更年轻优先）</span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-height"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按身高排序"
          >
            <input
              id="idol-sort-height"
              type="radio"
              name="idol-sort"
              value="height"
              checked={idolSortInput === 'height'}
              onChange={() => onIdolSortChange?.('height')}
            />
            <div className="text-sm font-semibold">
              身高
              <span className="ml-2 text-xs font-normal text-gray-500">由低到高</span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-bust"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按胸围排序"
          >
            <input
              id="idol-sort-bust"
              type="radio"
              name="idol-sort"
              value="bust"
              checked={idolSortInput === 'bust'}
              onChange={() => onIdolSortChange?.('bust')}
            />
            <div className="text-sm font-semibold">
              胸围
              <span className="ml-2 text-xs font-normal text-gray-500">由大到小</span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-hips"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按臀围排序"
          >
            <input
              id="idol-sort-hips"
              type="radio"
              name="idol-sort"
              value="hips"
              checked={idolSortInput === 'hips'}
              onChange={() => onIdolSortChange?.('hips')}
            />
            <div className="text-sm font-semibold">
              臀围
              <span className="ml-2 text-xs font-normal text-gray-500">由大到小</span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-waist"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按腰围排序"
          >
            <input
              id="idol-sort-waist"
              type="radio"
              name="idol-sort"
              value="waist"
              checked={idolSortInput === 'waist'}
              onChange={() => onIdolSortChange?.('waist')}
            />
            <div className="text-sm font-semibold">
              腰围
              <span className="ml-2 text-xs font-normal text-gray-500">由小到大</span>
            </div>
          </label>
          <label
            htmlFor="idol-sort-cup"
            className="flex cursor-pointer items-center gap-3 rounded border px-3 py-1.5 hover:border-blue-500"
            aria-label="按罩杯排序"
          >
            <input
              id="idol-sort-cup"
              type="radio"
              name="idol-sort"
              value="cup"
              checked={idolSortInput === 'cup'}
              onChange={() => onIdolSortChange?.('cup')}
            />
            <div className="text-sm font-semibold">
              罩杯
              <span className="ml-2 text-xs font-normal text-gray-500">由大到小</span>
            </div>
          </label>
        </div>
        <div className="mt-3 flex justify-end">
          <button onClick={onClose} className="rounded border px-3 py-1 text-sm hover:bg-gray-50">
            取消
          </button>
          <button
            onClick={onSave}
            className="ml-2 rounded bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700"
          >
            保存
          </button>
        </div>
      </div>
    </div>
  )
}
