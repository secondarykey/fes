import { useState, useEffect } from 'react'
import { useParams, useSearchParams, useNavigate, Link as RouterLink } from 'react-router'
import {
  Box, Paper, Typography, TextField, Select, MenuItem,
  FormControl, InputLabel, FormControlLabel, Switch, Checkbox,
  Button, Breadcrumbs, Link, Divider, CircularProgress,
  Snackbar, Alert, Chip,
  Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions,
  IconButton, Tooltip,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import PublishIcon from '@mui/icons-material/Publish'
import UnpublishedIcon from '@mui/icons-material/Unpublished'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import PreviewIcon from '@mui/icons-material/Preview'
import AccountTreeIcon from '@mui/icons-material/AccountTree'
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import BookmarkAddIcon from '@mui/icons-material/BookmarkAdd'
import AddIcon from '@mui/icons-material/Add'
import ImageIcon from '@mui/icons-material/Image'
import DeleteIcon from '@mui/icons-material/Delete'
import UnfoldMoreIcon from '@mui/icons-material/UnfoldMore'
import UnfoldLessIcon from '@mui/icons-material/UnfoldLess'
import WrapTextIcon from '@mui/icons-material/WrapText'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import TextIncreaseIcon from '@mui/icons-material/TextIncrease'
import TextDecreaseIcon from '@mui/icons-material/TextDecrease'
import ManageSearchIcon from '@mui/icons-material/ManageSearch'
import { useRef } from 'react'
import { getPage, getChildren, newPage, updatePage, deletePage, publishPage, unpublishPage, getPageImage, uploadPageImage, deletePageImage } from '../api/page'
import { getCurrentDraft, addDraftPage } from '../api/draft'
import TrashCopyDialog from '../components/TrashCopyDialog'

const SITE_TEMPLATE = 1
const PAGE_TEMPLATE = 2
//子ページの Select で「子ページの追加」を表す値（ページ ID と重ならない）
const ADD_CHILD = '__add_child__'

//前・次のページボタンのページ名（長い名前は省略する）
const siblingLabelSx = { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }

//ページ画像に重ねる操作ボタン
const imageOverlayButtonSx = {
  bgcolor: 'background.paper',
  opacity: 0.85,
  boxShadow: 1,
  '&:hover': { opacity: 1, bgcolor: 'background.paper' },
}

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
  //公開日の打ち直しは既定でオフ
  const [publishUpdate, setPublishUpdate] = useState(false)
  const [deleteDialog, setDeleteDialog] = useState(false)
  const [imageDeleteDialog, setImageDeleteDialog] = useState(false)
  const [trashDialog, setTrashDialog] = useState(false)
  const [disableDialog, setDisableDialog] = useState(false)
  const [image, setImage] = useState(null)
  const [imageUploading, setImageUploading] = useState(false)
  const imageInputRef = useRef(null)
  const contentRef = useRef(null)
  const contentBoxRef = useRef(null)
  const tagBarRef = useRef(null)
  const actionsBarRef = useRef(null)
  const [contentExpanded, setContentExpanded] = useState(false)
  const [expandedRows, setExpandedRows] = useState(30)
  const [wordWrap, setWordWrap] = useState(true)
  const [contentFontSize, setContentFontSize] = useState(13)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const [draftRegistering, setDraftRegistering] = useState(false)
  //同じ親を持つページのうち、並び順で前と次のページ
  const [siblings, setSiblings] = useState({ prev: null, next: null })
  const parentId = data?.page?.parent || ''

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
    if (isNew || !key) {
      setImage(null)
      return
    }
    getPageImage(key)
      .then(d => setImage(d))
      .catch(() => {})
  }, [key, isNew])

  useEffect(() => {
    if (isNew || !key || !parentId) {
      setSiblings({ prev: null, next: null })
      return
    }
    let canceled = false
    getChildren(parentId)
      .then(list => {
        if (canceled) return
        const items = list || []
        const i = items.findIndex(c => c.id === key)
        setSiblings({
          prev: i > 0 ? items[i - 1] : null,
          next: i >= 0 && i < items.length - 1 ? items[i + 1] : null,
        })
      })
      .catch(() => {
        if (!canceled) setSiblings({ prev: null, next: null })
      })
    return () => { canceled = true }
  }, [key, isNew, parentId])

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
      await publishPage(key, publishUpdate)
      showSnack('HTMLを公開しました')
      await Promise.all([refreshPage(), refreshImage(key)])
    } catch (e) { showSnack(e.message, 'error') }
  }

  const handleCopyContent = async () => {
    const text = form.content ?? ''
    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text)
      } else {
        //http 経由など secure context でない場合のフォールバック
        const textarea = contentRef.current?.querySelector('textarea')
        if (!textarea) throw new Error('textarea not found')
        const { selectionStart, selectionEnd } = textarea
        textarea.select()
        const ok = document.execCommand('copy')
        textarea.setSelectionRange(selectionStart, selectionEnd)
        if (!ok) throw new Error('execCommand copy failed')
      }
      showSnack('コンテンツをコピーしました')
    } catch {
      showSnack('コピーに失敗しました', 'error')
    }
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
    setImageDeleteDialog(false)
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

  const wrapSelection = (tag) => {
    const el = contentRef.current?.querySelector('textarea')
    if (!el) return
    const { selectionStart, selectionEnd } = el
    if (selectionStart === selectionEnd) return
    const text = el.value
    const selected = text.slice(selectionStart, selectionEnd)
    const wrapped = `<${tag}>${selected}</${tag}>`
    const updated = text.slice(0, selectionStart) + wrapped + text.slice(selectionEnd)
    const { scrollTop } = el
    setForm(prev => ({ ...prev, content: updated }))
    requestAnimationFrame(() => {
      el.focus()
      el.scrollTop = scrollTop
      // 挿入したタグを含めて選択を維持し、続けて操作できるようにする
      el.setSelectionRange(selectionStart, selectionStart + wrapped.length)
    })
  }

  const insertBr = () => {
    const el = contentRef.current?.querySelector('textarea')
    if (!el) return
    const { selectionStart, selectionEnd } = el
    if (selectionStart === selectionEnd) return
    const text = el.value
    const selected = text.slice(selectionStart, selectionEnd)
    const replaced = selected.replace(/\n/g, '<br>\n')
    const updated = text.slice(0, selectionStart) + replaced + text.slice(selectionEnd)
    const { scrollTop } = el
    setForm(prev => ({ ...prev, content: updated }))
    requestAnimationFrame(() => {
      el.focus()
      el.scrollTop = scrollTop
      // 挿入したタグを含めて選択を維持し、続けて操作できるようにする
      el.setSelectionRange(selectionStart, selectionStart + replaced.length)
    })
  }

  //Trash のページから説明・コンテンツを写す（保存はしない）
  const handleApplyFromTrash = ({ description, content }) => {
    setForm(prev => ({ ...prev, description, content }))
    setTrashDialog(false)
    showSnack('説明・コンテンツを設定しました（未保存）')
  }

  const handleAddChild = async () => {
    try {
      const data = await newPage(key)
      if (data) navigate(`/page/new?parent=${key}&key=${data.page.id}`)
    } catch (e) {
      showSnack(e.message, 'error')
    }
  }

  const handleAddToDraft = async () => {
    setDraftRegistering(true)
    try {
      const cur = await getCurrentDraft()
      if (!cur?.id) {
        showSnack('ドラフトが設定されていません', 'error')
        return
      }
      await addDraftPage(cur.id, key)
      showSnack('ドラフトに登録しました')
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
        <Box sx={{ display: 'flex', gap: 3, mb: 2, alignItems: 'stretch' }}>

          {/* 左カラム */}
          <Box sx={{ flex: 1, minWidth: 0 }}>
            {/* ページ名 + Trash から説明・コンテンツを設定 */}
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
              <TextField
                label="ページ名"
                value={form.name}
                onChange={handleText('name')}
                fullWidth
                size="small"
              />
              <Tooltip title="Trash から説明・コンテンツを設定">
                <IconButton onClick={() => setTrashDialog(true)} aria-label="Trash から説明・コンテンツを設定">
                  <ManageSearchIcon />
                </IconButton>
              </Tooltip>
            </Box>

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
            <Box id="target-section" sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
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

          {/* 右カラム: ページ画像（枠は左カラムと同じ高さ） */}
          <Box
            sx={{
              width: 260,
              flexShrink: 0,
              position: 'relative',
              borderWidth: '1px',
              borderStyle: image?.url ? 'solid' : 'dashed',
              borderColor: 'divider',
              borderRadius: 1,
              overflow: 'hidden',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {isNew ? (
              <Typography variant="body2" color="text.disabled">保存後に設定できます</Typography>
            ) : (
              <>
                {image?.url ? (
                  <>
                    {/* 画像の大きさで枠が伸びないよう、絶対配置で枠いっぱいに収める */}
                    <Box
                      component="img"
                      src={image.url}
                      alt="ページ画像"
                      sx={{ position: 'absolute', inset: 0, width: '100%', height: '100%', objectFit: 'contain' }}
                    />
                    <Chip
                      label={image.draft ? '下書き' : '公開中'}
                      size="small"
                      color={image.draft ? 'warning' : 'success'}
                      sx={{ position: 'absolute', top: 8, left: 8 }}
                    />
                  </>
                ) : (
                  <Typography variant="body2" color="text.disabled">画像なし</Typography>
                )}
                <input
                  type="file"
                  accept="image/*"
                  ref={imageInputRef}
                  style={{ display: 'none' }}
                  onChange={handleImageSelect}
                />
                <Box sx={{ position: 'absolute', bottom: 8, right: 8, display: 'flex', gap: 0.5 }}>
                  <Tooltip title="画像を選択">
                    <span>
                      <IconButton
                        size="small"
                        onClick={() => imageInputRef.current?.click()}
                        disabled={imageUploading}
                        aria-label="画像を選択"
                        sx={imageOverlayButtonSx}
                      >
                        {imageUploading ? <CircularProgress size={18} /> : <ImageIcon fontSize="small" />}
                      </IconButton>
                    </span>
                  </Tooltip>
                  {image?.url && (
                    <Tooltip title="画像を削除">
                      <span>
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => setImageDeleteDialog(true)}
                          disabled={imageUploading}
                          aria-label="画像を削除"
                          sx={imageOverlayButtonSx}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </span>
                    </Tooltip>
                  )}
                </Box>
              </>
            )}
          </Box>
        </Box>

        {/* 子ページ・コンテンツの表示 + 保存/公開 */}
        <Box ref={actionsBarRef} sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {!isNew && (
              <FormControl size="small" sx={{ minWidth: 200 }}>
                <InputLabel>子ページ</InputLabel>
                <Select
                  value=""
                  label="子ページ"
                  onChange={e => {
                    if (e.target.value === ADD_CHILD) handleAddChild()
                    else navigate(`/page/${e.target.value}`)
                  }}
                >
                  <MenuItem value={ADD_CHILD}>
                    <AddIcon fontSize="small" sx={{ mr: 1, color: 'primary.main' }} />
                    子ページの追加
                  </MenuItem>
                  {(data.children || []).length > 0 && <Divider />}
                  {(data.children || []).map(c => (
                    <MenuItem
                      key={c.id}
                      value={c.id}
                      sx={c.deleted ? { textDecoration: 'line-through', color: 'text.disabled' } : undefined}
                    >
                      {c.name || '(名前なし)'}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}
            <Tooltip title="フォント縮小">
              <span>
                <IconButton
                  size="small"
                  onClick={() => setContentFontSize(v => Math.max(9, v - 1))}
                  disabled={contentFontSize <= 9}
                  aria-label="フォント縮小"
                >
                  <TextDecreaseIcon fontSize="small" />
                </IconButton>
              </span>
            </Tooltip>
            <Tooltip title="フォント拡大">
              <span>
                <IconButton
                  size="small"
                  onClick={() => setContentFontSize(v => Math.min(22, v + 1))}
                  disabled={contentFontSize >= 22}
                  aria-label="フォント拡大"
                >
                  <TextIncreaseIcon fontSize="small" />
                </IconButton>
              </span>
            </Tooltip>
            <Tooltip title={wordWrap ? '折り返しOFF' : '折り返しON'}>
              <IconButton
                size="small"
                onClick={() => setWordWrap(v => !v)}
                color={wordWrap ? 'primary' : 'default'}
                aria-label={wordWrap ? '折り返しOFF' : '折り返しON'}
                aria-pressed={wordWrap}
              >
                <WrapTextIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          </Box>
          <Box sx={{ ml: 'auto', display: 'flex', gap: 1 }}>
            {!isNew && data.page.canPublish && (
              <>
                <Tooltip title="オンにすると公開日を現在時刻で打ち直します。オフの場合は最終公開日のみ更新されます">
                  <FormControlLabel
                    sx={{ mr: 0 }}
                    control={
                      <Checkbox
                        size="small"
                        checked={publishUpdate}
                        onChange={e => setPublishUpdate(e.target.checked)}
                      />
                    }
                    label={<Typography variant="body2">公開日を更新</Typography>}
                  />
                </Tooltip>
                <Button
                  variant="outlined"
                  size="small"
                  color="success"
                  startIcon={<PublishIcon />}
                  onClick={handlePublish}
                >
                  公開
                </Button>
              </>
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
        <Box ref={contentBoxRef} sx={{ position: 'relative', mb: 0.5 }}>
          <TextField
            ref={contentRef}
            label="コンテンツ"
            value={form.content}
            onChange={handleText('content')}
            fullWidth
            multiline
            rows={contentExpanded ? expandedRows : 4}
            size="small"
            slotProps={{ htmlInput: { style: { fontFamily: 'monospace', fontSize: contentFontSize, whiteSpace: wordWrap ? 'pre-wrap' : 'pre', overflowX: wordWrap ? 'hidden' : 'auto' } } }}
            sx={{
              '& textarea': {
                transition: 'height 0.3s ease',
              },
            }}
          />
          <Tooltip title={contentExpanded ? '縮小' : '拡大'}>
            <IconButton
              size="small"
              onClick={() => {
                const next = !contentExpanded
                if (next && contentBoxRef.current) {
                  // タグ挿入バーが見える高さから行数を計算（1行 ≒ 19px）
                  const lineH = 19
                  const viewH = window.innerHeight
                  const textareaTop = contentBoxRef.current.getBoundingClientRect().top
                  const available = viewH - textareaTop
                  setExpandedRows(Math.max(10, Math.floor(available / lineH) + 5))
                }
                setContentExpanded(next)
                // アニメーション完了後にスクロール
                if (next) {
                  const textarea = contentRef.current?.querySelector('textarea')
                  if (textarea) {
                    const onEnd = () => {
                      textarea.removeEventListener('transitionend', onEnd)
                      const el = document.getElementById('target-section')
                      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
                    }
                    textarea.addEventListener('transitionend', onEnd, { once: true })
                  }
                }
              }}
              sx={{
                position: 'absolute',
                top: 4,
                right: 8,
                bgcolor: 'background.paper',
                opacity: 0.7,
                '&:hover': { opacity: 1, bgcolor: 'background.paper' },
              }}
            >
              {contentExpanded ? <UnfoldLessIcon fontSize="small" /> : <UnfoldMoreIcon fontSize="small" />}
            </IconButton>
          </Tooltip>
          <Tooltip title="コンテンツをコピー">
            <IconButton
              size="small"
              onClick={handleCopyContent}
              sx={{
                position: 'absolute',
                bottom: 8,
                right: 8,
                bgcolor: 'background.paper',
                opacity: 0.7,
                '&:hover': { opacity: 1, bgcolor: 'background.paper' },
              }}
            >
              <ContentCopyIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Box>
        <Box ref={tagBarRef} sx={{ display: 'flex', gap: 0.5, mb: 3, flexWrap: 'wrap' }}>
          <Button size="small" variant="outlined" sx={{ minWidth: 0, px: 1, textTransform: 'none', fontFamily: 'monospace', fontSize: 12 }}
            onClick={() => wrapSelection('p')}>
            &lt;p&gt;
          </Button>
          <Button size="small" variant="outlined" sx={{ minWidth: 0, px: 1, textTransform: 'none', fontFamily: 'monospace', fontSize: 12 }}
            onClick={insertBr}>
            &lt;br&gt;
          </Button>
          {!isNew && (
            <Button
              variant="text"
              size="small"
              startIcon={<BookmarkAddIcon />}
              onClick={handleAddToDraft}
              disabled={draftRegistering}
              sx={{ ml: 'auto' }}
            >
              ドラフトに設定
            </Button>
          )}
        </Box>

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

            <Divider sx={{ my: 2 }} />
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
            </Box>

            {parentId && (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 1 }}>
                <Box sx={{ flex: 1, minWidth: 0, display: 'flex' }}>
                  {siblings.prev && (
                    <Tooltip title="前のページ">
                      <Button
                        variant="text"
                        size="small"
                        startIcon={<ChevronLeftIcon />}
                        component={RouterLink}
                        to={`/page/${siblings.prev.id}`}
                        sx={{ maxWidth: '100%', textTransform: 'none' }}
                      >
                        <Box component="span" sx={{ ...siblingLabelSx, textDecoration: siblings.prev.deleted ? 'line-through' : 'none' }}>
                          {siblings.prev.name || '(名前なし)'}
                        </Box>
                      </Button>
                    </Tooltip>
                  )}
                </Box>
                <Button
                  variant="text"
                  size="small"
                  startIcon={<AccountTreeIcon />}
                  component={RouterLink}
                  to={`/page/${parentId}/children`}
                  sx={{ flexShrink: 0 }}
                >
                  親の子ページ管理
                </Button>
                <Box sx={{ flex: 1, minWidth: 0, display: 'flex', justifyContent: 'flex-end' }}>
                  {siblings.next && (
                    <Tooltip title="次のページ">
                      <Button
                        variant="text"
                        size="small"
                        endIcon={<ChevronRightIcon />}
                        component={RouterLink}
                        to={`/page/${siblings.next.id}`}
                        sx={{ maxWidth: '100%', textTransform: 'none' }}
                      >
                        <Box component="span" sx={{ ...siblingLabelSx, textDecoration: siblings.next.deleted ? 'line-through' : 'none' }}>
                          {siblings.next.name || '(名前なし)'}
                        </Box>
                      </Button>
                    </Tooltip>
                  )}
                </Box>
              </Box>
            )}
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

      <TrashCopyDialog
        open={trashDialog}
        initialQuery={form.name}
        onClose={() => setTrashDialog(false)}
        onApply={handleApplyFromTrash}
      />

      <Dialog open={imageDeleteDialog} onClose={() => setImageDeleteDialog(false)}>
        <DialogTitle>
          {image?.draft ? '下書きの画像を削除しますか？' : 'ページ画像を削除しますか？'}
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            {image?.draft
              ? '下書きの画像を削除します。公開中の画像がある場合は、その画像に戻ります。'
              : '公開中のページ画像を削除します。この操作は取り消せません。'}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setImageDeleteDialog(false)}>キャンセル</Button>
          <Button onClick={handleImageDelete} color="error" variant="contained">削除</Button>
        </DialogActions>
      </Dialog>

{/* 保存結果スナックバー */}
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
