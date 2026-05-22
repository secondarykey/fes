const BASE = '/manage/v2/api/site'

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options)
  if (res.status === 401) { location.href = '/login'; return null }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export function getSite() {
  return apiFetch(`${BASE}/`)
}

export function updateSite(body) {
  return apiFetch(`${BASE}/`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}
