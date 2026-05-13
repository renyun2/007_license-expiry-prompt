const json = (path, opt = {}) =>
  fetch(path, {
    headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
    ...opt,
  }).then(async (r) => {
    if (!r.ok) throw new Error((await r.text()) || r.statusText)
    const ct = r.headers.get('content-type') || ''
    if (ct.includes('application/json')) return r.json()
    return r.text()
  })

export const api = {
  health: () => json('/api/health'),
  dashboard: () => json('/api/dashboard'),
  certificates: (q) => json('/api/certificates' + (q ? `?${q}` : '')),
  urgent: () => json('/api/certificates/urgent'),
  exportCsv: (category) => {
    window.open('/api/certificates/export?' + new URLSearchParams({ category: category || '' }).toString())
  },
  getCert: (id) => json(`/api/certificates/${id}`),
  createCert: (body) => json('/api/certificates', { method: 'POST', body: JSON.stringify(body) }),
  updateCert: (id, body) => json(`/api/certificates/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteCert: (id) => fetch(`/api/certificates/${id}`, { method: 'DELETE' }).then((r) => {
    if (!r.ok) throw new Error(r.statusText)
  }),
  reminders: () => json('/api/reminders'),
  updateReminder: (id, body) => json(`/api/reminders/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  renewals: () => json('/api/renewals'),
  createRenewal: (body) => json('/api/renewals', { method: 'POST', body: JSON.stringify(body) }),
  patchRenewal: (id, body) => json(`/api/renewals/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  inspections: (cid) => json('/api/inspections' + (cid ? `?certificate_id=${cid}` : '')),
  createInspection: (body) => json('/api/inspections', { method: 'POST', body: JSON.stringify(body) }),
  fees: (cid) => json('/api/fees' + (cid ? `?certificate_id=${cid}` : '')),
  createFee: (body) => json('/api/fees', { method: 'POST', body: JSON.stringify(body) }),
  feesSummary: (year) => json('/api/fees/summary?' + new URLSearchParams({ year: String(year) }).toString()),
  feesByCategory: (year) => json('/api/fees/by-category?' + new URLSearchParams({ year: String(year) }).toString()),
  calendar: (year) => json(`/api/calendar/${year}`),
  todos: () => json('/api/todos'),
  patchTodo: (id, done) => json(`/api/todos/${id}`, { method: 'PATCH', body: JSON.stringify({ done }) }),
}
