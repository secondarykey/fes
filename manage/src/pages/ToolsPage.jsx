import { useState, useRef } from 'react'
import PropTypes from 'prop-types'
import {
  Typography, Box, Grid, Card, CardContent, CardActions,
  Button, Divider, Dialog, DialogTitle, DialogContent,
  DialogContentText, DialogActions, Alert, CircularProgress,
  Chip, LinearProgress,
} from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import MemoryIcon from '@mui/icons-material/Memory'
import CleaningServicesIcon from '@mui/icons-material/CleaningServices'
import RestoreIcon from '@mui/icons-material/Restore'

async function apiFetch(path, options = {}) {
  const res = await fetch(`/manage/api${path}`, options)
  if (!res.ok) {
    const b = await res.json().catch(() => ({}))
    throw new Error(b.error || `HTTP ${res.status}`)
  }
  return res.json()
}

const runGC = () => apiFetch('/tool/gc', { method: 'POST' })
const runClean = () => apiFetch('/tool/clean', { method: 'POST' })

async function runRestore(file) {
  const form = new FormData()
  form.append('file', file)
  const res = await fetch('/manage/api/tool/restore', { method: 'POST', body: form })
  if (!res.ok) {
    const b = await res.json().catch(() => ({}))
    throw new Error(b.error || `HTTP ${res.status}`)
  }
  return res.json()
}

