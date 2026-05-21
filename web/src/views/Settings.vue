<template>
  <div class="settings-page">
    <div class="page-header">
      <div class="page-header-icon">
        <Settings :size="24" :stroke-width="1.5" />
      </div>
      <div>
        <h1>{{ t('settings.title') }}</h1>
        <p class="page-description">{{ t('settings.description') }}</p>
      </div>
    </div>

    <div class="settings-content">
      <div class="settings-sidebar">
        <nav class="settings-nav">
          <a
            v-for="section in settingsSections"
            :key="section.id"
            href="#"
            @click.prevent="activeSection = section.id"
            :class="{ active: activeSection === section.id }"
          >
            <component :is="section.icon" :size="16" :stroke-width="1.5" />
            <span>{{ section.title }}</span>
          </a>
        </nav>
      </div>

      <div class="settings-main">
        <!-- General -->
        <div v-if="activeSection === 'general'" class="settings-section">
          <h2>{{ t('settings.general') }}</h2>
          <div class="setting-group">
            <label>{{ t('settings.systemName') }}</label>
            <input v-model="settings.systemName" type="text" class="input" />
          </div>
          <div class="setting-group">
            <label>{{ t('settings.systemDesc') }}</label>
            <textarea v-model="settings.systemDescription" rows="3" class="input"></textarea>
          </div>
          <div class="setting-group">
            <label>{{ t('settings.defaultLanguage') }}</label>
            <select v-model="settings.language" class="input">
              <option value="zh-CN">简体中文</option>
              <option value="en-US">English</option>
            </select>
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.enableNotifications" type="checkbox" id="notifications" />
            <label for="notifications">{{ t('settings.enableNotifications') }}</label>
          </div>
        </div>

        <!-- Docker -->
        <div v-if="activeSection === 'docker'" class="settings-section">
          <h2>{{ t('settings.docker') }}</h2>
          <div class="setting-group">
            <label>{{ t('settings.dockerConnectionType') }}</label>
            <select v-model="settings.docker.type" class="input">
              <option value="local">{{ t('settings.localDocker') }}</option>
              <option value="remote">{{ t('settings.remoteDocker') }}</option>
            </select>
          </div>
          <div v-if="settings.docker.type === 'remote'" class="setting-group">
            <label>{{ t('settings.remoteHost') }}</label>
            <input v-model="settings.docker.host" type="text" placeholder="tcp://remote-host:2375" class="input" />
          </div>
          <div class="setting-group">
            <label>{{ t('settings.defaultRegistry') }}</label>
            <input v-model="settings.docker.defaultRegistry" type="text" placeholder="docker.io" class="input" />
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.docker.autoPrune" type="checkbox" id="autoPrune" />
            <label for="autoPrune">{{ t('settings.autoPrune') }}</label>
          </div>
        </div>

        <!-- Security -->
        <div v-if="activeSection === 'security'" class="settings-section">
          <h2>{{ t('settings.security') }}</h2>
          <div class="setting-group">
            <label>{{ t('settings.sessionTimeout') }}</label>
            <input v-model="settings.sessionTimeout" type="number" min="5" max="1440" class="input" />
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.enableTwoFactor" type="checkbox" id="2fa" />
            <label for="2fa">{{ t('settings.enableTwoFactor') }}</label>
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.enableAuditLog" type="checkbox" id="audit" />
            <label for="audit">{{ t('settings.enableAuditLog') }}</label>
          </div>
        </div>

        <!-- Monitoring -->
        <div v-if="activeSection === 'monitoring'" class="settings-section">
          <h2>{{ t('settings.monitoring') }}</h2>
          <div class="setting-group checkbox">
            <input v-model="settings.monitoring.enabled" type="checkbox" id="monitoring" />
            <label for="monitoring">{{ t('settings.enableSystemMonitoring') }}</label>
          </div>
          <div class="setting-group">
            <label>{{ t('settings.dataCollectionInterval') }}</label>
            <input v-model="settings.monitoring.interval" type="number" min="10" max="300" class="input" />
          </div>
          <div class="setting-group">
            <label>{{ t('settings.dataRetentionDays') }}</label>
            <input v-model="settings.monitoring.retention" type="number" min="1" max="365" class="input" />
          </div>
        </div>

        <!-- AI -->
        <div v-if="activeSection === 'ai'" class="settings-section">
          <h2>{{ t('settings.ai.title') }}</h2>
          <div class="setting-group">
            <label>{{ t('settings.ai.apiEndpoint') }}</label>
            <input v-model="settings.ai.apiEndpoint" type="url" placeholder="https://api.openai.com" class="input" />
          </div>
          <div class="setting-group">
            <label>{{ t('settings.ai.apiKey') }}</label>
            <div class="password-field">
              <input
                v-model="settings.ai.apiKey"
                :type="showApiKey ? 'text' : 'password'"
                class="input"
                placeholder="sk-..."
              />
              <button type="button" class="password-toggle" @click="showApiKey = !showApiKey">
                <Eye v-if="!showApiKey" :size="16" :stroke-width="1.5" />
                <EyeOff v-else :size="16" :stroke-width="1.5" />
              </button>
            </div>
          </div>
          <div class="setting-group">
            <label>{{ t('settings.ai.modelName') }}</label>
            <input v-model="settings.ai.modelName" type="text" placeholder="gpt-4o" class="input" />
          </div>
        </div>

        <!-- Actions -->
        <div class="settings-actions">
          <button @click="saveSettings" class="btn btn-primary" :disabled="saving">
            <Save :size="14" :stroke-width="1.5" />
            {{ saving ? t('settings.saving') : t('settings.saveSettings') }}
          </button>
          <button @click="resetSettings" class="btn btn-secondary">
            <RotateCcw :size="14" :stroke-width="1.5" />
            {{ t('settings.resetDefaults') }}
          </button>
        </div>
      </div>
    </div>

    <ConfirmModal
      :visible="confirmState !== null"
      :title="confirmState?.title || ''"
      :message="confirmState?.message || ''"
      :confirm-text="confirmState?.confirmText"
      :danger="confirmState?.danger ?? false"
      :disabled="confirmBusy"
      @confirm="confirmAction"
      @cancel="cancelConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, markRaw, onMounted, type Component } from 'vue'
