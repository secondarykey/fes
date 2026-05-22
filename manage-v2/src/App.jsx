import { useMemo, useState } from 'react'
import { createTheme, ThemeProvider, useMediaQuery } from '@mui/material'
import CssBaseline from '@mui/material/CssBaseline'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ColorModeContext } from './context/ColorMode'
import Layout from './components/Layout'
import PageList from './pages/PageList'
import PageEdit from './pages/PageEdit'
import FileList from './pages/FileList'
import TemplateList from './pages/TemplateList'
import TemplateEdit from './pages/TemplateEdit'
import SiteEdit from './pages/SiteEdit'
import VariableList from './pages/VariableList'
import VariableEdit from './pages/VariableEdit'
import DraftList from './pages/DraftList'
import DraftEdit from './pages/DraftEdit'
import ToolsPage from './pages/ToolsPage'
import PageChildren from './pages/PageChildren'

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
        <BrowserRouter basename="/manage/v2">
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
              <Route path="tools" element={<ToolsPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ThemeProvider>
    </ColorModeContext.Provider>
  )
}
