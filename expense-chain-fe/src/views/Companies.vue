<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api/client.js'

const companies = ref([])
const form = ref({ name: '', vat_id: '', address: '', username: '', password: '' })
const error = ref('')
const success = ref('')

async function load() {
  error.value = ''
  try {
    companies.value = await api.listCompanies()
  } catch (e) {
    error.value = e.message
  }
}

async function create() {
  error.value = ''; success.value = ''
  try {
    await api.createCompany(form.value)
    success.value = 'Azienda creata'
    form.value = { name: '', vat_id: '', address: '', username: '', password: '' }
    await load()
  } catch (e) {
    error.value = e.message
  }
}

async function remove(id) {
  if (!confirm('Eliminare azienda?')) return
  try {
    await api.deleteCompany(id)
    await load()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <h2>Aziende</h2>
    <p class="subtitle">Gestione anagrafica aziende (creazione/eliminazione: solo ADMIN)</p>

    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>

    <div class="card">
      <h3>Nuova azienda</h3>
      <div class="form-grid">
        <div><label>Nome</label><input v-model="form.name" /></div>
        <div><label>Partita IVA</label><input v-model="form.vat_id" /></div>
        <div class="full"><label>Indirizzo</label><input v-model="form.address" /></div>
        <div><label>Username utente azienda</label><input v-model="form.username" autocomplete="off" /></div>
        <div><label>Password utente azienda</label><input v-model="form.password" type="password" autocomplete="new-password" /></div>
      </div>
      <p class="subtitle" style="margin-top:8px">Username + password creano l'utente COMPANY per il login dell'azienda.</p>
      <div style="margin-top:14px"><button @click="create">Crea azienda</button></div>
    </div>

    <div class="card">
      <h3>Elenco ({{ companies.length }})</h3>
      <table>
        <thead><tr><th>ID</th><th>Nome</th><th>P.IVA</th><th>Indirizzo</th><th></th></tr></thead>
        <tbody>
          <tr v-for="c in companies" :key="c.id">
            <td class="mono">{{ c.id.slice(0, 8) }}</td>
            <td>{{ c.name }}</td>
            <td>{{ c.vat_id }}</td>
            <td>{{ c.address }}</td>
            <td><button class="danger" @click="remove(c.id)">Elimina</button></td>
          </tr>
          <tr v-if="!companies.length"><td colspan="5" class="subtitle">Nessuna azienda</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
