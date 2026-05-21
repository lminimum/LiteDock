<template>
  <!-- Unified container: narrow handle bar when closed, expands to chat panel when open -->
  <div
    class="chat-container"
    :class="{ 'is-open': isOpen, 'is-dragging': isDragging, 'is-handle-dragging': isHandleDragging }"
    :style="containerStyle"
  >
    <!-- CLOSED: prominent handle on edge -->
    <div
      v-if="!isOpen"
      class="chat-handle"
      :aria-label="t('assistant.chat.toggle')"
      title="AI Assistant"
      @mousedown="onHandleDragStart"
      @click="openChat"
    >
      <Bot :size="16" :stroke-width="2" />
    </div>

    <!-- OPEN: full chat panel -->
    <template v-if="isOpen">
      <!-- Header (drag handle) -->
      <div
        class="chat-header flex items-center justify-between p-4"
        :class="{ 'chat-header--dragging': isDragging }"
        @mousedown="onDragStart"
      >
        <div class="flex items-center gap-2">
          <Bot :size="16" :stroke-width="2" />
          <span class="font-semibold text-sm">{{ t('assistant.title') }}</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="chat-mode-toggle">
            <button
              class="mode-btn"
              :class="{ active: !agentMode }"
              @click="agentMode = false"
            >
              {{ t('ai.mode.assistant') }}
            </button>
            <button
              class="mode-btn"
              :class="{ active: agentMode }"
              @click="agentMode = true"
            >
              <Zap :size="10" />
              {{ t('ai.mode.agent') }}
            </button>
          </div>
          <button
            class="btn btn-ghost btn-sm"
            :aria-label="t('assistant.chat.close')"
            @click="closeChat"
          >
            <X :size="16" />
          </button>
        </div>
      </div>

      <!-- Quick actions row -->
      <div class="chat-quick-actions">
        <button
          v-for="action in quickActions"
          :key="action.id"
          class="quick-action-btn"
          :title="action.label"
          :aria-label="action.label"
          :disabled="loading"
          @click="runQuickAction(action)"
        >
          <component :is="action.icon" :size="14" />
        </button>
      </div>

      <!-- Messages area (scrollable) -->
      <div ref="messagesContainer" class="chat-messages">
        <template v-for="(msg, i) in messages" :key="i">
          <!-- User message (right-aligned) -->
          <div v-if="msg.role === 'user'" class="message-user">
            <span>{{ msg.text }}</span>
          </div>
          <!-- Assistant response (left-aligned) -->
          <div v-else class="message-assistant">
            <div class="assistant-markdown" v-html="renderMarkdown(msg.text)"></div>
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
            <div v-else-if="msg.status === 'autonomous_executed'" class="msg-status msg-status-agent">
              <Zap :size="12" />
              <span>Auto-executed</span>
            </div>
          </div>
        </template>
      </div>

      <!-- Input area -->
      <div class="chat-input-area flex items-center gap-2 p-4">
        <input
          ref="inputRef"
          v-model="inputText"
          class="input flex-1"
          :placeholder="t('assistant.input.placeholder')"
          :disabled="loading"
          maxlength="500"
          @keyup.enter="sendMessage"
        />
        <button
          class="btn btn-primary"
          :disabled="loading || !inputText.trim()"
          :aria-label="t('assistant.input.send')"
          @click="sendMessage"
        >
          <Send :size="16" />
        </button>
      </div>
    </template>

    <ActionConfirmationModal
      :show="showConfirmModal"
      :executing="executingConfirm"
      :message="pendingConfirmMessage"
      :action-name="pendingActionName"
      :action-params="pendingActionParams"
      :risk-level="pendingRiskLevel"
      :typed-required="typedConfirmationRequired"
      :expected-text="typedConfirmationExpected"
      :model-value="typedConfirmationInput"
      @update:model-value="typedConfirmationInput = $event"
      @confirm="composableConfirm"
      @cancel="composableCancel"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Bot, Play, FileText, Activity, Globe, HardDrive, Network,
  X, Send, Loader2, CheckCircle2, XCircle, AlertTriangle, Zap,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import api from '@/utils/api'
