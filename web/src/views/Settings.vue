<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>系统设置</h1>
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
            {{ section.icon }} {{ section.title }}
          </a>
        </nav>
      </div>
      
      <div class="settings-main">
        <!-- 常规设置 -->
        <div v-if="activeSection === 'general'" class="settings-section">
          <h2>常规设置</h2>
          <div class="setting-group">
            <label>系统名称</label>
            <input v-model="settings.systemName" type="text" />
          </div>
          <div class="setting-group">
            <label>系统描述</label>
            <textarea v-model="settings.systemDescription" rows="3"></textarea>
          </div>
          <div class="setting-group">
            <label>默认语言</label>
            <select v-model="settings.language">
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
            <select v-model="settings.docker.type">
              <option value="local">本地 Docker</option>
              <option value="remote">远程 Docker</option>
            </select>
          </div>
          <div v-if="settings.docker.type === 'remote'" class="setting-group">
            <label>远程主机地址</label>
            <input v-model="settings.docker.host" type="text" placeholder="tcp://remote-host:2375" />
          </div>
          <div class="setting-group">
            <label>默认镜像仓库</label>
            <input v-model="settings.docker.defaultRegistry" type="text" placeholder="docker.io" />
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
            <input v-model="settings.sessionTimeout" type="number" min="5" max="1440" />
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
            <input v-model="settings.monitoring.interval" type="number" min="10" max="300" />
          </div>
          <div class="setting-group">
            <label>历史数据保留天数</label>
            <input v-model="settings.monitoring.retention" type="number" min="1" max="365" />
          </div>
        </div>
        
        <!-- 保存按钮 -->
        <div class="settings-actions">
          <button @click="saveSettings" class="btn-primary" :disabled="saving">
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
          <button @click="resetSettings" class="btn-secondary">
            重置为默认
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const activeSection = ref('general')
const saving = ref(false)

const settingsSections = [
  { id: 'general', title: '常规', icon: '⚙️' },
  { id: 'docker', title: 'Docker', icon: '🐳' },
  { id: 'security', title: '安全', icon: '🔒' },
  { id: 'monitoring', title: '监控', icon: '📊' }
]

const settings = reactive({
  systemName: 'LiteDock',
  systemDescription: '轻量级 Docker 容器管理平台',
  language: 'zh-CN',
  enableNotifications: true,
  docker: {
    type: 'local',
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

const saveSettings = async () => {
  saving.value = true
  try {
    // 这里应该调用API保存设置
    await new Promise(resolve => setTimeout(resolve, 1000))
    alert('设置已保存')
  } finally {
    saving.value = false
  }
}

const resetSettings = () => {
  if (confirm('确定要重置所有设置为默认值吗？')) {
    // 重置设置
    Object.assign(settings, {
      systemName: 'LiteDock',
      systemDescription: '轻量级 Docker 容器管理平台',
      language: 'zh-CN',
      enableNotifications: true,
      docker: {
        type: 'local',
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
  }
}
</script>

<style scoped>
.settings-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 32px;
}

.page-header h1 {
  margin: 0;
  color: #1e293b;
  font-size: 1.875rem;
  font-weight: 600;
}

.settings-content {
  display: grid;
  grid-template-columns: 250px 1fr;
  gap: 32px;
}

.settings-sidebar {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  height: fit-content;
}

.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.settings-nav a {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  color: #64748b;
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.2s;
}

.settings-nav a:hover {
  background: #f8fafc;
  color: #374151;
}

.settings-nav a.active {
  background: #f0f4ff;
  color: #667eea;
}

.settings-main {
  background: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.settings-section h2 {
  margin: 0 0 24px 0;
  color: #1e293b;
  font-size: 1.5rem;
  font-weight: 600;
}

.setting-group {
  margin-bottom: 24px;
}

.setting-group label {
  display: block;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
}

.setting-group input,
.setting-group select,
.setting-group textarea {
  width: 100%;
  padding: 10px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
}

.setting-group.checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
}

.setting-group.checkbox input {
  width: auto;
  margin: 0;
}

.setting-group.checkbox label {
  margin: 0;
  cursor: pointer;
}

.settings-actions {
  display: flex;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid #f1f5f9;
  margin-top: 32px;
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

@media (max-width: 768px) {
  .settings-content {
    grid-template-columns: 1fr;
  }
  
  .settings-sidebar {
    order: 2;
  }
  
  .settings-main {
    order: 1;
  }
}
</style>