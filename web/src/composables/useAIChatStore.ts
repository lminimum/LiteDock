import { computed, ref, watch } from 'vue'

export interface AIChatMessage {
  role: 'user' | 'assistant'
  text: string
  status?: 'sending' | 'executing' | 'completed' | 'failed' | 'requires_confirmation' | 'autonomous_executed'
  confirmationMessage?: string
  actionName?: string
  actionParams?: Record<string, string>
}

export interface AIConversation {
  id: string
  title: string
  messages: AIChatMessage[]
  createdAt: number
  updatedAt: number
}

const STORAGE_KEY = 'litdock-ai-conversations'
const AGENT_MODE_KEY = 'litdock-ai-agent-mode'
export const DEFAULT_TITLE = 'New Chat'
const DEFAULT_GREETING = '你好有什么可以帮助你'

function isClient(): boolean {
  return typeof window !== 'undefined' && typeof localStorage !== 'undefined'
}

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}

function createGreetingMessage(greeting: string): AIChatMessage {
  return {
    role: 'assistant',
    text: greeting,
    status: 'completed',
  }
}

function loadConversations(): AIConversation[] {
  if (!isClient()) return []

  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []

    const parsed = JSON.parse(raw) as AIConversation[]
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function loadAgentMode(): boolean {
  if (!isClient()) return false
  return localStorage.getItem(AGENT_MODE_KEY) === 'true'
}

function persistConversations(conversations: AIConversation[]): void {
  if (!isClient()) return
  localStorage.setItem(STORAGE_KEY, JSON.stringify(conversations))
}

function persistAgentMode(value: boolean): void {
  if (!isClient()) return
  localStorage.setItem(AGENT_MODE_KEY, String(value))
}

const conversations = ref<AIConversation[]>(loadConversations())
const activeConversationId = ref<string>(conversations.value[0]?.id || '')
const agentMode = ref<boolean>(loadAgentMode())
let greetingText = DEFAULT_GREETING

function ensureActiveConversation(): AIConversation {
  let conversation = conversations.value.find(conv => conv.id === activeConversationId.value)

  if (!conversation) {
    if (conversations.value.length > 0) {
      activeConversationId.value = conversations.value[0]!.id
      conversation = conversations.value[0]!
    } else {
      const created = createConversationInternal(greetingText)
      conversations.value.unshift(created)
      activeConversationId.value = created.id
      persistConversations(conversations.value)
      return created
    }
  }

  return conversation
}

function createConversationInternal(greeting: string): AIConversation {
  const now = Date.now()
  return {
    id: generateId(),
    title: DEFAULT_TITLE,
    messages: [createGreetingMessage(greeting)],
    createdAt: now,
    updatedAt: now,
  }
}

function createConversation(greeting = greetingText): string {
  const conversation = createConversationInternal(greeting)
  conversations.value.unshift(conversation)
  activeConversationId.value = conversation.id
  persistConversations(conversations.value)
  return conversation.id
}

function deleteConversation(id: string): void {
  const idx = conversations.value.findIndex(conv => conv.id === id)
  if (idx === -1) return

  conversations.value.splice(idx, 1)

  if (conversations.value.length === 0) {
    const created = createConversationInternal(greetingText)
    conversations.value.unshift(created)
    activeConversationId.value = created.id
    persistConversations(conversations.value)
    return
  }

  if (activeConversationId.value === id) {
    const nextIdx = Math.min(idx, conversations.value.length - 1)
    activeConversationId.value = conversations.value[nextIdx]!.id
  }

  persistConversations(conversations.value)
}

function setActiveConversation(id: string): void {
  if (conversations.value.some(conv => conv.id === id)) {
    activeConversationId.value = id
  }
}

function ensureGreeting(greeting: string): AIConversation {
  greetingText = greeting

  if (conversations.value.length === 0) {
    createConversation(greeting)
    return conversations.value[0]!
  }

  const conversation = ensureActiveConversation()
  if (conversation.messages.length === 0) {
    conversation.messages.push(createGreetingMessage(greeting))
    conversation.updatedAt = Date.now()
    persistConversations(conversations.value)
  }

  return conversation
}

watch(conversations, () => persistConversations(conversations.value), { deep: true })
watch(agentMode, value => persistAgentMode(value))

export function useAIChatStore() {
  const currentConversation = computed(() => {
    const found = conversations.value.find(conv => conv.id === activeConversationId.value)
    return found || conversations.value[0] || null
  })

  return {
    conversations,
    activeConversationId,
    agentMode,
    currentConversation,
    bootstrap: ensureGreeting,
    createConversation,
    deleteConversation,
    setActiveConversation,
  }
}
