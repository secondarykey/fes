import { useState, useEffect, useCallback } from 'react'
import PropTypes from 'prop-types'
import { Link as RouterLink, useNavigate } from 'react-router'
import {
  Typography, Box, Paper, Table, TableHead, TableBody, TableRow, TableCell,
  TableContainer, IconButton, Button, Tooltip, CircularProgress, Alert,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions,
  Chip, Tabs, Tab, Snackbar,
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import AddIcon from '@mui/icons-material/Add'
import SaveIcon from '@mui/icons-material/Save'
import DragIndicatorIcon from '@mui/icons-material/DragIndicator'
import {
  DndContext, closestCenter, PointerSensor, useSensor, useSensors,
} from '@dnd-kit/core'
import {
  SortableContext, verticalListSortingStrategy, useSortable, arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { getTemplates, newTemplate, deleteTemplate, sequenceTemplates } from '../api/template'

const TYPE_LABELS = { 1: 'サイト', 2: 'ページ' }
const TYPE_COLORS = { 1: 'secondary', 2: 'primary' }

function fmtDate(iso) {
  return new Date(iso).toLocaleString('ja-JP', { dateStyle: 'short', timeStyle: 'short' })
}

SortableRow.propTypes = { t: PropTypes.object.isRequired, onDelete: PropTypes.func }
function SortableRow({ t, onDelete }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: t.id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    background: isDragging ? 'rgba(0,0,0,0.04)' : undefined,
  }
  return (
    <TableRow ref={setNodeRef} style={style} hover>
      <TableCell sx={{ width: 40, cursor: 'grab', color: 'text.disabled', px: 1 }}
        {...attributes} {...listeners}>
        <DragIndicatorIcon fontSize="small" />
      </TableCell>
      <TableCell>
        <RouterLink to={`/template/${t.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
          <Typography variant="body2" color="primary" sx={{ '&:hover': { textDecoration: 'underline' } }}>
            {t.name || '(名前なし)'}
          </Typography>
        </RouterLink>
      </TableCell>
      <TableCell>
        <Chip label={TYPE_LABELS[t.type] || `type${t.type}`} size="small"
          color={TYPE_COLORS[t.type] || 'default'} variant="outlined" />
      </TableCell>
      <TableCell>
        <Typography variant="body2" color="text.secondary">
          {t.updatedAt ? fmtDate(t.updatedAt) : '-'}
        </Typography>
      </TableCell>
      <TableCell align="center">
        <Tooltip title="削除">
          <IconButton size="small" color="error" onClick={() => onDelete(t)}>
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      </TableCell>
    </TableRow>
  )
}

StaticRow.propTypes = { t: PropTypes.object.isRequired, onDelete: PropTypes.func }
function StaticRow({ t, onDelete }) {
  return (
    <TableRow hover>
      <TableCell sx={{ width: 40 }} />
      <TableCell>
        <RouterLink to={`/template/${t.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
          <Typography variant="body2" color="primary" sx={{ '&:hover': { textDecoration: 'underline' } }}>
            {t.name || '(名前なし)'}
          </Typography>
        </RouterLink>
      </TableCell>
      <TableCell>
        <Chip label={TYPE_LABELS[t.type] || `type${t.type}`} size="small"
          color={TYPE_COLORS[t.type] || 'default'} variant="outlined" />
      </TableCell>
      <TableCell>
        <Typography variant="body2" color="text.secondary">
          {t.updatedAt ? fmtDate(t.updatedAt) : '-'}
        </Typography>
      </TableCell>
      <TableCell align="center">
        <Tooltip title="削除">
          <IconButton size="small" color="error" onClick={() => onDelete(t)}>
            <DeleteIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      </TableCell>
    </TableRow>
  )
}

export default function TemplateList() {
  const [templates, setTemplates] = useState([])
  const [nextCursor, setNextCursor] = useState('')
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(null)
  const [typeTab, setTypeTab] = useState('2')
  const [dirty, setDirty] = useState(false)
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const navigate = useNavigate()

  const sensors = useSensors(useSensor(PointerSensor))
  const isSingleType = typeTab !== 'all'

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  const load = useCallback(async (type, cursor = '', append = false) => {
    setLoading(true)
    try {
      // タイプ別タブでは全件取得してドラッグ並び替えを有効にする
      const cur = type !== 'all' ? '' : cursor
      const data = await getTemplates(type, cur)
      if (!data) return
      setTemplates(prev => append ? [...prev, ...(data.templates || [])] : (data.templates || []))
      setNextCursor(data.nextCursor || '')
      setLoadError(null)
      setDirty(false)
    } catch (e) {
      setLoadError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load(typeTab) }, [typeTab, load])

  const handleTabChange = (_, v) => {
    setTemplates([])
    setNextCursor('')
    setTypeTab(v)
  }

  const handleAdd = async () => {
    try {
      const data = await newTemplate()
      if (data) navigate(`/template/new?key=${data.id}`)
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const handleDragEnd = ({ active, over }) => {
    if (!over || active.id === over.id) return
    setTemplates(prev => {
      const oldIdx = prev.findIndex(t => t.id === active.id)
      const newIdx = prev.findIndex(t => t.id === over.id)
      return arrayMove(prev, oldIdx, newIdx)
    })
    setDirty(true)
  }

  const handleSaveSequence = async () => {
    setSaving(true)
    try {
      await sequenceTemplates(templates.map(t => ({ id: t.id, version: t.version || 0 })))
      showSnack('順序を保存しました')
      await load(typeTab)
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return
    try {
      await deleteTemplate(deleteTarget.id)
      showSnack('削除しました')
      setTemplates(prev => prev.filter(t => t.id !== deleteTarget.id))
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDeleteTarget(null)
    }
  }

  const tableHead = (
    <TableHead>
      <TableRow>
        <TableCell sx={{ width: 40 }} />
        <TableCell>テンプレート名</TableCell>
        <TableCell>種別</TableCell>
        <TableCell>更新日時</TableCell>
        <TableCell align="center" sx={{ width: 64 }}>操作</TableCell>
      </TableRow>
    </TableHead>
  )

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 1 }}>
        <Typography variant="h5" sx={{ fontWeight: 600, flexGrow: 1 }}>Templates</Typography>
        {isSingleType && (
          <Button
            variant="outlined"
            size="small"
            startIcon={saving ? <CircularProgress size={14} /> : <SaveIcon />}
            onClick={handleSaveSequence}
            disabled={!dirty || saving}
          >
            {saving ? '保存中...' : '順序を保存'}
          </Button>
        )}
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
          新規作成
        </Button>
      </Box>

      <Tabs value={typeTab} onChange={handleTabChange} sx={{ mb: 2 }}>
        <Tab label="すべて" value="all" />
        <Tab label="サイトテンプレート" value="1" />
        <Tab label="ページテンプレート" value="2" />
      </Tabs>

      {loadError && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setLoadError(null)}>
          {loadError}
        </Alert>
      )}

      {isSingleType ? (
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
          <SortableContext items={templates.map(t => t.id)} strategy={verticalListSortingStrategy}>
            <TableContainer component={Paper} variant="outlined">
              <Table size="small">
                {tableHead}
                <TableBody>
                  {loading && templates.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                        <CircularProgress size={24} />
                      </TableCell>
                    </TableRow>
                  ) : templates.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5}>
                        <Alert severity="info" sx={{ border: 'none' }}>テンプレートがありません</Alert>
                      </TableCell>
                    </TableRow>
                  ) : (
                    templates.map(t => (
                      <SortableRow key={t.id} t={t} onDelete={setDeleteTarget} />
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </SortableContext>
        </DndContext>
      ) : (
        <TableContainer component={Paper} variant="outlined">
          <Table size="small">
            {tableHead}
            <TableBody>
              {loading && templates.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                    <CircularProgress size={24} />
                  </TableCell>
                </TableRow>
              ) : templates.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5}>
                    <Alert severity="info" sx={{ border: 'none' }}>テンプレートがありません</Alert>
                  </TableCell>
                </TableRow>
              ) : (
                templates.map(t => (
                  <StaticRow key={t.id} t={t} onDelete={setDeleteTarget} />
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {typeTab === 'all' && nextCursor && (
        <Box sx={{ mt: 2, display: 'flex', justifyContent: 'center' }}>
          <Button variant="outlined" onClick={() => load(typeTab, nextCursor, true)} disabled={loading}>
            さらに読み込む
          </Button>
        </Box>
      )}

      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogTitle>テンプレートを削除しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            「{deleteTarget?.name || '(名前なし)'}」を削除します。ページで使用中の場合は削除できません。
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
