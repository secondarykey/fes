const BASE = '/manage/api/draft'

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options)
  if (res.status === 401) { location.href = '/login'; return null }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

function json(body) {
  return { headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }
}

export const getCurrentDraft = () =>
  apiFetch(`${BASE}/current`)

export const setCurrentDraft = (id) =>
  apiFetch(`${BASE}/current/${encodeURIComponent(id)}`, { method: 'POST' })

export const getDrafts = (cursor = '') =>
  apiFetch(`${BASE}/${cursor ? '?cursor=' + encodeURIComponent(cursor) : ''}`)

export const createDraft = (name) =>
  apiFetch(`${BASE}/`, { method: 'POST', ...json({ name }) })

export const getDraft = (id) =>
  apiFetch(`${BASE}/${encodeURIComponent(id)}`)

export const updateDraft = (id, body) =>
  apiFetch(`${BASE}/${encodeURIComponent(id)}`, { method: 'POST', ...json(body) })

export const deleteDraft = (id) =>
  apiFetch(`${BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' })

export const publishDraft = (id) =>
  apiFetch(`${BASE}/${encodeURIComponent(id)}/publish`, { method: 'POST' })

export const addDraftPage = (id, pageId) =>
  apiFetch(`${BASE}/${encodeURIComponent(id)}/page`, { method: 'POST', ...json({ pageId }) })

export const addDraftPages = (id, pageIds) =>
  apiFetch(`${BASE}/${encodeURIComponent(id)}/pages`, { method: 'POST', ...json({ pageIds }) })

export const removeDraftPage = (id, pageKey) =>
  apiFetch(`${BASE}/${encodeURIComponent(id)}/page/${encodeURIComponent(pageKey)}`, { method: 'DELETE' })
