# Remote Machines Frontend - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add frontend interfaces for remote SSH machine management — list/add/edit/delete machines, view containers on remote machines, view logs, exec commands, and container lifecycle controls.

**Architecture:** Follow existing Vue3 + TypeScript patterns. Machine list page (like Containers.vue) + Machine detail page (containers list + logs/exec). All API calls through new `remoteMachineService.ts`. Full i18n support.

**Tech Stack:** Vue 3 `<script setup>`, TypeScript, axios, lucide-vue-next icons, existing CSS design tokens (IBM Plex Mono, monospace aesthetic).

---

## Design Decisions

- **MachineDetail as separate route** (`/machines/:id`) — cleaner separation, browser back works naturally
- **Logs displayed in expandable panel** — below container list, no modal needed
- **Exec command inline** — textarea + run button in logs panel
- **Container lifecycle** — start/stop/restart/remove buttons on each container card (same pattern as Containers.vue)

---

## Task 1: Add TypeScript Types

**Files:**
- Modify: `web/src/types/index.ts`

**Step 1: Add RemoteMachine and Container types**

Append to `web/src/types/index.ts`:

```typescript
// Remote Machine types
export type AuthMethod = 'password' | 'key'

export interface RemoteMachine {
  id: string
  name: string
  host: string
  port: number
  username: string
  auth_method: AuthMethod
  ssh_key_path?: string
  docker_host: string
  status: 'online' | 'offline' | 'unknown'
  created_at: string
  updated_at: string
}

export interface CreateMachineRequest {
  name: string
  host: string
  port?: number
  username: string
  auth_method: AuthMethod
  password?: string
  ssh_key?: string
  ssh_key_path?: string
  docker_host?: string
}

export interface UpdateMachineRequest {
  name?: string
  host?: string
  port?: number
  username?: string
  auth_method?: AuthMethod
  password?: string
  ssh_key?: string
  ssh_key_path?: string
  docker_host?: string
}

// Remote Container type (matches entity.Container from backend)
export interface RemoteContainer {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused' | 'restarting' | 'exited' | 'created'
  ports: string[]
  createdAt: string
  startedAt?: string
  labels?: Record<string, string>
  mounts?: Array<{
    type: string
    source: string
    destination: string
  }>
}
```

**Step 2: Verify no TypeScript errors**

Run: `cd web && npm run type-check` (or `npx vue-tsc --noEmit`)
Expected: No errors related to new types

**QA:** Confirm `RemoteMachine`, `CreateMachineRequest`, `RemoteContainer` are exported and usable.

---

## Task 2: Create API Service

**Files:**
- Create: `web/src/services/remoteMachineService.ts`

**Step 1: Write the service**

```typescript
import api from '@/utils/api'
import type {
  RemoteMachine,
  CreateMachineRequest,
  UpdateMachineRequest,
  RemoteContainer
} from '@/types'

export const remoteMachineService = {
  // Machine CRUD
  list(): Promise<RemoteMachine[]> {
    return api.get('/machines').then(r => r.data.data ?? [])
  },

  get(id: string): Promise<RemoteMachine> {
    return api.get(`/machines/${id}`).then(r => r.data.data)
  },

  create(data: CreateMachineRequest): Promise<RemoteMachine> {
    return api.post('/machines', data).then(r => r.data.data)
  },

  update(id: string, data: UpdateMachineRequest): Promise<RemoteMachine> {
    return api.put(`/machines/${id}`, data).then(r => r.data.data)
  },

  delete(id: string): Promise<void> {
    return api.delete(`/machines/${id}`).then(() => {})
  },

  testConnection(id: string): Promise<void> {
    return api.post(`/machines/${id}/test`).then(() => {})
  },

  // Container operations
  listContainers(machineId: string): Promise<RemoteContainer[]> {
    return api.get(`/machines/${machineId}/containers`).then(r => r.data.containers ?? [])
  },

  getContainerLogs(machineId: string, containerId: string, tail = '100'): Promise<string> {
    return api.get(`/machines/${machineId}/containers/${containerId}/logs`, {
      params: { tail }
    }).then(r => r.data.logs ?? '')
  },

  execContainer(machineId: string, containerId: string, cmd: string[]): Promise<string> {
    return api.post(`/machines/${machineId}/containers/${containerId}/exec`, { cmd })
      .then(r => r.data.output ?? '')
  },

  startContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/start`).then(() => {})
  },

  stopContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/stop`).then(() => {})
  },

  restartContainer(machineId: string, containerId: string): Promise<void> {
    return api.post(`/machines/${machineId}/containers/${containerId}/restart`).then(() => {})
  },

  removeContainer(machineId: string, containerId: string, force = false): Promise<void> {
    return api.delete(`/machines/${machineId}/containers/${containerId}`, { params: { force } })
      .then(() => {})
  },

  inspectContainer(machineId: string, containerId: string): Promise<any> {
    return api.get(`/machines/${machineId}/containers/${containerId}`)
      .then(r => r.data.container)
  }
}
```

