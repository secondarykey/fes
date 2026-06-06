const BASE = '/manage/api'

async function apiFetch(path, options = {}) {
  const headers = options.body instanceof FormData
    ? { ...options.headers }
    : { 'Content-Type': 'application/json', ...options.headers }
  const res = await fetch(`${BASE}${path}`, { ...options, headers })
  if (res.status === 401) {
    window.location.href = '/login'
    return null
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || res.statusText)
  }
  return res.json()
}

export const getRootPage = () => apiFetch('/page/')

export const getPage = (key) => apiFetch(`/page/${key}`)

export const newPage = (parentKey) => apiFetch(`/page/new/${parentKey}`)

export const getChildren = (key) => apiFetch(`/page/children/${key}`)

export const updatePage = (key, data) =>
  apiFetch(`/page/${key}`, { method: 'POST', body: JSON.stringify(data) })

export const deletePage = (key) =>
  apiFetch(`/page/${key}`, { method: 'DELETE' })

export const deletePages = (ids) =>
  apiFetch('/page/', { method: 'DELETE', body: JSON.stringify({ ids }) })

export const publishPage = (key) =>
  apiFetch(`/html/publish/${key}`, { method: 'POST' })

export const unpublishPage = (key) =>
  apiFetch(`/html/unpublish/${key}`, { method: 'POST' })

export const sequencePages = (key, items) =>
  apiFetch(`/page/${key}/sequence`, { method: 'POST', body: JSON.stringify(items) })

export const sortPages = (key) =>
  apiFetch(`/page/${key}/sort`, { method: 'POST' })

export const movePage = (key, targetId) =>
  apiFetch(`/page/${key}/move`, { method: 'POST', body: JSON.stringify({ targetId }) })

export const getPageListHtml = (parentId) =>
  apiFetch(`/tool/page-list?parent=${encodeURIComponent(parentId)}`)

export const getPageImage = (key) =>
  apiFetch(`/page/${key}/image`)

export const uploadPageImage = (key, file) => {
  const form = new FormData()
  form.append('file', file)
  return apiFetch(`/page/${key}/image`, {
    method: 'POST',
    headers: {},
    body: form,
  })
}

export const deletePageImage = (key) =>
  apiFetch(`/page/${key}/image`, { method: 'DELETE' })
