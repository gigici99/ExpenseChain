import { createRouter, createWebHistory } from 'vue-router'
import { getAccess, decodeToken } from './api/client.js'

import Login from './views/Login.vue'
import Dashboard from './views/Dashboard.vue'
import Companies from './views/Companies.vue'
import Employees from './views/Employees.vue'
import Cards from './views/Cards.vue'
import Policies from './views/Policies.vue'
import Transactions from './views/Transactions.vue'
import Ledger from './views/Ledger.vue'

const routes = [
  { path: '/login', component: Login, meta: { public: true } },
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', component: Dashboard },
  { path: '/companies', component: Companies },
  { path: '/employees', component: Employees },
  { path: '/cards', component: Cards },
  { path: '/policies', component: Policies },
  { path: '/transactions', component: Transactions },
  { path: '/ledger', component: Ledger },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Redirect to login if not authenticated.
router.beforeEach((to) => {
  const claims = decodeToken(getAccess())
  if (!to.meta.public && !claims) {
    return '/login'
  }
  if (to.path === '/login' && claims) {
    return '/dashboard'
  }
})

export default router
