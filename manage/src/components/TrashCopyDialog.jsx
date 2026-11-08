import { useState, useEffect } from 'react'
import PropTypes from 'prop-types'
import {
  Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField,
  List, ListItemButton, ListItemText, Box, Typography, CircularProgress,
  InputAdornment, IconButton, Alert,
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import ClearIcon from '@mui/icons-material/Clear'
import { getChildren, getPage } from '../api/page'

const TRASH_ID = 'Trash'

//登録日（作成日）を 2026/11/09 の形で出す
const fmtDate = (iso) => (iso ? new Date(iso).toLocaleDateString('ja-JP', { year: 'numeric', month: '2-digit', day: '2-digit' }) : '')

TrashCopyDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  initialQuery: PropTypes.string,
  onClose: PropTypes.func.isRequired,
  onApply: PropTypes.func.isRequired,
}

//Trash 内のページを名前で探し、選んだページの説明・コンテンツを渡すダイアログ。保存はしない。
export default function TrashCopyDialog({ open, initialQuery = '', onClose, onApply }) {
  const [pages, setPages] = useState(null)
  const [loadError, setLoadError] = useState(null)
  const [query, setQuery] = useState('')
  const [selectedId, setSelectedId] = useState(null)
  const [preview, setPreview] = useState(null) // { description, content }
  const [previewLoading, setPreviewLoading] = useState(false)

  //開くたびに Trash を読み直し、検索語をページ名で初期化する
  useEffect(() => {
    if (!open) return
    setQuery(initialQuery)
    setSelectedId(null)
    setPreview(null)
    setLoadError(null)
    setPages(null)
    let canceled = false
    getChildren(TRASH_ID)
      .then(list => {
        if (canceled) return
        //登録日の新しい順
        const byCreatedDesc = (a, b) => new Date(b.createdAt || 0) - new Date(a.createdAt || 0)
        setPages([...(list || [])].sort(byCreatedDesc))
      })
      .catch(e => { if (!canceled) setLoadError(e.message) })
    return () => { canceled = true }
  }, [open, initialQuery])

  //選んだページのコンテンツを読み込む
  useEffect(() => {
    if (!selectedId) return
    let canceled = false
    setPreviewLoading(true)
    getPage(selectedId)
      .then(d => {
        if (canceled || !d) return
        setPreview({ description: d.page.description || '', content: d.pageData || '' })
      })
      .catch(e => { if (!canceled) setLoadError(e.message) })
      .finally(() => { if (!canceled) setPreviewLoading(false) })
    return () => { canceled = true }
  }, [selectedId])

  const words = query.trim().toLowerCase().split(/\s+/).filter(Boolean)
  const results = (pages || []).filter(p => {
    const name = (p.name || '').toLowerCase()
    return words.every(w => name.includes(w))
  })

  const handleSelect = (id) => {
    if (id === selectedId) return
    setPreview(null)
    setSelectedId(id)
  }

  const handleApply = () => {
    if (!preview) return
    onApply(preview)
  }

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>Trash から説明・コンテンツを設定</DialogTitle>
      <DialogContent dividers>
        {loadError && <Alert severity="error" sx={{ mb: 2 }}>{loadError}</Alert>}
        <TextField
          value={query}
          onChange={e => setQuery(e.target.value)}
          size="small"
          fullWidth
          autoFocus
          onFocus={e => e.target.select()}
          placeholder="ページ名で検索"
          sx={{ mb: 2 }}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" color="disabled" />
                </InputAdornment>
              ),
              endAdornment: query ? (
                <InputAdornment position="end">
                  <IconButton size="small" onClick={() => setQuery('')} aria-label="検索条件をクリア">
                    <ClearIcon fontSize="small" />
                  </IconButton>
                </InputAdornment>
              ) : null,
            },
          }}
        />

        <Box sx={{ display: 'flex', gap: 2, height: 420 }}>
          {/* 検索結果 */}
          <Box sx={{ width: 300, flexShrink: 0, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
            <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5 }}>
              {pages ? `${results.length} / ${pages.length} 件` : '読み込み中...'}
            </Typography>
            <Box sx={{ flex: 1, overflowY: 'auto', border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
              {!pages ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', pt: 4 }}>
                  <CircularProgress size={24} />
                </Box>
              ) : results.length === 0 ? (
                <Typography variant="body2" color="text.secondary" sx={{ p: 2, textAlign: 'center' }}>
                  一致するページがありません
                </Typography>
              ) : (
                <List dense disablePadding>
                  {results.map(p => (
                    <ListItemButton
                      key={p.id}
                      selected={p.id === selectedId}
                      onClick={() => handleSelect(p.id)}
                    >
                      <ListItemText
                        primary={
                          <Box component="span" sx={{ display: 'flex', alignItems: 'baseline', gap: 1 }}>
                            <Box component="span" sx={{ flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {p.name || '(名前なし)'}
                            </Box>
                            <Typography component="span" variant="caption" color="text.secondary" sx={{ flexShrink: 0 }}>
                              {fmtDate(p.createdAt)}
                            </Typography>
                          </Box>
                        }
                        secondary={p.description || null}
                        slotProps={{
                          secondary: { noWrap: true },
                        }}
                      />
                    </ListItemButton>
                  ))}
                </List>
              )}
            </Box>
          </Box>

          {/* 選んだページの説明・コンテンツ */}
          <Box sx={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 1.5, minHeight: 0 }}>
            {!selectedId ? (
              <Box sx={{ flex: 1, border: '1px dashed', borderColor: 'divider', borderRadius: 1 }} />
            ) : previewLoading || !preview ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', pt: 4 }}>
                <CircularProgress size={24} />
              </Box>
            ) : (
              <>
                <TextField
                  label="説明"
                  value={preview.description}
                  multiline
                  rows={2}
                  size="small"
                  fullWidth
                  slotProps={{ htmlInput: { readOnly: true } }}
                />
                <TextField
                  label="コンテンツ"
                  value={preview.content}
                  multiline
                  size="small"
                  fullWidth
                  sx={{
                    flex: 1,
                    minHeight: 0,
                    '& .MuiInputBase-root': { height: '100%', alignItems: 'flex-start' },
                    '& textarea': { height: '100% !important', overflowY: 'auto !important' },
                  }}
                  slotProps={{ htmlInput: { readOnly: true, style: { fontFamily: 'monospace', fontSize: 12 } } }}
                />
              </>
            )}
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>キャンセル</Button>
        <Button onClick={handleApply} variant="contained" disabled={!preview || previewLoading}>
          設定
        </Button>
      </DialogActions>
    </Dialog>
  )
}
