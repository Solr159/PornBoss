import { useState } from 'react'
import { Button, TextField } from '@mui/material'
import { postJavLibraryChat, postJavNlQuery } from '@/api'
import { directoryQueryIds, useStore } from '@/store'
import { zh } from '@/utils/i18n'

export default function JavNlSearchBar({ collectionId = null }) {
  const [deepseekOpen, setDeepseekOpen] = useState(false)
  const [mode, setMode] = useState('search')
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [chatInput, setChatInput] = useState('')
  const [chatBusy, setChatBusy] = useState(false)
  const [chatErr, setChatErr] = useState('')
  const [chatHistory, setChatHistory] = useState([])
  const [chatKind, setChatKind] = useState('snapshot')
  const [focusCode, setFocusCode] = useState('')
  const directoryIds = useStore(directoryQueryIds)
  const nlHint = useStore((s) => s.javNlHint)
  const javPageSize = useStore((s) => s.javPageSize)
  const javSort = useStore((s) => s.javSort)
  const javTempSort = useStore((s) => s.javTempSort)

  const applySearch = async () => {
    const q = text.trim()
    if (!q) return
    setBusy(true)
    setErr('')
    try {
      const resp = await postJavNlQuery({
        query: q,
        directoryIds,
        collectionId: collectionId || undefined,
        limit: javPageSize,
        offset: 0,
        sort: javTempSort || javSort || 'recent',
      })
      const r = resp.resolved || {}
      useStore.setState({
        javSearchTerm: String(r.search || '').trim(),
        javTags: Array.isArray(r.tag_ids) ? r.tag_ids : [],
        javIdolIds: Array.isArray(r.idol_ids) ? r.idol_ids : [],
        javStudioId: r.studio_id > 0 ? r.studio_id : null,
        javStudioName: '',
        javPage: 1,
        javTempSort: '',
        javRandomMode: false,
        javRandomSeed: null,
        javNlHint: [
          resp.interpretation,
          ...(resp.warnings || []).map((w) => zh(`未匹配标签/演员：${w}`, `Unmatched: ${w}`)),
        ]
          .filter(Boolean)
          .join(' '),
      })
      useStore.setState({
        javItems: resp.items || [],
        javTotal: resp.total ?? 0,
      })
    } catch (e) {
      setErr(e.message || zh('请求失败', 'Request failed'))
    } finally {
      setBusy(false)
    }
  }

  const sendChat = async () => {
    const q = chatInput.trim()
    if (!q) return
    setChatBusy(true)
    setChatErr('')
    try {
      const history = chatHistory.map(({ role, content }) => ({ role, content }))
      const apiMode = chatKind === 'model' ? 'code_reasoning' : 'library'
      const resp = await postJavLibraryChat({
        message: q,
        history,
        directoryIds,
        collectionId: collectionId || undefined,
        mode: apiMode,
        focusCode: chatKind === 'model' ? focusCode : '',
      })
      const reply = String(resp.reply || '').trim() || zh('（无回复）', '(No reply)')
      setChatHistory((prev) => [
        ...prev,
        { role: 'user', content: q },
        { role: 'assistant', content: reply },
      ])
      setChatInput('')
    } catch (e) {
      setChatErr(e.message || zh('请求失败', 'Request failed'))
    } finally {
      setChatBusy(false)
    }
  }

  const tabBtn =
    'rounded-md border px-3 py-1 text-xs font-medium transition disabled:opacity-50 disabled:cursor-not-allowed'
  const tabActive = 'border-indigo-500 bg-indigo-600 text-white shadow-sm'
  const tabIdle = 'border-indigo-200 bg-white text-indigo-900 hover:border-indigo-400'

  return (
    <div className="mb-4">
      <button
        type="button"
        onClick={() => setDeepseekOpen((v) => !v)}
        className="flex w-full items-center justify-between rounded-lg border border-indigo-200 bg-indigo-50/80 px-3 py-2 text-left text-sm font-medium text-indigo-950 shadow-sm hover:bg-indigo-100/90"
      >
        <span>{zh('DeepSeek：自然语言找片 / 库问答', 'DeepSeek: NL search & chat')}</span>
        <span className="text-xs text-indigo-700">
          {deepseekOpen ? zh('点击收起', 'Hide') : zh('点击展开', 'Show')}
        </span>
      </button>

      {deepseekOpen ? (
        <div className="mt-2 rounded-lg border border-indigo-100 bg-indigo-50/60 p-3">
          <div className="mb-3 flex flex-wrap items-center gap-2">
            <span className="text-xs font-medium text-indigo-900">
              {zh('DeepSeek', 'DeepSeek')}
            </span>
            <div className="flex gap-1">
              <button
                type="button"
                className={`${tabBtn} ${mode === 'search' ? tabActive : tabIdle}`}
                onClick={() => setMode('search')}
              >
                {zh('自然语言找片', 'Natural language search')}
              </button>
              <button
                type="button"
                className={`${tabBtn} ${mode === 'chat' ? tabActive : tabIdle}`}
                onClick={() => setMode('chat')}
              >
                {zh('对话', 'Chat')}
              </button>
            </div>
          </div>

          {mode === 'search' ? (
            <>
              <div className="mb-1 text-xs text-indigo-800">
                {zh(
                  '用自然语言描述想看的作品，将解析为筛选条件并刷新列表。',
                  'Describe what you want; filters update the list.'
                )}
              </div>
              <div className="flex flex-col gap-2 sm:flex-row sm:items-start">
                <TextField
                  size="small"
                  fullWidth
                  multiline
                  minRows={1}
                  maxRows={3}
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  placeholder={zh('描述你想看的…', 'Describe what you want to watch…')}
                  disabled={busy}
                />
                <Button
                  variant="contained"
                  onClick={() => void applySearch()}
                  disabled={busy || !text.trim()}
                >
                  {busy ? zh('分析中…', 'Searching…') : zh('搜索', 'Search')}
                </Button>
              </div>
              {nlHint ? <p className="mt-2 text-xs text-indigo-800">{nlHint}</p> : null}
              {err ? <p className="mt-1 text-xs text-red-700">{err}</p> : null}
            </>
          ) : (
            <>
              <div className="mb-2 flex flex-wrap gap-1">
                <button
                  type="button"
                  className={`${tabBtn} ${chatKind === 'snapshot' ? tabActive : tabIdle}`}
                  onClick={() => setChatKind('snapshot')}
                >
                  {zh('基于库快照', 'Library snapshot')}
                </button>
                <button
                  type="button"
                  className={`${tabBtn} ${chatKind === 'model' ? tabActive : tabIdle}`}
                  onClick={() => setChatKind('model')}
                >
                  {zh('番号推测（仅模型常识）', 'Code guess (model only)')}
                </button>
              </div>
              {chatKind === 'snapshot' ? (
                <div className="mb-2 text-xs text-indigo-800">
                  {zh(
                    '询问库的整体口味、推荐方向、标签结构等；每次附带当前可见库的统计快照（无文件路径）。',
                    'Ask about your library taste and patterns; each request includes an aggregate snapshot (no file paths).'
                  )}
                </div>
              ) : (
                <>
                  <div className="mb-2 text-xs text-indigo-800">
                    {zh(
                      '不读取你库里的条目，只根据番号与行业常识做推测；可能不准，仅供娱乐参考。',
                      'No data from your library—only public-style guesses from the code; may be wrong.'
                    )}
                  </div>
                  <TextField
                    className="mb-2"
                    size="small"
                    fullWidth
                    value={focusCode}
                    onChange={(e) => setFocusCode(e.target.value)}
                    placeholder={zh('番号（可选，会一并发给模型）', 'Catalog code (optional)')}
                    disabled={chatBusy}
                  />
                </>
              )}
              {chatHistory.length > 0 ? (
                <div className="mb-2 max-h-60 space-y-2 overflow-y-auto rounded border border-indigo-100 bg-white/80 p-2 text-xs">
                  {chatHistory.map((m, i) => (
                    <div
                      key={`${i}-${m.role}`}
                      className={
                        m.role === 'user'
                          ? 'ml-4 rounded-md bg-indigo-100 px-2 py-1.5 text-indigo-950'
                          : 'mr-4 rounded-md bg-gray-50 px-2 py-1.5 text-gray-800'
                      }
                    >
                      <div className="mb-0.5 text-[10px] font-semibold uppercase text-gray-500">
                        {m.role === 'user' ? zh('你', 'You') : zh('助手', 'Assistant')}
                      </div>
                      <div className="whitespace-pre-wrap">{m.content}</div>
                    </div>
                  ))}
                </div>
              ) : null}
              <div className="flex flex-col gap-2 sm:flex-row sm:items-start">
                <TextField
                  size="small"
                  fullWidth
                  multiline
                  minRows={2}
                  maxRows={8}
                  value={chatInput}
                  onChange={(e) => setChatInput(e.target.value)}
                  placeholder={
                    chatKind === 'model'
                      ? zh(
                          '例如：这个前缀通常是什么片商线？题材大概偏哪类？',
                          'e.g. What line/studio does this prefix suggest?'
                        )
                      : zh(
                          '例如：根据我的库推荐几部可能合我口味的？',
                          'e.g. Recommendations based on my library snapshot?'
                        )
                  }
                  disabled={chatBusy}
                />
                <div className="flex shrink-0 flex-col gap-1 sm:w-auto">
                  <Button
                    variant="contained"
                    onClick={() => void sendChat()}
                    disabled={chatBusy || !chatInput.trim()}
                  >
                    {chatBusy ? zh('思考中…', 'Thinking…') : zh('发送', 'Send')}
                  </Button>
                  {chatHistory.length > 0 ? (
                    <Button
                      variant="text"
                      size="small"
                      onClick={() => {
                        setChatHistory([])
                        setChatErr('')
                      }}
                    >
                      {zh('清空对话', 'Clear chat')}
                    </Button>
                  ) : null}
                </div>
              </div>
              {chatErr ? <p className="mt-1 text-xs text-red-700">{chatErr}</p> : null}
            </>
          )}
        </div>
      ) : null}
    </div>
  )
}
