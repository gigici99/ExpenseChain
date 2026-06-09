<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../store/auth.js'

const router = useRouter()
const { login } = useAuth()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await login(username.value, password.value)
    router.push('/dashboard')
  } catch (e) {
    error.value = e.message || 'Login fallito'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrap">
    <form class="login-box" @submit.prevent="submit">
      <h1>⛓ Expense Chain</h1>
      <p>Accedi per continuare</p>
      <div v-if="error" class="alert error">{{ error }}</div>
      <label>Username</label>
      <input v-model="username" autocomplete="username" />
      <label>Password</label>
      <input v-model="password" type="password" autocomplete="current-password" />
      <button :disabled="loading">{{ loading ? 'Accesso...' : 'Accedi' }}</button>
    </form>
  </div>
</template>
