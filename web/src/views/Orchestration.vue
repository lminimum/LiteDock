<template>
  <div class="orchestration-page">
    <PageHeader
      :title="t('pages.orchestration.title')"
      :description="t('pages.orchestration.description')"
    >
      <button @click="refreshProjects" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ spinning: loading }" />
        {{ t('pages.orchestration.actions.refresh') }}
      </button>
      <button
        v-if="machines.length > 0"
        @click="showCreateModal = true"
        class="btn btn-primary"
      >
        <Plus :size="16" />
        {{ t('pages.orchestration.actions.create') }}
      </button>
    </PageHeader>

    <!-- Filters -->
    <div v-if="!loading && !error && projects.length > 0" class="filters">
      <div class="search-box">
        <input
          v-model="searchQuery"
          :placeholder="t('common.searchPlaceholder')"
          type="text"
          class="input"
        />
      </div>
      <div class="filter-options">
        <select v-model="statusFilter" class="input">
          <option value="">All</option>
          <option value="running">{{ t('pages.orchestration.status.running') }}</option>
          <option value="stopped">{{ t('pages.orchestration.status.stopped') }}</option>
          <option value="paused">Paused</option>
          <option value="failed">{{ t('pages.orchestration.status.failed') }}</option>
          <option value="partial">{{ t('pages.orchestration.status.partial') }}</option>
        </select>
      </div>
      <ViewToggle v-model="viewMode" />
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="loading-state">
      <RefreshCw :size="24" class="spinning" />
      <span>{{ t('common.refresh') }}...</span>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="error-state card text-center">
      <p class="mb-4">{{ error }}</p>
      <button @click="refreshProjects" class="btn btn-secondary">{{ t('common.refresh') }}</button>
    </div>

    <!-- Empty state (no projects at all) -->
    <div v-else-if="projects.length === 0" class="empty-state card text-center">
      <GitBranch :size="48" class="mb-4" style="color: var(--color-text-weaker)" />
      <h3 class="mb-2">{{ t('pages.orchestration.empty.title') }}</h3>
      <p class="mb-4" style="color: var(--color-text-weak)">{{ t('pages.orchestration.empty.description') }}</p>
      <button
        v-if="machines.length > 0"
        @click="showCreateModal = true"
        class="btn btn-primary"
      >
        <Plus :size="16" />
        {{ t('pages.orchestration.empty.createAction') }}
      </button>
    </div>

    <!-- No-match state (filter has no results) -->
    <div v-else-if="filteredProjects.length === 0" class="empty-state card text-center">
      <p style="color: var(--color-text-weak)">No projects match your search.</p>
    </div>

    <template v-else>
      <Transition name="view-fade" mode="out-in">
        <div v-if="viewMode === 'card'" class="card-grid" key="card">
        <ComposeCard
          v-for="project in filteredProjects"
          :key="`${project.machineId}:${project.id}`"
          :project="project"
          @inspect="handleInspect"
          @delete="handleDelete"
          @up="handleUp"
          @down="handleDown"
          @logs="handleLogsClick"
        />
      </div>
      <div v-else class="item-list">
        <div
          v-for="project in filteredProjects"
          :key="`${project.machineId}:${project.id}`"
          class="item-list-row"
        >
          <div class="item-list-info">
            <div class="item-list-title">{{ project.name }}</div>
            <div class="item-list-meta">
              <span :class="['badge', statusBadgeClass(project.status)]">
                {{ project.status }}
              </span>
              <span>{{ project.machineName }}</span>
              <span>{{ project.services?.length ?? 0 }} services</span>
              <span class="text-muted">{{ formatDate(project.createdAt) }}</span>
            </div>
          </div>
          <div class="item-list-actions">
            <button
              class="btn btn-sm btn-primary"
              @click="handleUp(project.id)"
              :title="t('pages.orchestration.actions.up')"
            >
              <Play :size="14" />
            </button>
            <button
              class="btn btn-sm btn-warning"
              @click="handleDown(project.id)"
              :title="t('pages.orchestration.actions.down')"
            >
              <Square :size="14" />
            </button>
            <button
              class="btn btn-sm btn-ghost"
              @click="handleLogsClick(project.id)"
              :title="t('pages.orchestration.actions.logs')"
            >
              <ScrollText :size="14" />
            </button>
            <button
              class="btn btn-sm btn-ghost"
              @click="handleInspect(project.id)"
              :title="t('common.inspect')"
            >
              <Eye :size="14" />
            </button>
            <button
              class="btn btn-sm btn-ghost"
              @click="handleDelete(project.id)"
              :title="t('common.delete')"
            >
              <Trash2 :size="14" />
            </button>
          </div>
        </div>
      </div>
    </Transition></template>

    <!-- Detail panel -->
    <div v-if="selectedProject" class="detail-panel card">
      <div class="detail-header">
        <div class="flex items-center gap-3">
          <GitBranch :size="20" style="color: var(--color-accent)" />
          <h2 class="detail-title">{{ selectedProject.name }}</h2>
          <span :class="['badge', statusBadgeClass(selectedProject.status)]">
            {{ selectedProject.status }}
          </span>
          <span class="text-xs text-muted">
            {{ selectedProject.machineName }} / {{ selectedProject.projectName }}
          </span>
        </div>
        <div class="detail-actions">
          <button
            class="btn btn-success btn-sm"
            :disabled="operationLoading"
            @click="handleUpOnSelected"
          >
            <Play :size="14" />
            {{ t('pages.orchestration.actions.up') }}
          </button>
          <button
            class="btn btn-warning btn-sm"
            :disabled="operationLoading"
            @click="handleDownOnSelected"
          >
            <Square :size="14" />
            {{ t('pages.orchestration.actions.down') }}
          </button>
          <button
            class="btn btn-sm"
            :disabled="operationLoading"
            @click="handleBuildOnSelected"
          >
            {{ t('pages.orchestration.actions.build') }}
          </button>
          <button
            class="btn btn-sm"
            :disabled="operationLoading"
            @click="handleStartOnSelected"
          >
            {{ t('pages.orchestration.actions.start') }}
          </button>
          <button
            class="btn btn-sm"
            :disabled="operationLoading"
            @click="handleStopOnSelected"
          >
            {{ t('pages.orchestration.actions.stop') }}
          </button>
          <button
            class="btn btn-sm"
            :disabled="operationLoading"
            @click="handleRestartOnSelected"
          >
            {{ t('pages.orchestration.actions.restart') }}
          </button>
          <button class="btn btn-ghost btn-sm" @click="selectedProject = null">
            <X :size="14" />
            {{ t('common.close') }}
          </button>
        </div>
      </div>

      <div class="detail-meta">
        <span class="meta-item">
          <span class="meta-label">{{ t('pages.orchestration.inspect.filePath') }}</span>
          <span class="meta-value">{{ selectedProject.filePath || '-' }}</span>
        </span>
        <span class="meta-item">
          <span class="meta-label">{{ t('pages.orchestration.inspect.services') }}</span>
          <span class="meta-value">{{ selectedProject.services?.length ?? 0 }}</span>
        </span>
        <span class="meta-item">
          <span class="meta-label">{{ t('pages.orchestration.inspect.updatedAt') }}</span>
          <span class="meta-value">{{ formatDate(selectedProject.updatedAt) }}</span>
        </span>
      </div>

      <div class="detail-tabs">
        <button
          :class="['tab', { active: activeTab === 'services' }]"
          @click="activeTab = 'services'"
        >
          <Layers :size="14" />
          Services
        </button>
        <button
          :class="['tab', { active: activeTab === 'yaml' }]"
          @click="activeTab = 'yaml'"
        >
          <FileCode :size="14" />
          YAML
        </button>
        <button
          :class="['tab', { active: activeTab === 'logs' }]"
          @click="activeTab = 'logs'"
        >
          <ScrollText :size="14" />
          {{ t('pages.orchestration.actions.logs') }}
        </button>
      </div>

      <div class="detail-content">
        <ServiceStatusCard v-if="activeTab === 'services'" :services="selectedProject.services" />
        <ComposeEditor
          v-if="activeTab === 'yaml'"
          v-model="yamlContent"
          @save="handleSaveYaml"
        />
        <LogViewer v-if="activeTab === 'logs'" :logs="logs" />
      </div>
    </div>

    <!-- Create modal -->
    <ComposeCreateModal
      v-if="showCreateModal && machines.length > 0"
      :show="showCreateModal"
      :machine-id="machines[0]?.id ?? ''"
      :templates="templates"
      @close="showCreateModal = false"
      @created="onProjectCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, Plus, GitBranch, Play, Square, X, Layers, FileCode, ScrollText, Eye, Trash2 } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import { composeService } from '@/services/composeService'
