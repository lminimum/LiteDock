<template>
  <div class="ai-layout">
    <!-- Compact top bar -->
    <header class="ai-topbar">
      <Bot :size="18" :stroke-width="1.5" />
      <span class="ai-topbar-title">{{ t('assistant.title') }}</span>
    </header>

    <!-- Body: sidebar + main -->
    <div class="ai-body">
      <!-- Sidebar -->
      <aside class="ai-sidebar">
        <button class="new-chat-btn" @click="newChat">
          <Plus :size="14" />
          <span>New Chat</span>
        </button>

        <div v-if="conversations.length === 0" class="sidebar-empty">
          <span class="text-xs text-muted">{{ t('ai.emptyPrompt') }}</span>
        </div>

        <div v-else class="history-list">
          <div
            v-for="conv in conversations"
            :key="conv.id"
            class="history-item"
            :class="{ active: conv.id === activeConversationId }"
            @click="switchConversation(conv.id)"
          >
            <MessageSquare :size="14" class="history-icon" />
            <div class="history-content">
              <span class="history-title">{{ conv.title }}</span>
              <span class="history-time">{{ formatTime(conv.updatedAt) }}</span>
            </div>
            <button
              v-if="conv.id === activeConversationId"
              class="delete-btn"
              @click.stop="deleteConversation(conv.id)"
            >
              <Trash2 :size="12" />
            </button>
          </div>
        </div>
      </aside>

      <!-- Main area -->
      <main class="ai-main">
        <!-- Messages -->
        <div ref="messagesRef" class="ai-messages">
          <Transition name="msg-fade" mode="out-in">
            <div :key="activeConversationId" class="msg-inner">
              <div v-if="currentMessages.length === 0 && !loading" class="empty-state">
                <Bot :size="40" :stroke-width="1" class="empty-icon" />
                <p>{{ t('ai.emptyPrompt') }}</p>
              </div>
              <template v-for="(msg, i) in currentMessages" :key="i">
                <div v-if="msg.role === 'user'" class="msg msg-user">
                  <span>{{ msg.text }}</span>
                </div>
                <div v-else class="msg msg-assistant">
                  <p>{{ msg.text }}</p>
                  <div v-if="msg.status === 'executing'" class="msg-status msg-status-executing">
                    <Loader2 :size="12" class="spin" />
                    <span>{{ t('assistant.status.executing') }}</span>
                  </div>
                  <div v-else-if="msg.status === 'completed'" class="msg-status msg-status-completed">
                    <CheckCircle2 :size="12" />
                    <span>{{ t('assistant.status.completed') }}</span>
                  </div>
                  <div v-else-if="msg.status === 'failed'" class="msg-status msg-status-failed">
                    <XCircle :size="12" />
                    <span>{{ t('assistant.status.failed') }}</span>
                  </div>
                  <div v-else-if="msg.status === 'requires_confirmation'" class="msg-status msg-status-warning">
                    <AlertTriangle :size="12" />
                    <span>Awaiting confirmation</span>
                  </div>
                </div>
              </template>
            </div>
          </Transition>
        </div>

        <!-- Bottom: quick chips + input -->
        <div class="ai-bottom">
          <div class="quick-chips">
            <button
              v-for="action in quickActions"
              :key="action.id"
              class="chip"
              :disabled="loading"
              @click="runQuickAction(action)"
            >
              <component :is="action.icon" :size="12" />
              <span>{{ action.label }}</span>
            </button>
          </div>
          <div class="ai-input-row">
            <input
              ref="inputRef"
              v-model="inputText"
              class="ai-input"
              :placeholder="t('assistant.input.placeholder')"
              :disabled="loading"
              maxlength="500"
              @keyup.enter="sendMessage"
            />
            <button
              v-if="loading"
              class="btn btn-secondary send-btn"
              aria-label="Stop"
              @click="stopStreaming"
            >
              <Square :size="16" />
            </button>
            <button
              v-else
              class="btn btn-primary send-btn"
              :disabled="!inputText.trim()"
              :aria-label="t('assistant.input.send')"
              @click="sendMessage"
            >
              <Send :size="16" />
            </button>
          </div>
        </div>
      </main>
    </div>
    <!-- Confirmation modal overlay -->
    <Teleport to="body">
      <div v-if="showConfirmModal" class="confirm-overlay" @click.self="cancelAction">
        <div class="confirm-modal card">
          <div class="confirm-header">
            <AlertTriangle :size="20" class="confirm-icon" />
            <span class="confirm-title">{{ t('assistant.confirmation.title') }}</span>
          </div>
          <div class="confirm-body">
            <p class="confirm-message">{{ pendingConfirmMessage }}</p>
          </div>
          <div class="confirm-footer">
            <button class="btn btn-ghost" :disabled="executingConfirm" @click="cancelAction">
              {{ t('assistant.confirmation.cancel') }}
            </button>
            <button class="btn btn-danger" :disabled="executingConfirm" @click="confirmAction">
              <Loader2 v-if="executingConfirm" :size="14" class="spin" />
              <span>{{ executingConfirm ? t('assistant.confirmation.waiting') : t('assistant.confirmation.confirm') }}</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onUnmounted, onMounted, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Bot, Play, FileText, Activity, Globe, HardDrive, Network,
  Send, Loader2, Plus, MessageSquare, Trash2, AlertTriangle,
  CheckCircle2, XCircle, Square,
} from 'lucide-vue-next'
import api from '@/utils/api'
import { useWebSocket } from '@/composables/useWebSocket'
import { stripShellChars } from '@/utils/sanitize'
import type { Component } from 'vue'

