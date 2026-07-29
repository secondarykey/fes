import { useState, useEffect, useCallback } from 'react'
import PropTypes from 'prop-types'
import { useParams, Link as RouterLink, useNavigate } from 'react-router'
import {
  Box, Paper, Typography, Breadcrumbs, Link, CircularProgress,
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow,
  Switch, Button, Tooltip, Snackbar, Alert,
  Dialog, DialogTitle, DialogContent, DialogContentText,
  DialogActions, TextField, Checkbox, Divider,
  IconButton, Menu, MenuItem,
} from '@mui/material'
import DragIndicatorIcon from '@mui/icons-material/DragIndicator'
import MoreVertIcon from '@mui/icons-material/MoreVert'
import VerticalAlignTopIcon from '@mui/icons-material/VerticalAlignTop'
import VerticalAlignBottomIcon from '@mui/icons-material/VerticalAlignBottom'
import SortByAlphaIcon from '@mui/icons-material/SortByAlpha'
import DriveFileMoveIcon from '@mui/icons-material/DriveFileMove'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import SaveIcon from '@mui/icons-material/Save'
import ListIcon from '@mui/icons-material/List'
import AddIcon from '@mui/icons-material/Add'
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { getPage, getChildren, newPage, sequencePages, movePage, deletePages, getPageListHtml } from '../api/page'
import { addDraftPages, getCurrentDraft } from '../api/draft'

SortableRow.propTypes = {
  item: PropTypes.object.isRequired,
  idx: PropTypes.number.isRequired,
  selected: PropTypes.bool,
  onToggleSelect: PropTypes.func,
  onToggleEnabled: PropTypes.func,
  onMoveToTop: PropTypes.func,
  onMoveToBottom: PropTypes.func,
}
function SortableRow({ item, idx, selected, onToggleSelect, onToggleEnabled, onMoveToTop, onMoveToBottom }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: item.id })
  const [menuAnchor, setMenuAnchor] = useState(null)

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    background: isDragging ? 'rgba(0,0,0,0.04)' : undefined,
  }

  const handleMenuOpen = (e) => {
    e.stopPropagation()
    setMenuAnchor(e.currentTarget)
  }
  const handleMenuClose = () => setMenuAnchor(null)

  return (
    <TableRow ref={setNodeRef} style={style} hover selected={selected}>
      <TableCell padding="checkbox">
        <Checkbox size="small" checked={selected} onChange={() => onToggleSelect(item.id)} />
      </TableCell>
      <TableCell sx={{ width: 40, cursor: 'grab', color: 'text.disabled', px: 1 }}
        {...attributes} {...listeners}>
        <DragIndicatorIcon fontSize="small" />
      </TableCell>
      <TableCell>
        <Link component={RouterLink} to={`/page/${item.id}`} underline="hover">
          {item.name}
        </Link>
      </TableCell>
      <TableCell sx={{ width: 60 }} align="center">
        <Switch
          size="small"
          checked={item.enabled}
          onChange={() => onToggleEnabled(idx)}
          color="success"
        />
      </TableCell>
      <TableCell sx={{ width: 100 }} align="center">
        <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'nowrap' }}>
          {new Date(item.createdAt).toLocaleDateString('ja-JP')}
        </Typography>
      </TableCell>
      <TableCell sx={{ width: 40, px: 0.5 }} align="center">
        <IconButton size="small" onClick={handleMenuOpen}>
          <MoreVertIcon fontSize="small" />
        </IconButton>
        <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={handleMenuClose}>
          <MenuItem onClick={() => { handleMenuClose(); onMoveToTop(item.id) }}>
            <VerticalAlignTopIcon fontSize="small" sx={{ mr: 1 }} />
            先頭へ移動
          </MenuItem>
          <MenuItem onClick={() => { handleMenuClose(); onMoveToBottom(item.id) }}>
            <VerticalAlignBottomIcon fontSize="small" sx={{ mr: 1 }} />
            末尾へ移動
          </MenuItem>
        </Menu>
      </TableCell>
    </TableRow>
  )
}

