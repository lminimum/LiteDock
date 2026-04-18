<template>
  <div class="settings-page">
    <div class="page-header">
      <div class="page-header-icon">
        <Settings :size="24" :stroke-width="1.5" />
      </div>
      <div>
        <h1>系统设置</h1>
        <p class="page-description">管理 LiteDock 平台配置</p>
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
        <!-- 常规设置 -->
        <div v-if="activeSection === 'general'" class="settings-section">
          <h2>常规设置</h2>
          <div class="setting-group">
            <label>系统名称</label>
            <input v-model="settings.systemName" type="text" class="input" />
          </div>
          <div class="setting-group">
            <label>系统描述</label>
            <textarea v-model="settings.systemDescription" rows="3" class="input"></textarea>
          </div>
          <div class="setting-group">
            <label>默认语言</label>
            <select v-model="settings.language" class="input">
              <option value="zh-CN">简体中文</option>
              <option value="en-US">English</option>
            </select>
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.enableNotifications" type="checkbox" id="notifications" />
            <label for="notifications">启用系统通知</label>
          </div>
        </div>

        <!-- Docker 设置 -->
        <div v-if="activeSection === 'docker'" class="settings-section">
          <h2>Docker 设置</h2>
          <div class="setting-group">
            <label>Docker 连接类型</label>
            <select v-model="settings.docker.type" class="input">
              <option value="local">本地 Docker</option>
              <option value="remote">远程 Docker</option>
            </select>
          </div>
          <div v-if="settings.docker.type === 'remote'" class="setting-group">
            <label>远程主机地址</label>
            <input v-model="settings.docker.host" type="text" placeholder="tcp://remote-host:2375" class="input" />
          </div>
          <div class="setting-group">
            <label>默认镜像仓库</label>
            <input v-model="settings.docker.defaultRegistry" type="text" placeholder="docker.io" class="input" />
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.docker.autoPrune" type="checkbox" id="autoPrune" />
            <label for="autoPrune">自动清理未使用的镜像</label>
          </div>
        </div>

        <!-- 安全设置 -->
        <div v-if="activeSection === 'security'" class="settings-section">
          <h2>安全设置</h2>
          <div class="setting-group">
            <label>会话超时 (分钟)</label>
            <input v-model="settings.sessionTimeout" type="number" min="5" max="1440" class="input" />
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.enableTwoFactor" type="checkbox" id="2fa" />
            <label for="2fa">启用双因素认证</label>
          </div>
          <div class="setting-group checkbox">
            <input v-model="settings.enableAuditLog" type="checkbox" id="audit" />
            <label for="audit">启用操作审计日志</label>
          </div>
        </div>

        <!-- 监控设置 -->
        <div v-if="activeSection === 'monitoring'" class="settings-section">
          <h2>监控设置</h2>
          <div class="setting-group checkbox">
            <input v-model="settings.monitoring.enabled" type="checkbox" id="monitoring" />
            <label for="monitoring">启用系统监控</label>
          </div>
          <div class="setting-group">
            <label>数据收集间隔 (秒)</label>
            <input v-model="settings.monitoring.interval" type="number" min="10" max="300" class="input" />
          </div>
          <div class="setting-group">
            <label>历史数据保留天数</label>
            <input v-model="settings.monitoring.retention" type="number" min="1" max="365" class="input" />
          </div>
        </div>

        <!-- 保存按钮 -->
        <div class="settings-actions">
          <button @click="saveSettings" class="btn btn-primary" :disabled="saving">
            <Save :size="14" :stroke-width="1.5" />
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
          <button @click="resetSettings" class="btn btn-secondary">
            <RotateCcw :size="14" :stroke-width="1.5" />
            重置为默认
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, markRaw, type Component } from 'vue'
import { Settings, Container, Shield, Activity, Save, RotateCcw } from 'lucide-vue-next'

const activeSection = ref('general')
const saving = ref(false)

const settingsSections: { id: string; title: string; icon: Component }[] = [
  { id: 'general', title: '常规', icon: markRaw(Settings) },
  { id: 'docker', title: 'Docker', icon: markRaw(Container) },
  { id: 'security', title: '安全', icon: markRaw(Shield) },
  { id: 'monitoring', title: '监控', icon: markRaw(Activity) }
]

const defaultSettings = () => ({
  systemName: 'LiteDock',
  systemDescription: '轻量级 Docker 容器管理平台',
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
  }
})

const settings = reactive(defaultSettings())

const saveSettings = async () => {
  saving.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 1000))
    alert('设置已保存')
  } finally {
    saving.value = false
  }
}

const resetSettings = () => {
  if (confirm('确定要重置所有设置为默认值吗？')) {
    Object.assign(settings, defaultSettings())
  }
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
  border-radius: var(--radius-md);
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
  border-radius: var(--radius-md);
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
  border-radius: var(--radius-md);
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
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.setting-group .input::placeholder {
  color: var(--color-text-weaker);
}

.setting-group .input:focus {
  outline: none;
  border-color: var(--color-background-strong);
  box-shadow: 0 0 0 3px var(--color-background-interactive-weaker);
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