import { formatDate } from '@/utils/format'
import type { ComposeProject, ComposeTemplate, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import ComposeCard from '@/components/compose/ComposeCard.vue'
import ComposeCreateModal from '@/components/compose/ComposeCreateModal.vue'
import ServiceStatusCard from '@/components/compose/ServiceStatusCard.vue'
import ComposeEditor from '@/components/compose/ComposeEditor.vue'
import LogViewer from '@/components/compose/LogViewer.vue'
import ViewToggle from '@/components/ui/ViewToggle.vue'
import { useViewMode } from '@/composables/useViewMode'

interface ProjectWithMachine extends ComposeProject {
  machineName: string
}

const viewMode = useViewMode('orchestration')
const projects = ref<ProjectWithMachine[]>([])
const machines = ref<RemoteMachine[]>([])
const loading = ref(true)
const error = ref('')
const searchQuery = ref('')
const statusFilter = ref('')
const selectedProject = ref<ProjectWithMachine | null>(null)
const showCreateModal = ref(false)
const activeTab = ref<'services' | 'yaml' | 'logs'>('services')
const yamlContent = ref('')
const logs = ref('')
const operationLoading = ref(false)

const templates: ComposeTemplate[] = [
  {
    name: t('pages.orchestration.templates.wordpress'),
    description: 'WordPress with MySQL database',
    category: 'CMS',
    content: `version: '3.8'

services:
  wordpress:
    image: wordpress:latest
    ports:
      - "8080:80"
    environment:
      WORDPRESS_DB_HOST: db
      WORDPRESS_DB_USER: wordpress
      WORDPRESS_DB_PASSWORD: wordpress
      WORDPRESS_DB_NAME: wordpress
    depends_on:
      - db

  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: wordpress
      MYSQL_USER: wordpress
      MYSQL_PASSWORD: wordpress
`,
  },
  {
    name: t('pages.orchestration.templates.nginxPhp'),
    description: 'Nginx reverse proxy with PHP-FPM',
    category: 'Web',
    content: `version: '3.8'

services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
    volumes:
      - ./html:/usr/share/nginx/html
    depends_on:
      - php

  php:
    image: php:fpm-alpine
    volumes:
      - ./html:/var/www/html
`,
  },
  {
    name: t('pages.orchestration.templates.nodeRedis'),
    description: 'Node.js application with Redis caching',
    category: 'Web',
    content: `version: '3.8'

services:
  app:
    image: node:20-alpine
    ports:
      - "3000:3000"
    environment:
      REDIS_HOST: redis
    depends_on:
      - redis

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
`,
  },
  {
    name: t('pages.orchestration.templates.postgres'),
    description: 'PostgreSQL database with pgAdmin',
    category: 'Database',
    content: `version: '3.8'

services:
  db:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: app
    volumes:
      - pgdata:/var/lib/postgresql/data

  admin:
    image: dpage/pgadmin4
    ports:
      - "5050:80"
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@litedock.io
      PGADMIN_DEFAULT_PASSWORD: admin

volumes:
  pgdata:
`,
  },
  {
    name: t('pages.orchestration.templates.lamp'),
    description: 'Linux Apache MySQL PHP stack',
    category: 'Web',
    content: `version: '3.8'

services:
  web:
    image: php:8.2-apache
    ports:
      - "8080:80"
    volumes:
      - ./app:/var/www/html
    depends_on:
      - db

  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: app
      MYSQL_USER: appuser
      MYSQL_PASSWORD: apppassword
`,
  },
]

// Computed

const filteredProjects = computed(() => {
  let result = projects.value

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.projectName.toLowerCase().includes(q) ||
        (p.machineName && p.machineName.toLowerCase().includes(q))
    )
  }

  if (statusFilter.value) {
    result = result.filter((p) => p.status === statusFilter.value)
  }

  return result
})

