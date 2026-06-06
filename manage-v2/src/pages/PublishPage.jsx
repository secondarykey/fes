import { useState, useEffect, useCallback } from 'react'
import {
  Typography, Box, Paper, Button, CircularProgress, Alert,
  Snackbar, Dialog, DialogTitle, DialogContent,
  DialogContentText, DialogActions, Chip, Collapse, Link,
  List, ListItem, ListItemText, ListItemIcon, Divider,
} from '@mui/material'
import PublishIcon from '@mui/icons-material/Publish'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ExpandLessIcon from '@mui/icons-material/ExpandLess'
import ArticleIcon from '@mui/icons-material/Article'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import LockIcon from '@mui/icons-material/Lock'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import { getDrafts, getDraft, publishDraft } from '../api/draft'

function fmtDate(iso) {
  return new Date(iso).toLocaleString('ja-JP', { dateStyle: 'short', timeStyle: 'short' })
}

function DraftCard({ draft, onPublished }) {
  const [expanded, setExpanded] = useState(false)
  const [pages, setPages] = useState(null)
  const [loadingPages, setLoadingPages] = useState(false)
  const [publishing, setPublishing] = useState(false)
  const [publishDialog, setPublishDialog] = useState(false)
  const [published, setPublished] = useState(false)
  const [error, setError] = useState(null)
  const isLocked = draft.lock

  const handleExpand = async () => {
    if (expanded) {
      setExpanded(false)
      return
    }
    setExpanded(true)
    if (pages !== null) return
    setLoadingPages(true)
    try {
      const data = await getDraft(draft.id)
      setPages(data?.pages || [])
    } catch (e) {
      setError(e.message)
    } finally {
      setLoadingPages(false)
    }
  }

  const handlePublish = async () => {
    setPublishDialog(false)
    setPublishing(true)
    setError(null)
    try {
      await publishDraft(draft.id)
      setPublished(true)
      if (onPublished) onPublished(draft.id)
    } catch (e) {
      setError(e.message)
    } finally {
      setPublishing(false)
    }
  }

  if (published) {
    return (
      <Paper variant="outlined" sx={{ p: 2.5, opacity: 0.6 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <CheckCircleIcon color="success" />
          <Typography variant="subtitle1" fontWeight={600}>
            {draft.name || '(名前なし)'}
          </Typography>
          <Chip label="公開済み" size="small" color="success" />
        </Box>
      </Paper>
    )
  }

  return (
    <Paper variant="outlined" sx={{ overflow: 'hidden' }}>
      <Box
        sx={{
          p: 2.5,
          display: 'flex',
          alignItems: 'center',
          gap: 2,
          cursor: 'pointer',
          '&:hover': { bgcolor: 'action.hover' },
        }}
        onClick={handleExpand}
      >
        <Box sx={{ flexGrow: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography variant="subtitle1" fontWeight={600}>
              {draft.name || '(名前なし)'}
            </Typography>
            {isLocked && <Chip icon={<LockIcon />} label="作業中" size="small" color="warning" variant="outlined" />}
          </Box>
          <Typography variant="body2" color="text.secondary">
            更新: {draft.updatedAt ? fmtDate(draft.updatedAt) : '-'}
          </Typography>
        </Box>

        {isLocked ? (
          <Typography variant="body2" color="warning.main" sx={{ fontWeight: 500, whiteSpace: 'nowrap' }}>
            作業中の為公開不可
          </Typography>
        ) : (
          <Button
            variant="contained"
            color="success"
            size="small"
            startIcon={publishing ? <CircularProgress size={16} color="inherit" /> : <PublishIcon />}
            onClick={e => { e.stopPropagation(); setPublishDialog(true) }}
            disabled={publishing}
            sx={{ minWidth: 100 }}
          >
            {publishing ? '公開中...' : '公開'}
          </Button>
        )}

        {expanded ? <ExpandLessIcon color="action" /> : <ExpandMoreIcon color="action" />}
      </Box>

      <Collapse in={expanded}>
        <Divider />
        <Box sx={{ px: 2.5, py: 2 }}>
          {draft.note && (
            <Typography
              variant="body2"
              sx={{ mb: 2, whiteSpace: 'pre-wrap', color: 'text.secondary', bgcolor: 'action.hover', p: 1.5, borderRadius: 1 }}
            >
              {draft.note}
            </Typography>
          )}
          {loadingPages ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
              <CircularProgress size={20} />
            </Box>
          ) : pages !== null && pages.length === 0 ? (
            <Typography variant="body2" color="text.secondary">
              ページが追加されていません
            </Typography>
          ) : pages !== null ? (
            <List dense disablePadding>
              {pages.map((p) => (
                <ListItem key={p.id} disablePadding sx={{ py: 0.3 }}>
                  <ListItemIcon sx={{ minWidth: 28 }}>
                    <ArticleIcon fontSize="small" color="action" />
                  </ListItemIcon>
                  <ListItemText
                    primary={
                      <Link
                        href={`/manage/page/view/${p.pageId}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        underline="hover"
                        variant="body2"
                      >
                        {p.name || '(名前なし)'}
                      </Link>
                    }
                  />
                  {p.updatedAt && (
                    <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'nowrap', ml: 1 }}>
                      {fmtDate(p.updatedAt)}
                    </Typography>
                  )}
                </ListItem>
              ))}
            </List>
          ) : null}
        </Box>
      </Collapse>

      {error && (
        <Alert severity="error" sx={{ m: 1 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <Dialog open={publishDialog} onClose={() => setPublishDialog(false)}>
        <DialogTitle>ドラフトを公開しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            「{draft.name || '(名前なし)'}」に含まれるページの HTML を生成・公開します。
            公開後このドラフトは削除されます。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPublishDialog(false)}>キャンセル</Button>
          <Button onClick={handlePublish} color="success" variant="contained">公開する</Button>
        </DialogActions>
      </Dialog>
    </Paper>
  )
}

export default function PublishPage() {
  const [drafts, setDrafts] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getDrafts()
      if (!data) return
      setDrafts(data.drafts || [])
      setError(null)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handlePublished = (id) => {
    setSnack({ open: true, message: '公開しました', severity: 'success' })
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5" fontWeight={600} sx={{ flexGrow: 1 }}>公開</Typography>
        <Button
          variant="outlined"
          size="small"
          startIcon={<OpenInNewIcon />}
          component="a"
          href="/manage/page/view/"
          target="_blank"
          rel="noopener noreferrer"
        >
          プレビュー
        </Button>
      </Box>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', pt: 6 }}>
          <CircularProgress />
        </Box>
      ) : drafts.length === 0 ? (
        <Alert severity="info">
          公開待ちのドラフトはありません。
        </Alert>
      ) : (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {drafts.map(d => (
            <DraftCard key={d.id} draft={d} onPublished={handlePublished} />
          ))}
        </Box>
      )}

      <Snackbar
        open={snack.open}
        autoHideDuration={3000}
        onClose={() => setSnack(s => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          severity={snack.severity}
          onClose={() => setSnack(s => ({ ...s, open: false }))}
          sx={{ width: '100%' }}
          elevation={6}
          variant="filled"
        >
          {snack.message}
        </Alert>
      </Snackbar>
    </Box>
  )
}
