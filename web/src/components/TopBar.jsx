import { useEffect, useRef } from 'react'
import { Button } from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import LocalOfferOutlinedIcon from '@mui/icons-material/LocalOfferOutlined'
import ShuffleOutlinedIcon from '@mui/icons-material/ShuffleOutlined'
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined'
import SwapHorizOutlinedIcon from '@mui/icons-material/SwapHorizOutlined'

export default function TopBar({
  onHome,
  isJavMode,
  onToggleMode,
  videoSearchInput,
  onVideoSearchInputChange,
  onSubmitVideoSearch,
  videoSearchHref,
  randomHref,
  onRandomClick,
  onOpenTagModal,
  onOpenJavTagModal,
  onOpenVideoSettings,
  onOpenJavSettings,
  onOpenGlobalSettings,
  javSearchInput,
  onJavSearchInputChange,
  onSubmitJavSearch,
  javSearchHref,
  javRandomHref,
  javRandomMode,
  onJavRandomClick,
  isModifiedClick,
  javTab,
  onSwitchJavTab,
  filterSummary,
}) {
  const headerRef = useRef(null)
  const headerClassName = ['sticky top-0 z-40 border-b bg-white/80 backdrop-blur']
    .filter(Boolean)
    .join(' ')

  const updateTopbarOffset = () => {
    const height = headerRef.current?.getBoundingClientRect().height || 0
    document.documentElement.style.setProperty('--topbar-height', `${Math.round(height)}px`)
  }

  useEffect(() => {
    updateTopbarOffset()
    window.addEventListener('resize', updateTopbarOffset)
    return () => window.removeEventListener('resize', updateTopbarOffset)
  }, [])

  useEffect(() => {
    updateTopbarOffset()
  }, [isJavMode, javTab, javRandomMode])

  const handleSettingsClick = () => {
    if (isJavMode) {
      onOpenJavSettings?.()
    } else {
      onOpenVideoSettings?.()
    }
  }

  return (
    <header ref={headerRef} className={headerClassName}>
      <div className="mx-auto flex max-w-screen-2xl flex-col gap-3 px-6 py-3">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex min-w-0 flex-1 flex-wrap items-center gap-3">
            <button
              type="button"
              onClick={onHome}
              className="cursor-pointer select-none rounded text-left text-xl font-bold focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
            >
              Pornboss
            </button>

            <div className="flex flex-wrap items-center gap-2">
              {isJavMode ? (
                <div className="flex items-center gap-2">
                  <Button
                    variant={javTab === 'list' ? 'contained' : 'outlined'}
                    onClick={() => onSwitchJavTab('list')}
                  >
                    作品
                  </Button>
                  <Button
                    variant={javTab === 'idol' ? 'contained' : 'outlined'}
                    onClick={() => onSwitchJavTab('idol')}
                  >
                    女优
                  </Button>
                  <form
                    onSubmit={onSubmitJavSearch}
                    className="flex items-center overflow-hidden rounded-full border border-gray-200 bg-white shadow-sm"
                  >
                    <input
                      value={javSearchInput}
                      onChange={(e) => onJavSearchInputChange(e.target.value)}
                      placeholder={javTab === 'idol' ? '搜索女优名称' : '搜索番号或标题'}
                      className="h-10 flex-1 border-0 bg-white px-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      aria-label="搜索JAV"
                    />
                    <Button
                      component="a"
                      href={javSearchHref}
                      startIcon={<SearchIcon fontSize="small" />}
                      variant="contained"
                      size="medium"
                      onClick={(e) => {
                        if (isModifiedClick(e)) return
                        onSubmitJavSearch(e)
                      }}
                      sx={{
                        borderTopLeftRadius: 0,
                        borderBottomLeftRadius: 0,
                        minHeight: '40px',
                        height: '40px',
                        px: 2.5,
                      }}
                    >
                      搜索
                    </Button>
                  </form>
                  <Button
                    component="a"
                    href={javRandomHref}
                    startIcon={<ShuffleOutlinedIcon fontSize="small" />}
                    variant="outlined"
                    onClick={(e) => {
                      if (isModifiedClick(e)) return
                      e.preventDefault()
                      onJavRandomClick?.()
                    }}
                  >
                    随机
                  </Button>
                  <Button
                    startIcon={<LocalOfferOutlinedIcon fontSize="small" />}
                    variant="outlined"
                    onClick={onOpenJavTagModal}
                  >
                    标签
                  </Button>
                </div>
              ) : (
                <>
                  <form
                    onSubmit={onSubmitVideoSearch}
                    className="flex items-center overflow-hidden rounded-full border border-gray-200 bg-white shadow-sm"
                  >
                    <input
                      value={videoSearchInput}
                      onChange={(e) => onVideoSearchInputChange(e.target.value)}
                      placeholder="搜索文件名"
                      className="h-10 flex-1 border-0 bg-white px-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      aria-label="搜索视频"
                    />
                    <Button
                      component="a"
                      href={videoSearchHref}
                      startIcon={<SearchIcon fontSize="small" />}
                      variant="contained"
                      size="medium"
                      onClick={(e) => {
                        if (isModifiedClick(e)) return
                        onSubmitVideoSearch(e)
                      }}
                      sx={{
                        borderTopLeftRadius: 0,
                        borderBottomLeftRadius: 0,
                        minHeight: '40px',
                        height: '40px',
                        px: 2.5,
                      }}
                    >
                      搜索
                    </Button>
                  </form>
                  <div className="flex items-center gap-2">
                    <Button
                      component="a"
                      href={randomHref}
                      startIcon={<ShuffleOutlinedIcon fontSize="small" />}
                      variant="outlined"
                      onClick={(e) => {
                        if (isModifiedClick(e)) return
                        e.preventDefault()
                        onRandomClick()
                      }}
                    >
                      随机
                    </Button>
                  </div>
                  <Button
                    startIcon={<LocalOfferOutlinedIcon fontSize="small" />}
                    variant="outlined"
                    onClick={onOpenTagModal}
                  >
                    标签
                  </Button>
                </>
              )}

              <Button
                startIcon={<SettingsOutlinedIcon fontSize="small" />}
                variant="outlined"
                onClick={handleSettingsClick}
                title="全局设置"
              >
                设置
              </Button>
              <span className="ml-1 max-w-[280px] truncate text-xs text-gray-500">
                筛选条件：
                <span className="font-semibold text-gray-700">{filterSummary || '无'}</span>
              </span>
            </div>
          </div>

          <div className="flex flex-shrink-0 items-center gap-2">
            <Button
              startIcon={<SettingsOutlinedIcon fontSize="small" />}
              variant="outlined"
              onClick={onOpenGlobalSettings}
              title="全局设置"
            >
              全局设置
            </Button>
            <Button
              variant="contained"
              color={isJavMode ? 'secondary' : 'primary'}
              startIcon={<SwapHorizOutlinedIcon fontSize="small" />}
              onClick={onToggleMode}
            >
              {isJavMode ? '切换到视频' : '切换到 JAV'}
            </Button>
          </div>
        </div>
      </div>
    </header>
  )
}
