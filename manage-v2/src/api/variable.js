const BASE = '/manage/v2/api/variable'

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options)
  if (res.status === 401) { location.href = '/login'; return null }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export function getVariables(cursor = '') {
  const params = cursor ? `?cursor=${encodeURIComponent(cursor)}` : ''
  return apiFetch(`${BASE}/${params}`)
}

export function getVariable(id) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}`)
}

export function createVariable(id, content) {
  return apiFetch(`${BASE}/`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, content }),
  })
}

export function updateVariable(id, body) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export function deleteVariable(id) {
  return apiFetch(`${BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' })
}