ToolCard.propTypes = { title: PropTypes.string.isRequired, description: PropTypes.string, actions: PropTypes.node, chip: PropTypes.string }
function ToolCard({ title, description, actions, chip }) {
  return (
    <Card variant="outlined" sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <CardContent sx={{ flexGrow: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>{title}</Typography>
          {chip && <Chip label={chip} size="small" color="warning" variant="outlined" />}
        </Box>
        <Typography variant="body2" color="text.secondary">{description}</Typography>
      </CardContent>
      <Divider />
      <CardActions sx={{ p: 1.5, gap: 0.5, flexWrap: 'wrap' }}>{actions}</CardActions>
    </Card>
  )
}

export default function ToolsPage() {
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

  const handleGC = async () => {
    setGcDialog(false)
    setGcRunning(true)
    setGcResult(null)
    setGcError(null)
    try {
      const data = await runGC()
      setGcResult(data.result || 'GC 完了')
    } catch (e) {
      setGcError(e.message)
    } finally {
      setGcRunning(false)
    }
  }

  const handleClean = async () => {
    setCleanDialog(false)
    setCleanRunning(true)
    setCleanResult(null)
    setCleanError(null)
    try {
      const data = await runClean()
      setCleanResult(data)
    } catch (e) {
      setCleanError(e.message)
    } finally {
      setCleanRunning(false)
    }
  }

  const handleRestoreFileChange = (e) => {
    const f = e.target.files?.[0] ?? null
    setRestoreFile(f)
    if (f) setRestoreDialog(true)
  }

  const handleRestore = async () => {
    setRestoreDialog(false)
    setRestoreRunning(true)
    setRestoreResult(null)
    setRestoreError(null)
    try {
      await runRestore(restoreFile)
      setRestoreResult('リストアが完了しました。')
    } catch (e) {
      setRestoreError(e.message)
    } finally {
      setRestoreRunning(false)
      setRestoreFile(null)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 600 }} gutterBottom>Tools</Typography>

      <Grid container spacing={2}>

        {/* サイトクリーン */}
        <Grid size={{ xs: 12, sm: 6, md: 4 }}>
          <ToolCard
            title="サイトクリーン"
            description="孤立した HTML・ページ画像と、無効なページに残っている公開 HTML を検出して削除します。"
            actions={
              <Button
                size="small"
                variant="outlined"
                color="warning"
                startIcon={cleanRunning ? <CircularProgress size={14} /> : <CleaningServicesIcon />}
                onClick={() => setCleanDialog(true)}
                disabled={cleanRunning}
              >
                {cleanRunning ? '実行中...' : 'クリーン実行'}
              </Button>
            }
          />
        </Grid>

        {/* GC */}
        <Grid size={{ xs: 12, sm: 6, md: 4 }}>
          <ToolCard
            title="ガベージコレクション"
            description="Go ランタイムの GC を手動実行します。メモリ使用量の統計をログに記録します。"
            actions={
              <Button
                size="small"
                variant="outlined"
                color="warning"
                startIcon={gcRunning ? <CircularProgress size={14} /> : <MemoryIcon />}
                onClick={() => setGcDialog(true)}
                disabled={gcRunning}
              >
                {gcRunning ? '実行中...' : 'GC 実行'}
              </Button>
            }
          />
        </Grid>

        {/* データ管理 */}
        <Grid size={{ xs: 12, sm: 6, md: 4 }}>
          <ToolCard
            title="データ管理"
            description="Datastore のバックアップ（ZIP ダウンロード）とリストアを行います。リストアは全データを上書きします。"
            chip="注意"
            actions={
              <>
                <Button
                  size="small"
                  variant="outlined"
                  startIcon={<DownloadIcon />}
                  component="a"
                  href="/manage/api/tool/backup"
                >
                  バックアップ
                </Button>
                <Button
                  size="small"
                  variant="outlined"
                  color="warning"
                  startIcon={restoreRunning ? <CircularProgress size={14} /> : <RestoreIcon />}
                  onClick={() => fileInputRef.current?.click()}
                  disabled={restoreRunning}
                >
                  {restoreRunning ? '実行中...' : 'リストア'}
                </Button>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".zip"
                  style={{ display: 'none' }}
                  onChange={handleRestoreFileChange}
                />
              </>
            }
          />
        </Grid>

      </Grid>

      {/* リストア進行中 */}
      {restoreRunning && <LinearProgress sx={{ mt: 3 }} />}

      {/* リストア結果 */}
      {restoreResult && (
        <Alert severity="success" sx={{ mt: 3 }} onClose={() => setRestoreResult(null)}>
          {restoreResult}
        </Alert>
      )}
      {restoreError && (
        <Alert severity="error" sx={{ mt: 3 }} onClose={() => setRestoreError(null)}>
          {restoreError}
        </Alert>
      )}

      {/* クリーン結果 */}
      {cleanResult && (
        <Alert severity="success" sx={{ mt: 3 }} onClose={() => setCleanResult(null)}>
          <Typography variant="body2" sx={{ mb: 0.5 }}>
            孤立HTML削除: {cleanResult.deletedHtmls?.length ?? 0} 件 / 無効ページのHTML削除: {cleanResult.disabledHtmls?.length ?? 0} 件 / 画像削除: {cleanResult.deletedImages?.length ?? 0} 件
          </Typography>
          {cleanResult.errors?.length > 0 && (
            <Typography variant="body2" color="error.main">
              エラー: {cleanResult.errors.join(', ')}
            </Typography>
          )}
          {!cleanResult.deletedHtmls?.length && !cleanResult.disabledHtmls?.length && !cleanResult.deletedImages?.length && (
            <Typography variant="body2">削除対象はありませんでした。</Typography>
          )}
        </Alert>
      )}
      {cleanError && (
        <Alert severity="error" sx={{ mt: 3 }} onClose={() => setCleanError(null)}>
          {cleanError}
        </Alert>
      )}

      {/* GC 結果 */}
      {gcResult && (
        <Alert severity="success" sx={{ mt: 3 }} onClose={() => setGcResult(null)}>
          <Typography variant="body2" component="pre" sx={{ fontFamily: 'monospace', whiteSpace: 'pre-wrap', m: 0 }}>
            {gcResult}
          </Typography>
        </Alert>
      )}
      {gcError && (
        <Alert severity="error" sx={{ mt: 3 }} onClose={() => setGcError(null)}>
          {gcError}
        </Alert>
      )}

      {/* クリーン確認ダイアログ */}
      <Dialog open={cleanDialog} onClose={() => setCleanDialog(false)}>
        <DialogTitle>サイトクリーンを実行しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            ページが存在しない孤立した HTML エンティティとページ画像に加え、無効なページに残っている公開 HTML を削除します。対象ページは公開側で 404 になります。この操作は取り消せません。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCleanDialog(false)}>キャンセル</Button>
          <Button onClick={handleClean} color="warning" variant="contained">実行</Button>
        </DialogActions>
      </Dialog>

      {/* リストア確認ダイアログ */}
      <Dialog open={restoreDialog} onClose={() => { setRestoreDialog(false); setRestoreFile(null); if (fileInputRef.current) fileInputRef.current.value = '' }}>
        <DialogTitle>リストアを実行しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            <strong>{restoreFile?.name}</strong> を使用してリストアを実行します。
            <br /><br />
            この操作は <strong>現在の全データを上書き</strong> します。取り消すことはできません。実行前にバックアップを取得することをお勧めします。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setRestoreDialog(false); setRestoreFile(null); if (fileInputRef.current) fileInputRef.current.value = '' }}>キャンセル</Button>
          <Button onClick={handleRestore} color="error" variant="contained">リストア実行</Button>
        </DialogActions>
      </Dialog>

      {/* GC 確認ダイアログ */}
      <Dialog open={gcDialog} onClose={() => setGcDialog(false)}>
        <DialogTitle>GC を実行しますか？</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Go ランタイムのガベージコレクションを手動実行します。通常は自動で実行されますが、メモリ使用量を即座に解放したい場合に使用します。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setGcDialog(false)}>キャンセル</Button>
          <Button onClick={handleGC} color="warning" variant="contained">実行</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
