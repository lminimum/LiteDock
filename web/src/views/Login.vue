<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo">
          <span class="logo-text">LiteDock</span>
        </div>
        <h1>欢迎回来</h1>
        <p>登录到 LiteDock 管理平台</p>
      </div>

      <div class="success-message" v-if="showRegisteredSuccess">
        <CheckCircle :size="16" />
        管理员账户创建成功，请登录
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group" :class="{ error: errors.username }">
          <label for="username">用户名</label>
          <input
            id="username"
            v-model="credentials.username"
            type="text"
            placeholder="请输入用户名"
            :disabled="loading"
            required
            class="input"
          />
          <span class="error-text" v-if="errors.username">{{ errors.username }}</span>
        </div>

        <div class="form-group" :class="{ error: errors.password }">
          <label for="password">密码</label>
          <div class="password-input">
            <input
              id="password"
              v-model="credentials.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="请输入密码"
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
            <span>记住我</span>
          </label>
        </div>

        <button type="submit" class="btn btn-primary btn-lg" :disabled="loading" style="width: 100%">
          <span v-if="!loading">登录</span>
          <span v-else class="loading">
            <Loader2 :size="16" class="spinning" />
            登录中...
          </span>
        </button>
      </form>

      <div class="login-footer">
        <p v-if="!setupComplete">
          还没有账户？
          <a href="#" @click.prevent="goToSetup">创建管理员</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { CheckCircle, Eye, EyeOff, Loader2 } from 'lucide-vue-next'

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
    errors.username = '请输入用户名'
    return false
  }

  if (!credentials.password.trim()) {
    errors.password = '请输入密码'
    return false
  }

  if (credentials.password.length < 6) {
    errors.password = '密码长度至少6位'
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
      errors.password = result.message || '登录失败'
    }
  } catch (error: any) {
    errors.password = error.response?.data?.message || '登录失败，请重试'
  } finally {
    loading.value = false
  }
}

const goToSetup = () => {
  router.push('/setup')
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  background: var(--color-background);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
}

.login-card {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  max-width: 380px;
  width: 100%;
  padding: var(--space-8);
}

.login-header {
  text-align: center;
  margin-bottom: var(--space-6);
}

.logo {
  margin-bottom: var(--space-4);
}

.logo-text {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  letter-spacing: -0.02em;
}

.login-header h1 {
  margin: 0 0 var(--space-1) 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.login-header p {
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
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-3);
  margin-bottom: var(--space-6);
  font-size: var(--font-size-sm);
  color: var(--color-success);
}

.login-form {
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

.input {
  width: 100%;
  padding: var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--color-text-strong);
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.input::placeholder {
  color: var(--color-text-weaker);
}

.input:focus {
  outline: none;
  border-color: var(--color-background-strong);
  box-shadow: 0 0 0 3px var(--color-background-interactive-weaker);
}

.form-group.error .input {
  border-color: var(--color-error);
}

.error-text {
  font-size: var(--font-size-xs);
  color: var(--color-error);
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

.btn-primary {
  background: var(--color-background-strong);
  color: var(--color-background);
  border-color: var(--color-background-strong);
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-text-weak);
  border-color: var(--color-text-weak);
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

.login-footer {
  margin-top: var(--space-6);
  text-align: center;
  padding-top: var(--space-6);
  border-top: 1px solid var(--color-border-weak);
}

.login-footer p {
  margin: 0;
  color: var(--color-text);
  font-size: var(--font-size-sm);
}

.login-footer a {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
}

.login-footer a:hover {
  color: var(--color-text-weak);
}
</style>