export default function PageChildren() {
  const { key } = useParams()
  const navigate = useNavigate()
  const [page, setPage] = useState(null)
  const [breadcrumbs, setBreadcrumbs] = useState([])
  const [items, setItems] = useState([])
  const [selected, setSelected] = useState(new Set())
  const [dirty, setDirty] = useState(false)
  const [saving, setSaving] = useState(false)
  const [moveTargetId, setMoveTargetId] = useState('Trash')
  const [moveDialog, setMoveDialog] = useState(false)
  const [deleteDialog, setDeleteDialog] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [listHtml, setListHtml] = useState(null)
  const [listLoading, setListLoading] = useState(false)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const [draftRegistering, setDraftRegistering] = useState(false)

  const sensors = useSensors(useSensor(PointerSensor))

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  const load = useCallback(async () => {
    try {
      const [detail, children] = await Promise.all([
        getPage(key),
        getChildren(key),
      ])
      if (detail) {
        setPage(detail.page)
        setBreadcrumbs(detail.breadcrumbs || [])
      }
      setItems((children || []).map(c => ({ ...c, enabled: !c.deleted })))
      setSelected(new Set())
      setDirty(false)
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }, [key])

  useEffect(() => { load() }, [load])

  const handleDragEnd = ({ active, over }) => {
    if (!over || active.id === over.id) return
    setItems(prev => {
      const oldIdx = prev.findIndex(i => i.id === active.id)
      const newIdx = prev.findIndex(i => i.id === over.id)
      return arrayMove(prev, oldIdx, newIdx)
    })
    setDirty(true)
  }

  const toggleEnabled = (idx) => {
    setItems(prev => {
      const next = [...prev]
      next[idx] = { ...next[idx], enabled: !next[idx].enabled }
      return next
    })
    setDirty(true)
  }

  const toggleSelect = (id) => {
    setSelected(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const allSelected = items.length > 0 && selected.size === items.length
  const someSelected = selected.size > 0 && selected.size < items.length

  const toggleSelectAll = () => {
    if (allSelected) {
      setSelected(new Set())
    } else {
      setSelected(new Set(items.map(i => i.id)))
    }
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await sequencePages(key, items.map(item => ({
        id: item.id,
        enabled: item.enabled,
        version: item.version,
      })))
      showSnack('順序を保存しました')
      await load()
    } catch (e) {
      showSnack(e.message, 'error')
    }
    setSaving(false)
  }

  const handleSort = () => {
    setItems(prev => {
      const sorted = [...prev].sort((a, b) => a.name.localeCompare(b.name, 'ja'))
      return sorted
    })
    setDirty(true)
  }

  const handleAddChild = async () => {
    try {
      const data = await newPage(key)
      if (data) navigate(`/page/new?parent=${key}&key=${data.page.id}`)
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const handleGenerateList = async () => {
    setListLoading(true)
    try {
      const data = await getPageListHtml(key)
      setListHtml(data.html || '')
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setListLoading(false)
    }
  }

  const handleMove = async () => {
    setMoveDialog(false)
    try {
      await Promise.all([...selected].map(id => movePage(id, moveTargetId)))
      showSnack(`${selected.size} 件移動しました`)
      setMoveTargetId('Trash')
      await load()
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const handleMoveToTop = (id) => {
    setItems(prev => {
      const idx = prev.findIndex(i => i.id === id)
      if (idx <= 0) return prev
      return arrayMove(prev, idx, 0)
    })
    setDirty(true)
  }

  const handleMoveToBottom = (id) => {
    setItems(prev => {
      const idx = prev.findIndex(i => i.id === id)
      if (idx < 0 || idx === prev.length - 1) return prev
      return arrayMove(prev, idx, prev.length - 1)
    })
    setDirty(true)
  }

  const setAllEnabled = (enabled) => {
    setItems(prev => prev.map(item => ({ ...item, enabled })))
    setDirty(true)
  }

  const handleAddToDraft = async () => {
    setDraftRegistering(true)
    const ids = [...selected]
    try {
      const cur = await getCurrentDraft()
      if (!cur?.id) {
        showSnack('ドラフトが設定されていません', 'error')
        return
      }
      await addDraftPages(cur.id, ids)
      showSnack(`${ids.length} ページをドラフトに登録しました`)
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDraftRegistering(false)
    }
  }

  const handleDelete = async () => {
    setDeleteDialog(false)
    setDeleting(true)
    const ids = [...selected]
    try {
      await deletePages(ids)
      showSnack(`${ids.length} 件削除しました`)
      await load()
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDeleting(false)
    }
  }

  if (!page) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', pt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ maxWidth: 960 }}>
      <Breadcrumbs sx={{ mb: 2 }}>
        <Link component={RouterLink} to="/page" underline="hover" color="inherit">Pages</Link>
        {breadcrumbs.map(b => (
          <Link key={b.id} component={RouterLink} to={`/page/${b.id}`} underline="hover" color="inherit">
            {b.name}
          </Link>
        ))}
        <Link component={RouterLink} to={`/page/${key}`} underline="hover" color="inherit">
          {page.name}
        </Link>
        <Typography color="text.primary">子ページ管理</Typography>
      </Breadcrumbs>

      <Paper sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2, flexWrap: 'wrap' }}>
          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            子ページ管理 — {page.name}
          </Typography>
          {key === 'Trash' && (
            <>
              <Button size="small" variant="outlined" color="success" onClick={() => setAllEnabled(true)}>
                全て有効
              </Button>
              <Button size="small" variant="outlined" color="warning" onClick={() => setAllEnabled(false)}>
                全て無効
              </Button>
            </>
          )}
          {key !== 'Trash' && (
            <Button size="small" variant="contained" color="success" startIcon={<AddIcon />} onClick={handleAddChild}>
              追加
            </Button>
          )}
          {key === 'Trash' && (
            <Tooltip title="名前でアルファベット順に自動ソート">
              <Button size="small" variant="outlined" startIcon={<SortByAlphaIcon />} onClick={handleSort}>
                自動ソート
              </Button>
            </Tooltip>
          )}
          <Button
            size="small"
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            disabled={!dirty || saving}
          >
            {saving ? '保存中...' : '保存'}
          </Button>
        </Box>

        {items.length === 0 ? (
          <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>
            子ページがありません
          </Typography>
        ) : (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={items.map(i => i.id)} strategy={verticalListSortingStrategy}>
              <TableContainer>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell padding="checkbox">
                        <Checkbox
                          size="small"
                          checked={allSelected}
                          indeterminate={someSelected}
                          onChange={toggleSelectAll}
                        />
                      </TableCell>
                      <TableCell sx={{ width: 40 }} />
                      <TableCell>ページ名</TableCell>
                      <TableCell sx={{ width: 60 }} align="center">有効</TableCell>
                      <TableCell sx={{ width: 100 }} align="center">作成日</TableCell>
                      <TableCell sx={{ width: 40 }} />
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {items.map((item, idx) => (
                      <SortableRow
                        key={item.id}
                        item={item}
                        idx={idx}
                        selected={selected.has(item.id)}
                        onToggleSelect={toggleSelect}
                        onToggleEnabled={toggleEnabled}
                        onMoveToTop={handleMoveToTop}
                        onMoveToBottom={handleMoveToBottom}
                      />
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </SortableContext>
          </DndContext>
        )}

        {/* 移動・削除エリア */}
        {items.length > 0 && (
          <>
            <Divider sx={{ mt: 3, mb: 2 }} />
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, flexWrap: 'wrap' }}>
              <Typography variant="body2" color="text.secondary" sx={{ flexShrink: 0 }}>
                選択中: {selected.size} 件
              </Typography>
              {key !== 'Trash' && (
                <>
                  <TextField
                    label="移動先ページID"
                    value={moveTargetId}
                    onChange={e => setMoveTargetId(e.target.value)}
                    size="small"
                    sx={{ width: 260 }}
                    slotProps={{ htmlInput: { style: { fontFamily: 'monospace' } } }}
                  />
                  <Button
                    variant="outlined"
                    color="warning"
                    startIcon={<DriveFileMoveIcon />}
                    disabled={selected.size === 0 || !moveTargetId.trim()}
                    onClick={() => setMoveDialog(true)}
                  >
                    移動
                  </Button>
                </>
              )}
              <Button
                variant="outlined"
                color="error"
                startIcon={deleting ? <CircularProgress size={14} /> : null}
                disabled={selected.size === 0 || deleting}
                onClick={() => setDeleteDialog(true)}
              >
                削除
              </Button>
              {key !== 'Trash' && (
                <Button
                  variant="outlined"
                  color="info"
                  disabled={selected.size === 0 || draftRegistering}
                  onClick={handleAddToDraft}
                >
                  ドラフトに登録
                </Button>
              )}
            </Box>
          </>
        )}

        {/* ページリスト生成 */}
        <Divider sx={{ mt: 3, mb: 2 }} />
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: listHtml !== null ? 1.5 : 0 }}>
          <Typography variant="body2" color="text.secondary" sx={{ flexGrow: 1 }}>
            子ページを &lt;li&gt; タグで囲んだ HTML を生成します
          </Typography>
          <Button
            size="small"
            variant="outlined"
            startIcon={listLoading ? <CircularProgress size={14} /> : <ListIcon />}
            onClick={handleGenerateList}
            disabled={listLoading || items.length === 0}
          >
            {listLoading ? '生成中...' : 'リスト生成'}
          </Button>
        </Box>
        {listHtml !== null && (
          <TextField
            value={listHtml}
            multiline
            fullWidth
            minRows={4}
            maxRows={20}
            slotProps={{ htmlInput: { readOnly: true, style: { fontFamily: 'monospace', fontSize: 13 } } }}
          />
        )}

        <Box sx={{ mt: 2 }}>
          <Button
            variant="text"
            startIcon={<ArrowBackIcon />}
            component={RouterLink}
            to={`/page/${key}`}
          >
            ページ編集へ戻る
          </Button>
        </Box>
      </Paper>

      {/* 移動確認ダイアログ */}
      <Dialog open={moveDialog} onClose={() => setMoveDialog(false)}>
        <DialogTitle>ページを移動しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            選択した {selected.size} 件のページをページID「{moveTargetId}」の配下へ移動します。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMoveDialog(false)}>キャンセル</Button>
          <Button onClick={handleMove} variant="contained" color="warning">移動</Button>
        </DialogActions>
      </Dialog>

      {/* 削除確認ダイアログ */}
      <Dialog open={deleteDialog} onClose={() => setDeleteDialog(false)}>
        <DialogTitle>ページを削除しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            選択した {selected.size} 件のページを完全に削除します。この操作は取り消せません。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialog(false)}>キャンセル</Button>
          <Button onClick={handleDelete} variant="contained" color="error">削除</Button>
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