**QA:** File created at correct path, exports `remoteMachineService`.

---

## Task 3: Create Machines.vue (List Page)

**Files:**
- Create: `web/src/views/Machines.vue`

**Step 1: Write the component**

Follow the exact pattern of `Containers.vue` — page header with refresh + add buttons, search/filter bar, grid of machine cards, modal for add/edit form.

Key features:
- List all machines in card grid (like Containers.vue)
- Each card shows: name, host:port, username, auth_method badge, status badge, docker_host
- Actions: Edit (pencil), Delete (trash), Test Connection (wifi), View Containers (box)
- "Add Machine" button opens modal with form
- Form fields: name, host, port (default 22), username, auth_method toggle (password/key), password field OR ssh_key textarea, docker_host (default /var/run/docker.sock)
- When auth_method='key', show ssh_key textarea; when 'password', show password field
- Edit form pre-populated with existing values
- Test Connection shows success/error alert

```vue
<template>
  <div class="machines-page">
    <div class="page-header">
      <h1>{{ t('machines.title') }}</h1>
      <div class="header-actions">
        <button @click="refreshMachines" class="btn btn-secondary" :disabled="loading">
          <RefreshCw :size="16" :class="{ 'spinning': loading }" />
          {{ t('machines.refresh') }}
        </button>
        <button @click="openAddModal()" class="btn btn-primary">
          <Plus :size="16" />
          {{ t('machines.addMachine') }}
        </button>
      </div>
    </div>

    <!-- Search -->
    <div class="filters">
      <div class="search-box">
        <input v-model="searchQuery" :placeholder="t('machines.searchPlaceholder')" type="text" class="input" />
      </div>
    </div>

    <!-- Machine Grid -->
    <div v-if="filteredMachines.length > 0" class="machines-grid">
      <div v-for="machine in filteredMachines" :key="machine.id" class="machine-card">
        <div class="machine-header">
          <div class="machine-name">{{ machine.name }}</div>
          <div class="badge" :class="getStatusClass(machine.status)">{{ getStatusText(machine.status) }}</div>
        </div>
        <div class="machine-info">
          <div class="info-item"><span class="label">{{ t('machines.host') }}</span><span class="value">{{ machine.host }}:{{ machine.port }}</span></div>
          <div class="info-item"><span class="label">{{ t('machines.username') }}</span><span class="value">{{ machine.username }}</span></div>
          <div class="info-item"><span class="label">{{ t('machines.auth') }}</span><span class="value">{{ machine.auth_method === 'password' ? t('machines.password') : t('machines.sshKey') }}</span></div>
          <div class="info-item"><span class="label">{{ t('machines.dockerHost') }}</span><span class="value">{{ machine.docker_host }}</span></div>
        </div>
        <div class="machine-actions">
          <button @click="testConnection(machine.id)" class="btn btn-sm btn-secondary">
            <Wifi :size="14" /> {{ t('machines.test') }}
          </button>
          <button @click="openEditModal(machine)" class="btn btn-sm btn-secondary">
            <Pencil :size="14" /> {{ t('machines.edit') }}
          </button>
          <button @click="viewContainers(machine.id)" class="btn btn-sm btn-secondary">
            <Box :size="14" /> {{ t('machines.containers') }}
          </button>
          <button @click="deleteMachine(machine.id)" class="btn btn-sm btn-ghost btn-danger-text">
            <Trash2 :size="14" /> {{ t('machines.delete') }}
          </button>
        </div>
      </div>
    </div>
    <div v-else-if="!loading" class="empty-state">
      <ServerIcon :size="48" :stroke-width="1" />
      <p>{{ t('machines.empty') }}</p>
      <button @click="openAddModal()" class="btn btn-primary">{{ t('machines.addFirst') }}</button>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ editingMachine ? t('machines.editMachine') : t('machines.addMachine') }}</h2>
          <button @click="closeModal" class="btn btn-ghost"><X :size="16" /></button>
        </div>
        <form @submit.prevent="saveMachine" class="modal-body">
          <div class="form-group"><label>{{ t('machines.name') }}</label><input v-model="form.name" type="text" class="input" required /></div>
          <div class="form-row">
            <div class="form-group"><label>{{ t('machines.host') }}</label><input v-model="form.host" type="text" class="input" required /></div>
            <div class="form-group" style="width:100px"><label>{{ t('machines.port') }}</label><input v-model.number="form.port" type="number" class="input" min="1" max="65535" /></div>
          </div>
          <div class="form-group"><label>{{ t('machines.username') }}</label><input v-model="form.username" type="text" class="input" required /></div>
          <div class="form-group">
            <label>{{ t('machines.authMethod') }}</label>
            <select v-model="form.auth_method" class="input">
              <option value="password">{{ t('machines.password') }}</option>
              <option value="key">{{ t('machines.sshKey') }}</option>
            </select>
          </div>
          <div v-if="form.auth_method === 'password'" class="form-group">
            <label>{{ t('machines.password') }}</label>
            <input v-model="form.password" type="password" class="input" :placeholder="editingMachine ? t('machines.leaveBlank') : ''" />
          </div>
          <div v-else class="form-group">
            <label>{{ t('machines.sshKey') }}</label>
            <textarea v-model="form.ssh_key" class="input" rows="5" :placeholder="t('machines.sshKeyPlaceholder')"></textarea>
          </div>
          <div class="form-group"><label>{{ t('machines.dockerHost') }}</label><input v-model="form.docker_host" type="text" class="input" /></div>
          <div v-if="formError" class="form-error">{{ formError }}</div>
          <div class="modal-actions">
            <button type="button" @click="closeModal" class="btn btn-secondary">{{ t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('machines.saving') : t('common.confirm') }}</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Test Result Toast -->
    <div v-if="testResult" class="toast" :class="testResult.type">
      {{ testResult.message }}
      <button @click="testResult = null" class="toast-close"><X :size="14" /></button>
    </div>
  </div>
</template>
```

