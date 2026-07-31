import { lazy, Suspense, useMemo, useState } from 'react'
import { createTheme, ThemeProvider, useMediaQuery } from '@mui/material'
import Box from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'
import CssBaseline from '@mui/material/CssBaseline'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import { ColorModeContext } from './context/ColorMode'
import Layout from './components/Layout'

// 各ページはルート単位で遅延ロードし、初回のチャンクサイズを抑える。
// 特に DraftEdit / PageChildren は @dnd-kit を引き込むため分離の効果が大きい。
const PageList = lazy(() => import('./pages/PageList'))
const PageEdit = lazy(() => import('./pages/PageEdit'))
const FileList = lazy(() => import('./pages/FileList'))
const TemplateList = lazy(() => import('./pages/TemplateList'))
const TemplateEdit = lazy(() => import('./pages/TemplateEdit'))
const SiteEdit = lazy(() => import('./pages/SiteEdit'))
const VariableList = lazy(() => import('./pages/VariableList'))
const VariableEdit = lazy(() => import('./pages/VariableEdit'))
const DraftList = lazy(() => import('./pages/DraftList'))
const DraftEdit = lazy(() => import('./pages/DraftEdit'))
const PublishPage = lazy(() => import('./pages/PublishPage'))
const PageChildren = lazy(() => import('./pages/PageChildren'))

function PageFallback() {
  return (
    <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
      <CircularProgress size={32} />
    </Box>
  )
}

const STORAGE_KEY = 'fes-color-mode'

export default function App() {
  const prefersDark = useMediaQuery('(prefers-color-scheme: dark)')

  // null = システム設定に従う / 'light' | 'dark' = ユーザー指定
  const [userMode, setUserMode] = useState(() => {
    const saved = localStorage.getItem(STORAGE_KEY)
    return saved === 'light' || saved === 'dark' ? saved : null
  })

  const mode = userMode ?? (prefersDark ? 'dark' : 'light')

  const colorModeContext = useMemo(() => ({
    mode,
    toggleColorMode: () => {
      setUserMode(prev => {
        const current = prev ?? (prefersDark ? 'dark' : 'light')
        const next = current === 'dark' ? 'light' : 'dark'
        // システム設定と同じ値になる場合は override を解除する
        if (next === (prefersDark ? 'dark' : 'light')) {
          localStorage.removeItem(STORAGE_KEY)
          return null
        }
        localStorage.setItem(STORAGE_KEY, next)
        return next
      })
    },
  }), [mode, prefersDark])

  const theme = useMemo(() =>
    createTheme({
      palette: {
        mode,
        primary: { main: '#1976d2' },
      },
      components: {
        MuiListItemButton: {
          styleOverrides: { root: { borderRadius: 6 } },
        },
      },
    }),
    [mode]
  )

  return (
    <ColorModeContext.Provider value={colorModeContext}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <BrowserRouter basename="/manage">
          <Suspense fallback={<PageFallback />}>
            <Routes>
              <Route path="/" element={<Layout />}>
                <Route index element={<Navigate to="/page" replace />} />
                <Route path="page" element={<PageList />} />
                <Route path="page/new" element={<PageEdit />} />
                <Route path="page/:key/children" element={<PageChildren />} />
                <Route path="page/:key" element={<PageEdit />} />
                <Route path="file" element={<FileList />} />
                <Route path="template" element={<TemplateList />} />
                <Route path="template/new" element={<TemplateEdit />} />
                <Route path="template/:key" element={<TemplateEdit />} />
                <Route path="site" element={<SiteEdit />} />
                <Route path="variable" element={<VariableList />} />
                <Route path="variable/:key" element={<VariableEdit />} />
                <Route path="draft" element={<DraftList />} />
                <Route path="draft/:key" element={<DraftEdit />} />
                <Route path="publish" element={<PublishPage />} />
              </Route>
            </Routes>
          </Suspense>
        </BrowserRouter>
      </ThemeProvider>
    </ColorModeContext.Provider>
  )
}
