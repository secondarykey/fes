import { useState } from 'react'
import { Outlet, Link as RouterLink, useLocation } from 'react-router-dom'
import {
  AppBar, Toolbar, IconButton, Typography, Drawer,
  List, ListItem, ListItemButton, ListItemIcon, ListItemText,
  Box, Divider, Tooltip, useTheme,
} from '@mui/material'
import MenuIcon from '@mui/icons-material/Menu'
import ArticleIcon from '@mui/icons-material/Article'
import FolderIcon from '@mui/icons-material/Folder'
import CodeIcon from '@mui/icons-material/Code'
import SettingsIcon from '@mui/icons-material/Settings'
import DataObjectIcon from '@mui/icons-material/DataObject'
import DraftsIcon from '@mui/icons-material/Drafts'
import BuildIcon from '@mui/icons-material/Build'
import DeleteIcon from '@mui/icons-material/Delete'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import LightModeIcon from '@mui/icons-material/LightMode'
import { useColorMode } from '../context/ColorMode'

const DRAWER_WIDTH = 240

export default function Layout() {
  const [open, setOpen] = useState(true)
  const location = useLocation()
  const theme = useTheme()
  const { mode, toggleColorMode } = useColorMode()

  const isDark = mode === 'dark'

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar
        position="fixed"
        elevation={1}
        sx={{ zIndex: t => t.zIndex.drawer + 1 }}
      >
        <Toolbar variant="dense">
          <IconButton
            color="inherit"
            onClick={() => setOpen(v => !v)}
            edge="start"
            sx={{ mr: 2 }}
            aria-label="メニューを開閉"
          >
            <MenuIcon />
          </IconButton>

          <Typography variant="h6" noWrap component="div" sx={{ fontWeight: 600, flexGrow: 1 }}>
            FES 管理画面
          </Typography>

          <Tooltip title={isDark ? 'ライトモードに切り替え' : 'ダークモードに切り替え'}>
            <IconButton color="inherit" onClick={toggleColorMode} aria-label="テーマ切り替え">
              {isDark ? <LightModeIcon /> : <DarkModeIcon />}
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="persistent"
        open={open}
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: DRAWER_WIDTH,
            boxSizing: 'border-box',
            bgcolor: isDark ? 'grey.900' : 'grey.200',
            color: isDark ? 'grey.100' : 'grey.900',
            borderRight: 'none',
          },
        }}
      >
        <Toolbar variant="dense" />
        <Box sx={{ overflow: 'auto', flex: 1, px: 1, pt: 1 }}>
          <List dense>
            {[
              { to: '/page', label: 'Pages', icon: <ArticleIcon fontSize="small" /> },
              { to: '/file', label: 'Files', icon: <FolderIcon fontSize="small" /> },
              { to: '/template', label: 'Templates', icon: <CodeIcon fontSize="small" /> },
              { to: '/draft', label: 'Drafts', icon: <DraftsIcon fontSize="small" /> },
              { to: '/variable', label: 'Variables', icon: <DataObjectIcon fontSize="small" /> },
            ].map(({ to, label, icon }) => (
              <ListItem key={to} disablePadding>
                <ListItemButton
                  component={RouterLink}
                  to={to}
                  selected={
                    to === '/page'
                      ? location.pathname.startsWith(to) && !location.pathname.startsWith('/page/Trash')
                      : location.pathname.startsWith(to)
                  }
                  sx={{
                    color: isDark ? 'grey.300' : 'grey.800',
                    '&.Mui-selected': { bgcolor: 'primary.dark', color: 'white' },
                    '&.Mui-selected:hover': { bgcolor: 'primary.dark' },
                    '&:hover': { bgcolor: isDark ? 'grey.800' : 'grey.300' },
                  }}
                >
                  <ListItemIcon sx={{ minWidth: 36, color: 'inherit' }}>
                    {icon}
                  </ListItemIcon>
                  <ListItemText primary={label} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Box>

        <Divider sx={{ borderColor: isDark ? 'grey.700' : 'grey.400' }} />
        <Box sx={{ px: 1, pb: 1 }}>
          <List dense>
            {[
              { to: '/page/Trash/children', label: 'Trash', icon: <DeleteIcon fontSize="small" /> },
              { to: '/tools', label: 'Tools', icon: <BuildIcon fontSize="small" /> },
              { to: '/site', label: 'Site Settings', icon: <SettingsIcon fontSize="small" /> },
            ].map(({ to, label, icon }) => (
              <ListItem key={to} disablePadding>
                <ListItemButton
                  component={RouterLink}
                  to={to}
                  selected={location.pathname.startsWith(to)}
                  sx={{
                    color: isDark ? 'grey.300' : 'grey.800',
                    '&.Mui-selected': { bgcolor: 'primary.dark', color: 'white' },
                    '&.Mui-selected:hover': { bgcolor: 'primary.dark' },
                    '&:hover': { bgcolor: isDark ? 'grey.800' : 'grey.300' },
                  }}
                >
                  <ListItemIcon sx={{ minWidth: 36, color: 'inherit' }}>
                    {icon}
                  </ListItemIcon>
                  <ListItemText primary={label} />
                </ListItemButton>
              </ListItem>
            ))}
            <ListItem disablePadding>
              <ListItemButton
                component="a"
                href="/manage/"
                sx={{
                  color: isDark ? 'grey.400' : 'grey.600',
                  '&:hover': { bgcolor: isDark ? 'grey.800' : 'grey.300', color: isDark ? 'grey.100' : 'grey.900' },
                }}
              >
                <ListItemIcon sx={{ minWidth: 36, color: 'inherit' }}>
                  <ArrowBackIcon fontSize="small" />
                </ListItemIcon>
                <ListItemText primary="v1 に戻る" />
              </ListItemButton>
            </ListItem>
          </List>
        </Box>
      </Drawer>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          transition: theme.transitions.create('margin', {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
          }),
          marginLeft: open ? 0 : `-${DRAWER_WIDTH}px`,
          minHeight: '100vh',
          bgcolor: 'background.default',
        }}
      >
        <Toolbar variant="dense" />
        <Outlet />
      </Box>
    </Box>
  )
}