import { Settings, Container, Shield, Activity, Save, RotateCcw, Bot, Eye, EyeOff } from 'lucide-vue-next'
import { t } from '@/i18n'
import api from '@/utils/api'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'

const activeSection = ref('general')
const saving = ref(false)
const showApiKey = ref(false)
const confirmState = ref<{
  title: string
  message: string
  confirmText?: string
  danger?: boolean
  action: 'reset'
} | null>(null)
const confirmBusy = ref(false)

const settingsSections = computed<{ id: string; title: string; icon: Component }[]>(() => [
  { id: 'general', title: t('settings.general'), icon: markRaw(Settings) },
  { id: 'docker', title: t('settings.docker'), icon: markRaw(Container) },
  { id: 'security', title: t('settings.security'), icon: markRaw(Shield) },
  { id: 'monitoring', title: t('settings.monitoring'), icon: markRaw(Activity) },
  { id: 'ai', title: t('settings.ai.title'), icon: markRaw(Bot) }
])

const defaultSettings = () => ({
  systemName: 'LiteDock',
  systemDescription: '',
  language: 'zh-CN',
  enableNotifications: true,
  docker: {
    type: 'local' as 'local' | 'remote',
    host: '',
    defaultRegistry: 'docker.io',
    autoPrune: false
  },
  sessionTimeout: 60,
  enableTwoFactor: false,
  enableAuditLog: true,
  monitoring: {
    enabled: true,
    interval: 30,
    retention: 30
  },
  ai: {
    apiEndpoint: '',
    apiKey: '',
    modelName: 'gpt-4o'
  }
})

const settings = reactive(defaultSettings())

