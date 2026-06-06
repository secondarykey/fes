import { useState, useEffect } from 'react'
import { useParams, Link as RouterLink, useNavigate } from 'react-router-dom'
import PropTypes from 'prop-types'
import {
  Box, Paper, Typography, TextField, Button, Breadcrumbs, Link,
  Divider, CircularProgress, Snackbar, Alert,
  Table, TableHead, TableBody, TableRow, TableCell, TableContainer,
  Checkbox, IconButton, Tooltip, Dialog, DialogTitle, DialogContent,
  DialogContentText, DialogActions, Chip, FormControlLabel, Switch,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import PublishIcon from '@mui/icons-material/Publish'
import DeleteIcon from '@mui/icons-material/Delete'
import DragIndicatorIcon from '@mui/icons-material/DragIndicator'
import {
  DndContext, closestCenter, PointerSensor, useSensor, useSensors,
} from '@dnd-kit/core'
import {
  SortableContext, verticalListSortingStrategy, useSortable, arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { getDraft, updateDraft, publishDraft, removeDraftPage } from '../api/draft'

SortableDraftRow.propTypes = { p: PropTypes.object.isRequired, idx: PropTypes.number.isRequired, onToggle: PropTypes.func, onRemove: PropTypes.func }
function SortableDraftRow({ p, idx, onToggle, onRemove }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: p.id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    background: isDragging ? 'rgba(0,0,0,0.04)' : undefined,
  }
  return (
    <TableRow ref={setNodeRef} style={style} hover>
      <TableCell sx={{ width: 40, cursor: 'grab', color: 'text.disabled', px: 1 }} {...attributes} {...listeners}>
        <DragIndicatorIcon fontSize="small" />
      </TableCell>
      <TableCell sx={{ width: 40 }}>
        <Typography variant="body2" color="text.secondary">{idx + 1}</Typography>
      </TableCell>
      <TableCell>
        <Link component={RouterLink} to={`/page/${p.pageId}`} underline="hover" variant="body2">
          {p.name || '(名前なし)'}
        </Link>
      </TableCell>
      <TableCell sx={{ width: 130 }} align="center">
        <Checkbox size="small" checked={p.publishUpdate} onChange={() => onToggle(p.id)} />
      </TableCell>
      <TableCell align="center" sx={{ width: 56 }}>
        <Tooltip title="このドラフトから削除">
          <IconButton size="small" color="error" onClick={() => onRemove(p.id)}>
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      </TableCell>
    </TableRow>
  )
}

export default function DraftEdit() {
  const { key } = useParams()
  const navigate = useNavigate()

  const [draft, setDraft] = useState(null)
  const [pages, setPages] = useState([])
  const [name, setName] = useState('')
  const [note, setNote] = useState('')
  const [lock, setLock] = useState(false)
  const [saving, setSaving] = useState(false)
  const [publishing, setPublishing] = useState(false)
  const [publishDialog, setPublishDialog] = useState(false)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const sensors = useSensors(useSensor(PointerSensor))

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  const handleDragEnd = ({ active, over }) => {
    if (!over || active.id === over.id) return
    setPages(prev => {
      const oldIdx = prev.findIndex(p => p.id === active.id)
      const newIdx = prev.findIndex(p => p.id === over.id)
      return arrayMove(prev, oldIdx, newIdx)
    })
  }

  useEffect(() => {
    getDraft(key)
      .then(d => {
        if (!d) return
        setDraft(d.draft)
        setPages(d.pages || [])
        setName(d.draft.name || '')
        setNote(d.draft.note || '')
        setLock(d.draft.lock || false)
      })
      .catch(err => showSnack(err.message, 'error'))
  }, [key])

  const handleSave = async () => {
    setSaving(true)
    try {
      const body = {
        name,
        note,
        lock,
        version: draft.version,
        pages: pages.map(p => ({ id: p.id, publishUpdate: p.publishUpdate })),
      }
      const res = await updateDraft(key, body)
      if (res?.draft) {
        setDraft(res.draft)
        setPages(res.pages || [])
        setName(res.draft.name || '')
        setNote(res.draft.note || '')
        setLock(res.draft.lock || false)
      }
      showSnack('保存しました')
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  const handlePublish = async () => {
    setPublishDialog(false)
    setPublishing(true)
    try {
      await publishDraft(key)
      showSnack('公開しました')
      setTimeout(() => navigate('/draft'), 1500)
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setPublishing(false)
    }
  }

  const handleRemovePage = async (pageId) => {
    try {
      await removeDraftPage(key, pageId)
      setPages(prev => prev.filter(p => p.id !== pageId))
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const togglePublishUpdate = (id) => {
    setPages(prev => prev.map(p =>
      p.id === id ? { ...p, publishUpdate: !p.publishUpdate } : p
    ))
  }

  if (!draft) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', pt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ maxWidth: 960 }}>
      <Breadcrumbs sx={{ mb: 2 }}>
        <Link component={RouterLink} to="/draft" underline="hover" color="inherit">Drafts</Link>
        <Typography color="text.primary">{name || '(名前なし)'}</Typography>
      </Breadcrumbs>

      <Paper sx={{ p: 3, mb: 3 }}>
        <TextField
          label="ドラフト名"
          value={name}
          onChange={e => setName(e.target.value)}
          fullWidth
          size="small"
          sx={{ mb: 2 }}
        />

        <TextField
          label="説明（Note）"
          value={note}
          onChange={e => setNote(e.target.value)}
          fullWidth
          size="small"
          multiline
          minRows={2}
          maxRows={8}
          sx={{ mb: 2 }}
        />

        <FormControlLabel
          control={<Switch checked={lock} onChange={e => setLock(e.target.checked)} />}
          label="ロック（作業中 — 公開を禁止）"
          sx={{ mb: 3 }}
        />

        <Typography variant="subtitle2" sx={{ mb: 1 }}>
          ページ一覧 <Chip label={pages.length} size="small" sx={{ ml: 0.5 }} />
        </Typography>

        <TableContainer variant="outlined" component={Paper} sx={{ mb: 3 }}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell sx={{ width: 40 }} />
                <TableCell sx={{ width: 40 }}>順</TableCell>
                <TableCell>ページ名</TableCell>
                <TableCell sx={{ width: 130 }} align="center">
                  <Tooltip title="オンにすると公開日時を現在時刻で更新します">
                    <span>公開日を更新</span>
                  </Tooltip>
                </TableCell>
                <TableCell align="center" sx={{ width: 56 }}>操作</TableCell>
              </TableRow>
            </TableHead>
            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
              <SortableContext items={pages.map(p => p.id)} strategy={verticalListSortingStrategy}>
                <TableBody>
                  {pages.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} align="center" sx={{ color: 'text.disabled', py: 2 }}>
                        ページが追加されていません
                      </TableCell>
                    </TableRow>
                  ) : (
                    pages.map((p, i) => (
                      <SortableDraftRow
                        key={p.id}
                        p={p}
                        idx={i}
                        onToggle={togglePublishUpdate}
                        onRemove={handleRemovePage}
                      />
                    ))
                  )}
                </TableBody>
              </SortableContext>
            </DndContext>
          </Table>
        </TableContainer>

        <Divider sx={{ mb: 2 }} />

        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            disabled={saving || publishing}
          >
            {saving ? '保存中...' : '保存'}
          </Button>
          <Button
            variant="contained"
            color="success"
            startIcon={<PublishIcon />}
            onClick={() => setPublishDialog(true)}
            disabled={pages.length === 0 || saving || publishing}
          >
            {publishing ? '公開中...' : '公開'}
          </Button>
        </Box>
      </Paper>

      {/* 公開確認ダイアログ */}
      <Dialog open={publishDialog} onClose={() => setPublishDialog(false)}>
        <DialogTitle>ドラフトを公開しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {pages.length} 件のページの HTML を生成・保存します。公開後このドラフトは削除されます。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPublishDialog(false)}>キャンセル</Button>
          <Button onClick={handlePublish} color="success" variant="contained">公開する</Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snack.open}
        autoHideDuration={3000}
        onClose={() => setSnack(s => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity={snack.severity} onClose={() => setSnack(s => ({ ...s, open: false }))}
          sx={{ width: '100%' }} elevation={6} variant="filled">
          {snack.message}
        </Alert>
      </Snackbar>
    </Box>
  )
}
