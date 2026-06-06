const BASE = '/manage/api/template'

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options)
  if (res.status === 401) { location.href = '/login'; return null }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export function getTemplates(type = 'all', cursor = '') {
  const params = new URLSearchParams()
  if (type) params.set('type', type)
  if (cursor) params.set('cursor', cursor)
  return apiFetch(`${BASE}/?${params.toString()}`)
}

export function newTemplate() {
  return apiFetch(`${BASE}/new`)
}

export function getTemplate(id) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}`)
}

export function updateTemplate(id, body) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export function deleteTemplate(id) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export function sequenceTemplates(items) {
  return apiFetch(`${BASE}/sequence`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(items),
  })
}

export function getTemplateReferences(id) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}/references`)
}

export function publishPages(ids) {
  return apiFetch('/manage/api/html/publish-pages', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ids }),
  })
}
