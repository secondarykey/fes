import { useState, useEffect, useCallback, useRef } from 'react'
import {
  Typography, Box, Paper, Table, TableHead, TableBody, TableRow, TableCell,
  TableContainer, IconButton, Button, Tooltip, CircularProgress, Alert,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, TextField,
  Chip, Tabs, Tab, Snackbar, LinearProgress, Checkbox,
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import UploadIcon from '@mui/icons-material/Upload'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import EditIcon from '@mui/icons-material/Edit'
import SwapHorizIcon from '@mui/icons-material/SwapHoriz'
import { getFiles, uploadFile, deleteFiles, renameFile, replaceFile, checkFile } from '../api/file'

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
  const [renameDialog, setRenameDialog] = useState(false)
  const [renameTarget, setRenameTarget] = useState(null)
  const [renameName, setRenameName] = useState('')
  const [replaceTarget, setReplaceTarget] = useState(null)
  const replaceInputRef = useRef()
  const [uploadConfirmDialog, setUploadConfirmDialog] = useState(false)
  const [pendingUploadFile, setPendingUploadFile] = useState(null)
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
    try {
      const existing = await checkFile(file.name)
      if (existing) {
        setPendingUploadFile(file)
        setUploadConfirmDialog(true)
        return
      }
    } catch {
      // チェック失敗時はそのままアップロード
    }
    doUpload(file)
  }

  const doUpload = async (file) => {
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

  const handleUploadConfirm = () => {
    setUploadConfirmDialog(false)
    if (pendingUploadFile) {
      doUpload(pendingUploadFile)
      setPendingUploadFile(null)
    }
  }

  const handleUploadCancel = () => {
    setUploadConfirmDialog(false)
    setPendingUploadFile(null)
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

  const openRename = (f) => {
    setRenameTarget(f)
    setRenameName(f.id)
    setRenameDialog(true)
  }

  const handleRename = async () => {
    const name = renameName.trim()
    if (!name || !renameTarget) return
    setRenameDialog(false)
    try {
      await renameFile(renameTarget.id, name)
      showSnack('名前を変更しました')
      load(typeTab)
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const openReplace = (f) => {
    setReplaceTarget(f)
    replaceInputRef.current?.click()
  }

  const handleReplaceFile = async (e) => {
    const file = e.target.files?.[0]
    if (!file || !replaceTarget) return
    e.target.value = ''
    setUploading(true)
    try {
      await replaceFile(replaceTarget.id, file)
      showSnack('ファイルを更新しました')
      load(typeTab)
    } catch (err) {
      showSnack(err.message, 'error')
    } finally {
      setUploading(false)
      setReplaceTarget(null)
    }
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 1 }}>
        <Typography variant="h5" sx={{ fontWeight: 600, flexGrow: 1 }}>Files</Typography>
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
              <TableCell sx={{ width: 80 }} />
            </TableRow>
          </TableHead>
          <TableBody>
            {loading && files.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                  <CircularProgress size={24} />
                </TableCell>
              </TableRow>
            ) : files.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6}>
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
                  <TableCell sx={{ px: 0.5, whiteSpace: 'nowrap' }}>
                    <Tooltip title="名前変更">
                      <IconButton size="small" onClick={e => { e.stopPropagation(); openRename(f) }}>
                        <EditIcon fontSize="inherit" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="ファイル更新">
                      <IconButton size="small" onClick={e => { e.stopPropagation(); openReplace(f) }}>
                        <SwapHorizIcon fontSize="inherit" />
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

      {/* アップロード上書き確認ダイアログ */}
      <Dialog open={uploadConfirmDialog} onClose={handleUploadCancel}>
        <DialogTitle>ファイルの上書き確認</DialogTitle>
        <DialogContent>
          <DialogContentText>
            同一のファイル名「{pendingUploadFile?.name}」が存在します。更新しますか？
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleUploadCancel}>キャンセル</Button>
          <Button onClick={handleUploadConfirm} variant="contained" color="warning">更新</Button>
        </DialogActions>
      </Dialog>

      {/* 名前変更ダイアログ */}
      <Dialog open={renameDialog} onClose={() => setRenameDialog(false)} fullWidth maxWidth="xs">
        <DialogTitle>ファイル名を変更</DialogTitle>
        <DialogContent>
          <TextField
            label="新しいファイル名"
            value={renameName}
            onChange={e => setRenameName(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleRename() }}
            fullWidth
            size="small"
            autoFocus
            sx={{ mt: 1 }}
            slotProps={{ htmlInput: { style: { fontFamily: 'monospace' } } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRenameDialog(false)}>キャンセル</Button>
          <Button onClick={handleRename} variant="contained" disabled={!renameName.trim() || renameName.trim() === renameTarget?.id}>変更</Button>
        </DialogActions>
      </Dialog>

      {/* ファイル更新用の非表示input */}
      <input
        type="file"
        ref={replaceInputRef}
        style={{ display: 'none' }}
        onChange={handleReplaceFile}
      />

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
