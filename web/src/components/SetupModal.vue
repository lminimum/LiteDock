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
  background: var(--color-background-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: var(--space-4);
}

.setup-modal {
  background: var(--color-background);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-sm);
  max-width: 600px;
  width: 100%;
  overflow: hidden;
  max-height: 90vh;
  overflow-y: auto;
}

.setup-modal-header {
  background: var(--color-background-weak);
  padding: var(--space-6);
  text-align: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}

.setup-modal-header h2 {
  margin: 0 0 var(--space-2) 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  font-family: var(--font-mono);
}

.setup-modal-header p {
  margin: 0;
  color: var(--color-text);
  font-size: var(--font-size-sm);
}

.setup-modal-content {
  padding: var(--space-6);
}

.step-indicator {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-6);
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
  top: 18px;
  left: 50%;
  width: 100%;
  height: 2px;
  background: var(--color-border);
  z-index: 0;
}

.step.completed:not(:last-child)::after {
  background: var(--color-success);
}

.step-number {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  background: var(--color-background-weak);
  color: var(--color-text-weak);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-weight-semibold);
  margin-bottom: var(--space-2);
  position: relative;
  z-index: 1;
  transition: all var(--transition-fast);
}

.step.active .step-number {
  background: var(--color-accent);
  color: #fdfcfc;
}

.step.completed .step-number {
  background: var(--color-success);
  color: #fdfcfc;
}

.step-title {
  font-size: var(--font-size-xs);
  color: var(--color-text-weak);
  text-align: center;
  font-family: var(--font-mono);
}

.step.active .step-title {
  color: var(--color-accent);
  font-weight: var(--font-weight-medium);
}

.step-content {
  min-height: 300px;
}

.step-form h3 {
  margin: 0 0 var(--space-5) 0;
  color: var(--color-text-strong);
  font-size: var(--font-size-lg);
  font-family: var(--font-mono);
}

.form-group {
  margin-bottom: var(--space-4);
}

.form-group label {
  display: block;
  margin-bottom: var(--space-2);
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-sm);
  font-family: var(--font-mono);
}

.form-group input,
.form-group select {
  width: 100%;
  padding: var(--space-3);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  transition: border-color var(--transition-fast);
  background: var(--color-background);
  color: var(--color-text-strong);
  font-family: var(--font-mono);
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--color-accent);
}

.form-group.checkbox {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.form-group.checkbox input {
  width: auto;
  margin: 0;
  accent-color: var(--color-accent);
}

.step-complete {
  text-align: center;
  padding: var(--space-6) 0;
}

.success-icon {
  width: 60px;
  height: 60px;
  background: var(--color-success);
  color: #fdfcfc;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-xl);
  margin: 0 auto var(--space-4) auto;
  font-weight: var(--font-weight-bold);
}

.config-summary {
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-4);
  margin-top: var(--space-4);
  text-align: left;
}

.config-summary h4 {
  margin: 0 0 var(--space-3) 0;
  color: var(--color-text-strong);
  font-size: var(--font-size-base);
  font-family: var(--font-mono);
}

.config-summary ul {
  margin: 0;
  padding: 0;
  list-style: none;
}

.config-summary li {
  padding: var(--space-1) 0;
  color: var(--color-text);
  font-size: var(--font-size-sm);
  font-family: var(--font-mono);
}

.setup-modal-actions {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  margin-top: var(--space-6);
  padding-top: var(--space-5);
  border-top: 1px solid var(--color-border-weak);
}

.btn-primary,
.btn-secondary {
  padding: var(--space-2) var(--space-5);
  border-radius: var(--radius-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  border: 1px solid transparent;
  transition: all var(--transition-fast);
  font-size: var(--font-size-sm);
  font-family: var(--font-mono);
}

.btn-primary {
  background: var(--color-background-strong);
  color: var(--color-text-strong);
  border-color: var(--color-text-weaker);
}

.btn-primary:hover:not(:disabled) {
  background: #3d3939;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: transparent;
  color: var(--color-text-strong);
  border-color: var(--color-border);
}

.btn-secondary:hover {
  background: var(--color-background-weak);
}

@media (max-width: 767px) {
  .modal-overlay {
    padding: var(--space-2);
  }

  .setup-modal-header {
    padding: var(--space-4);
  }

  .setup-modal-content {
    padding: var(--space-4);
  }

  .step-title {
    font-size: 0.65rem;
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