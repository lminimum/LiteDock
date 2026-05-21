<template>
  <AuthCard>
    <template #header>
      <div class="logo">
        <span class="logo-text">LiteDock</span>
      </div>
      <h1>{{ t('auth.welcomeBack') }}</h1>
      <p>{{ t('auth.loginSubtitle') }}</p>
    </template>

    <div class="success-message" v-if="showRegisteredSuccess">
      <CheckCircle :size="16" />
      {{ t('auth.adminCreatedSuccess') }}
    </div>

    <form @submit.prevent="handleLogin" class="auth-form">
      <div class="form-group" :class="{ error: errors.username }">
        <label for="username">{{ t('auth.username') }}</label>
        <input
          id="username"
          v-model="credentials.username"
          type="text"
          autocomplete="username"
          :placeholder="t('auth.usernamePlaceholder')"
          :disabled="loading"
          required
          class="input"
        />
        <span class="error-text" v-if="errors.username">{{ errors.username }}</span>
      </div>

      <div class="form-group" :class="{ error: errors.password }">
        <label for="password">{{ t('auth.password') }}</label>
        <div class="password-input">
          <input
            id="password"
            v-model="credentials.password"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="current-password"
            :placeholder="t('auth.passwordPlaceholder')"
            :disabled="loading"
            required
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

      <div class="form-options">
        <label class="checkbox">
          <input v-model="rememberMe" type="checkbox" />
          <span>{{ t('auth.rememberMe') }}</span>
        </label>
      </div>

      <button type="submit" class="btn btn-primary btn-lg" :disabled="loading" style="width: 100%">
        <span v-if="!loading">{{ t('auth.login') }}</span>
        <span v-else class="loading">
          <Loader2 :size="16" class="spinning" />
          {{ t('auth.loggingIn') }}
        </span>
      </button>
    </form>

    <template #footer>
      <p v-if="!setupComplete">
        {{ t('auth.noAccount') }}
        <router-link to="/setup">{{ t('auth.createAdmin') }}</router-link>
      </p>
    </template>
  </AuthCard>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { CheckCircle, Eye, EyeOff, Loader2 } from 'lucide-vue-next'
import { t } from '@/i18n'
import AuthCard from '@/components/auth/AuthCard.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const credentials = reactive({ username: '', password: '' })
const errors = reactive({ username: '', password: '' })
const loading = ref(false)
const showPassword = ref(false)
const rememberMe = ref(false)
const setupComplete = ref(false)
const showRegisteredSuccess = ref(false)

onMounted(async () => {
  setupComplete.value = await authStore.checkSetupStatus()

  if (route.query.registered === 'true') {
    showRegisteredSuccess.value = true
  }

  if (authStore.token) {
    const isAuth = await authStore.checkAuth()
    if (isAuth) {
      router.push('/')
    }
  }
})

const validateForm = () => {
  errors.username = ''
  errors.password = ''

  if (!credentials.username.trim()) {
    errors.username = t('auth.usernameRequired')
    return false
  }

  if (!credentials.password.trim()) {
    errors.password = t('auth.passwordRequired')
    return false
  }

  if (credentials.password.length < 6) {
    errors.password = t('auth.passwordMinLength6')
    return false
  }

  return true
}

const handleLogin = async () => {
  if (!validateForm()) return

  loading.value = true

  try {
    const result = await authStore.login(credentials)

    if (result.success) {
      if (rememberMe.value) {
        localStorage.setItem('litedock-remember', 'true')
      }
      router.push('/')
    } else {
      errors.password = result.message || t('auth.loginFailed')
    }
  } catch (error: any) {
    errors.password = error.message || t('auth.loginFailedRetry')
  } finally {
    loading.value = false
  }
}

const goToSetup = () => {
  router.push('/setup')
}
</script>

<style scoped>
.logo {
  margin-bottom: var(--space-4);
}

.logo-text {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  letter-spacing: -0.02em;
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

.success-message {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  background: var(--color-success-bg);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  padding: var(--space-3);
  margin-bottom: var(--space-6);
  font-size: var(--font-size-sm);
  color: var(--color-success);
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

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.checkbox {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
  font-size: var(--font-size-sm);
  color: var(--color-text);
}

.checkbox input {
  width: auto;
  margin: 0;
  accent-color: var(--color-background-strong);
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

p {
  margin: 0;
  color: var(--color-text);
  font-size: var(--font-size-sm);
}

a {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
}

a:hover {
  color: var(--color-text-weak);
}
</style>
