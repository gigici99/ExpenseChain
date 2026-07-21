<script setup>
import { ref, onMounted, computed } from 'vue'
import { api } from '../api/client.js'
import { useAuth } from '../store/auth.js'

const { can, username, role, state } = useAuth()

const isEmployee = computed(() => role.value === 'EMPLOYEE')
const employeeId = computed(() => state.claims?.employee_id || '')

const stats = ref({ companies: 0, employees: 0, policies: 0, transactions: 0 })
const myTransactions = ref([])
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
    if (isEmployee.value && employeeId.value) {
      myTransactions.value = await api.getTransactionsByEmployee(employeeId.value).catch(() => [])
    }
  } catch (e) {
    error.value = e.message
  }
})

const approvedCount = computed(() => myTransactions.value.filter(t => t.status === 'APPROVED').length)
const rejectedCount = computed(() => myTransactions.value.filter(t => t.status === 'REJECTED').length)
const totalSpent = computed(() => myTransactions.value.filter(t => t.status === 'APPROVED').reduce((s, t) => s + t.amount, 0))
</script>

<template>
  <div>
    <h2>Dashboard</h2>
    <p class="subtitle">Benvenuto {{ username }} — ruolo {{ role }}</p>

    <div v-if="error" class="alert error">{{ error }}</div>

    <!-- COMPANY / ADMIN stats -->
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

    <!-- EMPLOYEE stats -->
    <div v-if="isEmployee" class="form-grid">
      <div class="card"><h3>Transazioni totali</h3><div style="font-size:32px;font-weight:700">{{ myTransactions.length }}</div></div>
      <div class="card"><h3>Approvate</h3><div style="font-size:32px;font-weight:700;color:var(--green)">{{ approvedCount }}</div></div>
      <div class="card"><h3>Rifiutate</h3><div style="font-size:32px;font-weight:700;color:var(--red)">{{ rejectedCount }}</div></div>
      <div class="card"><h3>Speso totale</h3><div style="font-size:32px;font-weight:700">€ {{ totalSpent.toFixed(2) }}</div></div>
    </div>

    <div v-if="isEmployee" class="card">
      <p>Vai su <router-link to="/transactions">Transazioni</router-link> per inserire una spesa o su <router-link to="/cards">Carte</router-link> per vedere le tue carte.</p>
    </div>
  </div>
</template>
