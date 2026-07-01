import { useState, useEffect, useCallback } from 'react'
import { Link as RouterLink, useNavigate } from 'react-router'
import {
  Typography, Box, Paper, Table, TableHead, TableBody, TableRow, TableCell,
  TableContainer, IconButton, Button, Tooltip, CircularProgress, Alert,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions,
  TextField, Snackbar,
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import AddIcon from '@mui/icons-material/Add'
import { getVariables, createVariable, deleteVariable } from '../api/variable'

function fmtDate(iso) {
  return new Date(iso).toLocaleString('ja-JP', { dateStyle: 'short', timeStyle: 'short' })
}

export default function VariableList() {
  const [variables, setVariables] = useState([])
  const [nextCursor, setNextCursor] = useState('')
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [newDialog, setNewDialog] = useState(false)
  const [newId, setNewId] = useState('')
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const navigate = useNavigate()

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  const load = useCallback(async (cursor = '', append = false) => {
    setLoading(true)
    try {
      const data = await getVariables(cursor)
      if (!data) return
      setVariables(prev => append ? [...prev, ...(data.variables || [])] : (data.variables || []))
      setNextCursor(data.nextCursor || '')
      setLoadError(null)
    } catch (e) {
      setLoadError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleCreate = async () => {
    const id = newId.trim()
    if (!id) return
    try {
      const data = await createVariable(id, '')
      if (data) navigate(`/variable/${encodeURIComponent(data.id)}`)
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setNewDialog(false)
      setNewId('')
    }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return
    try {
      await deleteVariable(deleteTarget)
      showSnack('削除しました')
      setVariables(prev => prev.filter(v => v.id !== deleteTarget))
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDeleteTarget(null)
    }
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 1 }}>
        <Typography variant="h5" fontWeight={600} sx={{ flexGrow: 1 }}>Variables</Typography>
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
              <TableCell>変数名（ID）</TableCell>
              <TableCell>更新日時</TableCell>
              <TableCell align="center" sx={{ width: 64 }}>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading && variables.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3} align="center" sx={{ py: 4 }}>
                  <CircularProgress size={24} />
                </TableCell>
              </TableRow>
            ) : variables.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3}>
                  <Alert severity="info" sx={{ border: 'none' }}>変数がありません</Alert>
                </TableCell>
              </TableRow>
            ) : (
              variables.map(v => (
                <TableRow key={v.id} hover>
                  <TableCell>
                    <RouterLink
                      to={`/variable/${encodeURIComponent(v.id)}`}
                      style={{ textDecoration: 'none' }}
                    >
                      <Typography
                        variant="body2"
                        color="primary"
                        fontFamily="monospace"
                        sx={{ '&:hover': { textDecoration: 'underline' } }}
                      >
                        {v.id}
                      </Typography>
                    </RouterLink>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      {v.updatedAt ? fmtDate(v.updatedAt) : '-'}
                    </Typography>
                  </TableCell>
                  <TableCell align="center">
                    <Tooltip title="削除">
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => setDeleteTarget(v.id)}
                      >
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
      <Dialog open={newDialog} onClose={() => { setNewDialog(false); setNewId('') }} fullWidth maxWidth="xs">
        <DialogTitle>新規変数を作成</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            テンプレート内で参照するキー名を入力してください。
          </DialogContentText>
          <TextField
            label="変数名（ID）"
            value={newId}
            onChange={e => setNewId(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleCreate() }}
            fullWidth
            size="small"
            autoFocus
            inputProps={{ style: { fontFamily: 'monospace' } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setNewDialog(false); setNewId('') }}>キャンセル</Button>
          <Button onClick={handleCreate} variant="contained" disabled={!newId.trim()}>作成</Button>
        </DialogActions>
      </Dialog>

      {/* 削除確認ダイアログ */}
      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogTitle>変数を削除しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ fontFamily: 'monospace' }}>
            「{deleteTarget}」を削除します。この操作は取り消せません。
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
