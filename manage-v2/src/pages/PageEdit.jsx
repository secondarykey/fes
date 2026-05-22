import { useState, useEffect } from 'react'
import { useParams, useSearchParams, useNavigate, Link as RouterLink } from 'react-router-dom'
import {
  Box, Paper, Typography, TextField, Select, MenuItem,
  FormControl, InputLabel, FormControlLabel, Switch,
  Button, Breadcrumbs, Link, Divider, CircularProgress,
  Snackbar, Alert as MuiAlert, Chip,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import PublishIcon from '@mui/icons-material/Publish'
import UnpublishedIcon from '@mui/icons-material/Unpublished'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import PreviewIcon from '@mui/icons-material/Preview'
import AccountTreeIcon from '@mui/icons-material/AccountTree'
import BookmarkAddIcon from '@mui/icons-material/BookmarkAdd'
import ImageIcon from '@mui/icons-material/Image'
import DeleteIcon from '@mui/icons-material/Delete'
import { useRef } from 'react'
import { getPage, newPage, updatePage, deletePage, publishPage, unpublishPage, getPageImage, uploadPageImage, deletePageImage } from '../api/page'
import { getDrafts, getCurrentDraft, addDraftPage } from '../api/draft'

const SITE_TEMPLATE = 1
const PAGE_TEMPLATE = 2

function toForm(d) {
  return {
    name: d.page.name || '',
    description: d.page.description || '',
    content: d.pageData || '',
    parentId: d.page.parent || '',
    siteTemplate: d.page.siteTemplate || '',
    pageTemplate: d.page.pageTemplate || '',
    paging: d.page.paging || 0,
    published: !d.page.deleted,
    version: d.page.version || 0,
    seq: d.page.seq || 0,
  }
}

export default function PageEdit() {
  const { key } = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()

  const isNew = !key
  const parentKey = searchParams.get('parent') || ''
  const presetKey = searchParams.get('key') || ''

  const [data, setData] = useState(null)
  const [form, setForm] = useState({})
  const [saving, setSaving] = useState(false)
  const [deleteDialog, setDeleteDialog] = useState(false)
  const [disableDialog, setDisableDialog] = useState(false)
  const [image, setImage] = useState(null)
  const [imageUploading, setImageUploading] = useState(false)
  const imageInputRef = useRef(null)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const [draftDialog, setDraftDialog] = useState(false)
  const [drafts, setDrafts] = useState([])
  const [currentDraft, setCurrentDraft] = useState(null)
  const [draftsLoading, setDraftsLoading] = useState(false)
  const [draftRegistering, setDraftRegistering] = useState(false)

  useEffect(() => {
    const load = isNew ? newPage(parentKey) : getPage(key)
    load
      .then(d => {
        if (!d) return
        setData(d)
        const f = toForm(d)
        if (isNew) {
          const templates = d.templates || []
          if (!f.siteTemplate) {
            const first = templates.find(t => t.type === SITE_TEMPLATE)
            if (first) f.siteTemplate = first.id
          }
          if (!f.pageTemplate) {
            const first = templates.find(t => t.type === PAGE_TEMPLATE)
            if (first) f.pageTemplate = first.id
          }
        }
        setForm(f)
      })
      .catch(err => showSnack(err.message, 'error'))
  }, [key, isNew, parentKey])

  useEffect(() => {
    if (isNew || !key) return
    getPageImage(key)
      .then(d => setImage(d))
      .catch(() => {})
  }, [key, isNew])

  const showSnack = (message, severity = 'success') =>
    setSnack({ open: true, message, severity })

  const handleText = name => e => setForm(prev => ({ ...prev, [name]: e.target.value }))

  const handlePublishedToggle = e => {
    if (!e.target.checked) {
      setDisableDialog(true)
    } else {
      setForm(prev => ({ ...prev, published: true }))
    }
  }

  const confirmDisable = () => {
    setForm(prev => ({ ...prev, published: false }))
    setDisableDialog(false)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const pageKey = isNew ? presetKey : key
      const updated = await updatePage(pageKey, form)
      if (updated) {
        setData(prev => ({ ...prev, page: updated }))
        setForm(prev => ({ ...prev, version: updated.version }))
      }
      showSnack('保存しました')
      if (isNew) navigate(`/page/${pageKey}`, { replace: true })
    } catch (e) {
      showSnack(e.message, 'error')
    }
    setSaving(false)
  }

  const refreshPage = async () => {
    if (!key) return
    const r = await getPage(key)
    if (r) { setData(r); setForm(f => ({ ...f, version: r.page.version })) }
  }

  const refreshImage = async (pageKey) => {
    const result = await getPageImage(pageKey)
    setImage(result)
  }

  const handlePublish = async () => {
    try {
      await publishPage(key)
      showSnack('HTMLを公開しました')
      await Promise.all([refreshPage(), refreshImage(key)])
    } catch (e) { showSnack(e.message, 'error') }
  }

  const handleImageSelect = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    e.target.value = ''
    setImageUploading(true)
    try {
      const pageKey = isNew ? presetKey : key
      await uploadPageImage(pageKey, file)
      await Promise.all([refreshImage(pageKey), refreshPage()])
      showSnack('画像をアップロードしました')
    } catch (err) {
      showSnack(err.message, 'error')
    } finally {
      setImageUploading(false)
    }
  }

  const handleImageDelete = async () => {
    setImageUploading(true)
    try {
      const pageKey = isNew ? presetKey : key
      const result = await deletePageImage(pageKey)
      setImage(result)
      await refreshPage()
      showSnack('画像を削除しました')
    } catch (err) {
      showSnack(err.message, 'error')
    } finally {
      setImageUploading(false)
    }
  }

  const handleDelete = async () => {
    try {
      await deletePage(key)
      navigate('/page', { replace: true })
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDeleteDialog(false)
    }
  }

  const handleOpenDraftDialog = async () => {
    setDraftsLoading(true)
    setDraftDialog(true)
    try {
      const [draftsData, cur] = await Promise.all([getDrafts(), getCurrentDraft()])
      setDrafts(draftsData?.drafts || [])
      setCurrentDraft(cur || null)
    } catch (e) {
      showSnack(e.message, 'error')
      setDraftDialog(false)
    } finally {
      setDraftsLoading(false)
    }
  }

  const handleRegisterToDraft = async (draftId) => {
    setDraftRegistering(true)
    try {
      await addDraftPage(draftId, key)
      showSnack('ドラフトに登録しました')
      setDraftDialog(false)
    } catch (e) {
      showSnack(e.message, 'error')
    } finally {
      setDraftRegistering(false)
    }
  }

  const handleUnpublish = async () => {
    try {
      await unpublishPage(key)
      showSnack('非公開にしました')
      const r = await getPage(key)
      if (r) { setData(r); setForm(f => ({ ...f, version: r.page.version })) }
    } catch (e) { showSnack(e.message, 'error') }
  }

  if (!data) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', pt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  const siteTemplates = (data.templates || []).filter(t => t.type === SITE_TEMPLATE)
  const pageTemplates = (data.templates || []).filter(t => t.type === PAGE_TEMPLATE)

  return (
    <Box sx={{ maxWidth: 960 }}>
      {/* パンくずリスト + プレビュー/公開ページ */}
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Breadcrumbs sx={{ flexGrow: 1 }}>
          <Link component={RouterLink} to="/page" underline="hover" color="inherit">
            Pages
          </Link>
          {(data.breadcrumbs || []).map(b => (
            <Link key={b.id} component={RouterLink} to={`/page/${b.id}`} underline="hover" color="inherit">
              {b.name}
            </Link>
          ))}
          <Typography color="text.primary">{form.name || '(新規)'}</Typography>
        </Breadcrumbs>
        {!isNew && (
          <Box sx={{ display: 'flex', gap: 1, ml: 2, flexShrink: 0 }}>
            <Button
              variant="text"
              size="small"
              startIcon={<PreviewIcon />}
              href={`/manage/page/view/${key}`}
              target="_blank"
              rel="noopener noreferrer"
              component="a"
            >
              プレビュー
            </Button>
            <Button
              variant="text"
              size="small"
              startIcon={<OpenInNewIcon />}
              href={`/page/${key}`}
              target="_blank"
              rel="noopener noreferrer"
              component="a"
            >
              公開ページ
            </Button>
          </Box>
        )}
      </Box>

      <Paper sx={{ p: 3 }}>

        {/* 上段: 左=基本情報、右=ページ画像 */}
        <Box sx={{ display: 'flex', gap: 3, mb: 2, alignItems: 'flex-start' }}>

          {/* 左カラム */}
          <Box sx={{ flex: 1, minWidth: 0 }}>
            {/* ページ名 */}
            <TextField
              label="ページ名"
              value={form.name}
              onChange={handleText('name')}
              fullWidth
              size="small"
              sx={{ mb: 2 }}
            />

            {/* 説明 */}
            <TextField
              label="説明"
              value={form.description}
              onChange={handleText('description')}
              fullWidth
              multiline
              rows={2}
              size="small"
              sx={{ mb: 2 }}
            />

            {/* テンプレート + ページング */}
            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
              <FormControl size="small" sx={{ flex: 1, minWidth: 160 }}>
                <InputLabel>サイトテンプレート</InputLabel>
                <Select
                  value={form.siteTemplate}
                  label="サイトテンプレート"
                  onChange={handleText('siteTemplate')}
                >
                  <MenuItem value=""><em>選択してください...</em></MenuItem>
                  {siteTemplates.map(t => (
                    <MenuItem key={t.id} value={t.id}>{t.name}</MenuItem>
                  ))}
                </Select>
              </FormControl>

              <FormControl size="small" sx={{ flex: 1, minWidth: 160 }}>
                <InputLabel>ページテンプレート</InputLabel>
                <Select
                  value={form.pageTemplate}
                  label="ページテンプレート"
                  onChange={handleText('pageTemplate')}
                >
                  <MenuItem value=""><em>選択してください...</em></MenuItem>
                  {pageTemplates.map(t => (
                    <MenuItem key={t.id} value={t.id}>{t.name}</MenuItem>
                  ))}
                </Select>
              </FormControl>

              <TextField
                label="ページング"
                type="number"
                value={form.paging}
                onChange={handleText('paging')}
                size="small"
                sx={{ width: 110, flexShrink: 0 }}
              />
            </Box>
          </Box>

          {/* 右カラム: ページ画像 */}
          <Box sx={{ width: 260, flexShrink: 0 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <Typography variant="subtitle2">ページ画像</Typography>
              {image?.url && (
                <Chip
                  label={image.draft ? '下書き' : '公開中'}
                  size="small"
                  color={image.draft ? 'warning' : 'success'}
                  variant="outlined"
                />
              )}
            </Box>
            {image?.url ? (
              <Box
                component="img"
                src={image.url}
                alt="ページ画像"
                sx={{ width: '100%', maxHeight: 160, objectFit: 'contain', border: '1px solid', borderColor: 'divider', borderRadius: 1, display: 'block', mb: 1 }}
              />
            ) : (
              <Box sx={{ width: '100%', height: 120, border: '1px dashed', borderColor: 'divider', borderRadius: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', mb: 1 }}>
                <Typography variant="body2" color="text.disabled">画像なし</Typography>
              </Box>
            )}
            <input
              type="file"
              accept="image/*"
              ref={imageInputRef}
              style={{ display: 'none' }}
              onChange={handleImageSelect}
            />
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button
                variant="outlined"
                size="small"
                sx={{ flex: 1 }}
                startIcon={imageUploading ? <CircularProgress size={14} /> : <ImageIcon />}
                onClick={() => imageInputRef.current?.click()}
                disabled={imageUploading}
              >
                {imageUploading ? '処理中...' : '画像を選択'}
              </Button>
              {image?.url && (
                <Button
                  variant="outlined"
                  size="small"
                  color="error"
                  startIcon={<DeleteIcon />}
                  onClick={handleImageDelete}
                  disabled={imageUploading}
                >
                  削除
                </Button>
              )}
            </Box>
          </Box>
        </Box>

        {/* 子ページ管理 + 保存/公開 */}
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
          {!isNew ? (
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button
                variant="text"
                size="small"
                startIcon={<AccountTreeIcon />}
                component={RouterLink}
                to={`/page/${key}/children`}
              >
                子ページ管理
              </Button>
              <Button
                variant="text"
                size="small"
                startIcon={<BookmarkAddIcon />}
                onClick={handleOpenDraftDialog}
              >
                ドラフトに設定
              </Button>
            </Box>
          ) : <Box />}
          <Box sx={{ ml: 'auto', display: 'flex', gap: 1 }}>
            {!isNew && data.page.canPublish && (
              <Button
                variant="outlined"
                size="small"
                color="success"
                startIcon={<PublishIcon />}
                onClick={handlePublish}
              >
                公開
              </Button>
            )}
            <Button
              variant="contained"
              size="small"
              startIcon={<SaveIcon />}
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? '保存中...' : '保存'}
            </Button>
          </Box>
        </Box>

        {/* コンテンツ */}
        <TextField
          label="コンテンツ"
          value={form.content}
          onChange={handleText('content')}
          fullWidth
          multiline
          rows={20}
          size="small"
          inputProps={{ style: { fontFamily: 'monospace', fontSize: 13 } }}
          sx={{ mb: 3 }}
        />

        {!isNew && (
          <>
            <Divider sx={{ mb: 2 }} />
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              {!data.page.deleted && (
                <Button
                  variant="outlined"
                  color="warning"
                  startIcon={<UnpublishedIcon />}
                  onClick={handleUnpublish}
                >
                  非公開にする
                </Button>
              )}
              <Button
                variant="outlined"
                color="error"
                startIcon={<DeleteIcon />}
                onClick={() => setDeleteDialog(true)}
              >
                削除
              </Button>
              <FormControlLabel
                control={
                  <Switch
                    checked={form.published}
                    onChange={handlePublishedToggle}
                    color="success"
                  />
                }
                label="有効"
                sx={{ ml: 'auto' }}
              />
            </Box>
          </>
        )}
      </Paper>

      <Dialog open={disableDialog} onClose={() => setDisableDialog(false)}>
        <DialogTitle>ページを無効にしますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            無効にすると一覧などで表示されなくなります。保存後に反映されます。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisableDialog(false)}>キャンセル</Button>
          <Button onClick={confirmDisable} color="warning" variant="contained">無効にする</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={deleteDialog} onClose={() => setDeleteDialog(false)}>
        <DialogTitle>ページを削除しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            「{form.name || '(名前なし)'}」を削除します。この操作は取り消せません。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialog(false)}>キャンセル</Button>
          <Button onClick={handleDelete} color="error" variant="contained">削除</Button>
        </DialogActions>
      </Dialog>

      {/* ドラフト選択ダイアログ */}
      <Dialog open={draftDialog} onClose={() => setDraftDialog(false)} fullWidth maxWidth="xs">
        <DialogTitle>ドラフトに設定</DialogTitle>
        <DialogContent>
          {draftsLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
              <CircularProgress size={24} />
            </Box>
          ) : drafts.length === 0 ? (
            <DialogContentText>ドラフトがありません。先にドラフトを作成してください。</DialogContentText>
          ) : (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1, pt: 1 }}>
              {currentDraft && (
                <Button
                  variant="contained"
                  fullWidth
                  disabled={draftRegistering}
                  onClick={() => handleRegisterToDraft(currentDraft.id)}
                >
                  {currentDraft.name || '(名前なし)'} (現在)
                </Button>
              )}
              {drafts.filter(d => !currentDraft || d.id !== currentDraft.id).map(d => (
                <Button
                  key={d.id}
                  variant="outlined"
                  fullWidth
                  disabled={draftRegistering}
                  onClick={() => handleRegisterToDraft(d.id)}
                >
                  {d.name || '(名前なし)'}
                </Button>
              ))}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDraftDialog(false)}>キャンセル</Button>
        </DialogActions>
      </Dialog>

{/* 保存結果スナックバー */}
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
