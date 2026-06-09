<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api/client.js'
import { useAuth } from '../store/auth.js'

const { can, username, role } = useAuth()
const stats = ref({ companies: 0, employees: 0, policies: 0, transactions: 0 })
const ledgerStatus = ref(null)
const error = ref('')

onMounted(async () => {
  try {
    if (can('COMPANY')) {
      const [c, e, p, t] = await Promise.all([
        api.listCompanies().catch(() => []),
        api.listEmployees().catch(() => []),
        api.listPolicies().catch(() => []),
        api.listTransactions().catch(() => []),
      ])
      stats.value = {
        companies: c?.length || 0,
        employees: e?.length || 0,
        policies: p?.length || 0,
        transactions: t?.length || 0,
      }
      ledgerStatus.value = await api.verifyLedger().catch(() => null)
    }
  } catch (e) {
    error.value = e.message
  }
})
</script>

<template>
  <div>
    <h2>Dashboard</h2>
    <p class="subtitle">Benvenuto {{ username }} — ruolo {{ role }}</p>

    <div v-if="error" class="alert error">{{ error }}</div>

    <div v-if="can('COMPANY')" class="form-grid">
      <div class="card"><h3>Aziende</h3><div style="font-size:32px;font-weight:700">{{ stats.companies }}</div></div>
      <div class="card"><h3>Dipendenti</h3><div style="font-size:32px;font-weight:700">{{ stats.employees }}</div></div>
      <div class="card"><h3>Policy</h3><div style="font-size:32px;font-weight:700">{{ stats.policies }}</div></div>
      <div class="card"><h3>Transazioni</h3><div style="font-size:32px;font-weight:700">{{ stats.transactions }}</div></div>
    </div>

    <div v-if="ledgerStatus" class="card">
      <h3>Integrità Ledger</h3>
      <p>
        <span class="badge" :class="ledgerStatus.valid ? 'valid' : 'invalid'">
          {{ ledgerStatus.valid ? 'INTEGRO' : 'COMPROMESSO' }}
        </span>
        — {{ ledgerStatus.entries }} entry
        <span v-if="!ledgerStatus.valid"> · rotto a seq {{ ledgerStatus.broken_at }}: {{ ledgerStatus.broken_why }}</span>
      </p>
    </div>

    <div v-if="!can('COMPANY')" class="card">
      <h3>Dipendente</h3>
      <p class="subtitle">Vai su <router-link to="/transactions">Transazioni</router-link> per inserire una spesa.</p>
    </div>
  </div>
</template>
