import { useState, useEffect } from 'react'
import {
  Box, Paper, Typography, TextField, Button, Divider,
  CircularProgress, Snackbar, Alert as MuiAlert,
  IconButton, List, ListItem, ListItemText, InputAdornment,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import { getSite, updateSite } from '../api/site'

export default function SiteEdit() {
  const [form, setForm] = useState(null)
  const [newManager, setNewManager] = useState('')
  const [newArchiveName, setNewArchiveName] = useState('')
  const [saving, setSaving] = useState(false)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })

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

  if (!form) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', pt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ maxWidth: 720 }}>
      <Typography variant="h5" fontWeight={600} gutterBottom>Site Settings</Typography>

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
