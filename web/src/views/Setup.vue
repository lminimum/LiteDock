<template>
  <AuthCard>
    <template #header>
      <div class="logo-text">LD</div>
      <h1>{{ t('auth.initSetup') }}</h1>
      <p>{{ t('auth.setupDescription') }}</p>
    </template>

    <form @submit.prevent="handleSubmit" class="auth-form">
      <div class="form-group" :class="{ error: errors.username }">
        <label for="username">{{ t('auth.adminUsername') }}</label>
        <input
          id="username"
          v-model="form.username"
          type="text"
          autocomplete="username"
          :placeholder="t('auth.usernamePlaceholder')"
          :disabled="loading"
          class="input"
        />
        <span class="error-text" v-if="errors.username">{{ errors.username }}</span>
      </div>

      <div class="form-group" :class="{ error: errors.password }">
        <label for="password">{{ t('auth.adminPassword') }}</label>
        <div class="password-input">
          <input
            id="password"
            v-model="form.password"
            :type="showPassword ? 'text' : 'password'"
            :placeholder="t('auth.adminPasswordPlaceholder')"
            :disabled="loading"
            autocomplete="new-password"
            class="input"
          />
          <button
            type="button"
            @click="showPassword = !showPassword"
            class="password-toggle"
          >
            <Eye v-if="!showPassword" :size="16" />
            <EyeOff v-else :size="16" />
          </button>
        </div>
        <span class="error-text" v-if="errors.password">{{ errors.password }}</span>
      </div>

      <div class="form-group" :class="{ error: errors.confirmPassword }">
        <label for="confirmPassword">{{ t('auth.confirmPassword') }}</label>
        <input
          id="confirmPassword"
          v-model="form.confirmPassword"
          :type="showPassword ? 'text' : 'password'"
          :placeholder="t('auth.confirmPasswordPlaceholder')"
          :disabled="loading"
          autocomplete="new-password"
          class="input"
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
          class="input"
        />
      </div>

      <div class="error-message" v-if="submitError">
        {{ submitError }}
      </div>

      <button type="submit" class="btn btn-primary btn-lg" :disabled="loading" style="width: 100%">
        <span v-if="!loading">{{ t('auth.createAccount') }}</span>
        <span v-else class="loading">
          <Loader2 :size="16" class="spinning" />
          {{ t('auth.creating') }}
        </span>
      </button>
    </form>
  </AuthCard>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { Eye, EyeOff, Loader2 } from 'lucide-vue-next'
import api from '@/utils/api'
import { t } from '@/i18n'
import AuthCard from '@/components/auth/AuthCard.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

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
const showPassword = ref(false)

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

    await authStore.refreshSetupStatus()
    router.push('/login?registered=true')
  } catch (error: any) {
    submitError.value = error.message || t('auth.createFailed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.logo-text {
  width: 56px;
  height: 56px;
  background: var(--color-background-strong);
  color: var(--color-background);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  margin: 0 auto var(--space-4);
  font-family: var(--font-mono);
}

h1 {
  margin: 0 0 var(--space-1) 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

p {
  margin: 0;
  color: var(--color-text);
  font-size: var(--font-size-sm);
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-group label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-strong);
}

.password-input {
  position: relative;
}

.password-input .input {
  padding-right: var(--space-10);
}

.password-toggle {
  position: absolute;
  right: var(--space-3);
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--color-text-weaker);
  cursor: pointer;
  padding: var(--space-1);
  display: flex;
  align-items: center;
  justify-content: center;
}

.password-toggle:hover {
  color: var(--color-text);
}

.form-group.error .input {
  border-color: var(--color-error);
}

.error-text {
  font-size: var(--font-size-xs);
  color: var(--color-error);
}

.error-message {
  background: var(--color-error-bg);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  padding: var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--color-error);
  text-align: center;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
