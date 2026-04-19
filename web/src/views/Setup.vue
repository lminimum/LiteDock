<template>
  <div class="setup-container">
    <div class="setup-card">
      <div class="setup-header">
        <div class="logo-text">LD</div>
        <h1>{{ t('auth.initSetup') }}</h1>
        <p>{{ t('auth.setupDescription') }}</p>
      </div>

      <form @submit.prevent="handleSubmit" class="setup-form">
        <div class="form-group" :class="{ error: errors.username }">
          <label for="username">{{ t('auth.adminUsername') }}</label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            :placeholder="t('auth.usernamePlaceholder')"
            :disabled="loading"
          />
          <span class="error-text" v-if="errors.username">{{ errors.username }}</span>
        </div>

        <div class="form-group" :class="{ error: errors.password }">
          <label for="password">{{ t('auth.adminPassword') }}</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            :placeholder="t('auth.adminPasswordPlaceholder')"
            :disabled="loading"
          />
          <span class="error-text" v-if="errors.password">{{ errors.password }}</span>
        </div>

        <div class="form-group" :class="{ error: errors.confirmPassword }">
          <label for="confirmPassword">{{ t('auth.confirmPassword') }}</label>
          <input
            id="confirmPassword"
            v-model="form.confirmPassword"
            type="password"
            :placeholder="t('auth.confirmPasswordPlaceholder')"
            :disabled="loading"
          />
          <span class="error-text" v-if="errors.confirmPassword">{{ errors.confirmPassword }}</span>
        </div>

        <div class="form-group">
          <label for="email">{{ t('auth.emailOptional') }}</label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            placeholder="admin@example.com"
            :disabled="loading"
          />
        </div>

        <div class="error-message" v-if="submitError">
          {{ submitError }}
        </div>

        <button type="submit" class="submit-btn" :disabled="loading">
          <span v-if="!loading">{{ t('auth.createAccount') }}</span>
          <span v-else class="loading">
            <span class="spinner"></span>
            {{ t('auth.creating') }}
          </span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/utils/api'
import { t } from '@/i18n'

const router = useRouter()

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: ''
})

const errors = reactive({
  username: '',
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const submitError = ref('')

const validate = (): boolean => {
  errors.username = ''
  errors.password = ''
  errors.confirmPassword = ''
  let valid = true

  if (!form.username.trim()) {
    errors.username = t('auth.usernameRequired')
    valid = false
  } else if (form.username.length < 3) {
    errors.username = t('auth.usernameMinLength')
    valid = false
  }

  if (!form.password) {
    errors.password = t('auth.passwordRequired')
    valid = false
  } else if (form.password.length < 8) {
    errors.password = t('auth.passwordMinLength')
    valid = false
  }

  if (form.password !== form.confirmPassword) {
    errors.confirmPassword = t('auth.passwordMismatch')
    valid = false
  }

  return valid
}

const handleSubmit = async () => {
  if (!validate()) return

  loading.value = true
  submitError.value = ''

  try {
    await api.post('/auth/register', {
      username: form.username,
      password: form.password,
      email: form.email || '',
      role: 'admin'
    })

    router.push('/login?registered=true')
  } catch (error: any) {
    submitError.value = error.response?.data?.error || t('auth.createFailed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.setup-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary);
  padding: var(--spacing-lg);
}

.setup-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 400px;
  overflow: hidden;
}

.setup-header {
  padding: var(--spacing-xl);
  text-align: center;
  border-bottom: 1px solid var(--border-light);
}

.logo-text {
  width: 56px;
  height: 56px;
  background: var(--primary-gradient);
  color: white;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-xl);
  font-weight: var(--font-bold);
  margin: 0 auto var(--spacing-lg);
}

.setup-header h1 {
  font-size: var(--text-xl);
  margin-bottom: var(--spacing-xs);
}

.setup-header p {
  color: var(--text-secondary);
  font-size: var(--text-sm);
  margin: 0;
}

.setup-form {
  padding: var(--spacing-xl);
}

.form-group {
  margin-bottom: var(--spacing-lg);
}

.form-group label {
  display: block;
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--text-primary);
  margin-bottom: var(--spacing-xs);
}

.form-group input {
  width: 100%;
  padding: 12px var(--spacing-md);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  color: var(--text-primary);
  background: var(--bg-primary);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.form-group input::placeholder {
  color: var(--text-muted);
}

.form-group input:focus {
  outline: none;
  border-color: var(--accent-color);
  box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.1);
}

.form-group input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-group.error input {
  border-color: var(--error-color);
}

.error-text {
  display: block;
  font-size: var(--text-xs);
  color: var(--error-color);
  margin-top: var(--spacing-xs);
}

.error-message {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  font-size: var(--text-sm);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-lg);
  text-align: center;
}

.submit-btn {
  width: 100%;
  padding: 14px var(--spacing-lg);
  background: var(--primary-gradient);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-weight: var(--font-medium);
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg);
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>