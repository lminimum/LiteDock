<template>
  <div v-if="show" class="modal-overlay" @click="closeModal">
    <div class="setup-modal" @click.stop :class="{ 'dark-theme': isDarkMode }">
      <div class="setup-modal-header">
        <h2>LiteDock 初始配置</h2>
        <p>欢迎使用 LiteDock Docker 管理平台</p>
      </div>
      
      <div class="setup-modal-content">
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
        
        <div class="setup-modal-actions">
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
          >
            开始使用
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

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

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits(['close', 'complete'])

const router = useRouter()

const steps = [
  { title: 'Docker 配置' },
  { title: '管理员账户' },
  { title: '系统设置' },
  { title: '完成' }
]

const currentStep = ref(0)

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

const isDarkMode = ref(false)

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

const nextStep = () => {
  if (canProceed.value) {
    currentStep.value++
  }
}

const previousStep = () => {
  currentStep.value--
}

const completeSetup = async () => {
  // 模拟配置保存
  localStorage.setItem('litedock-configured', 'true')
  localStorage.setItem('litedock-config', JSON.stringify(config.value))
  
  // 直接跳转到登录页面
  router.push('/login')
  
  // 发出完成事件
  emit('complete')
}

const closeModal = () => {
  emit('close')
}

// 恢复主题设置
onMounted(() => {
  const savedTheme = localStorage.getItem('litedock-theme')
  if (savedTheme) {
    isDarkMode.value = savedTheme === 'dark'
  }
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.setup-modal {
  background: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  max-width: 600px;
  width: 100%;
  overflow: hidden;
  max-height: 90vh;
  overflow-y: auto;
}

.setup-modal.dark-theme {
  background: #1e293b;
  color: #f1f5f9;
}

.setup-modal-header {
  background: #f8fafc;
  padding: 24px;
  text-align: center;
  border-bottom: 1px solid #e2e8f0;
}

.setup-modal.dark-theme .setup-modal-header {
  background: #334155;
  border-bottom: 1px solid #475569;
}

.setup-modal-header h2 {
  margin: 0 0 8px 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: #1e293b;
}

.setup-modal.dark-theme .setup-modal-header h2 {
  color: #f1f5f9;
}

.setup-modal-header p {
  margin: 0;
  color: #64748b;
  font-size: 0.875rem;
}

.setup-modal.dark-theme .setup-modal-header p {
  color: #cbd5e1;
}

.setup-modal-content {
  padding: 24px;
}

.step-indicator {
  display: flex;
  justify-content: space-between;
  margin-bottom: 24px;
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
  left: 50%;
  width: 100%;
  height: 2px;
  background: #e2e8f0;
  z-index: 0;
}

.setup-modal.dark-theme .step:not(:last-child)::after {
  background: #475569;
}

.step.completed:not(:last-child)::after {
  background: #667eea;
}

.setup-modal.dark-theme .step.completed:not(:last-child)::after {
  background: #60a5fa;
}

.step-number {
  width: 36px;
  height: 36px;
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
  transition: all 0.2s;
}

.setup-modal.dark-theme .step-number {
  background: #475569;
  color: #cbd5e1;
}

.step.active .step-number {
  background: #667eea;
  color: white;
}

.setup-modal.dark-theme .step.active .step-number {
  background: #60a5fa;
  color: white;
}

.step.completed .step-number {
  background: #10b981;
  color: white;
}

.setup-modal.dark-theme .step.completed .step-number {
  background: #34d399;
  color: #0f172a;
}

.step-title {
  font-size: 0.75rem;
  color: #64748b;
  text-align: center;
}

.setup-modal.dark-theme .step-title {
  color: #94a3b8;
}

.step.active .step-title {
  color: #667eea;
  font-weight: 600;
}

.setup-modal.dark-theme .step.active .step-title {
  color: #60a5fa;
}

.step-content {
  min-height: 300px;
}

.setup-modal.dark-theme .step-content {
  color: #f1f5f9;
}

.step-form h3 {
  margin: 0 0 20px 0;
  color: #1e293b;
  font-size: 1.25rem;
}

.setup-modal.dark-theme .step-form h3 {
  color: #f1f5f9;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  color: #374151;
  font-weight: 500;
  font-size: 0.875rem;
}

.setup-modal.dark-theme .form-group label {
  color: #e2e8f0;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: border-color 0.2s;
  background: white;
  color: #111827;
}

.setup-modal.dark-theme .form-group input,
.setup-modal.dark-theme .form-group select {
  background: #475569;
  border-color: #64748b;
  color: #f1f5f9;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.setup-modal.dark-theme .form-group input:focus,
.setup-modal.dark-theme .form-group select:focus {
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
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

.step-complete {
  text-align: center;
  padding: 24px 0;
}

.success-icon {
  width: 60px;
  height: 60px;
  background: #10b981;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin: 0 auto 16px auto;
  font-weight: bold;
}

.setup-modal.dark-theme .success-icon {
  background: #34d399;
  color: #0f172a;
}

.config-summary {
  background: #f8fafc;
  border-radius: 6px;
  padding: 16px;
  margin-top: 16px;
  text-align: left;
}

.setup-modal.dark-theme .config-summary {
  background: #334155;
}

.config-summary h4 {
  margin: 0 0 12px 0;
  color: #1e293b;
  font-size: 1rem;
}

.setup-modal.dark-theme .config-summary h4 {
  color: #f1f5f9;
}

.config-summary ul {
  margin: 0;
  padding: 0;
  list-style: none;
}

.config-summary li {
  padding: 4px 0;
  color: #64748b;
  font-size: 0.875rem;
}

.setup-modal.dark-theme .config-summary li {
  color: #cbd5e1;
}

.setup-modal-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.setup-modal.dark-theme .setup-modal-actions {
  border-top: 1px solid #475569;
}

.btn-primary,
.btn-secondary {
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  font-size: 0.875rem;
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

.btn-primary.dark-theme {
  background: #60a5fa;
}

.btn-primary.dark-theme:hover:not(:disabled) {
  background: #3b82f6;
}

.btn-secondary {
  background: #f8fafc;
  color: #64748b;
  border: 1px solid #d1d5db;
}

.setup-modal.dark-theme .btn-secondary {
  background: #475569;
  color: #cbd5e1;
  border: 1px solid #64748b;
}

.btn-secondary:hover {
  background: #f1f5f9;
}

.setup-modal.dark-theme .btn-secondary:hover {
  background: #64748b;
}

/* Media Queries for Responsive Design */
@media (max-width: 767px) {
  .modal-overlay {
    padding: 10px;
  }

  .setup-modal {
    margin: 10px;
    max-height: 95vh;
  }

  .setup-modal-header {
    padding: 16px;
  }

  .setup-modal-header h2 {
    font-size: 1.25rem;
  }

  .setup-modal-content {
    padding: 16px;
  }

  .step-indicator {
    margin-bottom: 16px;
  }

  .step-title {
    font-size: 0.65rem;
  }

  .step-form h3 {
    font-size: 1.1rem;
  }

  .form-group input,
  .form-group select {
    font-size: 0.8rem;
    padding: 8px 10px;
  }

  .setup-modal-actions {
    flex-direction: column;
  }

  .btn-primary,
  .btn-secondary {
    width: 100%;
  }
}
</style>