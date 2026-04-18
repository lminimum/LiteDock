<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo">
          <img src="/src/assets/logo.svg" alt="LiteDock" v-if="hasLogo" />
          <div class="logo-text" v-else>LiteDock</div>
        </div>
        <h1>欢迎回来</h1>
        <p>登录到 LiteDock 管理平台</p>
      </div>

      <div class="success-message" v-if="showRegisteredSuccess">
        <span class="success-icon">✓</span>
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
             />
             <button
               type="button"
               @click="showPassword = !showPassword"
               class="password-toggle"
             >
               {{ showPassword ? '隐藏' : '显示' }}
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

        <button type="submit" class="login-btn" :disabled="loading">
          <span v-if="!loading">登录</span>
          <span v-else class="loading">
            <span class="spinner"></span>
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

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const credentials = reactive({
  username: '',
  password: ''
})

const errors = reactive({
  username: '',
  password: ''
})

const loading = ref(false)
const showPassword = ref(false)
const rememberMe = ref(false)
const hasLogo = ref(true)
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
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-lg);
}

.login-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  max-width: 400px;
  width: 100%;
  overflow: hidden;
}

.login-header {
  padding: var(--spacing-xl) var(--spacing-xl) var(--spacing-lg);
  text-align: center;
}

.logo {
  margin-bottom: var(--spacing-lg);
}

.logo img {
  width: 60px;
  height: 60px;
}

.logo-text {
  width: 60px;
  height: 60px;
  background: var(--primary-gradient);
  color: white;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-xl);
  font-weight: var(--font-bold);
  margin: 0 auto;
}

.login-header h1 {
  margin: 0 0 var(--spacing-xs);
  color: var(--text-primary);
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
}

.login-header p {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-sm);
}

.success-message {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin: 0 var(--spacing-xl) var(--spacing-lg);
  font-size: var(--text-sm);
  color: var(--text-secondary);
}

.success-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: var(--accent-color);
  color: white;
  border-radius: 50%;
  font-size: var(--text-xs);
}

.login-form {
  padding: 0 var(--spacing-xl) var(--spacing-xl);
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

.form-group.error input {
  border-color: var(--error-color);
}

.error-text {
  display: block;
  font-size: var(--text-xs);
  color: var(--error-color);
  margin-top: var(--spacing-xs);
}

.password-input {
  position: relative;
}

.password-toggle {
  position: absolute;
  right: var(--spacing-md);
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: var(--text-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
}

.password-toggle:hover {
  color: var(--text-primary);
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-xl);
}

.checkbox {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  cursor: pointer;
  font-size: var(--text-sm);
  color: var(--text-secondary);
}

.checkbox input {
  width: auto;
  margin: 0;
}

.login-btn {
  width: 100%;
  background: var(--primary-gradient);
  color: white;
  border: none;
  padding: 14px var(--spacing-lg);
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-weight: var(--font-medium);
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.login-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg);
}

.login-btn:disabled {
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

.login-footer {
  padding: var(--spacing-lg) var(--spacing-xl) var(--spacing-xl);
  text-align: center;
  border-top: 1px solid var(--border-light);
}

.login-footer p {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-sm);
}

.login-footer a {
  color: var(--accent-color);
  text-decoration: none;
  font-weight: var(--font-medium);
}

.login-footer a:hover {
  text-decoration: underline;
}
</style>