const fetchAISettings = async () => {
  try {
    const data = await api.get('/assistant/settings')
    if (data) Object.assign(settings.ai, data)
  } catch {
    // Keep defaults on error
  }
}

onMounted(fetchAISettings)

const saveSettings = async () => {
  saving.value = true
  try {
    const resp = await api.put('/assistant/settings', settings.ai)
    Object.assign(settings.ai, resp)
    alert(t('settings.saved'))
  } catch {
    alert('Failed to save settings')
  } finally {
    saving.value = false
  }
}

const cancelConfirm = () => {
  if (confirmBusy.value) return
  confirmState.value = null
}

const openResetConfirm = () => {
  confirmState.value = {
    title: t('settings.resetDefaults'),
    message: t('settings.confirmReset'),
    confirmText: t('settings.resetDefaults'),
    danger: true,
    action: 'reset',
  }
}

const confirmAction = async () => {
  const state = confirmState.value
  if (!state || confirmBusy.value) return
  confirmBusy.value = true
  confirmState.value = null

  try {
    if (state.action === 'reset') {
      performReset()
    }
  } finally {
    confirmBusy.value = false
  }
}

const performReset = () => {
  Object.assign(settings, defaultSettings())
}

const resetSettings = () => {
  openResetConfirm()
}
</script>

<style scoped>
.settings-page {
  max-width: var(--container-max);
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-bottom: var(--space-8);
}

.page-header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-sm);
  background: var(--color-background-interactive-weaker);
  color: var(--color-text-strong);
  flex-shrink: 0;
}

.page-header h1 {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  font-family: var(--font-mono);
}

.page-description {
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
  margin-top: var(--space-1);
}

.settings-content {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: var(--space-8);
}

.settings-sidebar {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-4);
  height: fit-content;
}

.settings-nav {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.settings-nav a {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text-weak);
  text-decoration: none;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  transition: all var(--transition-fast);
}

.settings-nav a:hover {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

.settings-nav a.active {
  background: var(--color-background-interactive-weaker);
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
}

.settings-main {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-8);
}

.settings-section h2 {
  margin: 0 0 var(--space-6) 0;
  color: var(--color-text-strong);
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  font-family: var(--font-mono);
}

.setting-group {
  margin-bottom: var(--space-6);
}

.setting-group label {
  display: block;
  margin-bottom: var(--space-2);
  color: var(--color-text);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-sm);
  font-family: var(--font-mono);
}

.setting-group .input {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text-strong);
  background: var(--color-background);
  border-radius: var(--radius-md);
  transition: border-color var(--transition-fast);
}

.setting-group .input::placeholder {
  color: var(--color-text-weaker);
}

.setting-group .input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.setting-group textarea.input {
  resize: vertical;
  min-height: 80px;
}

.setting-group select.input {
  cursor: pointer;
}

.setting-group.checkbox {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.setting-group.checkbox input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--color-background-strong);
  cursor: pointer;
  flex-shrink: 0;
}

.setting-group.checkbox label {
  margin: 0;
  cursor: pointer;
  color: var(--color-text);
}

.settings-actions {
  display: flex;
  gap: var(--space-3);
  padding-top: var(--space-6);
  border-top: 1px solid var(--color-border-weak);
  margin-top: var(--space-8);
}

.password-field {
  position: relative;
}

.password-field .input {
  padding-right: var(--space-8);
}

.password-toggle {
  position: absolute;
  right: var(--space-2);
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-weaker);
  padding: var(--space-1);
  border-radius: var(--radius-sm);
  transition: color var(--transition-fast);
}

.password-toggle:hover {
  color: var(--color-text);
}

@media (max-width: 767px) {
  .settings-content {
    grid-template-columns: 1fr;
    gap: var(--space-4);
  }

  .settings-main {
    padding: var(--space-6);
  }

  .page-header-icon {
    width: 40px;
    height: 40px;
  }

  .page-header h1 {
    font-size: var(--font-size-xl);
  }
}
</style>