interface ChatMessage {
  role: 'user' | 'assistant'
  text: string
  status?: 'sending' | 'executing' | 'completed' | 'failed' | 'requires_confirmation'
  confirmationMessage?: string
  actionName?: string
  actionParams?: Record<string, string>
}

interface Conversation {
  id: string
  title: string
  messages: ChatMessage[]
  createdAt: number
  updatedAt: number
}

interface QuickAction {
  id: string
  icon: Component
  label: string
  prompt: string
}

const STORAGE_KEY = 'litdock-ai-conversations'
const DEFAULT_TITLE = 'New Chat'

const { t } = useI18n()

/* ── State ────────────────────────────────────────────────── */

const showConfirmModal = ref(false)
const executingConfirm = ref(false)
const pendingConfirmMessage = ref('')
const pendingActionName = ref('')
const pendingActionParams = ref<Record<string, string>>({})
let pendingMessageIndex = -1

const inputText = ref('')
const loading = ref(false)
const messagesRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const conversations = ref<Conversation[]>(loadConversations())
const activeConversationId = ref<string>(getOrCreateActive())

/* ── WebSocket ─────────────────────────────────────────────── */

const token = localStorage.getItem('litedock-token') || ''
const wsUrl = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/v1/assistant/stream/ws${token ? '?token=' + encodeURIComponent(token) : ''}`

const { ws, isConnected, connect: connectWS, disconnect: disconnectWS } = useWebSocket({
  url: wsUrl,
  onError: () => {
    loading.value = false
    const conv = conversations.value.find(c => c.id === activeConversationId.value)
    if (conv) {
      const lastMsg = conv.messages[conv.messages.length - 1]
      if (lastMsg && lastMsg.role === 'assistant' && lastMsg.status === 'executing') {
        lastMsg.status = 'failed'
        lastMsg.text = 'Connection lost. Please try again.'
      }
    }
  }
})

/* ── Persistence ──────────────────────────────────────────── */

function loadConversations(): Conversation[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

let saveDebounceTimer: ReturnType<typeof setTimeout> | null = null

function saveConversations(): void {
  if (saveDebounceTimer) clearTimeout(saveDebounceTimer)
  saveDebounceTimer = setTimeout(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(conversations.value))
  }, 300)
}

function getOrCreateActive(): string {
  if (conversations.value.length === 0) {
    return createConversation()
  }
  return conversations.value[0].id
}

function createConversation(): string {
  const id = generateId()
  conversations.value.unshift({
    id,
    title: DEFAULT_TITLE,
    messages: [],
    createdAt: Date.now(),
    updatedAt: Date.now(),
  })
  saveConversations()
  return id
}

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}

/* ── Computed ─────────────────────────────────────────────── */

const currentMessages = computed<ChatMessage[]>(() => {
  const conv = conversations.value.find(c => c.id === activeConversationId.value)
  return conv ? conv.messages : []
})

const quickActions = computed<QuickAction[]>(() => [
  { id: 'start-container', icon: markRaw(Play), label: t('ai.quick.startContainer'), prompt: t('ai.prompts.startContainer') },
  { id: 'check-logs', icon: markRaw(FileText), label: t('ai.quick.checkLogs'), prompt: t('ai.prompts.checkLogs') },
  { id: 'diagnose', icon: markRaw(Activity), label: t('ai.quick.diagnose'), prompt: t('ai.prompts.diagnose') },
  { id: 'deploy-web', icon: markRaw(Globe), label: t('ai.quick.deployWeb'), prompt: t('ai.prompts.deployWeb') },
  { id: 'check-disk', icon: markRaw(HardDrive), label: t('ai.quick.checkDisk'), prompt: t('ai.prompts.checkDisk') },
  { id: 'network-status', icon: markRaw(Network), label: t('ai.quick.network'), prompt: t('ai.prompts.network') },
])

/* ── Conversation actions ─────────────────────────────────── */

function newChat(): void {
  const id = createConversation()
  activeConversationId.value = id
  inputText.value = ''
  nextTick(() => inputRef.value?.focus())
}

function switchConversation(id: string): void {
  if (id === activeConversationId.value) return
  activeConversationId.value = id
  nextTick(scrollToBottom)
}

function deleteConversation(id: string): void {
  const idx = conversations.value.findIndex(c => c.id === id)
  if (idx === -1) return

  conversations.value.splice(idx, 1)

  if (id === activeConversationId.value) {
    if (conversations.value.length > 0) {
      const nextIdx = Math.min(idx, conversations.value.length - 1)
      activeConversationId.value = conversations.value[nextIdx].id
    } else {
      activeConversationId.value = createConversation()
    }
  }

  saveConversations()
}

function formatTime(timestamp: number): string {
  const diff = Date.now() - timestamp
  const mins = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (mins < 1) return t('dashboard.justNow')
  if (mins < 60) return t('dashboard.minutesAgo', { n: mins })
  if (hours < 24) return t('dashboard.hoursAgo', { n: hours })
  if (days < 7) return t('dashboard.daysAgo', { n: days })
  return new Date(timestamp).toLocaleDateString()
}

/* ── Core chat functions ──────────────────────────────────── */

function scrollToBottom(): void {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

function runQuickAction(action: QuickAction): void {
  if (loading.value) return
  inputText.value = action.prompt
  sendMessage()
}

function stopStreaming(): void {
  disconnectWS()
  loading.value = false
}

async function sendMessage(): Promise<void> {
  let text = inputText.value.trim()
  if (!text || loading.value) return

  text = stripShellChars(text)
  if (!text) return

  const conv = conversations.value.find(c => c.id === activeConversationId.value)
  if (!conv) return

  // Add user message
  conv.messages.push({ role: 'user', text })
  inputText.value = ''
  loading.value = true

  // Auto-title from first user message
  if (conv.title === DEFAULT_TITLE) {
    conv.title = text.length > 40 ? text.slice(0, 40) + '...' : text
  }

  // Add placeholder for streaming assistant response
  const assistantMsg: ChatMessage = { role: 'assistant', text: '', status: 'executing' }
  conv.messages.push(assistantMsg)
  const assistantMsgIdx = conv.messages.length - 1

  conv.updatedAt = Date.now()
  saveConversations()
  scrollToBottom()

  // Build messages payload excluding the placeholder just added
  const apiMessages = conv.messages
    .slice(0, -1)
    .map(m => ({ role: m.role, content: m.text }))

  try {
    // Connect or reuse WebSocket
    if (!isConnected.value) {
      connectWS()
      // Wait for connection (up to 5s)
      let attempts = 0
      while (!isConnected.value && attempts < 100) {
        await new Promise(r => setTimeout(r, 50))
        attempts++
      }
    }

    if (!isConnected.value) {
      throw new Error('Failed to connect to AI service')
    }

    const wsConn = ws.value
    if (!wsConn || wsConn.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected')
    }

    // Wait for streaming response via temporary message listener
    await new Promise<void>((resolve) => {
      const messageHandler = (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data)

          if (data.done) {
            wsConn.removeEventListener('message', messageHandler)
            wsConn.removeEventListener('close', closeHandler)
            if (!conv.messages[assistantMsgIdx].text) {
              conv.messages[assistantMsgIdx].text = t('assistant.response.noMatch')
            }
            conv.messages[assistantMsgIdx].status = 'completed'
            resolve()
          } else if (data.error) {
            wsConn.removeEventListener('message', messageHandler)
            wsConn.removeEventListener('close', closeHandler)
            conv.messages[assistantMsgIdx].text = data.error
            conv.messages[assistantMsgIdx].status = 'failed'
            resolve()
          } else if (data.content) {
            conv.messages[assistantMsgIdx].text += data.content
            conv.updatedAt = Date.now()
            scrollToBottom()
          }
        } catch {
          // skip malformed chunks
        }
      }

      const closeHandler = () => {
        wsConn.removeEventListener('message', messageHandler)
        wsConn.removeEventListener('close', closeHandler)
        if (!conv.messages[assistantMsgIdx].text) {
          conv.messages[assistantMsgIdx].text = t('assistant.response.noMatch')
        }
        resolve()
      }

      wsConn.addEventListener('message', messageHandler)
      wsConn.addEventListener('close', closeHandler)

      // Send the request
      wsConn.send(JSON.stringify({ messages: apiMessages }))

      // Safety timeout
      setTimeout(() => {
        wsConn.removeEventListener('message', messageHandler)
        wsConn.removeEventListener('close', closeHandler)
        if (!conv.messages[assistantMsgIdx].text) {
          conv.messages[assistantMsgIdx].text = t('assistant.response.noMatch')
        }
        resolve()
      }, 60000)
    })
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    conv.messages[assistantMsgIdx].text = err?.message || t('assistant.error.general')
    conv.messages[assistantMsgIdx].status = 'failed'
  } finally {
    if (conv.messages[assistantMsgIdx].status === 'executing') {
      conv.messages[assistantMsgIdx].status = 'completed'
    }
    loading.value = false
    conv.updatedAt = Date.now()
    saveConversations()
    scrollToBottom()
  }
}

/* ── Confirmation actions ─────────────────────────────────── */

async function confirmAction(): Promise<void> {
  if (executingConfirm.value || pendingMessageIndex < 0) return
  executingConfirm.value = true

  const conv = conversations.value.find(c => c.id === activeConversationId.value)
  if (!conv || pendingMessageIndex >= conv.messages.length) {
    executingConfirm.value = false
    showConfirmModal.value = false
    return
  }

  const msg = conv.messages[pendingMessageIndex]
  msg.status = 'executing'
  msg.text = t('assistant.response.thinking')
  showConfirmModal.value = false
  saveConversations()

  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const resp = await api.post<any>('/assistant/execute', {
      action: pendingActionName.value,
      params: pendingActionParams.value,
    }, {
      timeout: 30000,
    })
    msg.text = resp?.message || 'Action executed successfully'
    msg.status = 'completed'
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    msg.text = err?.response?.data?.msg || err?.response?.data?.error || 'Failed to execute action'
    msg.status = 'failed'
  } finally {
    executingConfirm.value = false
    pendingMessageIndex = -1
    conv.updatedAt = Date.now()
    saveConversations()
    scrollToBottom()
  }
}

function cancelAction(): void {
  showConfirmModal.value = false
  executingConfirm.value = false

  if (pendingMessageIndex >= 0) {
    const conv = conversations.value.find(c => c.id === activeConversationId.value)
    if (conv && pendingMessageIndex < conv.messages.length) {
      const msg = conv.messages[pendingMessageIndex]
      msg.status = 'failed'
      msg.text = 'Action cancelled'
      saveConversations()
    }
  }

  pendingMessageIndex = -1
}

onUnmounted(() => {
  if (saveDebounceTimer) {
    clearTimeout(saveDebounceTimer)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(conversations.value))
  }
})

onMounted(() => {
  nextTick(scrollToBottom)
})
</script>

<style scoped>
/* ── Layout ────────────────────────────────────────────────── */

.ai-layout {
  display: flex;
  flex-direction: column;
  height: calc(100vh - var(--header-height) - var(--content-padding) * 2);
  max-width: 100%;
  overflow: hidden;
}

/* ── Top bar ───────────────────────────────────────────────── */

.ai-topbar {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-6);
  border-bottom: 1px solid var(--color-border-weak);
  flex-shrink: 0;
}

.ai-topbar-title {
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-strong);
}

/* ── Body (sidebar + main) ─────────────────────────────────── */

.ai-body {
  display: flex;
  flex: 1;
  min-height: 0;
}

/* ── Sidebar ───────────────────────────────────────────────── */

.ai-sidebar {
  width: 280px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--color-border-weak);
  padding: var(--space-4);
  gap: var(--space-2);
}

.sidebar-empty {
  padding: var(--space-4) var(--space-2);
}

/* New Chat button */
.new-chat-btn {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-sm);
  background: none;
  color: var(--color-text-weak);
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.new-chat-btn:hover {
  border-color: var(--color-accent);
  color: var(--color-accent);
  background: var(--color-background-interactive-weaker);
}

/* History list */
.history-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
  scrollbar-width: none;
}

.history-list::-webkit-scrollbar {
  display: none;
}

.history-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
  position: relative;
}

.history-item:hover {
  background: var(--color-background-weak);
}

.history-item.active {
  background: var(--color-background-interactive-weaker);
}

.history-icon {
  flex-shrink: 0;
  color: var(--color-text-weaker);
}

.history-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.history-title {
  font-size: var(--font-size-xs);
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-time {
  font-size: 11px;
  color: var(--color-text-weaker);
}

.delete-btn {
  opacity: 0;
  background: none;
  border: none;
  color: var(--color-text-weaker);
  cursor: pointer;
  padding: 2px;
  border-radius: var(--radius-sm);
  flex-shrink: 0;
  transition: opacity var(--transition-fast), color var(--transition-fast), background var(--transition-fast);
}

.history-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  color: var(--color-error);
  background: var(--color-background-weak);
}

/* ── Main area ─────────────────────────────────────────────── */

.ai-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

/* Messages container */
.ai-messages {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-8) var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  margin: 0 auto;
  width: 100%;
  scrollbar-width: none;
}

.ai-messages::-webkit-scrollbar {
  display: none;
}

/* ── Empty state ───────────────────────────────────────────── */

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  color: var(--color-text-weaker);
  text-align: center;
}

.empty-icon {
  opacity: 0.3;
}

.empty-state p {
  font-size: var(--font-size-sm);
  line-height: var(--line-height-relaxed);
  max-width: 340px;
}

/* ── Messages ──────────────────────────────────────────────── */

.msg {
  padding: 8px 14px;
  border-radius: 16px;
  max-width: 85%;
  word-break: break-word;
  font-size: var(--font-size-sm);
  line-height: var(--line-height-normal);
}

.msg p {
  margin: 0;
}

.msg-user {
  align-self: flex-end;
  background: var(--color-accent);
  color: #ffffff;
  border-bottom-right-radius: 4px;
}

.msg-assistant {
  align-self: flex-start;
  background: var(--color-background-strong);
  color: var(--color-text);
  border-bottom-left-radius: 4px;
}

/* ── Message status labels ─────────────────────────────────── */

.msg-inner {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.msg-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  line-height: 1;
}

.msg-status-executing {
  color: var(--color-info);
}

.msg-status-completed {
  color: var(--color-success);
}

.msg-status-failed {
  color: var(--color-error);
}

.msg-status-warning {
  color: var(--color-warning);
}

/* ── Conversation switching transition ─────────────────────── */

.msg-fade-enter-active,
.msg-fade-leave-active {
  transition: opacity var(--transition-base);
}

.msg-fade-enter-from,
.msg-fade-leave-to {
  opacity: 0;
}

/* ── Bottom area ───────────────────────────────────────────── */

.ai-bottom {
  flex-shrink: 0;
  padding: var(--space-3) var(--space-6) var(--space-6);
  margin: 0 auto;
  width: 100%;
}

/* Quick action chips */
.quick-chips {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.quick-chips::-webkit-scrollbar {
  height: 3px;
}

.quick-chips::-webkit-scrollbar-thumb {
  background: var(--color-text-weaker);
  border-radius: var(--radius-full);
}

.chip {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-full);
  background: var(--color-background);
  color: var(--color-text-weak);
  font-size: 11px;
  font-family: var(--font-mono);
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.chip:hover:not(:disabled) {
  border-color: var(--color-accent);
  color: var(--color-accent);
  background: var(--color-background-interactive-weaker);
}

.chip:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

/* Input row */
.ai-input-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  background: var(--color-background);
  transition: border-color var(--transition-fast);
}

.ai-input-row:focus-within {
  border-color: var(--color-accent);
}

.ai-input {
  border: none;
  outline: none;
  flex: 1;
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text-strong);
  background: none;
  padding: 0;
}

.ai-input::placeholder {
  color: var(--color-text-weaker);
}

.ai-input:disabled {
  cursor: not-allowed;
}

.send-btn {
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* ── Spin animation ────────────────────────────────────────── */

.spin {
  animation: ai-spin 1s linear infinite;
}

@keyframes ai-spin {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}

/* ── Responsive ────────────────────────────────────────────── */

/* ── Confirmation modal ───────────────────────────────────── */

.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
}

.confirm-modal {
  width: 420px;
  max-width: 90vw;
  padding: var(--space-6);
}

.confirm-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.confirm-icon {
  color: var(--color-warning);
  flex-shrink: 0;
}

.confirm-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.confirm-body {
  margin-bottom: var(--space-6);
}

.confirm-message {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  line-height: var(--line-height-relaxed);
  margin: 0;
}

.confirm-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}

@media (max-width: 767px) {
  .ai-layout {
    height: calc(100vh - var(--header-height) - var(--space-4) * 2);
  }

  .ai-topbar {
    padding: var(--space-3) var(--space-4);
  }

  .ai-sidebar {
    display: none;
  }

  .ai-messages {
    padding: var(--space-4);
    gap: var(--space-2);
  }

  .ai-bottom {
    padding: var(--space-2) var(--space-4) var(--space-4);
  }

  .msg {
    max-width: 85%;
  }
}
</style>