import { useWebSocket } from '@/composables/useWebSocket'
import { stripShellChars } from '@/utils/sanitize'
import { renderMarkdown } from '@/utils/markdown'
import { useActionConfirmation } from '@/composables/useActionConfirmation'
import ActionConfirmationModal from '@/components/ui/ActionConfirmationModal.vue'

interface ChatMessage {
  role: 'user' | 'assistant'
  text: string
  status?: 'executing' | 'completed' | 'failed' | 'requires_confirmation' | 'autonomous_executed'
}

interface QuickAction {
  id: string
  icon: Component
  label: string
  prompt: string
}

const { t } = useI18n()
const isOpen = ref(false)
const inputText = ref('')
const messages = ref<ChatMessage[]>([])
const loading = ref(false)
const agentMode = ref(localStorage.getItem('litdock-ai-agent-mode') === 'true')
const messagesContainer = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)

const {
  showConfirmModal,
  executingConfirm,
  pendingConfirmMessage,
  pendingActionName,
  pendingActionParams,
  pendingRiskLevel,
  typedConfirmationInput,
  typedConfirmationRequired,
  typedConfirmationExpected,
  triggerConfirmation,
  confirmAction: composableConfirm,
  cancelAction: composableCancel,
} = useActionConfirmation()

let pendingMessageIndex = -1

/* ── WebSocket ─────────────────────────────────────────────── */

