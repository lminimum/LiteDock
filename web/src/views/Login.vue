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
          <span class="error-message" v-if="errors.username">{{ errors.username }}</span>
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
          <span class="error-message" v-if="errors.password">{{ errors.password }}</span>
        </div>
        
        <div class="form-options">
          <label class="checkbox">
            <input v-model="rememberMe" type="checkbox" />
            <span>记住我</span>
          </label>
          
          <a href="#" class="forgot-password" @click.prevent="handleForgotPassword">
            忘记密码？
          </a>
        </div>
        
        <button type="submit" class="login-btn" :disabled="loading">
          <span v-if="!loading">登录</span>
          <span v-else class="loading-spinner">
            <div class="spinner"></div>
            登录中...
          </span>
        </button>
      </form>
      
      <div class="login-footer">
        <p>
          还没有账户？
          <a href="#" @click.prevent="handleSetup">重新配置</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
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
const hasLogo = ref(false) // 检查是否有logo文件

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

const handleForgotPassword = () => {
  // 这里可以实现忘记密码功能
  alert('请联系系统管理员重置密码')
}

const handleSetup = () => {
  // 清除配置状态，重新进入配置流程
  localStorage.removeItem('litedock-configured')
  router.push('/setup')
}

// 检查是否有记住的用户名
if (localStorage.getItem('litedock-remember') === 'true') {
  const savedUsername = localStorage.getItem('litedock-username')
  if (savedUsername) {
    credentials.username = savedUsername
    rememberMe.value = true
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  max-width: 400px;
  width: 100%;
  overflow: hidden;
}

.login-header {
  padding: 40px 40px 20px 40px;
  text-align: center;
}

.logo {
  margin-bottom: 20px;
}

.logo img {
  width: 60px;
  height: 60px;
}

.logo-text {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0 auto;
}

.login-header h1 {
  margin: 0 0 8px 0;
  color: #1e293b;
  font-size: 1.5rem;
  font-weight: 600;
}

.login-header p {
  margin: 0;
  color: #64748b;
}

.login-form {
  padding: 20px 40px 30px 40px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
  font-size: 0.875rem;
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-group.error input {
  border-color: #ef4444;
}

.error-message {
  color: #ef4444;
  font-size: 0.75rem;
  margin-top: 4px;
  display: block;
}

.password-input {
  position: relative;
}

.password-toggle {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: #64748b;
  cursor: pointer;
  font-size: 0.75rem;
  padding: 4px 8px;
}

.password-toggle:hover {
  color: #374151;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 0.875rem;
  color: #64748b;
}

.checkbox input {
  width: auto;
  margin: 0;
}

.forgot-password {
  color: #667eea;
  text-decoration: none;
  font-size: 0.875rem;
}

.forgot-password:hover {
  text-decoration: underline;
}

.login-btn {
  width: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 14px;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.login-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
}

.login-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.loading-spinner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.login-footer {
  padding: 20px 40px 40px 40px;
  text-align: center;
  border-top: 1px solid #f1f5f9;
}

.login-footer p {
  margin: 0;
  color: #64748b;
  font-size: 0.875rem;
}

.login-footer a {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
}

.login-footer a:hover {
  text-decoration: underline;
}
</style>