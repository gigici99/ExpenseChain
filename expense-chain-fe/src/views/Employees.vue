<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api/client.js'

const employees = ref([])
const companies = ref([])
const policies = ref([])
const form = ref({ company_id: '', first_name: '', last_name: '', email: '', policy_id: '', username: '', password: '' })
const error = ref('')
const success = ref('')

async function load() {
  error.value = ''
  try {
    employees.value = await api.listEmployees()
    companies.value = await api.listCompanies().catch(() => [])
    policies.value = await api.listPolicies().catch(() => [])
  } catch (e) {
    error.value = e.message
  }
}

async function create() {
  error.value = ''; success.value = ''
  try {
    await api.createEmployee(form.value)
    success.value = 'Dipendente creato'
    form.value = { company_id: '', first_name: '', last_name: '', email: '', policy_id: '', username: '', password: '' }
    await load()
  } catch (e) {
    error.value = e.message
  }
}

async function remove(id) {
  if (!confirm('Eliminare dipendente?')) return
  try { await api.deleteEmployee(id); await load() } catch (e) { error.value = e.message }
}

onMounted(load)
</script>

<template>
  <div>
    <h2>Dipendenti</h2>
    <p class="subtitle">Gestione dipendenti aziendali</p>

    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>

    <div class="card">
      <h3>Nuovo dipendente</h3>
      <div class="form-grid">
        <div>
          <label>Azienda</label>
          <select v-model="form.company_id">
            <option value="">— seleziona —</option>
            <option v-for="c in companies" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div>
          <label>Policy</label>
          <select v-model="form.policy_id">
            <option value="">— seleziona —</option>
            <option v-for="p in policies" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>
        <div><label>Nome</label><input v-model="form.first_name" /></div>
        <div><label>Cognome</label><input v-model="form.last_name" /></div>
        <div class="full"><label>Email</label><input v-model="form.email" /></div>
        <div><label>Username utente dipendente</label><input v-model="form.username" autocomplete="off" /></div>
        <div><label>Password utente dipendente</label><input v-model="form.password" type="password" autocomplete="new-password" /></div>
      </div>
      <p class="subtitle" style="margin-top:8px">Username + password creano l'utente EMPLOYEE per il login del dipendente.</p>
      <div style="margin-top:14px"><button @click="create">Crea dipendente</button></div>
    </div>

    <div class="card">
      <h3>Elenco ({{ employees.length }})</h3>
      <table>
        <thead><tr><th>ID</th><th>Nome</th><th>Email</th><th>Policy</th><th></th></tr></thead>
        <tbody>
          <tr v-for="e in employees" :key="e.id">
            <td class="mono">{{ e.id.slice(0, 8) }}</td>
            <td>{{ e.first_name }} {{ e.last_name }}</td>
            <td>{{ e.email }}</td>
            <td class="mono">{{ e.policy_id ? e.policy_id.slice(0, 8) : '—' }}</td>
            <td><button class="danger" @click="remove(e.id)">Elimina</button></td>
          </tr>
          <tr v-if="!employees.length"><td colspan="5" class="subtitle">Nessun dipendente</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
