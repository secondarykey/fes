import { useState, useEffect, useCallback } from 'react'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import {
  Typography, Box, Paper, Table, TableHead, TableBody, TableRow, TableCell,
  TableContainer, IconButton, Button, Tooltip, CircularProgress, Alert,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions,
  TextField, Snackbar, Alert as MuiAlert, Chip,
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import AddIcon from '@mui/icons-material/Add'
import BookmarkIcon from '@mui/icons-material/Bookmark'
import BookmarkBorderIcon from '@mui/icons-material/BookmarkBorder'
import LockIcon from '@mui/icons-material/Lock'
import LockOpenIcon from '@mui/icons-material/LockOpen'
import { getDrafts, createDraft, deleteDraft, getCurrentDraft, setCurrentDraft, toggleDraftLock } from '../api/draft'

function fmtDate(iso) {
  return new Date(iso).toLocaleString('ja-JP', { dateStyle: 'short', timeStyle: 'short' })
}

export default function DraftList() {
  const [drafts, setDrafts] = useState([])
  const [nextCursor, setNextCursor] = useState('')
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [newDialog, setNewDialog] = useState(false)
  const [newName, setNewName] = useState('')
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const [currentDraftId, setCurrentDraftId] = useState(null)
  const navigate = useNavigate()

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  const load = useCallback(async (cursor = '', append = false) => {
    setLoading(true)
    try {
      const data = await getDrafts(cursor)
      if (!data) return
      setDrafts(prev => append ? [...prev, ...(data.drafts || [])] : (data.drafts || []))
      setNextCursor(data.nextCursor || '')
      setLoadError(null)
    } catch (e) {
      setLoadError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    getCurrentDraft().then(d => setCurrentDraftId(d?.id ?? null)).catch(() => {})
  }, [load])

  const handleCreate = async () => {
    const name = newName.trim()
    if (!name) return
    try {
      const data = await createDraft(name)
      if (data) navigate(`/draft/${data.id}`)
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setNewDialog(false)
      setNewName('')
    }
  }

  const handleSetCurrent = async (d) => {
    try {
      await setCurrentDraft(d.id)
      setCurrentDraftId(d.id)
      showSnack(`「${d.name || '(名前なし)'}」を現在のドラフトに設定しました`)
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const handleToggleLock = async (d) => {
    try {
      const updated = await toggleDraftLock(d.id)
      setDrafts(prev => prev.map(dr => dr.id === d.id ? { ...dr, lock: updated.lock } : dr))
      showSnack(updated.lock ? 'ロックしました' : 'ロックを解除しました')
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return
    try {
      await deleteDraft(deleteTarget.id)
      showSnack('削除しました')
      setDrafts(prev => prev.filter(d => d.id !== deleteTarget.id))
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDeleteTarget(null)
    }
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 1 }}>
        <Typography variant="h5" fontWeight={600} sx={{ flexGrow: 1 }}>Drafts</Typography>
        <Button variant="outlined" color="success" component={RouterLink} to="/publish">
          公開ページ
        </Button>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setNewDialog(true)}>
          新規作成
        </Button>
      </Box>

      {loadError && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setLoadError(null)}>
          {loadError}
        </Alert>
      )}

      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>ドラフト名</TableCell>
              <TableCell sx={{ width: 70, px: 0.5 }} />
              <TableCell align="center" sx={{ width: 150 }}>更新日時</TableCell>
              <TableCell sx={{ width: 40, px: 0.5 }} />
            </TableRow>
          </TableHead>
          <TableBody>
            {loading && drafts.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} align="center" sx={{ py: 4 }}>
                  <CircularProgress size={24} />
                </TableCell>
              </TableRow>
            ) : drafts.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4}>
                  <Alert severity="info" sx={{ border: 'none' }}>ドラフトがありません</Alert>
                </TableCell>
              </TableRow>
            ) : (
              drafts.map(d => (
                <TableRow key={d.id} hover>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <RouterLink to={`/draft/${d.id}`} style={{ textDecoration: 'none' }}>
                        <Typography
                          variant="body2"
                          color="primary"
                          sx={{ '&:hover': { textDecoration: 'underline' } }}
                        >
                          {(d.name || '(名前なし)') + (d.lock ? '（作業中）' : '')}
                        </Typography>
                      </RouterLink>
                      {currentDraftId === d.id && (
                        <Chip label="現在" size="small" color="primary" variant="outlined" />
                      )}
                    </Box>
                  </TableCell>
                  <TableCell sx={{ px: 0.5, whiteSpace: 'nowrap' }}>
                    <Tooltip title={d.lock ? 'ロック解除' : 'ロック'}>
                      <IconButton size="small" onClick={() => handleToggleLock(d)}>
                        {d.lock
                          ? <LockIcon fontSize="small" color="warning" />
                          : <LockOpenIcon fontSize="small" color="action" />}
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={currentDraftId === d.id ? '現在のドラフト' : '現在のドラフトに設定'}>
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={() => handleSetCurrent(d)}
                        disabled={currentDraftId === d.id}
                      >
                        {currentDraftId === d.id
                          ? <BookmarkIcon fontSize="small" />
                          : <BookmarkBorderIcon fontSize="small" />}
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                  <TableCell align="center">
                    <Typography variant="body2" color="text.secondary">
                      {d.updatedAt ? fmtDate(d.updatedAt) : '-'}
                    </Typography>
                  </TableCell>
                  <TableCell sx={{ px: 0.5 }}>
                    <Tooltip title="削除">
                      <IconButton size="small" color="error" onClick={() => setDeleteTarget(d)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {nextCursor && (
        <Box sx={{ mt: 2, display: 'flex', justifyContent: 'center' }}>
          <Button variant="outlined" onClick={() => load(nextCursor, true)} disabled={loading}>
            さらに読み込む
          </Button>
        </Box>
      )}

      {/* 新規作成ダイアログ */}
      <Dialog open={newDialog} onClose={() => { setNewDialog(false); setNewName('') }} fullWidth maxWidth="xs">
        <DialogTitle>新規ドラフトを作成</DialogTitle>
        <DialogContent>
          <TextField
            label="ドラフト名"
            value={newName}
            onChange={e => setNewName(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleCreate() }}
            fullWidth
            size="small"
            autoFocus
            sx={{ mt: 1 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setNewDialog(false); setNewName('') }}>キャンセル</Button>
          <Button onClick={handleCreate} variant="contained" disabled={!newName.trim()}>作成</Button>
        </DialogActions>
      </Dialog>

      {/* 削除確認ダイアログ */}
      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogTitle>ドラフトを削除しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            「{deleteTarget?.name || '(名前なし)'}」を削除します。この操作は取り消せません。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)}>キャンセル</Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained">削除</Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snack.open}
        autoHideDuration={3000}
        onClose={() => setSnack(s => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <MuiAlert severity={snack.severity} onClose={() => setSnack(s => ({ ...s, open: false }))}
          sx={{ width: '100%' }} elevation={6} variant="filled">
          {snack.message}
        </MuiAlert>
      </Snackbar>
    </Box>
  )
}
