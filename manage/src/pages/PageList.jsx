import { useState, useEffect, useCallback } from 'react'
import PropTypes from 'prop-types'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import {
  Typography, Box, Paper, List, ListItem, ListItemButton,
  ListItemText, ListItemIcon, Chip, IconButton, Collapse,
  CircularProgress, Tooltip, Alert, Snackbar,
} from '@mui/material'
import ExpandLessIcon from '@mui/icons-material/ExpandLess'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import AddIcon from '@mui/icons-material/Add'
import ArticleIcon from '@mui/icons-material/Article'
import PreviewIcon from '@mui/icons-material/Preview'
import PublishIcon from '@mui/icons-material/Publish'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import { getRootPage, getChildren, newPage, publishPage } from '../api/page'

export default function PageList() {
  const [root, setRoot] = useState(null)
  const [childrenMap, setChildrenMap] = useState({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const navigate = useNavigate()

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  useEffect(() => {
    getRootPage()
      .then(data => {
        if (!data) return
        setRoot(data.page)
        setChildrenMap({ [data.page.id]: data.children || [] })
        setLoading(false)
      })
      .catch(err => { setError(err.message); setLoading(false) })
  }, [])

  const handleExpand = useCallback(async (id) => {
    if (childrenMap[id] !== undefined) return
    try {
      const children = await getChildren(id)
      setChildrenMap(prev => ({ ...prev, [id]: children || [] }))
    } catch (e) {
      console.error(e)
    }
  }, [childrenMap])

  const handlePublish = useCallback(async (id) => {
    try {
      await publishPage(id)
      setChildrenMap(prev => {
        const next = {}
        for (const [k, v] of Object.entries(prev)) {
          next[k] = v.map(p => p.id === id ? { ...p, canPublish: false } : p)
        }
        return next
      })
      setRoot(r => r?.id === id ? { ...r, canPublish: false } : r)
      showSnack('HTMLを公開しました')
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }, [])

  const handleAddChild = useCallback(async (parentId) => {
    try {
      const data = await newPage(parentId)
      if (data) navigate(`/page/new?parent=${parentId}&key=${data.page.id}`)
    } catch (e) {
      alert('エラー: ' + e.message)
    }
  }, [navigate])

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', pt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }
  if (error) return <Alert severity="error">{error}</Alert>
  if (!root) return <Alert severity="info">ルートページが設定されていません</Alert>

  return (
    <Box>
      <Typography variant="h5" fontWeight={600} gutterBottom>Pages</Typography>
      <Paper variant="outlined">
        <List dense disablePadding>
          <PageNode
            page={root}
            childrenMap={childrenMap}
            onExpand={handleExpand}
            onAddChild={handleAddChild}
            onPublish={handlePublish}
            isRoot
            depth={0}
          />
        </List>
      </Paper>
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

PageNode.propTypes = { page: PropTypes.object.isRequired, childrenMap: PropTypes.object.isRequired, onExpand: PropTypes.func, onAddChild: PropTypes.func, onPublish: PropTypes.func, isRoot: PropTypes.bool, depth: PropTypes.number }
function PageNode({ page, childrenMap, onExpand, onAddChild, onPublish, isRoot, depth }) {
  const [expanded, setExpanded] = useState(isRoot)
  const children = childrenMap[page.id]

  const handleToggle = () => {
    if (!expanded) onExpand(page.id)
    setExpanded(v => !v)
  }

  // secondaryAction: ステータスチップ + アクションボタンをまとめて右端に配置
  const secondaryAction = (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, pr: 0.5 }}>
      {page.deleted && (
        <Chip label="非表示" size="small" variant="outlined" />
      )}
      {/* HTML要更新: 保存後に HTML 公開がまだ行われていない */}
      {page.canPublish && (
        <Tooltip title="編集内容が保存されていますが、まだ「HTML公開」が実行されていません">
          <Chip label="HTML要更新" size="small" color="warning" variant="outlined" />
        </Tooltip>
      )}
      <Tooltip title="プレビュー">
        <IconButton size="small" component="a" href={`/manage/page/view/${page.id}`} target="_blank" rel="noopener noreferrer">
          <PreviewIcon fontSize="small" />
        </IconButton>
      </Tooltip>
      <Tooltip title="公開ページ">
        <IconButton size="small" component="a" href={`/page/${page.id}`} target="_blank" rel="noopener noreferrer">
          <OpenInNewIcon fontSize="small" />
        </IconButton>
      </Tooltip>
      {page.canPublish && (
        <Tooltip title="HTML公開">
          <IconButton size="small" color="success" onClick={() => onPublish(page.id)}>
            <PublishIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      )}
      <Tooltip title="子ページ追加">
        <IconButton size="small" color="success" onClick={() => onAddChild(page.id)}>
          <AddIcon fontSize="small" />
        </IconButton>
      </Tooltip>
    </Box>
  )

  const rightPadding = '295px'

  return (
    <>
      <ListItem disablePadding divider secondaryAction={secondaryAction}>
        {/* 展開トグル */}
        <IconButton onClick={handleToggle} size="small" sx={{ ml: depth * 3 + 0.5, flexShrink: 0 }}>
          {expanded
            ? <ExpandLessIcon fontSize="small" color="action" />
            : <ExpandMoreIcon fontSize="small" color="action" />
          }
        </IconButton>
        {/* ページ名リンク */}
        <ListItemButton
          component={RouterLink}
          to={`/page/${page.id}`}
          sx={{ pr: rightPadding, minWidth: 0 }}
        >
          <ListItemIcon sx={{ minWidth: 28 }}>
            <ArticleIcon fontSize="small" color={page.deleted ? 'disabled' : 'primary'} />
          </ListItemIcon>
          <ListItemText
            primary={page.name || '(名前なし)'}
            primaryTypographyProps={{
              sx: {
                textDecoration: page.deleted ? 'line-through' : 'none',
                color: page.deleted ? 'text.disabled' : 'text.primary',
              },
            }}
          />
        </ListItemButton>
      </ListItem>

      <Collapse in={expanded} timeout="auto" unmountOnExit>
        <List dense disablePadding>
          {children && children.map(child => (
            <PageNode
              key={child.id}
              page={child}
              childrenMap={childrenMap}
              onExpand={onExpand}
              onAddChild={onAddChild}
              onPublish={onPublish}
              depth={depth + 1}
            />
          ))}
          {children && children.length === 0 && (
            <ListItem sx={{ pl: (depth + 1) * 3 + 1 }}>
              <ListItemText
                primary="子ページなし"
                primaryTypographyProps={{ variant: 'caption', color: 'text.disabled' }}
              />
            </ListItem>
          )}
          {!children && expanded && (
            <ListItem sx={{ pl: (depth + 1) * 3 + 1 }}>
              <CircularProgress size={16} />
            </ListItem>
          )}
        </List>
      </Collapse>
    </>
  )
}
