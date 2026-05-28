import { useState, useEffect, useRef } from 'react'
import {
  Box, Paper, Typography, TextField, Button, Divider,
  CircularProgress, Snackbar, Alert as MuiAlert,
  IconButton, List, ListItem, ListItemText, InputAdornment,
  Card, CardContent, CardActions, Chip, Grid, LinearProgress,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import DownloadIcon from '@mui/icons-material/Download'
import MemoryIcon from '@mui/icons-material/Memory'
import CleaningServicesIcon from '@mui/icons-material/CleaningServices'
import RestoreIcon from '@mui/icons-material/Restore'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import LightModeIcon from '@mui/icons-material/LightMode'
import { getSite, updateSite } from '../api/site'
import { useColorMode } from '../context/ColorMode'

async function toolFetch(path, options = {}) {
  const res = await fetch(`/manage/api${path}`, options)
  if (!res.ok) {
    const b = await res.json().catch(() => ({}))
    throw new Error(b.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export default function SiteEdit() {
  const { mode, toggleColorMode } = useColorMode()
  const isDark = mode === 'dark'
  const [form, setForm] = useState(null)
  const [newManager, setNewManager] = useState('')
  const [newArchiveName, setNewArchiveName] = useState('')
  const [saving, setSaving] = useState(false)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })

  // Tools state
  const [gcDialog, setGcDialog] = useState(false)
  const [gcRunning, setGcRunning] = useState(false)
  const [gcResult, setGcResult] = useState(null)
  const [gcError, setGcError] = useState(null)
  const [cleanDialog, setCleanDialog] = useState(false)
  const [cleanRunning, setCleanRunning] = useState(false)
  const [cleanResult, setCleanResult] = useState(null)
  const [cleanError, setCleanError] = useState(null)
  const [restoreDialog, setRestoreDialog] = useState(false)
  const [restoreRunning, setRestoreRunning] = useState(false)
  const [restoreResult, setRestoreResult] = useState(null)
  const [restoreError, setRestoreError] = useState(null)
  const [restoreFile, setRestoreFile] = useState(null)
  const fileInputRef = useRef(null)

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  useEffect(() => {
    getSite()
      .then(d => { if (d) setForm(d) })
      .catch(err => showSnack(err.message, 'error'))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      const updated = await updateSite(form)
      if (updated && updated.version !== undefined) {
        setForm(updated)
      }
      showSnack('保存しました')
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleAddManager = () => {
    const email = newManager.trim()
    if (!email) return
    if (form.managers.includes(email)) {
      setNewManager('')
      return
    }
    setForm(f => ({ ...f, managers: [...f.managers, email] }))
    setNewManager('')
  }

  const handleRemoveManager = (email) => {
    setForm(f => ({ ...f, managers: f.managers.filter(m => m !== email) }))
  }

  const handleAddArchiveName = () => {
    const name = newArchiveName.trim()
    if (!name) return
    if (form.archiveNames.includes(name)) {
      setNewArchiveName('')
      return
    }
    setForm(f => ({ ...f, archiveNames: [...f.archiveNames, name] }))
    setNewArchiveName('')
  }

  const handleRemoveArchiveName = (name) => {
    setForm(f => ({ ...f, archiveNames: f.archiveNames.filter(n => n !== name) }))
  }

  const handleGC = async () => {
    setGcDialog(false); setGcRunning(true); setGcResult(null); setGcError(null)
    try { const d = await toolFetch('/tool/gc', { method: 'POST' }); setGcResult(d.result || 'GC 完了') }
    catch (e) { setGcError(e.message) } finally { setGcRunning(false) }
  }

  const handleClean = async () => {
    setCleanDialog(false); setCleanRunning(true); setCleanResult(null); setCleanError(null)
    try { setCleanResult(await toolFetch('/tool/clean', { method: 'POST' })) }
    catch (e) { setCleanError(e.message) } finally { setCleanRunning(false) }
  }

  const handleRestoreFileChange = (e) => {
    const f = e.target.files?.[0] ?? null; setRestoreFile(f); if (f) setRestoreDialog(true)
  }

  const handleRestore = async () => {
    setRestoreDialog(false); setRestoreRunning(true); setRestoreResult(null); setRestoreError(null)
    try {
      const fd = new FormData(); fd.append('file', restoreFile)
      const res = await fetch('/manage/api/tool/restore', { method: 'POST', body: fd })
      if (!res.ok) { const b = await res.json().catch(() => ({})); throw new Error(b.error || `HTTP ${res.status}`) }
      setRestoreResult('リストアが完了しました。')
    } catch (e) { setRestoreError(e.message) }
    finally { setRestoreRunning(false); setRestoreFile(null); if (fileInputRef.current) fileInputRef.current.value = '' }
  }

  if (!form) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', pt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ maxWidth: 720 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
        <Typography variant="h5" fontWeight={600}>Site Settings</Typography>
        <IconButton size="small" onClick={toggleColorMode} sx={{ color: isDark ? 'grey.400' : 'grey.600' }}>
          {isDark ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
        </IconButton>
      </Box>

      <Paper sx={{ p: 3 }}>
        <TextField
          label="サイト名"
          value={form.name}
          onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
          fullWidth
          size="small"
          sx={{ mb: 2 }}
        />

        <TextField
          label="説明"
          value={form.description}
          onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
          fullWidth
          multiline
          rows={3}
          size="small"
          sx={{ mb: 2 }}
        />

        <TextField
          label="ルートページ ID"
          value={form.root}
          onChange={e => setForm(f => ({ ...f, root: e.target.value }))}
          fullWidth
          size="small"
          sx={{ mb: 2 }}
        />

        <TextField
          label="管理 URL"
          value={form.manageURL}
          onChange={e => setForm(f => ({ ...f, manageURL: e.target.value }))}
          fullWidth
          size="small"
          sx={{ mb: 3 }}
        />

        <Divider sx={{ mb: 2 }} />
        <Typography variant="subtitle1" fontWeight={600} gutterBottom>Archive</Typography>

        <TextField
          label="GCS バケット名"
          value={form.archiveBucket || ''}
          onChange={e => setForm(f => ({ ...f, archiveBucket: e.target.value }))}
          fullWidth
          size="small"
          placeholder="hummingbird-archives"
          sx={{ mb: 2 }}
        />

        <Typography variant="subtitle2" gutterBottom>アーカイブ名一覧</Typography>
        <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, mb: 3 }}>
          {(form.archiveNames || []).length === 0 ? (
            <Typography variant="body2" color="text.disabled" sx={{ p: 1.5 }}>
              アーカイブ名が登録されていません
            </Typography>
          ) : (
            <List dense disablePadding>
              {form.archiveNames.map((n, i) => (
                <ListItem
                  key={n}
                  divider={i < form.archiveNames.length - 1}
                  secondaryAction={
                    <IconButton size="small" color="error" onClick={() => handleRemoveArchiveName(n)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  }
                >
                  <ListItemText primary={n} />
                </ListItem>
              ))}
            </List>
          )}
          <Divider />
          <Box sx={{ p: 1 }}>
            <TextField
              placeholder="アーカイブ名を追加（例: 2026-Fall）"
              value={newArchiveName}
              onChange={e => setNewArchiveName(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); handleAddArchiveName() } }}
              size="small"
              fullWidth
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton size="small" onClick={handleAddArchiveName} color="primary">
                      <AddIcon fontSize="small" />
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />
          </Box>
        </Box>

        <Divider sx={{ mb: 2 }} />
        <Typography variant="subtitle2" gutterBottom>管理者</Typography>
        <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, mb: 3 }}>
          {form.managers.length === 0 ? (
            <Typography variant="body2" color="text.disabled" sx={{ p: 1.5 }}>
              管理者が登録されていません
            </Typography>
          ) : (
            <List dense disablePadding>
              {form.managers.map((m, i) => (
                <ListItem
                  key={m}
                  divider={i < form.managers.length - 1}
                  secondaryAction={
                    <IconButton size="small" color="error" onClick={() => handleRemoveManager(m)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  }
                >
                  <ListItemText primary={m} />
                </ListItem>
              ))}
            </List>
          )}
          <Divider />
          <Box sx={{ p: 1 }}>
            <TextField
              placeholder="メールアドレスを追加"
              value={newManager}
              onChange={e => setNewManager(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); handleAddManager() } }}
              size="small"
              fullWidth
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton size="small" onClick={handleAddManager} color="primary">
                      <AddIcon fontSize="small" />
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />
          </Box>
        </Box>

        <Divider sx={{ mb: 2 }} />

        <Button
          variant="contained"
          startIcon={<SaveIcon />}
          onClick={handleSave}
          disabled={saving}
        >
          {saving ? '保存中...' : '保存'}
        </Button>
      </Paper>

      {/* Tools */}
      <Typography variant="h6" fontWeight={600} sx={{ mt: 4, mb: 2 }}>Tools</Typography>
      <Grid container spacing={2}>
        <Grid item xs={12} sm={6}>
          <Card variant="outlined" sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
            <CardContent sx={{ flexGrow: 1 }}>
              <Typography variant="subtitle1" fontWeight={600} sx={{ mb: 0.5 }}>サイトクリーン</Typography>
              <Typography variant="body2" color="text.secondary">ページが存在しない孤立した HTML・ページ画像を検出して削除します。</Typography>
            </CardContent>
            <Divider />
            <CardActions sx={{ p: 1.5 }}>
              <Button size="small" variant="outlined" color="warning"
                startIcon={cleanRunning ? <CircularProgress size={14} /> : <CleaningServicesIcon />}
                onClick={() => setCleanDialog(true)} disabled={cleanRunning}>
                {cleanRunning ? '実行中...' : 'クリーン実行'}
              </Button>
            </CardActions>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6}>
          <Card variant="outlined" sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
            <CardContent sx={{ flexGrow: 1 }}>
              <Typography variant="subtitle1" fontWeight={600} sx={{ mb: 0.5 }}>ガベージコレクション</Typography>
              <Typography variant="body2" color="text.secondary">Go ランタイムの GC を手動実行します。</Typography>
            </CardContent>
            <Divider />
            <CardActions sx={{ p: 1.5 }}>
              <Button size="small" variant="outlined" color="warning"
                startIcon={gcRunning ? <CircularProgress size={14} /> : <MemoryIcon />}
                onClick={() => setGcDialog(true)} disabled={gcRunning}>
                {gcRunning ? '実行中...' : 'GC 実行'}
              </Button>
            </CardActions>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6}>
          <Card variant="outlined" sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
            <CardContent sx={{ flexGrow: 1 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                <Typography variant="subtitle1" fontWeight={600}>データ管理</Typography>
                <Chip label="注意" size="small" color="warning" variant="outlined" />
              </Box>
              <Typography variant="body2" color="text.secondary">Datastore のバックアップ（ZIP）とリストアを行います。</Typography>
            </CardContent>
            <Divider />
            <CardActions sx={{ p: 1.5, gap: 0.5 }}>
              <Button size="small" variant="outlined" startIcon={<DownloadIcon />} component="a" href="/manage/api/tool/backup">
                バックアップ
              </Button>
              <Button size="small" variant="outlined" color="warning"
                startIcon={restoreRunning ? <CircularProgress size={14} /> : <RestoreIcon />}
                onClick={() => fileInputRef.current?.click()} disabled={restoreRunning}>
                {restoreRunning ? '実行中...' : 'リストア'}
              </Button>
              <input ref={fileInputRef} type="file" accept=".zip" style={{ display: 'none' }} onChange={handleRestoreFileChange} />
            </CardActions>
          </Card>
        </Grid>
      </Grid>

      {restoreRunning && <LinearProgress sx={{ mt: 2 }} />}
      {restoreResult && <MuiAlert severity="success" sx={{ mt: 2 }} onClose={() => setRestoreResult(null)}>{restoreResult}</MuiAlert>}
      {restoreError && <MuiAlert severity="error" sx={{ mt: 2 }} onClose={() => setRestoreError(null)}>{restoreError}</MuiAlert>}
      {cleanResult && (
        <MuiAlert severity="success" sx={{ mt: 2 }} onClose={() => setCleanResult(null)}>
          HTML削除: {cleanResult.deletedHtmls?.length ?? 0} 件 / 画像削除: {cleanResult.deletedImages?.length ?? 0} 件
          {(cleanResult.deletedHtmls?.length === 0 && cleanResult.deletedImages?.length === 0) && ' — 孤立データはありませんでした。'}
        </MuiAlert>
      )}
      {cleanError && <MuiAlert severity="error" sx={{ mt: 2 }} onClose={() => setCleanError(null)}>{cleanError}</MuiAlert>}
      {gcResult && (
        <MuiAlert severity="success" sx={{ mt: 2 }} onClose={() => setGcResult(null)}>
          <Typography variant="body2" component="pre" sx={{ fontFamily: 'monospace', whiteSpace: 'pre-wrap', m: 0 }}>{gcResult}</Typography>
        </MuiAlert>
      )}
      {gcError && <MuiAlert severity="error" sx={{ mt: 2 }} onClose={() => setGcError(null)}>{gcError}</MuiAlert>}

      {/* クリーン確認 */}
      <Dialog open={cleanDialog} onClose={() => setCleanDialog(false)}>
        <DialogTitle>サイトクリーンを実行しますか？</DialogTitle>
        <DialogContent><DialogContentText>孤立した HTML・ページ画像を削除します。取り消せません。</DialogContentText></DialogContent>
        <DialogActions>
          <Button onClick={() => setCleanDialog(false)}>キャンセル</Button>
          <Button onClick={handleClean} color="warning" variant="contained">実行</Button>
        </DialogActions>
      </Dialog>

      {/* GC 確認 */}
      <Dialog open={gcDialog} onClose={() => setGcDialog(false)}>
        <DialogTitle>GC を実行しますか？</DialogTitle>
        <DialogContent><DialogContentText>Go ランタイムの GC を手動実行します。</DialogContentText></DialogContent>
        <DialogActions>
          <Button onClick={() => setGcDialog(false)}>キャンセル</Button>
          <Button onClick={handleGC} color="warning" variant="contained">実行</Button>
        </DialogActions>
      </Dialog>

      {/* リストア確認 */}
      <Dialog open={restoreDialog} onClose={() => { setRestoreDialog(false); setRestoreFile(null); if (fileInputRef.current) fileInputRef.current.value = '' }}>
        <DialogTitle>リストアを実行しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            <strong>{restoreFile?.name}</strong> で全データを上書きします。取り消せません。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setRestoreDialog(false); setRestoreFile(null); if (fileInputRef.current) fileInputRef.current.value = '' }}>キャンセル</Button>
          <Button onClick={handleRestore} color="error" variant="contained">リストア実行</Button>
        </DialogActions>
      </Dialog>

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
