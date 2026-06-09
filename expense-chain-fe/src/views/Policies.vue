<script setup>
import { ref, onMounted } from 'vue'
import { api, CATEGORIES, CARD_TYPES } from '../api/client.js'

const policies = ref([])
const companies = ref([])
const form = ref({
  company_id: '',
  name: '',
  max_amount_per_tx: 0,
  max_amount_per_day: 0,
  max_amount_per_month: 0,
  blocked_categories: [],
  allowed_card_types: [],
})
const error = ref('')
const success = ref('')

async function load() {
  error.value = ''
  try {
    policies.value = await api.listPolicies()
    companies.value = await api.listCompanies().catch(() => [])
  } catch (e) {
    error.value = e.message
  }
}

async function create() {
  error.value = ''; success.value = ''
  try {
    await api.createPolicy({
      ...form.value,
      max_amount_per_tx: Number(form.value.max_amount_per_tx),
      max_amount_per_day: Number(form.value.max_amount_per_day),
      max_amount_per_month: Number(form.value.max_amount_per_month),
    })
    success.value = 'Policy creata'
    form.value = { company_id: '', name: '', max_amount_per_tx: 0, max_amount_per_day: 0, max_amount_per_month: 0, blocked_categories: [], allowed_card_types: [] }
    await load()
  } catch (e) {
    error.value = e.message
  }
}

async function remove(id) {
  if (!confirm('Eliminare policy?')) return
  try { await api.deletePolicy(id); await load() } catch (e) { error.value = e.message }
}

onMounted(load)
</script>

<template>
  <div>
    <h2>Policy</h2>
    <p class="subtitle">Regole di spesa che lo smart contract applica alle transazioni</p>

    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>

    <div class="card">
      <h3>Nuova policy</h3>
      <div class="form-grid">
        <div>
          <label>Azienda</label>
          <select v-model="form.company_id">
            <option value="">— seleziona —</option>
            <option v-for="c in companies" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div><label>Nome</label><input v-model="form.name" /></div>
        <div><label>Max per transazione (€)</label><input v-model="form.max_amount_per_tx" type="number" /></div>
        <div><label>Max giornaliero (€)</label><input v-model="form.max_amount_per_day" type="number" /></div>
        <div><label>Max mensile (€)</label><input v-model="form.max_amount_per_month" type="number" /></div>
        <div>
          <label>Categorie bloccate</label>
          <select v-model="form.blocked_categories" multiple style="height:auto;min-height:90px">
            <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
        <div>
          <label>Tipi carta ammessi (vuoto = tutti)</label>
          <select v-model="form.allowed_card_types" multiple style="height:auto;min-height:90px">
            <option v-for="t in CARD_TYPES" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
      </div>
      <div style="margin-top:14px"><button @click="create">Crea policy</button></div>
      <p class="subtitle" style="margin-top:8px">Limiti a 0 = controllo disabilitato.</p>
    </div>

    <div class="card">
      <h3>Elenco ({{ policies.length }})</h3>
      <table>
        <thead><tr><th>Nome</th><th>Max tx</th><th>Max gg</th><th>Max mese</th><th>Bloccate</th><th></th></tr></thead>
        <tbody>
          <tr v-for="p in policies" :key="p.id">
            <td>{{ p.name }}</td>
            <td>{{ p.max_amount_per_tx }}</td>
            <td>{{ p.max_amount_per_day }}</td>
            <td>{{ p.max_amount_per_month }}</td>
            <td>{{ (p.blocked_categories || []).join(', ') || '—' }}</td>
            <td><button class="danger" @click="remove(p.id)">Elimina</button></td>
          </tr>
          <tr v-if="!policies.length"><td colspan="6" class="subtitle">Nessuna policy</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
