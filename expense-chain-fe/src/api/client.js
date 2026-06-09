// Thin fetch wrapper over the Go backend.
// Handles Bearer auth, JSON, and automatic access-token refresh on 401.

const ACCESS_KEY = 'ec_access'
const REFRESH_KEY = 'ec_refresh'

export function getAccess() {
  return localStorage.getItem(ACCESS_KEY)
}
export function getRefresh() {
  return localStorage.getItem(REFRESH_KEY)
}
export function setTokens(access, refresh) {
  if (access) localStorage.setItem(ACCESS_KEY, access)
  if (refresh) localStorage.setItem(REFRESH_KEY, refresh)
}
export function clearTokens() {
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

// Decode the JWT payload (no verification — UI gating only).
export function decodeToken(token) {
  if (!token) return null
  try {
    const payload = token.split('.')[1]
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
    return JSON.parse(json)
  } catch {
    return null
  }
}

async function tryRefresh() {
  const refresh = getRefresh()
  if (!refresh) return false
  const res = await fetch('/api/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refresh }),
  })
  if (!res.ok) return false
  const data = await res.json()
  setTokens(data.access_token, data.refresh_token)
  return true
}

// Core request. Adds Bearer header, retries once after a refresh on 401.
async function request(method, path, body, _retried = false) {
  const headers = { 'Content-Type': 'application/json' }
  const access = getAccess()
  if (access) headers['Authorization'] = `Bearer ${access}`

  const res = await fetch(path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401 && !_retried) {
    const ok = await tryRefresh()
    if (ok) return request(method, path, body, true)
    clearTokens()
    throw new ApiError(401, 'Sessione scaduta, effettua di nuovo il login')
  }

  if (res.status === 204) return null

  const text = await res.text()
  const data = text ? JSON.parse(text) : null

  if (!res.ok) {
    throw new ApiError(res.status, data?.error || `Errore ${res.status}`)
  }
  return data
}

export class ApiError extends Error {
  constructor(status, message) {
    super(message)
    this.status = status
  }
}

export const api = {
  // --- Auth ---
  login: (username, password) => request('POST', '/api/auth/login', { username, password }),
  register: (payload) => request('POST', '/api/auth/register', payload),

  // --- Companies ---
  listCompanies: () => request('GET', '/api/companies'),
  getCompany: (id) => request('GET', `/api/companies/${id}`),
  createCompany: (payload) => request('POST', '/api/companies', payload),
  deleteCompany: (id) => request('DELETE', `/api/companies/${id}`),

  // --- Employees ---
  listEmployees: () => request('GET', '/api/employees'),
  createEmployee: (payload) => request('POST', '/api/employees', payload),
  deleteEmployee: (id) => request('DELETE', `/api/employees/${id}`),

  // --- Cards ---
  createCard: (payload) => request('POST', '/api/cards', payload),
  getCardsByEmployee: (employeeId) => request('GET', `/api/cards/employee/${employeeId}`),
  deleteCard: (id) => request('DELETE', `/api/cards/${id}`),

  // --- Policies ---
  listPolicies: () => request('GET', '/api/policies'),
  createPolicy: (payload) => request('POST', '/api/policies', payload),
  deletePolicy: (id) => request('DELETE', `/api/policies/${id}`),

  // --- Transactions ---
  listTransactions: () => request('GET', '/api/transactions'),
  submitTransaction: (payload) => request('POST', '/api/transactions', payload),
  incomingPayment: (payload) => request('POST', '/api/payments/incoming', payload),
  getTransactionsByEmployee: (employeeId) => request('GET', `/api/transactions/employee/${employeeId}`),

  // --- Ledger ---
  getLedger: () => request('GET', '/api/ledger'),
  verifyLedger: () => request('GET', '/api/ledger/verify'),
  getLedgerByEntity: (entityId) => request('GET', `/api/ledger/entity/${entityId}`),
}

// Constants mirrored from the backend models.
export const CATEGORIES = ['FOOD', 'TRAVEL', 'ACCOMMODATION', 'ENTERTAINMENT', 'ELECTRONICS', 'OTHER']
export const CARD_TYPES = ['PHYSICAL', 'VIRTUAL', 'PREPAID']
export const ROLES = ['ADMIN', 'COMPANY', 'EMPLOYEE']
