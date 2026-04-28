<template>
  <div v-if="show" class="modal-overlay" @click="closeModal">
    <div class="setup-modal" @click.stop :class="{ 'dark-theme': isDarkMode }">
      <div class="setup-modal-header">
        <h2>{{ t('setup.title') }}</h2>
        <p>{{ t('setup.welcome') }}</p>
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
            <h3>{{ t('setup.dockerConnectionConfig') }}</h3>
            <div class="form-group">
              <label>{{ t('setup.connectionType') }}</label>
              <select v-model="config.docker.type" @change="onDockerTypeChange">
                <option value="local">{{ t('setup.localDocker') }}</option>
                <option value="remote">{{ t('setup.remoteDocker') }}</option>
              </select>
            </div>
            
            <div v-if="config.docker.type === 'remote'" class="form-group">
              <label>{{ t('setup.remoteHostAddress') }}</label>
              <input 
                v-model="config.docker.host" 
                placeholder="tcp://remote-host:2375"
                type="text"
              />
            </div>
            
            <div class="form-group">
              <label>{{ t('setup.tlsCertPath') }}</label>
              <input 
                v-model="config.docker.tlsPath" 
                placeholder="/path/to/certs"
                type="text"
              />
            </div>
          </div>
          
          <!-- 步骤 2: 管理员账户设置 -->
          <div v-if="currentStep === 1" class="step-form">
            <h3>{{ t('setup.adminAccountSetup') }}</h3>
            <div class="form-group">
              <label>{{ t('setup.username') }}</label>
              <input v-model="config.admin.username" placeholder="admin" type="text" />
            </div>
            
            <div class="form-group">
              <label>{{ t('setup.password') }}</label>
              <input v-model="config.admin.password" :placeholder="t('setup.passwordPlaceholder')" type="password" />
            </div>
            
            <div class="form-group">
              <label>{{ t('setup.confirmPassword') }}</label>
              <input v-model="config.admin.confirmPassword" :placeholder="t('setup.confirmPasswordPlaceholder')" type="password" />
            </div>
            
            <div class="form-group">
              <label>{{ t('setup.emailOptional') }}</label>
              <input v-model="config.admin.email" placeholder="admin@example.com" type="email" />
            </div>
          </div>
          
          <!-- 步骤 3: 系统设置 -->
          <div v-if="currentStep === 2" class="step-form">
            <h3>{{ t('setup.systemSettingsTitle') }}</h3>
            <div class="form-group">
              <label>{{ t('setup.systemName') }}</label>
              <input v-model="config.system.name" placeholder="LiteDock" type="text" />
            </div>
            
            <div class="form-group">
              <label>{{ t('setup.listenPort') }}</label>
              <input v-model="config.system.port" placeholder="8080" type="number" min="1024" max="65535" />
            </div>
            
            <div class="form-group checkbox">
              <input v-model="config.system.enableMetrics" type="checkbox" id="metrics" />
              <label for="metrics">{{ t('setup.enableMetrics') }}</label>
            </div>
            
            <div class="form-group checkbox">
              <input v-model="config.system.enableSwagger" type="checkbox" id="swagger" />
              <label for="swagger">{{ t('setup.enableApiDocs') }}</label>
            </div>
          </div>
          
          <!-- 步骤 4: 完成配置 -->
          <div v-if="currentStep === 3" class="step-complete">
            <div class="success-icon">✓</div>
            <h3>{{ t('setup.setupComplete') }}</h3>
            <p>{{ t('setup.setupCompleteDesc') }}</p>
            
            <div class="config-summary">
              <h4>{{ t('setup.configSummary') }}</h4>
              <ul>
                <li>{{ t('setup.dockerConnection') }} {{ config.docker.type === 'local' ? t('setup.local') : t('setup.remote') }}</li>
                <li>{{ t('setup.adminUser') }} {{ config.admin.username }}</li>
                <li>{{ t('setup.systemPort') }} {{ config.system.port }}</li>
                <li>{{ t('setup.metricsStatus') }} {{ config.system.enableMetrics ? t('setup.enabled') : t('setup.disabled') }}</li>
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
            {{ t('setup.previousStep') }}
          </button>
          
          <button 
            v-if="currentStep < steps.length - 1" 
            @click="nextStep" 
            class="btn-primary"
            :disabled="!canProceed"
          >
            {{ t('setup.nextStep') }}
          </button>
          
          <button 
            v-if="currentStep === steps.length - 1" 
            @click="completeSetup" 
            class="btn-primary"
          >
            {{ t('setup.startUsing') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '@/i18n'

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
  { title: t('setup.dockerConfig') },
  { title: t('setup.adminAccount') },
  { title: t('setup.systemSettings') },
  { title: t('setup.complete') }
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
  background: var(--setup-overlay-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.setup-modal {
  background: var(--setup-modal-bg);
  border-radius: 8px;
  box-shadow: var(--setup-modal-shadow);
  max-width: 600px;
  width: 100%;
  overflow: hidden;
  max-height: 90vh;
  overflow-y: auto;
}

.setup-modal.dark-theme {
  background: var(--setup-modal-bg);
  color: var(--setup-modal-text-on-dark);
}

.setup-modal-header {
  background: var(--setup-modal-header-bg);
  padding: 24px;
  text-align: center;
  border-bottom: 1px solid var(--setup-modal-header-border);
}

.setup-modal.dark-theme .setup-modal-header {
  background: var(--setup-modal-header-bg);
  border-bottom: 1px solid var(--setup-modal-border);
}

.setup-modal-header h2 {
  margin: 0 0 8px 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--setup-modal-text);
}

.setup-modal.dark-theme .setup-modal-header h2 {
  color: var(--setup-modal-text-on-dark);
}

.setup-modal-header p {
  margin: 0;
  color: var(--setup-modal-text-secondary);
  font-size: 0.875rem;
}

.setup-modal.dark-theme .setup-modal-header p {
  color: var(--setup-step-number-text);
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
  background: var(--setup-step-line);
  z-index: 0;
}

.setup-modal.dark-theme .step:not(:last-child)::after {
  background: var(--setup-step-line);
}

.step.completed:not(:last-child)::after {
  background: var(--setup-step-active-bg);
}

.setup-modal.dark-theme .step.completed:not(:last-child)::after {
  background: var(--setup-step-active-bg);
}

.step-number {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--setup-step-number-bg);
  color: var(--setup-step-number-text);
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
  background: var(--setup-step-number-bg);
  color: var(--setup-step-number-text);
}

.step.active .step-number {
  background: var(--setup-step-active-bg);
  color: var(--setup-step-active-text);
}

.setup-modal.dark-theme .step.active .step-number {
  background: var(--setup-step-active-bg);
  color: var(--setup-step-active-text);
}

.step.completed .step-number {
  background: var(--setup-step-completed-bg);
  color: var(--setup-step-completed-text);
}

.setup-modal.dark-theme .step.completed .step-number {
  background: var(--setup-step-completed-bg);
  color: var(--setup-step-completed-text);
}

.step-title {
  font-size: 0.75rem;
  color: var(--setup-step-title);
  text-align: center;
}

.setup-modal.dark-theme .step-title {
  color: var(--setup-step-title);
}

.step.active .step-title {
  color: var(--setup-step-active-bg);
  font-weight: 600;
}

.setup-modal.dark-theme .step.active .step-title {
  color: var(--setup-step-active-bg);
}

.step-content {
  min-height: 300px;
}

.setup-modal.dark-theme .step-content {
  color: var(--setup-modal-text-on-dark);
}

.step-form h3 {
  margin: 0 0 20px 0;
  color: var(--setup-modal-text);
  font-size: 1.25rem;
}

.setup-modal.dark-theme .step-form h3 {
  color: var(--setup-modal-text-on-dark);
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  color: var(--setup-form-label);
  font-weight: 500;
  font-size: 0.875rem;
}

.setup-modal.dark-theme .form-group label {
  color: var(--setup-form-label);
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--setup-input-border);
  border-radius: 6px;
  font-size: 0.875rem;
  transition: border-color 0.2s;
  background: var(--setup-input-bg);
  color: var(--setup-input-text);
}

