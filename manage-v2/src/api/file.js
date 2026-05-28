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
