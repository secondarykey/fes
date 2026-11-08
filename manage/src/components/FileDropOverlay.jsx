import PropTypes from 'prop-types'
import { Box, Typography, CircularProgress, alpha } from '@mui/material'

FileDropOverlay.propTypes = {
  dragging: PropTypes.bool,
  progress: PropTypes.shape({ done: PropTypes.number, total: PropTypes.number }),
  icon: PropTypes.node,
  label: PropTypes.string.isRequired,
  progressLabel: PropTypes.string.isRequired,
}

//ドロップを受け付ける要素（position: relative）に重ねる表示。
//ドラッグ中は受け付け枠を、処理中は進捗を出す。処理中は下の操作を止める。
export default function FileDropOverlay({ dragging, progress, icon, label, progressLabel }) {
  if (!dragging && !progress) return null
  return (
    <Box
      sx={{
        position: 'absolute',
        inset: 0,
        zIndex: 2,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 1,
        border: '2px dashed',
        borderColor: 'primary.main',
        borderRadius: 1,
        bgcolor: theme => alpha(theme.palette.background.paper, 0.9),
        pointerEvents: progress ? 'auto' : 'none',
      }}
    >
      {progress ? (
        <>
          <CircularProgress />
          <Typography color="text.secondary">
            {progressLabel} {progress.done} / {progress.total}
          </Typography>
        </>
      ) : (
        <>
          {icon}
          <Typography color="primary">{label}</Typography>
        </>
      )}
    </Box>
  )
}