function statusBadgeClass(status: string): string {
  switch (status) {
    case 'running':
      return 'badge-success'
    case 'paused':
      return 'badge-warning'
    case 'stopped':
    case 'failed':
    case 'exited':
      return 'badge-error'
    case 'partial':
      return 'badge-info'
    default:
      return 'badge'
  }
}

// Data fetching

async function fetchProjects() {
  loading.value = true
  error.value = ''
  try {
    const allMachines = await remoteMachineService.list()
    machines.value = allMachines

    const results = await Promise.allSettled(
      allMachines.map((m) =>
        composeService.listProjects(m.id).then((projs) =>
          projs.map((p) => ({
            ...p,
            machineId: m.id,
            machineName: m.name,
          }))
        )
      )
    )

    projects.value = results.flatMap(
      (r) => (r.status === 'fulfilled' ? r.value : [])
    )
  } catch (e: unknown) {
    const msg =
      e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  } finally {
    loading.value = false
  }
}

function refreshProjects() {
  selectedProject.value = null
  fetchProjects()
}

function findProjectById(id: string): ProjectWithMachine | undefined {
  return projects.value.find((p) => p.id === id)
}

// Card event handlers

function handleInspect(projectId: string) {
  const project = findProjectById(projectId)
  if (!project) return
  selectProject(project)
}