See full script and style sections below. Follow `Containers.vue` CSS patterns exactly.

**Step 2: Write script section**

```typescript
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { RefreshCw, Plus, Wifi, Pencil, Box, Trash2, X, ServerIcon } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import type { RemoteMachine, CreateMachineRequest } from '@/types'

const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const searchQuery = ref('')
const showModal = ref(false)
const editingMachine = ref<RemoteMachine | null>(null)
const testResult = ref<{ type: 'success' | 'error', message: string } | null>(null)
const formError = ref('')

const defaultForm = (): CreateMachineRequest & { port: number; password: string; ssh_key: string } => ({
  name: '', host: '', port: 22, username: '',
  auth_method: 'password', password: '', ssh_key: '', docker_host: '/var/run/docker.sock'
})

const form = ref(defaultForm())
const machines = ref<RemoteMachine[]>([])

const filteredMachines = computed(() => {
  if (!searchQuery.value) return machines.value
  const q = searchQuery.value.toLowerCase()
  return machines.value.filter(m =>
    m.name.toLowerCase().includes(q) || m.host.toLowerCase().includes(q)
  )
})

const getStatusClass = (status: string) => ({
  'badge-success': status === 'online',
  'badge-error': status === 'offline',
  'badge-warning': status === 'unknown'
})

const getStatusText = (status: string) => t(`machines.status.${status}`)

const refreshMachines = async () => {
  loading.value = true
  try {
    machines.value = await remoteMachineService.list()
  } finally {
    loading.value = false
  }
}

const openAddModal = () => {
  editingMachine.value = null
  form.value = defaultForm()
  formError.value = ''
  showModal.value = true
}

const openEditModal = (machine: RemoteMachine) => {
  editingMachine.value = machine
  form.value = {
    name: machine.name, host: machine.host, port: machine.port,
    username: machine.username, auth_method: machine.auth_method,
    password: '', ssh_key: '', docker_host: machine.docker_host
  }
  formError.value = ''
  showModal.value = true
}

const closeModal = () => { showModal.value = false }

const saveMachine = async () => {
  formError.value = ''
  saving.value = true
  try {
    const data: CreateMachineRequest = {
      name: form.value.name, host: form.value.host, port: form.value.port,
      username: form.value.username, auth_method: form.value.auth_method,
      docker_host: form.value.docker_host || '/var/run/docker.sock'
    }
    if (form.value.auth_method === 'password' && form.value.password) {
      data.password = form.value.password
    }
    if (form.value.auth_method === 'key' && form.value.ssh_key) {
      data.ssh_key = form.value.ssh_key
    }
    if (editingMachine.value) {
      await remoteMachineService.update(editingMachine.value.id, data)
    } else {
      await remoteMachineService.create(data)
    }
    closeModal()
    await refreshMachines()
  } catch (e: any) {
    formError.value = e.response?.data?.message || e.message || 'Failed'
  } finally {
    saving.value = false
  }
}

const testConnection = async (id: string) => {
  try {
    await remoteMachineService.testConnection(id)
    testResult.value = { type: 'success', message: t('machines.testSuccess') }
  } catch (e: any) {
    testResult.value = { type: 'error', message: e.response?.data?.message || t('machines.testFailed') }
  }
  setTimeout(() => { testResult.value = null }, 4000)
}

const viewContainers = (id: string) => {
  router.push(`/machines/${id}`)
}

const deleteMachine = async (id: string) => {
  if (!confirm(t('machines.confirmDelete'))) return
  await remoteMachineService.delete(id)
  await refreshMachines()
}

onMounted(() => refreshMachines())
</script>
```

