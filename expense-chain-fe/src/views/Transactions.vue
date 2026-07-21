<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { api, CATEGORIES } from '../api/client.js'
import { useAuth } from '../store/auth.js'

const { can, role, state } = useAuth()

const isEmployee = computed(() => role.value === 'EMPLOYEE')
const employeeId = computed(() => state.claims?.employee_id || '')

const transactions = ref([])
const employees = ref([])
const cards = ref([])
const form = ref({ employee_id: '', card_id: '', amount: 0, currency: 'EUR', category: 'FOOD', description: '' })
const mode = ref('manual')
const error = ref('')
const result = ref(null)

async function loadEmployees() {
  if (isEmployee.value) return
  try { employees.value = await api.listEmployees() } catch { /* ignore */ }
}

async function loadTransactions() {
  try { transactions.value = await api.listTransactions() } catch { /* ignore */ }
}

// When employee_id changes, load cards for that employee
const activeEmployeeId = computed(() => isEmployee.value ? employeeId.value : form.value.employee_id)

watch(activeEmployeeId, async (empId) => {
  cards.value = []
  form.value.card_id = ''
  if (empId) {
    try { cards.value = await api.getCardsByEmployee(empId) } catch (e) { error.value = e.message }
  }
})

async function submit() {
  error.value = ''; result.value = null
  try {
    const payload = {
      ...form.value,
      amount: Number(form.value.amount),
      employee_id: isEmployee.value ? employeeId.value : form.value.employee_id,
    }
    const fn = mode.value === 'payment' ? api.incomingPayment : api.submitTransaction
    result.value = await fn(payload)
    await loadTransactions()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(async () => {
  await loadEmployees()
  // Employee: auto-load own cards
  if (isEmployee.value && employeeId.value) {
    try { cards.value = await api.getCardsByEmployee(employeeId.value) } catch { /* ignore */ }
  }
  await loadTransactions()
})
</script>

<template>
  <div>
    <h2>Transazioni</h2>
    <p class="subtitle">Invio spese — validate dallo smart contract e registrate nel ledger</p>

    <div v-if="error" class="alert error">{{ error }}</div>

    <div class="card">
      <h3>Nuova spesa</h3>
      <div class="form-grid">
        <div>
          <label>Modalità</label>
          <select v-model="mode">
            <option value="manual">Manuale (richiesta rimborso)</option>
            <option value="payment">Pagamento carta simulato</option>
          </select>
        </div>
        <!-- COMPANY/ADMIN: choose employee. EMPLOYEE: auto-set -->
        <div v-if="!isEmployee">
          <label>Dipendente</label>
          <select v-model="form.employee_id">
            <option value="">— seleziona —</option>
            <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.first_name }} {{ e.last_name }}</option>
          </select>
        </div>
        <div>
          <label>Carta</label>
          <select v-model="form.card_id">
            <option value="">— seleziona —</option>
            <option v-for="c in cards" :key="c.id" :value="c.id">{{ c.type }} **** {{ c.last4 }}</option>
          </select>
        </div>
        <div>
          <label>Categoria</label>
          <select v-model="form.category">
            <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
        <div><label>Importo</label><input v-model="form.amount" type="number" /></div>
        <div><label>Valuta</label><input v-model="form.currency" /></div>
        <div class="full"><label>Descrizione</label><input v-model="form.description" /></div>
      </div>
      <div style="margin-top:14px"><button @click="submit">Invia spesa</button></div>
    </div>

    <div v-if="result" class="card">
      <h3>Esito validazione</h3>
      <p>
        <span class="badge" :class="result.status">{{ result.status }}</span>
        <span v-if="result.reject_reason" style="margin-left:10px;color:var(--red)">{{ result.reject_reason }}</span>
      </p>
    </div>

    <div class="card">
      <h3>{{ isEmployee ? 'Le mie transazioni' : 'Storico transazioni' }} ({{ transactions.length }})</h3>
      <table>
        <thead><tr><th>ID</th><th>Importo</th><th>Categoria</th><th>Stato</th><th>Motivo</th></tr></thead>
        <tbody>
          <tr v-for="t in transactions" :key="t.id">
            <td class="mono">{{ t.id.slice(0, 8) }}</td>
            <td>{{ t.amount }} {{ t.currency }}</td>
            <td>{{ t.category }}</td>
            <td><span class="badge" :class="t.status">{{ t.status }}</span></td>
            <td>{{ t.reject_reason || '—' }}</td>
          </tr>
          <tr v-if="!transactions.length"><td colspan="5" class="subtitle">Nessuna transazione</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