function handleLogsClick(projectId: string) {
  const project = findProjectById(projectId)
  if (!project) return
  selectProject(project, 'logs')
}

async function handleDelete(projectId: string) {
  const project = findProjectById(projectId)
  if (!project) return

  if (!confirm(t('pages.orchestration.confirm.delete'))) return

  try {
    await composeService.deleteProject(project.machineId, project.projectName)
    if (selectedProject.value?.id === project.id) {
      selectedProject.value = null
    }
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(msg)
  }
}

async function handleUp(projectId: string) {
  const project = findProjectById(projectId)
  if (!project) return

  operationLoading.value = true
  try {
    await composeService.up(project.machineId, project.projectName)
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(t('pages.orchestration.actions.up') + ' failed: ' + msg)
  } finally {
    operationLoading.value = false
  }
}

async function handleDown(projectId: string) {
  const project = findProjectById(projectId)
  if (!project) return

  if (!confirm(t('pages.orchestration.confirm.down'))) return

  const removeVolumes = confirm(t('pages.orchestration.confirm.downVolumes'))

  operationLoading.value = true
  try {
    await composeService.down(project.machineId, project.projectName, removeVolumes)
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(t('pages.orchestration.actions.down') + ' failed: ' + msg)
  } finally {
    operationLoading.value = false
  }
}

// Detail panel actions

async function handleUpOnSelected() {
  if (!selectedProject.value) return
  await handleUp(selectedProject.value.id)
}

async function handleDownOnSelected() {
  if (!selectedProject.value) return
  await handleDown(selectedProject.value.id)
}

async function handleBuildOnSelected() {
  if (!selectedProject.value) return

  operationLoading.value = true
  try {
    await composeService.build(selectedProject.value.machineId, selectedProject.value.projectName)
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(t('pages.orchestration.actions.build') + ' failed: ' + msg)
  } finally {
    operationLoading.value = false
  }
}

async function handleStartOnSelected() {
  if (!selectedProject.value) return

  operationLoading.value = true
  try {
    await composeService.start(selectedProject.value.machineId, selectedProject.value.projectName)
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(t('pages.orchestration.actions.start') + ' failed: ' + msg)
  } finally {
    operationLoading.value = false
  }
}