**Step 3: Write style section** (follow Containers.vue patterns, replace "container" with "machine")

Key CSS differences from Containers.vue:
- `.machine-card` instead of `.container-card`
- `.machine-header`, `.machine-info`, `.machine-actions` 
- Modal styles: `.modal-overlay`, `.modal`, `.modal-header`, `.modal-body`, `.modal-actions`
- Toast: `.toast`, `.toast-success`, `.toast-error`, `.toast-close`
- Form: `.form-group`, `.form-row`, `.form-error`
- Empty state: `.empty-state`

**QA:** Page loads, machines list displays, add modal opens, form validation works.

---

## Task 4: Create MachineDetail.vue (Container Management Page)

**Files:**
- Create: `web/src/views/MachineDetail.vue`

**Step 1: Write the component**

This page shows:
- Machine info header (name, host, status, back button)
- Container list grid (same style as Containers.vue)
- Logs panel (below containers or in expandable section)
- Exec command section within logs panel

```typescript
// Script key logic:
const machineId = route.params.id as string
const [machine, setMachine] = useState<RemoteMachine | null>(null)
const [containers, setContainers] = useState<RemoteContainer[]>([])
const [loading, setLoading] = useState(false)
const [selectedContainer, setSelectedContainer] = useState<string | null>(null)
const [logs, setLogs] = useState('')
const [logsLoading, setLogsLoading] = useState(false)
const [execCmd, setExecCmd] = useState('ls -la')
const [execOutput, setExecOutput] = useState('')
const [execLoading, setExecLoading] = useState(false)
const [searchQuery, setSearchQuery] = useState('')
const [statusFilter, setStatusFilter] = useState('')
```

Container card actions: start, stop, restart, logs (fetches + shows in panel), remove

Logs panel: shows `logs` ref content in `<pre>` block, fetch on "Show Logs" click, auto-refresh toggle

Exec section: single-line input for command, "Run" button, output in `<pre>` block

**Step 2: Write template and styles**

Follow `Containers.vue` layout patterns + `Settings.vue` for panel layout.

**QA:** Navigating to `/machines/:id` loads machine + containers. Logs fetch works. Exec works.

---

## Task 5: Add Routes

**Files:**
- Modify: `web/src/router/index.ts`

**Step 1: Add Machines and MachineDetail routes**

In the `children` array under the main layout route, add:

```typescript
// Remote Machines
{
  path: 'machines',
  name: 'Machines',
  component: () => import('@/views/Machines.vue')
},
{
  path: 'machines/:id',
  name: 'MachineDetail',
  component: () => import('@/views/MachineDetail.vue')
},
```

Place before the `settings` route.

**QA:** `/machines` and `/machines/:id` routes exist and load correct components.

---

## Task 6: Add Navigation Item

**Files:**
- Modify: `web/src/components/nav/SidebarNav.vue`

**Step 1: Add Server icon import and nav item**

Add `Server` (from lucide-vue-next) to imports:

