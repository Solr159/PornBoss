import { useEffect, useMemo, useRef } from 'react'
import flvjs from 'flv.js'
import videojs from 'video.js'
import 'video.js/dist/video-js.css'
import 'videojs-flvjs'
import { getVideoDisplayName } from '../utils/display'
import {
  PLAYER_HOTKEY_ACTIONS,
  normalizePlayerHotkeyKey,
  normalizePlayerHotkeysList,
} from '../utils/playerHotkeys'

const VOLUME_STORAGE_KEY = 'pornboss.player.volume'

if (typeof window !== 'undefined' && !window.flvjs) {
  window.flvjs = flvjs
}

export default function PlayerModal({ video, onClose, hotkeys = [] }) {
  const videoRef = useRef(null)
  const playerRef = useRef(null)
  const hotkeyMapRef = useRef(new Map())
  const normalizedHotkeys = useMemo(() => normalizePlayerHotkeysList(hotkeys), [hotkeys])
  const isFlv = useMemo(() => {
    const rawPath = String(video?.path || video?.filename || '')
    const parts = rawPath.split('.')
    if (parts.length < 2) return false
    return parts.at(-1)?.toLowerCase() === 'flv'
  }, [video])
  const streamParams = new URLSearchParams()
  const directoryPath = video?.directory?.path || video?.directory_path || ''
  if (video?.path && directoryPath) {
    streamParams.set('path', video.path)
    streamParams.set('dir_path', directoryPath)
  }
  const streamSrc = video
    ? `/videos/${video.id}/stream${streamParams.toString() ? `?${streamParams.toString()}` : ''}`
    : ''

  useEffect(() => {
    hotkeyMapRef.current = new Map(normalizedHotkeys.map((item) => [item.key, item]))
  }, [normalizedHotkeys])

  useEffect(() => {
    if (!video || !videoRef.current) return

    const player = videojs(videoRef.current, {
      controls: true,
      autoplay: true,
      preload: 'auto',
      ...(isFlv
        ? {
            techOrder: ['html5', 'flvjs'],
            flvjs: {
              mediaDataSource: { cors: true },
            },
          }
        : {}),
      sources: [
        {
          src: streamSrc,
          type: isFlv ? 'video/x-flv' : 'video/mp4',
        },
      ],
    })

    playerRef.current = player

    const playerEl = player.el()
    const savedVolume = (() => {
      try {
        const raw = localStorage.getItem(VOLUME_STORAGE_KEY)
        if (raw == null) return null
        const value = Number.parseFloat(raw)
        return Number.isFinite(value) ? value : null
      } catch {
        return null
      }
    })()

    if (savedVolume != null) {
      player.volume(Math.min(1, Math.max(0, savedVolume)))
    }

    const seekBy = (offsetSeconds) => {
      const current = player.currentTime() || 0
      const duration = player.duration()
      let next = current + offsetSeconds
      if (Number.isFinite(duration)) {
        next = Math.min(Math.max(0, next), duration)
      } else {
        next = Math.max(0, next)
      }
      player.currentTime(next)
    }

    const adjustVolume = (delta) => {
      const current = player.volume()
      const next = Math.min(1, Math.max(0, current + delta))
      player.volume(next)
    }

    const handleKeyDown = (event) => {
      const target = event.target
      if (
        target instanceof HTMLElement &&
        (target.isContentEditable ||
          target.closest('input, textarea, select, [contenteditable="true"]'))
      ) {
        return
      }
      const key = normalizePlayerHotkeyKey(event.key || '')
      const configured = hotkeyMapRef.current.get(key)
      if (configured) {
        event.preventDefault()
        if (configured.action === PLAYER_HOTKEY_ACTIONS.SEEK) {
          seekBy(configured.amount)
        } else if (configured.action === PLAYER_HOTKEY_ACTIONS.VOLUME) {
          adjustVolume(configured.amount / 100)
        }
        event.stopPropagation()
        return
      }
      switch (key) {
        case ' ':
        case 'Spacebar': {
          event.preventDefault()
          if (player.paused()) {
            player.play()
          } else {
            player.pause()
          }
          break
        }
        case 'Escape':
          event.preventDefault()
          onClose()
          break
        default:
          return
      }
      event.stopPropagation()
    }

    const focusPlayer = () => {
      playerEl?.focus({ preventScroll: true })
    }

    if (playerEl && !playerEl.hasAttribute('tabindex')) {
      playerEl.setAttribute('tabindex', '-1')
    }

    window.addEventListener('keydown', handleKeyDown, true)

    const handleVolumeChange = () => {
      try {
        localStorage.setItem(VOLUME_STORAGE_KEY, String(player.volume()))
      } catch {
        return
      }
    }

    player.ready(focusPlayer)
    player.on('fullscreenchange', focusPlayer)
    player.on('volumechange', handleVolumeChange)

    return () => {
      window.removeEventListener('keydown', handleKeyDown, true)
      player.off('fullscreenchange', focusPlayer)
      player.off('volumechange', handleVolumeChange)
      playerRef.current?.dispose()
      playerRef.current = null
    }
  }, [video, onClose, streamSrc, isFlv])

  if (!video) return null

  const displayName = getVideoDisplayName(video)

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
      <div className="relative mx-4 w-full max-w-6xl rounded-lg bg-white shadow-lg">
        <button
          aria-label="关闭"
          onClick={onClose}
          className="absolute right-3 top-3 rounded-full bg-black/60 px-2 py-1 text-sm text-white hover:bg-black/80"
        >
          ×
        </button>
        <div className="flex flex-col gap-4 p-4">
          <h2 className="truncate text-lg font-semibold" title={displayName}>
            {displayName}
          </h2>
          <div className="player-shell relative w-full bg-black">
            <div data-vjs-player className="h-full w-full">
              <video
                ref={videoRef}
                className="video-js vjs-big-play-centered h-full w-full"
                playsInline
              >
                <track kind="captions" />
              </video>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
