import { useState, useEffect } from 'react'
import { useParams, Link as RouterLink } from 'react-router-dom'
import {
  Box, Paper, Typography, TextField, Button,
  Breadcrumbs, Link, Divider, CircularProgress,
  Snackbar, Alert as MuiAlert,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import { getVariable, updateVariable } from '../api/variable'

export default function VariableEdit() {
  const { key } = useParams()
  const [data, setData] = useState(null)
  const [content, setContent] = useState('')
  const [version, setVersion] = useState(0)
  const [saving, setSaving] = useState(false)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  useEffect(() => {
    getVariable(key)
      .then(d => {
        if (!d) return
        setData(d)
        setContent(d.content || '')
        setVersion(d.version || 0)
      })
      .catch(err => showSnack(err.message, 'error'))
  }, [key])

  const handleSave = async () => {
    setSaving(true)
    try {
      const updated = await updateVariable(key, { content, version })
      if (updated && updated.version !== undefined) {
        setVersion(updated.version)
        setData(updated)
      }
      showSnack('保存しました')
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setSaving(false)
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
        <Link component={RouterLink} to="/variable" underline="hover" color="inherit">
          Variables
        </Link>
        <Typography color="text.primary" fontFamily="monospace">{data.id}</Typography>
      </Breadcrumbs>

      <Paper sx={{ p: 3 }}>
        <TextField
          label="変数名（ID）"
          value={data.id}
          fullWidth
          size="small"
          disabled
          inputProps={{ style: { fontFamily: 'monospace' } }}
          sx={{ mb: 2 }}
        />

        <TextField
          label="値"
          value={content}
          onChange={e => setContent(e.target.value)}
          fullWidth
          multiline
          rows={20}
          size="small"
          inputProps={{ style: { fontFamily: 'monospace', fontSize: 13 } }}
          sx={{ mb: 3 }}
        />

        <Divider sx={{ mb: 2 }} />

        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? '保存中...' : '保存'}
          </Button>
          <Button
            variant="text"
            startIcon={<ArrowBackIcon />}
            component={RouterLink}
            to="/variable"
          >
            一覧へ
          </Button>
        </Box>
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
