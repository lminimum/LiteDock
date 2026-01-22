<template>
  <div class="setup-container">
    <div class="setup-card">
      <div class="setup-header">
        <h1>LiteDock 初始配置</h1>
        <p>欢迎使用 LiteDock Docker 管理平台</p>
      </div>
      
      <div class="setup-content">
        <div class="step-indicator">
          <div 
            v-for="(step, index) in steps" 
            :key="index"
            :class="['step', { active: currentStep === index, completed: currentStep > index }]"
          >
            <div class="step-number">{{ index + 1 }}</div>
            <div class="step-title">{{ step.title }}</div>
          </div>
        </div>
        
        <div class="step-content">
          <!-- 步骤 1: Docker 连接配置 -->
          <div v-if="currentStep === 0" class="step-form">
            <h3>Docker 连接配置</h3>
            <div class="form-group">
              <label>连接类型</label>
              <select v-model="config.docker.type" @change="onDockerTypeChange">
                <option value="local">本地 Docker</option>
                <option value="remote">远程 Docker</option>
              </select>
            </div>
            
            <div v-if="config.docker.type === 'remote'" class="form-group">
              <label>远程主机地址</label>
              <input 
                v-model="config.docker.host" 
                placeholder="tcp://remote-host:2375"
                type="text"
              />
            </div>
            
            <div class="form-group">
              <label>TLS 证书路径 (可选)</label>
              <input 
                v-model="config.docker.tlsPath" 
                placeholder="/path/to/certs"
                type="text"
              />
            </div>
            
            <button @click="testDockerConnection" class="test-btn" :disabled="testing">
              {{ testing ? '测试中...' : '测试连接' }}
            </button>
          </div>
          
          <!-- 步骤 2: 管理员账户设置 -->
          <div v-if="currentStep === 1" class="step-form">
            <h3>管理员账户设置</h3>
            <div class="form-group">
              <label>用户名</label>
              <input v-model="config.admin.username" placeholder="admin" type="text" />
            </div>
            
            <div class="form-group">
              <label>密码</label>
              <input v-model="config.admin.password" placeholder="请输入密码" type="password" />
            </div>
            
            <div class="form-group">
              <label>确认密码</label>
              <input v-model="config.admin.confirmPassword" placeholder="请再次输入密码" type="password" />
            </div>
            
            <div class="form-group">
              <label>邮箱 (可选)</label>
              <input v-model="config.admin.email" placeholder="admin@example.com" type="email" />
            </div>
          </div>
          
          <!-- 步骤 3: 系统设置 -->
          <div v-if="currentStep === 2" class="step-form">
            <h3>系统设置</h3>
            <div class="form-group">
              <label>系统名称</label>
              <input v-model="config.system.name" placeholder="LiteDock" type="text" />
            </div>
            
            <div class="form-group">
              <label>监听端口</label>
              <input v-model="config.system.port" placeholder="8080" type="number" min="1024" max="65535" />
            </div>
            
            <div class="form-group checkbox">
              <input v-model="config.system.enableMetrics" type="checkbox" id="metrics" />
              <label for="metrics">启用监控指标</label>
            </div>
            
            <div class="form-group checkbox">
              <input v-model="config.system.enableSwagger" type="checkbox" id="swagger" />
              <label for="swagger">启用 API 文档</label>
            </div>
          </div>
          
          <!-- 步骤 4: 完成配置 -->
          <div v-if="currentStep === 3" class="step-complete">
            <div class="success-icon">✓</div>
            <h3>配置完成！</h3>
            <p>LiteDock 已成功配置，现在可以开始使用了。</p>
            
            <div class="config-summary">
              <h4>配置摘要</h4>
              <ul>
                <li>Docker 连接: {{ config.docker.type === 'local' ? '本地' : '远程' }}</li>
                <li>管理员用户: {{ config.admin.username }}</li>
                <li>系统端口: {{ config.system.port }}</li>
                <li>监控指标: {{ config.system.enableMetrics ? '启用' : '禁用' }}</li>
              </ul>
            </div>
          </div>
        </div>
        
        <div class="setup-actions">
          <button 
            v-if="currentStep > 0" 
            @click="previousStep" 
            class="btn-secondary"
          >
            上一步
          </button>
          
          <button 
            v-if="currentStep < steps.length - 1" 
            @click="nextStep" 
            class="btn-primary"
            :disabled="!canProceed"
          >
            下一步
          </button>
          
          <button 
            v-if="currentStep === steps.length - 1" 
            @click="completeSetup" 
            class="btn-primary"
            :disabled="completing"
          >
            {{ completing ? '配置中...' : '开始使用' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

interface Config {
  docker: {
    type: 'local' | 'remote'
    host: string
    tlsPath: string
  }
  admin: {
    username: string
    password: string
    confirmPassword: string
    email: string
  }
  system: {
    name: string
    port: string
    enableMetrics: boolean
    enableSwagger: boolean
  }
}

const steps = [
  { title: 'Docker 配置' },
  { title: '管理员账户' },
  { title: '系统设置' },
  { title: '完成' }
]

const currentStep = ref(0)
const testing = ref(false)
const completing = ref(false)

const config = ref<Config>({
  docker: {
    type: 'local',
    host: '',
    tlsPath: ''
  },
  admin: {
    username: 'admin',
    password: '',
    confirmPassword: '',
    email: ''
  },
  system: {
    name: 'LiteDock',
    port: '8080',
    enableMetrics: true,
    enableSwagger: true
  }
})

const canProceed = computed(() => {
  switch (currentStep.value) {
    case 0: // Docker 配置
      return config.value.docker.type === 'local' || config.value.docker.host.trim() !== ''
    case 1: // 管理员账户
      return (
        config.value.admin.username.trim() !== '' &&
        config.value.admin.password.length >= 6 &&
        config.value.admin.password === config.value.admin.confirmPassword
      )
    case 2: // 系统设置
      return (
        config.value.system.name.trim() !== '' &&
        config.value.system.port.trim() !== ''
      )
    default:
      return true
  }
})

const onDockerTypeChange = () => {
  if (config.value.docker.type === 'local') {
    config.value.docker.host = ''
  }
}

const testDockerConnection = async () => {
  testing.value = true
  try {
    // 这里应该调用后端API测试Docker连接
    await new Promise(resolve => setTimeout(resolve, 2000))
    alert('Docker 连接测试成功！')
  } catch (error) {
    alert('Docker 连接测试失败，请检查配置')
  } finally {
    testing.value = false
  }
}

const nextStep = () => {
  if (canProceed.value) {
    currentStep.value++
  }
}

const previousStep = () => {
  currentStep.value--
}

const completeSetup = async () => {
  completing.value = true
  try {
    // 这里应该调用后端API保存配置
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    // 标记为已配置
    localStorage.setItem('litedock-configured', 'true')
    localStorage.setItem('litedock-config', JSON.stringify(config.value))
    
    // 跳转到登录页面
    router.push('/login')
  } catch (error) {
    alert('配置保存失败，请重试')
  } finally {
    completing.value = false
  }
}
</script>

<style scoped>
.setup-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.setup-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  max-width: 800px;
  width: 100%;
  overflow: hidden;
}

.setup-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 40px;
  text-align: center;
}

