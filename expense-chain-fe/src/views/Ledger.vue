<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api/client.js'

const entries = ref([])
const verify = ref(null)
const error = ref('')

async function load() {
  error.value = ''
  try {
    entries.value = await api.getLedger()
  } catch (e) {
    error.value = e.message
  }
}

async function runVerify() {
  error.value = ''
  try {
    verify.value = await api.verifyLedger()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <h2>Ledger</h2>
    <p class="subtitle">Catena hash append-only — audit + blockchain unificati</p>

    <div v-if="error" class="alert error">{{ error }}</div>

    <div class="card">
      <div class="toolbar">
        <h3 style="margin:0">Verifica integrità</h3>
        <button @click="runVerify">Verifica catena</button>
      </div>
      <p v-if="verify">
        <span class="badge" :class="verify.valid ? 'valid' : 'invalid'">
          {{ verify.valid ? 'CATENA INTEGRA' : 'MANOMISSIONE RILEVATA' }}
        </span>
        — {{ verify.entries }} entry
        <span v-if="!verify.valid"> · rotto a seq {{ verify.broken_at }}: {{ verify.broken_why }}</span>
      </p>
      <p v-else class="subtitle">Premi "Verifica catena" per ricalcolare gli hash.</p>
    </div>

    <div class="card">
      <h3>Entry ({{ entries.length }})</h3>
      <table>
        <thead><tr><th>#</th><th>Entità</th><th>Azione</th><th>Entity ID</th><th>Hash</th><th>Prev</th></tr></thead>
        <tbody>
          <tr v-for="e in entries" :key="e.id">
            <td>{{ e.sequence }}</td>
            <td>{{ e.entity_type }}</td>
            <td><span class="badge" :class="e.action">{{ e.action }}</span></td>
            <td class="mono">{{ e.entity_id.slice(0, 8) }}</td>
            <td class="mono">{{ e.hash.slice(0, 12) }}…</td>
            <td class="mono">{{ e.prev_hash === 'GENESIS' ? 'GENESIS' : e.prev_hash.slice(0, 12) + '…' }}</td>
          </tr>
          <tr v-if="!entries.length"><td colspan="6" class="subtitle">Ledger vuoto</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
