<script setup>
import { useRouter, useRoute } from 'vue-router'
import { computed } from 'vue'
import { useAuth } from './store/auth.js'

const router = useRouter()
const route = useRoute()
const { isAuthenticated, role, username, can, logout } = useAuth()

const showShell = computed(() => isAuthenticated.value && route.path !== '/login')

function doLogout() {
  logout()
  router.push('/login')
}
</script>

<template>
  <div v-if="showShell" class="app-shell">
    <aside class="sidebar">
      <h1>⛓ Expense Chain</h1>
      <nav>
        <router-link to="/dashboard">Dashboard</router-link>
        <router-link v-if="can('COMPANY')" to="/companies">Aziende</router-link>
        <router-link v-if="can('COMPANY')" to="/employees">Dipendenti</router-link>
        <router-link v-if="can('COMPANY', 'EMPLOYEE')" to="/cards">Carte</router-link>
        <router-link v-if="can('COMPANY')" to="/policies">Policy</router-link>
        <router-link v-if="can('COMPANY', 'EMPLOYEE')" to="/transactions">Transazioni</router-link>
        <router-link v-if="can('COMPANY')" to="/ledger">Ledger</router-link>
      </nav>
      <div class="spacer"></div>
      <div class="user-box">
        <div>{{ username }}</div>
        <div class="role">{{ role }}</div>
        <button class="secondary" style="margin-top:10px;width:100%" @click="doLogout">Logout</button>
      </div>
    </aside>
    <main class="main">
      <router-view />
    </main>
  </div>
  <router-view v-else />
</template>