.setup-header h1 {
  margin: 0 0 10px 0;
  font-size: 2rem;
  font-weight: 600;
}

.setup-header p {
  margin: 0;
  opacity: 0.9;
}

.setup-content {
  padding: 40px;
}

.step-indicator {
  display: flex;
  justify-content: space-between;
  margin-bottom: 40px;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  position: relative;
}

.step:not(:last-child)::after {
  content: '';
  position: absolute;
  top: 20px;
  left: 60%;
  width: 80%;
  height: 2px;
  background: #e2e8f0;
  z-index: 0;
}

.step.completed:not(:last-child)::after {
  background: #667eea;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e2e8f0;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  margin-bottom: 8px;
  position: relative;
  z-index: 1;
}

.step.active .step-number {
  background: #667eea;
  color: white;
}

.step.completed .step-number {
  background: #10b981;
  color: white;
}

.step-title {
  font-size: 0.875rem;
  color: #64748b;
  text-align: center;
}

.step.active .step-title {
  color: #667eea;
  font-weight: 600;
}

.step-content {
  min-height: 300px;
}

.step-form h3 {
  margin: 0 0 24px 0;
  color: #1e293b;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-group.checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group.checkbox input {
  width: auto;
  margin: 0;
}

.test-btn {
  background: #f59e0b;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  margin-top: 10px;
}

.test-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.step-complete {
  text-align: center;
  padding: 40px 0;
}

.success-icon {
  width: 80px;
  height: 80px;
  background: #10b981;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  margin: 0 auto 24px auto;
}

.config-summary {
  background: #f8fafc;
  border-radius: 8px;
  padding: 20px;
  margin-top: 24px;
  text-align: left;
}

.config-summary h4 {
  margin: 0 0 12px 0;
  color: #1e293b;
}

.config-summary ul {
  margin: 0;
  padding: 0;
  list-style: none;
}

.config-summary li {
  padding: 4px 0;
  color: #64748b;
}

.setup-actions {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-top: 40px;
  padding-top: 24px;
  border-top: 1px solid #e2e8f0;
}

.btn-primary,
.btn-secondary {
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #5a67d8;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f8fafc;
  color: #64748b;
  border: 1px solid #d1d5db;
}

.btn-secondary:hover {
  background: #f1f5f9;
}
</style>