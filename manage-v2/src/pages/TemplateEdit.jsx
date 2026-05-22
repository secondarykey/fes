import { useState, useEffect } from 'react'
import { useParams, useSearchParams, useNavigate, Link as RouterLink } from 'react-router-dom'
import {
  Box, Paper, Typography, TextField, Select, MenuItem,
  FormControl, InputLabel, Button, Breadcrumbs, Link,
  Divider, CircularProgress, Snackbar, Alert as MuiAlert,
  Dialog, DialogTitle, DialogContent, DialogActions,
  List, ListItem, ListItemText, ListItemIcon, Checkbox,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import HtmlIcon from '@mui/icons-material/Html'
import { newTemplate, getTemplate, updateTemplate, getTemplateReferences, publishPages } from '../api/template'

const TYPE_OPTIONS = [
  { value: 1, label: 'サイトテンプレート' },
  { value: 2, label: 'ページテンプレート' },
]

export default function TemplateEdit() {
  const { key } = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()

  const isNew = !key
  const presetKey = searchParams.get('key') || ''

  const [data, setData] = useState(null)
  const [form, setForm] = useState({ name: '', type: 2, content: '', version: 0 })
  const [saving, setSaving] = useState(false)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })

  const [refDialog, setRefDialog] = useState(false)
  const [refPages, setRefPages] = useState([])
  const [refLoading, setRefLoading] = useState(false)
  const [refSelected, setRefSelected] = useState(new Set())
  const [refPublishing, setRefPublishing] = useState(false)

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  useEffect(() => {
    const load = isNew ? newTemplate() : getTemplate(key)
    load
      .then(d => {
        if (!d) return
        setData(d)
        setForm({
          name: d.name || '',
          type: d.type || 2,
          content: d.content || '',
          version: d.version || 0,
        })
      })
      .catch(err => showSnack(err.message, 'error'))
  }, [key, isNew])

  const handleSave = async () => {
    setSaving(true)
    try {
      const targetKey = isNew ? presetKey : key
      const updated = await updateTemplate(targetKey, form)
      if (updated) {
        setData(updated)
        setForm(f => ({ ...f, version: updated.version }))
      }
      showSnack('保存しました')
      if (isNew) navigate(`/template/${targetKey}`, { replace: true })
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleOpenRefDialog = async () => {
    setRefDialog(true)
    setRefLoading(true)
    try {
      const res = await getTemplateReferences(key)
      const pages = res?.pages || []
      setRefPages(pages)
      setRefSelected(new Set(pages.map(p => p.id)))
    } catch (e) {
      showSnack(e.message, 'error')
      setRefDialog(false)
    } finally {
      setRefLoading(false)
    }
  }

  const toggleRefPage = (id) => {
    setRefSelected(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const handlePublishRefs = async () => {
    setRefPublishing(true)
    try {
      const res = await publishPages([...refSelected])
      showSnack(`${res.count} 件のHTMLを更新しました`)
      setRefDialog(false)
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setRefPublishing(false)
    }
  }

  if (!data) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', pt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ maxWidth: 960 }}>
      <Breadcrumbs sx={{ mb: 2 }}>
        <Link component={RouterLink} to="/template" underline="hover" color="inherit">
          Templates
        </Link>
        <Typography color="text.primary">{form.name || '(新規)'}</Typography>
      </Breadcrumbs>

      <Paper sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', gap: 2, mb: 2, flexWrap: 'wrap' }}>
          <TextField
            label="テンプレート名"
            value={form.name}
            onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
            size="small"
            sx={{ flex: 1, minWidth: 200 }}
          />
          <FormControl size="small" sx={{ minWidth: 200 }}>
            <InputLabel>種別</InputLabel>
            <Select
              value={form.type}
              label="種別"
              onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
            >
              {TYPE_OPTIONS.map(o => (
                <MenuItem key={o.value} value={o.value}>{o.label}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </Box>

        <TextField
          label="テンプレート内容"
          value={form.content}
          onChange={e => setForm(f => ({ ...f, content: e.target.value }))}
          fullWidth
          multiline
          rows={30}
          size="small"
          inputProps={{ style: { fontFamily: 'monospace', fontSize: 13 } }}
          sx={{ mb: 3 }}
        />

        <Divider sx={{ mb: 2 }} />

        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? '保存中...' : '保存'}
          </Button>
          {!isNew && (
            <Button
              variant="outlined"
              startIcon={<HtmlIcon />}
              onClick={handleOpenRefDialog}
            >
              参照ページのHTML更新
            </Button>
          )}
          <Button
            variant="text"
            startIcon={<ArrowBackIcon />}
            component={RouterLink}
            to="/template"
          >
            一覧へ
          </Button>
        </Box>
      </Paper>

      {/* 参照ページHTML更新ダイアログ */}
      <Dialog open={refDialog} onClose={() => setRefDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>参照ページのHTML更新</DialogTitle>
        <DialogContent dividers>
          {refLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          ) : refPages.length === 0 ? (
            <Typography color="text.secondary">このテンプレートを参照しているページはありません。</Typography>
          ) : (
            <List dense disablePadding>
              {refPages.map(p => (
                <ListItem
                  key={p.id}
                  button
                  onClick={() => toggleRefPage(p.id)}
                  sx={{ px: 0 }}
                >
                  <ListItemIcon sx={{ minWidth: 36 }}>
                    <Checkbox
                      size="small"
                      checked={refSelected.has(p.id)}
                      onChange={() => toggleRefPage(p.id)}
                      onClick={e => e.stopPropagation()}
                    />
                  </ListItemIcon>
                  <ListItemText
                    primary={
                      <Typography variant="body2">{p.name}</Typography>
                    }
                    secondary={p.id}
                  />
                </ListItem>
              ))}
            </List>
          )}
        </DialogContent>
        <DialogActions>
          <Typography variant="body2" color="text.secondary" sx={{ flex: 1, ml: 1 }}>
            {refSelected.size} 件選択中
          </Typography>
          <Button onClick={() => setRefDialog(false)}>キャンセル</Button>
          <Button
            variant="contained"
            onClick={handlePublishRefs}
            disabled={refPublishing || refSelected.size === 0}
            startIcon={refPublishing ? <CircularProgress size={14} /> : <HtmlIcon />}
          >
            {refPublishing ? '更新中...' : 'HTML更新'}
          </Button>
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