.setup-modal.dark-theme .form-group input,
.setup-modal.dark-theme .form-group select {
  background: var(--setup-input-bg);
  border-color: var(--setup-input-border);
  color: var(--setup-input-text);
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--setup-focus-border);
  box-shadow: var(--setup-focus-shadow);
}

.setup-modal.dark-theme .form-group input:focus,
.setup-modal.dark-theme .form-group select:focus {
  border-color: var(--setup-focus-border);
  box-shadow: var(--setup-focus-shadow);
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
  background: var(--setup-success-bg);
  color: var(--setup-success-text);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin: 0 auto 16px auto;
  font-weight: bold;
}

.setup-modal.dark-theme .success-icon {
  background: var(--setup-success-bg);
  color: var(--setup-success-text);
}

.config-summary {
  background: var(--setup-config-summary-bg);
  border-radius: 6px;
  padding: 16px;
  margin-top: 16px;
  text-align: left;
}

.setup-modal.dark-theme .config-summary {
  background: var(--setup-config-summary-bg);
}

.config-summary h4 {
  margin: 0 0 12px 0;
  color: var(--setup-modal-text);
  font-size: 1rem;
}

.setup-modal.dark-theme .config-summary h4 {
  color: var(--setup-modal-text-on-dark);
}

.config-summary ul {
  margin: 0;
  padding: 0;
  list-style: none;
}

.config-summary li {
  padding: 4px 0;
  color: var(--setup-modal-text-secondary);
  font-size: 0.875rem;
}

.setup-modal.dark-theme .config-summary li {
  color: var(--setup-step-number-text);
}

.setup-modal-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--setup-modal-border);
}

.setup-modal.dark-theme .setup-modal-actions {
  border-top: 1px solid var(--setup-modal-border);
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
  background: var(--setup-btn-primary-bg);
  color: var(--setup-btn-primary-text);
}

.btn-primary:hover:not(:disabled) {
  background: var(--setup-btn-primary-hover);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary.dark-theme {
  background: var(--setup-btn-primary-bg);
}

.btn-primary.dark-theme:hover:not(:disabled) {
  background: var(--setup-btn-primary-hover);
}

.btn-secondary {
  background: var(--setup-btn-secondary-bg);
  color: var(--setup-btn-secondary-text);
  border: 1px solid var(--setup-btn-secondary-border);
}

.setup-modal.dark-theme .btn-secondary {
  background: var(--setup-btn-secondary-bg);
  color: var(--setup-btn-secondary-text);
  border: 1px solid var(--setup-btn-secondary-border);
}

.btn-secondary:hover {
  background: var(--setup-btn-secondary-hover);
}

.setup-modal.dark-theme .btn-secondary:hover {
  background: var(--setup-btn-secondary-hover);
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