```typescript
import {
  LayoutDashboard,
  Box,
  GitBranch,
  Image,
  Network,
  HardDrive,
  Settings,
  Server,  // ADD THIS
} from 'lucide-vue-next'
```

Add to `mainNavItems`:

```typescript
{ name: 'Machines', path: '/machines', label: t('nav.machines'), icon: Server },
```

Place after `Containers` or at the end of mainNavItems.

**QA:** "Remote Machines" appears in sidebar navigation with Server icon.

---

## Task 7: Add i18n Keys (Chinese)

**Files:**
- Modify: `web/src/i18n/locales/zh-CN.json`

**Step 1: Add machines section**

Add after `"containers": { ... }` section:

```json
"machines": {
  "title": "远程机器",
  "searchPlaceholder": "搜索远程机器...",
  "addMachine": "添加机器",
  "editMachine": "编辑机器",
  "empty": "还没有配置任何远程机器",
  "addFirst": "添加第一台机器",
  "refresh": "刷新",
  "host": "主机地址:",
  "port": "端口",
  "username": "用户名:",
  "authMethod": "认证方式",
  "password": "密码",
  "sshKey": "SSH 密钥",
  "sshKeyPlaceholder": "粘贴 SSH 私钥内容...",
  "dockerHost": "Docker Socket:",
  "leaveBlank": "留空以保持不变",
  "test": "测试连接",
  "testSuccess": "连接成功",
  "testFailed": "连接失败",
  "edit": "编辑",
  "delete": "删除",
  "containers": "容器",
  "logs": "日志",
  "exec": "执行命令",
  "confirmDelete": "确定要删除这台远程机器吗？",
  "saving": "保存中...",
  "status": {
    "online": "在线",
    "offline": "离线",
    "unknown": "未知"
  }
}
```

Also add to nav section:

```json
"nav": {
  "machines": "远程机器",
  ...
}
```

**QA:** Chinese labels appear correctly on machines page.

---

## Task 8: Add i18n Keys (English)

**Files:**
- Modify: `web/src/i18n/locales/en.json`

**Step 1: Add machines section**

Mirror the zh-CN structure with English values.

**QA:** English labels appear correctly on machines page.

---

## Task 9: Build Verification

**Step 1: Run frontend build**

```bash
cd web && npm install && npm run build
```

Expected: Exit code 0, no TypeScript errors, production build in `dist/`

**Step 2: Run type check**

```bash
cd web && npm run type-check
```

Expected: No type errors

---

## Dependency Graph

```
Task 1 (types) ─────────────────────────────────────┐
                                                     ├──► Task 3 (Machines.vue)
Task 2 (service) ───────────────────────────────────┤
                                                     ├──► Task 4 (MachineDetail.vue)
Task 5 (router) ────────────────────────────────────┤
Task 6 (nav) ────────────────────────────────────────┤
Task 7 (i18n) ──────────────────────────────────────┘
Task 9 (build)
```

**Parallel execution:** Tasks 1, 2, 5, 6, 7, 8 are fully independent — can run in parallel.
Tasks 3 and 4 depend on 1, 2.
Task 9 depends on all others.

---

## Atomic Commit Strategy

| # | Message | Tasks |
|---|---------|-------|
| 1 | `feat(frontend): add RemoteMachine TypeScript types` | Task 1 |
| 2 | `feat(frontend): add remoteMachineService API client` | Task 2 |
| 3 | `feat(frontend): add Machines list page with CRUD` | Task 3 |
| 4 | `feat(frontend): add MachineDetail page with containers, logs, exec` | Task 4 |
| 5 | `feat(frontend): add /machines routes to router` | Task 5 |
| 6 | `feat(frontend): add Remote Machines nav item` | Task 6 |
| 7 | `feat(i18n): add machines page translations (zh-CN + en)` | Tasks 7+8 |
| 8 | `chore(frontend): verify build passes` | Task 9 |

---

## Open Questions for User

1. **Container lifecycle** — Start/Stop/Restart/Remove buttons on container cards (matching Containers.vue) — included in this plan. OK?

2. **Logs auto-refresh** — Should logs panel auto-refresh every N seconds when open? Or manual refresh only? (Plan uses manual refresh.)

3. **Auth method UI** — Plan uses password field OR SSH key textarea based on auth_method toggle. OK?

Please confirm these choices, then I'll begin implementation.
