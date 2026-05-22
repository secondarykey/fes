import { useState, useEffect, useCallback, useRef } from 'react'
import {
  Typography, Box, Paper, Table, TableHead, TableBody, TableRow, TableCell,
  TableContainer, IconButton, Button, Tooltip, CircularProgress, Alert,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions,
  Chip, Tabs, Tab, Snackbar, Alert as MuiAlert, LinearProgress, Checkbox,
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import UploadIcon from '@mui/icons-material/Upload'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import { getFiles, uploadFile, deleteFiles } from '../api/file'

const TYPE_LABELS = { 1: 'データ', 2: 'ページ画像', 3: 'システム' }
const TYPE_COLORS = { 1: 'default', 2: 'primary', 3: 'warning' }

function fmtSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function fmtDate(iso) {
  return new Date(iso).toLocaleString('ja-JP', { dateStyle: 'short', timeStyle: 'short' })
}

export default function FileList() {
  const [files, setFiles] = useState([])
  const [nextCursor, setNextCursor] = useState('')
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(null)
  const [uploading, setUploading] = useState(false)
  const [typeTab, setTypeTab] = useState('1') // '' | '1' | '2'
  const [selected, setSelected] = useState(new Set())
  const [deleteDialog, setDeleteDialog] = useState(false)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const fileInputRef = useRef()

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  const load = useCallback(async (type, cursor = '', append = false) => {
    setLoading(true)
    try {
      const data = await getFiles(type, cursor)
      if (!data) return
      setFiles(prev => append ? [...prev, ...data.files] : (data.files || []))
      setNextCursor(data.nextCursor || '')
      setLoadError(null)
    } catch (e) {
      setLoadError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load(typeTab)
  }, [typeTab, load])

  const handleTabChange = (_, v) => {
    setFiles([])
    setNextCursor('')
    setSelected(new Set())
    setTypeTab(v)
  }

  const toggleSelect = (id) => {
    setSelected(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const handleUpload = async (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    e.target.value = ''
    setUploading(true)
    try {
      await uploadFile(file)
      showSnack('アップロードしました')
      load(typeTab)
    } catch (err) {
      showSnack(err.message, 'error')
    } finally {
      setUploading(false)
    }
  }

  const handleDeleteConfirm = async () => {
    try {
      await deleteFiles([...selected])
      showSnack('削除しました')
      setFiles(prev => prev.filter(f => !selected.has(f.id)))
      setSelected(new Set())
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDeleteDialog(false)
    }
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 1 }}>
        <Typography variant="h5" fontWeight={600} sx={{ flexGrow: 1 }}>Files</Typography>
        <input
          type="file"
          ref={fileInputRef}
          style={{ display: 'none' }}
          onChange={handleUpload}
        />
        <Button
          variant="contained"
          startIcon={<UploadIcon />}
          onClick={() => fileInputRef.current?.click()}
          disabled={uploading}
        >
          {uploading ? 'アップロード中...' : 'アップロード'}
        </Button>
      </Box>

      <Tabs value={typeTab} onChange={handleTabChange} sx={{ mb: 2 }}>
        <Tab label="すべて" value="" />
        <Tab label="データ" value="1" />
        <Tab label="ページ画像" value="2" />
      </Tabs>

      {uploading && <LinearProgress sx={{ mb: 1 }} />}

      {loadError && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setLoadError(null)}>
          {loadError}
        </Alert>
      )}

      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell sx={{ width: 40, px: 1 }} />
              <TableCell>ファイル名</TableCell>
              <TableCell align="center">種別</TableCell>
              <TableCell align="right">サイズ</TableCell>
              <TableCell align="center">更新日時</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading && files.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                  <CircularProgress size={24} />
                </TableCell>
              </TableRow>
            ) : files.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5}>
                  <Alert severity="info" sx={{ border: 'none' }}>ファイルがありません</Alert>
                </TableCell>
              </TableRow>
            ) : (
              files.map(f => (
                <TableRow
                  key={f.id}
                  hover
                  selected={selected.has(f.id)}
                  onClick={() => toggleSelect(f.id)}
                  sx={{ cursor: 'pointer' }}
                >
                  <TableCell sx={{ px: 1 }}>
                    <Checkbox
                      size="small"
                      checked={selected.has(f.id)}
                      onChange={() => toggleSelect(f.id)}
                      onClick={e => e.stopPropagation()}
                    />
                  </TableCell>
                  <TableCell sx={{ maxWidth: 320, wordBreak: 'break-all' }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      <Typography variant="body2">{f.id}</Typography>
                      <Tooltip title="プレビュー">
                        <IconButton
                          size="small"
                          component="a"
                          href={`/file/${encodeURIComponent(f.id)}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          onClick={e => e.stopPropagation()}
                        >
                          <OpenInNewIcon fontSize="inherit" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                  <TableCell align="center">
                    <Chip
                      label={TYPE_LABELS[f.type] || `type${f.type}`}
                      size="small"
                      color={TYPE_COLORS[f.type] || 'default'}
                      variant="outlined"
                    />
                  </TableCell>
                  <TableCell align="right">
                    <Typography variant="body2" color="text.secondary">
                      {fmtSize(f.size)}
                    </Typography>
                  </TableCell>
                  <TableCell align="center">
                    <Typography variant="body2" color="text.secondary">
                      {fmtDate(f.updatedAt)}
                    </Typography>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {nextCursor && (
        <Box sx={{ mt: 2, display: 'flex', justifyContent: 'center' }}>
          <Button
            variant="outlined"
            onClick={() => load(typeTab, nextCursor, true)}
            disabled={loading}
          >
            さらに読み込む
          </Button>
        </Box>
      )}

      {(() => {
        const selectedFiles = files.filter(f => selected.has(f.id))
        const totalSize = selectedFiles.reduce((sum, f) => sum + f.size, 0)
        return (
          <>
            <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end' }}>
              <Button
                variant="outlined"
                color="error"
                startIcon={<DeleteIcon />}
                disabled={selected.size === 0}
                onClick={() => setDeleteDialog(true)}
              >
                削除（{selected.size}件 / {fmtSize(totalSize)}）
              </Button>
            </Box>

            {/* 削除確認ダイアログ */}
            <Dialog open={deleteDialog} onClose={() => setDeleteDialog(false)}>
              <DialogTitle>ファイルを削除しますか？</DialogTitle>
              <DialogContent>
                <DialogContentText>
                  選択中の {selected.size} 件（{fmtSize(totalSize)}）を削除します。この操作は取り消せません。
                </DialogContentText>
              </DialogContent>
              <DialogActions>
                <Button onClick={() => setDeleteDialog(false)}>キャンセル</Button>
                <Button onClick={handleDeleteConfirm} color="error" variant="contained">削除</Button>
              </DialogActions>
            </Dialog>
          </>
        )
      })()}

      <Snackbar
        open={snack.open}
        autoHideDuration={3000}
        onClose={() => setSnack(s => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <MuiAlert
          severity={snack.severity}
          onClose={() => setSnack(s => ({ ...s, open: false }))}
          sx={{ width: '100%' }}
          elevation={6}
          variant="filled"
        >
          {snack.message}
        </MuiAlert>
      </Snackbar>
    </Box>
  )
}
