const BASE = '/manage/api/file'

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options)
  if (res.status === 401) { location.href = '/login'; return null }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export function getFiles(type = '', cursor = '') {
  const params = new URLSearchParams()
  if (type) params.set('type', type)
  if (cursor) params.set('cursor', cursor)
  const qs = params.toString()
  return apiFetch(`${BASE}/${qs ? '?' + qs : ''}`)
}

export async function checkFile(id) {
  const res = await fetch(`${BASE}/${encodeURIComponent(id)}`)
  if (res.status === 404) return null
  if (res.status === 401) { location.href = '/login'; return null }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export function uploadFile(file) {
  const fd = new FormData()
  fd.append('file', file)
  return apiFetch(`${BASE}/`, { method: 'POST', body: fd })
}

export function deleteFile(id) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export function deleteFiles(ids) {
  return apiFetch(`${BASE}/`, { method: 'DELETE', body: JSON.stringify({ ids }) })
}

export function renameFile(id, name) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}/rename`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
}

export function replaceFile(id, file) {
  const fd = new FormData()
  fd.append('file', file)
  return apiFetch(`${BASE}/${encodeURIComponent(id)}/replace`, { method: 'POST', body: fd })
}