async function handleStopOnSelected() {
  if (!selectedProject.value) return

  operationLoading.value = true
  try {
    await composeService.stop(selectedProject.value.machineId, selectedProject.value.projectName)
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(t('pages.orchestration.actions.stop') + ' failed: ' + msg)
  } finally {
    operationLoading.value = false
  }
}

async function handleRestartOnSelected() {
  if (!selectedProject.value) return

  operationLoading.value = true
  try {
    await composeService.restart(selectedProject.value.machineId, selectedProject.value.projectName)
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(t('pages.orchestration.actions.restart') + ' failed: ' + msg)
  } finally {
    operationLoading.value = false
  }
}

// Project selection and tabs

async function selectProject(project: ProjectWithMachine, tab: 'services' | 'yaml' | 'logs' = 'services') {
  selectedProject.value = project
  activeTab.value = tab

  // Fetch logs for the project
  logs.value = ''
  composeService
    .getLogs(project.machineId, project.projectName)
    .then((result) => {
      logs.value = result || ''
    })
    .catch(() => {
      logs.value = 'Failed to load logs.'
    })

  // Try to load YAML content from the project
  yamlContent.value = ''
  if (project.filePath) {
    try {
      const detailed = await composeService.getProject(project.machineId, project.projectName)
      // The backend may include raw content in the response
      const raw = detailed as unknown as Record<string, unknown>
      if (typeof raw.content === 'string') {
        yamlContent.value = raw.content
      } else if (typeof raw.yaml === 'string') {
        yamlContent.value = raw.yaml
      } else if (typeof raw.config === 'string') {
        yamlContent.value = raw.config
      }
    } catch {
      // Could not fetch YAML content — user can paste manually
    }
  }
}

// YAML saving

async function handleSaveYaml(content: string) {
  if (!selectedProject.value) return

  try {
    await composeService.updateProject(
      selectedProject.value.machineId,
      selectedProject.value.projectName,
      content
    )
    alert(t('pages.orchestration.editor.saveSuccess'))
    await fetchProjects()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    alert(t('pages.orchestration.editor.saveError') + ': ' + msg)
  }
}

// Create modal

function onProjectCreated() {
  showCreateModal.value = false
  fetchProjects()
}

// Lifecycle

onMounted(() => fetchProjects())
</script>

<style scoped>
.orchestration-page {
  max-width: 1400px;
  margin: 0 auto;
}

.filters {
  display: flex;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
  padding: var(--space-4);
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
}

.search-box {
  flex: 1;
}

.filter-options {
  display: flex;
  gap: var(--space-3);
}

.filter-options select {
  min-width: 130px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-16) 0;
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
}

.error-state {
  padding: var(--space-10) var(--space-6);
}

.empty-state {
  padding: var(--space-10) var(--space-6);
}

/* --- Detail Panel --- */

.detail-panel {
  margin-top: var(--space-6);
  padding: var(--space-6);
  border-color: var(--color-accent);
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.detail-title {
  margin: 0;
  font-size: var(--font-size-xl);
  color: var(--color-text-strong);
}

.detail-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.detail-meta {
  display: flex;
  gap: var(--space-6);
  flex-wrap: wrap;
  padding: var(--space-3) 0;
  margin-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border-weak);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.meta-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-weak);
}

.meta-value {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* --- Tabs --- */

.detail-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--color-border-weak);
  margin-bottom: var(--space-4);
}

.tab {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-weak);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab:hover {
  color: var(--color-text);
}

.tab.active {
  color: var(--color-accent);
  border-bottom-color: var(--color-accent);
}

.detail-content {
  min-height: 200px;
}

/* --- Responsive --- */

@media (max-width: 768px) {
  .card-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .filters {
    flex-direction: column;
  }

  .filter-options {
    flex-direction: column;
  }

  .filter-options select {
    min-width: 100%;
  }

  .detail-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .detail-actions {
    width: 100%;
    overflow-x: auto;
  }

  .detail-meta {
    gap: var(--space-3);
  }
}
</style>
