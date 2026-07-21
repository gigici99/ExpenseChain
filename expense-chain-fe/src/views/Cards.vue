<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { api, CARD_TYPES } from '../api/client.js'
import { useAuth } from '../store/auth.js'

const { can, role, state } = useAuth()

const isEmployee = computed(() => role.value === 'EMPLOYEE')
const employeeId = computed(() => state.claims?.employee_id || '')

const employees = ref([])
const cards = ref([])
const selectedEmployee = ref('')
const form = ref({ employee_id: '', type: 'PHYSICAL', last4: '', holder: '', provider: '', expires_at: '' })
const error = ref('')
const success = ref('')

async function loadEmployees() {
  if (isEmployee.value) return
  try { employees.value = await api.listEmployees() } catch (e) { error.value = e.message }
}

async function loadCards() {
  const empId = isEmployee.value ? employeeId.value : selectedEmployee.value
  if (!empId) { cards.value = []; return }
  try { cards.value = await api.getCardsByEmployee(empId) } catch (e) { error.value = e.message }
}

watch(selectedEmployee, loadCards)

async function create() {
  error.value = ''; success.value = ''
  try {
    const payload = { ...form.value }
    if (payload.expires_at && payload.expires_at.length === 10) {
      payload.expires_at = payload.expires_at + 'T00:00:00Z'
    }
    await api.createCard(payload)
    success.value = 'Carta creata'
    await loadCards()
    form.value = { employee_id: '', type: 'PHYSICAL', last4: '', holder: '', provider: '', expires_at: '' }
  } catch (e) {
    error.value = e.message
  }
}

async function remove(id) {
  if (!confirm('Eliminare carta?')) return
  try { await api.deleteCard(id); await loadCards() } catch (e) { error.value = e.message }
}

onMounted(async () => {
  await loadEmployees()
  if (isEmployee.value) await loadCards()
})
</script>

<template>
  <div>
    <h2>Carte</h2>
    <p class="subtitle">{{ isEmployee ? 'Le tue carte di pagamento' : 'Carte di pagamento dei dipendenti (mai PAN completo, solo last4)' }}</p>

    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>

    <!-- Create form — only COMPANY/ADMIN -->
    <div v-if="can('COMPANY')" class="card">
      <h3>Nuova carta</h3>
      <div class="form-grid">
        <div>
          <label>Dipendente</label>
          <select v-model="form.employee_id">
            <option value="">— seleziona —</option>
            <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.first_name }} {{ e.last_name }}</option>
          </select>
        </div>
        <div>
          <label>Tipo</label>
          <select v-model="form.type">
            <option v-for="t in CARD_TYPES" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
        <div><label>Last 4</label><input v-model="form.last4" maxlength="4" /></div>
        <div><label>Intestatario</label><input v-model="form.holder" /></div>
        <div><label>Provider</label><input v-model="form.provider" placeholder="Visa, Google Pay..." /></div>
        <div><label>Scadenza</label><input v-model="form.expires_at" type="date" /></div>
      </div>
      <div style="margin-top:14px"><button @click="create">Crea carta</button></div>
    </div>

    <!-- Card list -->
    <div class="card">
      <h3>{{ isEmployee ? 'Le mie carte' : 'Carte per dipendente' }}</h3>
      <template v-if="!isEmployee">
        <label>Dipendente</label>
        <select v-model="selectedEmployee" style="max-width:300px">
          <option value="">— seleziona —</option>
          <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.first_name }} {{ e.last_name }}</option>
        </select>
      </template>
      <table style="margin-top:14px">
        <thead><tr><th>ID</th><th>Tipo</th><th>Last4</th><th>Provider</th><th v-if="can('COMPANY')"></th></tr></thead>
        <tbody>
          <tr v-for="c in cards" :key="c.id">
            <td class="mono">{{ c.id.slice(0, 8) }}</td>
            <td>{{ c.type }}</td>
            <td class="mono">**** {{ c.last4 }}</td>
            <td>{{ c.provider }}</td>
            <td v-if="can('COMPANY')"><button class="danger" @click="remove(c.id)">Elimina</button></td>
          </tr>
          <tr v-if="!cards.length"><td :colspan="can('COMPANY') ? 5 : 4" class="subtitle">Nessuna carta</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
