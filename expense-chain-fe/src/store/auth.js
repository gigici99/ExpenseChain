// Reactive auth state shared across the app.
import { reactive, computed } from 'vue'
import { api, setTokens, clearTokens, getAccess, decodeToken } from '../api/client.js'

const state = reactive({
  claims: decodeToken(getAccess()),
})

export function useAuth() {
  async function login(username, password) {
    const pair = await api.login(username, password)
    setTokens(pair.access_token, pair.refresh_token)
    state.claims = decodeToken(pair.access_token)
  }

  function logout() {
    clearTokens()
    state.claims = null
  }

  const isAuthenticated = computed(() => !!state.claims)
  const role = computed(() => state.claims?.role || null)
  const username = computed(() => state.claims?.username || null)
  const can = (...roles) => {
    if (!state.claims) return false
    if (state.claims.role === 'ADMIN') return true
    return roles.includes(state.claims.role)
  }

  return { state, login, logout, isAuthenticated, role, username, can }
}