const token = localStorage.getItem('litedock-token') || ''
const wsUrl = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/v1/assistant/stream/ws${token ? '?token=' + encodeURIComponent(token) : ''}`

const { ws, isConnected, connect: connectWS, disconnect: disconnectWS } = useWebSocket({
  url: wsUrl,
  onError: () => {
    loading.value = false
    const lastMsg = messages.value[messages.value.length - 1]
    if (lastMsg && lastMsg.role === 'assistant' && lastMsg.status === 'executing') {
      lastMsg.status = 'failed'
      lastMsg.text = 'Connection lost. Please try again.'
    }
  },
})

// Drag state
const DIALOG_WIDTH = 380
const DIALOG_ESTIMATED_HEIGHT = 560

const dialogX = ref(0)
const dialogY = ref(computeInitialY())
const isDragging = ref(false)
const dragOffsetX = ref(0)
const dragOffsetY = ref(0)
const isHandleDragging = ref(false)
const dragDistance = ref(0)
const dragStartX = ref(0)
const dragStartY = ref(0)

const containerStyle = computed(() => {
  if (!isOpen.value) {
    return {
      top: `${dialogY.value}px`,
      right: '0',
    }
  }
  return {
    left: `${dialogX.value}px`,
    top: `${dialogY.value}px`,
  }
})

function computeInitialX(): number {
  return window.innerWidth - DIALOG_WIDTH
}

function computeInitialY(): number {
  return Math.max(0, (window.innerHeight - DIALOG_ESTIMATED_HEIGHT) / 2)
}

// Quick action definitions
const quickActions: QuickAction[] = [
  { id: 'start-container', icon: markRaw(Play), label: 'Start Container', prompt: '启动一个容器' },
  { id: 'check-logs', icon: markRaw(FileText), label: 'Check Logs', prompt: '查看最近的容器日志' },
  { id: 'diagnose', icon: markRaw(Activity), label: 'Diagnose', prompt: '诊断容器状态' },
  { id: 'deploy-web', icon: markRaw(Globe), label: 'Deploy Web', prompt: '部署一个web服务' },
  { id: 'check-disk', icon: markRaw(HardDrive), label: 'Check Disk', prompt: '查看磁盘使用情况' },
  { id: 'network-status', icon: markRaw(Network), label: 'Network Status', prompt: '查看网络状态' },
]

function runQuickAction(action: QuickAction): void {
  if (loading.value) return
  inputText.value = action.prompt
  sendMessage()
}

/* ── Drag ──────────────────────────────────────────────────── */

function onHandleDragStart(e: MouseEvent) {
  if (e.button !== 0) return
  isHandleDragging.value = true
  dragDistance.value = 0
  dragStartX.value = e.clientX
  dragStartY.value = e.clientY
  dragOffsetY.value = e.clientY - dialogY.value

  const onMove = (moveEvent: MouseEvent) => {
    if (!isHandleDragging.value) return
    
    // Calculate distance to distinguish drag from click
    const dist = Math.sqrt(
      Math.pow(moveEvent.clientX - dragStartX.value, 2) +
      Math.pow(moveEvent.clientY - dragStartY.value, 2)
    )
    dragDistance.value = dist

    let newY = moveEvent.clientY - dragOffsetY.value
    newY = Math.max(0, Math.min(newY, window.innerHeight - 64))
    dialogY.value = newY
  }

  const onEnd = () => {
    isHandleDragging.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onEnd)
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onEnd)
  e.preventDefault()
  e.stopPropagation()
}

function onDragStart(e: MouseEvent) {
  if (e.button !== 0) return
  if ((e.target as HTMLElement).closest('button') !== null) return

  isDragging.value = true
  dragOffsetX.value = e.clientX - dialogX.value
  dragOffsetY.value = e.clientY - dialogY.value

  document.addEventListener('mousemove', onDragMove)
  document.addEventListener('mouseup', onDragEnd)
  e.preventDefault()
}

function onDragMove(e: MouseEvent) {
  if (!isDragging.value) return

  let newX = e.clientX - dragOffsetX.value
  let newY = e.clientY - dragOffsetY.value

  newX = Math.max(0, Math.min(newX, window.innerWidth - DIALOG_WIDTH))
  newY = Math.max(0, Math.min(newY, window.innerHeight - 24))

  dialogX.value = newX
  dialogY.value = newY
}

function onDragEnd() {
  isDragging.value = false
  document.removeEventListener('mousemove', onDragMove)
  document.removeEventListener('mouseup', onDragEnd)
}

function cleanupDrag() {
  if (isDragging.value) {
    onDragEnd()
  }
}

/* ── Chat ──────────────────────────────────────────────────── */

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

function formatActionResult(payload: Record<string, unknown>): string {
  const result = payload.result
  if (typeof result === 'string' && result.trim()) {
    return result
  }

  const message = payload.message
  if (typeof message === 'string' && message.trim()) {
    return message
  }

  const data = payload.data
  if (typeof data === 'string' && data.trim()) {
    return data
  }

  if (data != null) {
    try {
      return JSON.stringify(data, null, 2)
    } catch {
      return 'Done'
    }
  }

  return 'Done'
}

function openChat() {
  if (dragDistance.value > 5) return // Ignore if it was a drag
  dialogX.value = computeInitialX()
  isOpen.value = true
  nextTick(() => {
    inputRef.value?.focus()
  })
}

function closeChat() {
  cleanupDrag()
  isOpen.value = false
}

async function sendMessage() {
  let text = inputText.value.trim()
  if (!text || loading.value) return

  text = stripShellChars(text)
  if (!text) return

  messages.value.push({ role: 'user', text })
  inputText.value = ''
  loading.value = true
  scrollToBottom()

  // Add placeholder for streaming assistant response
  const assistantMsg: ChatMessage = { role: 'assistant', text: '', status: 'executing' }
  messages.value.push(assistantMsg)
  const assistantMsgIdx = messages.value.length - 1
  let currentMsg = messages.value[assistantMsgIdx]!

  // Build messages payload excluding the placeholder just added
  const apiMessages = messages.value
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
      let resolved = false

      function cleanup() {
        if (resolved) return
        resolved = true
        wsConn!.removeEventListener('message', messageHandler)
        wsConn!.removeEventListener('close', closeHandler)
      }

      const messageHandler = (event: MessageEvent) => {
        try {
          const raw = JSON.parse(event.data)

          // ── Versioned envelope (v:1) ──────────────────────────
          if (raw.v === 1 && raw.type) {
            switch (raw.type) {
              case 'content': {
                const p = raw.payload || {}
                if (p.content) {
                  currentMsg.text += p.content
                  scrollToBottom()
                }
                if (p.done) {
                  cleanup()
                  if (!currentMsg.text) {
                    currentMsg.text = t('assistant.response.noMatch')
                  }
                  currentMsg.status = 'completed'
                  resolve()
                }
                break
              }
              case 'action_required': {
                const intent = raw.payload || {}
                cleanup()
                currentMsg.status = 'requires_confirmation'
                currentMsg.text = intent.confirmation_message || `Action required: ${intent.action}`
                scrollToBottom()
                triggerActionConfirmation(
                  intent.action || '',
                  intent.params || {},
                  intent.confirmation_message || `Execute ${intent.action}?`,
                  intent.risk_level === 'dangerous' ? 'dangerous' : intent.risk_level === 'modify' ? 'caution' : 'safe',
                  intent.confirmation_token || '',
                  assistantMsgIdx,
                )
                resolve()
                break
              }
              case 'action_result': {
                const p = (raw.payload || {}) as Record<string, unknown>
                const actionName = typeof p.action === 'string' && p.action ? p.action : 'Action'
                currentMsg.text += `\n**${actionName}** result:\n${formatActionResult(p)}`
                currentMsg.status = 'autonomous_executed'
                scrollToBottom()
                break
              }
              case 'error': {
                const p = raw.payload || {}
                cleanup()
                currentMsg.text = p.message || t('assistant.error.general')
                currentMsg.status = 'failed'
                resolve()
                break
              }
              case 'done': {
                cleanup()
                if (!currentMsg.text) {
                  currentMsg.text = t('assistant.response.noMatch')
                }
                currentMsg.status = 'completed'
                resolve()
                break
              }
            }
            return
          }

          // ── Legacy unversioned format (backwards-compat) ─────
          if (raw.done) {
            cleanup()
            if (!currentMsg.text) {
              currentMsg.text = t('assistant.response.noMatch')
            }
            currentMsg.status = 'completed'
            resolve()
          } else if (raw.error) {
            cleanup()
            currentMsg.text = raw.error
            currentMsg.status = 'failed'
            resolve()
          } else if (raw.content) {
            currentMsg.text += raw.content
            scrollToBottom()
          }
        } catch {
          // skip malformed chunks
        }
      }

      const closeHandler = () => {
        cleanup()
        if (!currentMsg.text) {
          currentMsg.text = t('assistant.response.noMatch')
        }
        resolve()
      }

      wsConn.addEventListener('message', messageHandler)
      wsConn.addEventListener('close', closeHandler)

      // Send the request
      wsConn.send(JSON.stringify({ messages: apiMessages, autonomous: agentMode.value }))

      // Safety timeout
      setTimeout(() => {
        cleanup()
        if (!currentMsg.text) {
          currentMsg.text = t('assistant.response.noMatch')
        }
        resolve()
      }, 60000)
    })
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    currentMsg.text = err?.message || t('assistant.error.general')
    currentMsg.status = 'failed'
  } finally {
    if (currentMsg.status === 'executing') {
      currentMsg.status = 'completed'
    }
    if (currentMsg.status === 'autonomous_executed') {
      currentMsg.status = 'completed'
    }
    loading.value = false
    scrollToBottom()
  }
}

/* ── Confirmation actions ─────────────────────────────────── */

function triggerActionConfirmation(
  actionName: string,
  params: Record<string, string>,
  message: string,
  riskLevel: 'safe' | 'caution' | 'dangerous',
  confirmToken: string,
  messageIndex: number,
): void {
  pendingMessageIndex = messageIndex

  triggerConfirmation(
    actionName,
    params,
    message,
    riskLevel,
    confirmToken,
    async () => {
      if (pendingMessageIndex < 0 || pendingMessageIndex >= messages.value.length) return

      const msg = messages.value[pendingMessageIndex]!
      msg.status = 'executing'
      msg.text = t('assistant.response.thinking')

      try {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const resp = await api.post<any>('/assistant/execute', {
          action: actionName,
          params,
          confirmation_token: confirmToken,
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
        pendingMessageIndex = -1
        scrollToBottom()
      }
    },
    () => {
      if (pendingMessageIndex >= 0 && pendingMessageIndex < messages.value.length) {
        const msg = messages.value[pendingMessageIndex]!
        msg.status = 'failed'
        msg.text = 'Action cancelled'
      }
      pendingMessageIndex = -1
    },
  )
}

/* ── Keyboard ──────────────────────────────────────────────── */

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    closeChat()
  }
}

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  cleanupDrag()
  disconnectWS()
})
</script>

<style scoped>
/* ── Container (unified: closed bar → open panel) ──────────── */

.chat-container {
  position: fixed;
  z-index: 100;
  display: flex;
  flex-direction: column;
  right: 0;
  transition:
    width var(--transition-base),
    height var(--transition-base),
    border-radius var(--transition-base),
    opacity var(--transition-base),
    background var(--transition-base),
    right var(--transition-base);
}

/* Closed: prominent semi-circle handle on right edge */
.chat-container:not(.is-open) {
  width: 32px;
  height: 64px;
  right: 0;
  border-radius: 32px 0 0 32px;
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-right: none;
  opacity: 0.9;
  cursor: grab;
  overflow: hidden;
  box-shadow: -2px 0 12px rgba(0, 0, 0, 0.2);
}

.chat-container:not(.is-open).is-handle-dragging {
  cursor: grabbing;
  transition: none;
}

.chat-container:not(.is-open):hover {
  width: 40px;
  opacity: 1;
  background: var(--color-background-strong);
}

.chat-container.is-open {
  width: 380px;
  max-height: calc(100vh - 48px);
  border-radius: var(--radius-md) 0 0 var(--radius-md);
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-right: none;
  opacity: 1;
  overflow: hidden;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.3);
}

/* ── Handle (closed state icon wrapper) ────────────────────── */

.chat-handle {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-weak);
}

/* ── Header ────────────────────────────────────────────────── */

.chat-header {
  flex-shrink: 0;
  border-bottom: 1px solid var(--color-border-weak);
  cursor: grab;
  user-select: none;
}

.chat-header--dragging {
  cursor: grabbing;
}

.chat-mode-toggle {
  display: flex;
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: 1px;
  gap: 1px;
}

.chat-mode-toggle .mode-btn {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px 6px;
  border: none;
  border-radius: var(--radius-sm);
  background: none;
  color: var(--color-text-weaker);
  font-family: var(--font-mono);
  font-size: 10px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.chat-mode-toggle .mode-btn:hover:not(.active) {
  color: var(--color-text);
}

.chat-mode-toggle .mode-btn.active {
  background: var(--color-background-strong);
  color: var(--color-accent);
  font-weight: var(--font-weight-medium);
}

/* ── Quick actions ─────────────────────────────────────────── */

.chat-quick-actions {
  display: flex;
  gap: var(--space-1);
  padding: var(--space-2) var(--space-4);
  overflow-x: auto;
  flex-shrink: 0;
  border-bottom: 1px solid var(--color-border-weak);
  scrollbar-width: none;
}

.chat-quick-actions::-webkit-scrollbar {
  display: none;
}

.quick-action-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: none;
  background: transparent;
  color: var(--color-text-weak);
  cursor: pointer;
  flex-shrink: 0;
  transition:
    background var(--transition-fast),
    color var(--transition-fast);
}

.quick-action-btn:hover:not(:disabled) {
  background: var(--color-background-interactive-weaker);
  color: var(--color-accent);
}

.quick-action-btn:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

/* ── Messages ──────────────────────────────────────────────── */

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  scrollbar-width: none;
  background: var(--color-background);
}

.chat-messages::-webkit-scrollbar {
  display: none;
}

.message-user {
  align-self: flex-end;
  background: var(--color-accent);
  color: #ffffff;
  padding: 6px 12px;
  border-radius: 12px 12px 2px 12px;
  max-width: 85%;
  word-break: break-word;
  font-size: var(--font-size-sm);
  line-height: var(--line-height-normal);
}

.message-assistant {
  align-self: flex-start;
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  color: var(--color-text);
  padding: 6px 12px;
  border-radius: 12px 12px 12px 2px;
  max-width: 85%;
  word-break: break-word;
  font-size: var(--font-size-sm);
  line-height: var(--line-height-normal);
}

/* ── Message status labels ─────────────────────────────────── */

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

.msg-status-agent {
  color: var(--color-accent);
}

/* ── Input area ────────────────────────────────────────────── */

.chat-input-area {
  flex-shrink: 0;
  border-top: 1px solid var(--color-border-weak);
}

/* ── Loading spin ──────────────────────────────────────────── */

.spin {
  animation: assistant-spin 1s linear infinite;
}

@keyframes assistant-spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

/* ── Responsive ────────────────────────────────────────────── */

@media (max-width: 480px) {
  .chat-container.is-open {
    width: calc(100vw - 16px);
  }

  .chat-container:not(.is-open) {
    width: 6px;
    height: 80px;
  }
}
</style>
