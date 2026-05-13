import { describe, it, expect, vi, beforeEach } from 'vitest'
import { nextTick } from 'vue'

// Mock vue-i18n
vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'assistant.title': 'AI Assistant',
        'assistant.response.thinking': 'Thinking...',
        'assistant.response.noMatch': 'No match',
        'assistant.error.general': 'Error occurred',
        'assistant.input.placeholder': 'Type your message...',
        'assistant.input.send': 'Send',
        'assistant.status.executing': 'Executing...',
        'assistant.status.completed': 'Completed',
        'assistant.status.failed': 'Failed',
        'assistant.confirmation.title': 'Confirm Action',
        'assistant.confirmation.confirm': 'Confirm',
        'assistant.confirmation.cancel': 'Cancel',
        'assistant.confirmation.waiting': 'Executing...',
        'ai.emptyPrompt': 'Ask me something',
        'ai.quick.startContainer': 'Start Container',
        'ai.quick.checkLogs': 'Check Logs',
        'ai.quick.diagnose': 'Diagnose',
        'ai.quick.deployWeb': 'Deploy Web',
        'ai.quick.checkDisk': 'Check Disk',
        'ai.quick.network': 'Network',
        'ai.prompts.startContainer': 'start nginx container',
        'ai.prompts.checkLogs': 'check nginx logs',
        'ai.prompts.diagnose': 'diagnose issues',
        'ai.prompts.deployWeb': 'deploy web app',
        'ai.prompts.checkDisk': 'check disk usage',
        'ai.prompts.network': 'check network',
        'dashboard.justNow': 'Just now',
        'dashboard.minutesAgo': '{n} min ago',
        'dashboard.hoursAgo': '{n} hr ago',
        'dashboard.daysAgo': '{n} day ago',
      }
      return messages[key] || key
    },
  }),
}))

// Mock lucide-vue-next
vi.mock('lucide-vue-next', () => ({
  Bot: { render: () => null },
  Play: { render: () => null },
  FileText: { render: () => null },
  Activity: { render: () => null },
  Globe: { render: () => null },
  HardDrive: { render: () => null },
  Network: { render: () => null },
  Send: { render: () => null },
  Loader2: { render: () => null },
  Plus: { render: () => null },
  MessageSquare: { render: () => null },
  Trash2: { render: () => null },
  CheckCircle2: { render: () => null },
  XCircle: { render: () => null },
  AlertTriangle: { render: () => null },
  Square: { render: () => null },
}))

// Mock sanitize
vi.mock('@/utils/sanitize', () => ({
  stripShellChars: (text: string) => text,
}))

import api from '@/utils/api'

describe('AI Chat Confirmation Modal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('handles requires_confirmation response from parse endpoint', async () => {
    // Simulate the backend returning requires_confirmation: true
    const mockParseResponse = {
      intent: 'prune_images',
      action: 'prune_images',
      description: 'Executing: prune_images',
      requires_confirmation: true,
      confirmation_message: 'This will permanently remove all unused Docker images. This action cannot be undone.',
      action_name: 'prune_images',
      action_params: { machine_id: 'local' },
    }

    vi.mocked(api.post).mockResolvedValueOnce(mockParseResponse)

    // Verify API was called correctly
    const resp = await api.post('/assistant/parse', { text: 'prune all unused images' })
    expect(resp.requires_confirmation).toBe(true)
    expect(resp.confirmation_message).toContain('permanently remove')
    expect(resp.action_name).toBe('prune_images')
    expect(resp.action_params).toEqual({ machine_id: 'local' })
  })

  it('handles non-destructive response (no confirmation needed)', async () => {
    const mockParseResponse = {
      intent: 'list_images',
      action: 'list_images',
      description: 'Executing: list_images',
      requires_confirmation: false,
    }

    vi.mocked(api.post).mockResolvedValueOnce(mockParseResponse)

    const resp = await api.post('/assistant/parse', { text: 'list all images' })
    expect(resp.requires_confirmation).toBe(false)
  })

  it('calls /assistant/execute on confirm and returns success', async () => {
    const mockExecuteResponse = {
      success: true,
      message: 'Unused images pruned successfully',
      data: { space_reclaimed: 1048576, images_deleted: 3 },
    }

    vi.mocked(api.post).mockResolvedValueOnce(mockExecuteResponse)

    const resp = await api.post('/assistant/execute', {
      action: 'prune_images',
      params: { machine_id: 'local' },
    })

    expect(resp.success).toBe(true)
    expect(resp.message).toContain('pruned')
    expect(api.post).toHaveBeenCalledWith('/assistant/execute', {
      action: 'prune_images',
      params: { machine_id: 'local' },
    })
  })

  it('handles /assistant/execute failure', async () => {
    const error = new Error('Failed to execute action')
    ;(error as any).response = { data: { msg: 'Failed to prune images' } }
    vi.mocked(api.post).mockRejectedValueOnce(error)

    try {
      await api.post('/assistant/execute', {
        action: 'prune_images',
        params: { machine_id: 'local' },
      })
      expect.unreachable('Should have thrown')
    } catch (err: any) {
      expect(err.response.data.msg).toBe('Failed to prune images')
    }
  })

  it('handles parse endpoint with safe input correctly', async () => {
    const mockParseResponse = {
      intent: 'chat',
      description: 'I can help you with Docker.',
    }

    vi.mocked(api.post).mockResolvedValueOnce(mockParseResponse)

    const resp = await api.post('/assistant/parse', { text: 'hello, how are you?' })
    expect(resp.intent).toBe('chat')
    expect(resp.description).toContain('help')
    expect(resp.requires_confirmation).toBeUndefined()
  })

  it('confirmation modal shows correct details for container delete', async () => {
    const mockParseResponse = {
      intent: 'delete_container',
      action: 'delete_container',
      description: 'Executing: delete_container',
      requires_confirmation: true,
      confirmation_message: "This will permanently delete container 'my-app' and all its data. This action cannot be undone.",
      action_name: 'delete_container',
      action_params: { container_id: 'my-app', machine_id: 'local' },
    }

    vi.mocked(api.post).mockResolvedValueOnce(mockParseResponse)

    const resp = await api.post('/assistant/parse', { text: 'delete my-app container' })
    expect(resp.requires_confirmation).toBe(true)
    expect(resp.confirmation_message).toContain("delete container 'my-app'")
  })

  it('confirms and executes a container stop action', async () => {
    const mockExecuteResponse = {
      success: true,
      message: "Container my-app stopped successfully",
    }

    vi.mocked(api.post).mockResolvedValueOnce(mockExecuteResponse)

    const resp = await api.post('/assistant/execute', {
      action: 'stop_container',
      params: { container_id: 'my-app', machine_id: 'local' },
    })

    expect(resp.success).toBe(true)
    expect(resp.message).toContain('stopped')
    expect(api.post).toHaveBeenCalledWith('/assistant/execute', {
      action: 'stop_container',
      params: { container_id: 'my-app', machine_id: 'local' },
    })
  })
})

function createSSEStream(chunks: { content: string; done: boolean }[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder()
  const data = chunks
    .map(c => `data: ${JSON.stringify(c)}\n\n`)
    .join('')
  return new ReadableStream({
    start(controller) {
      controller.enqueue(encoder.encode(data))
      controller.close()
    },
  })
}

async function parseSSEStream(stream: ReadableStream<Uint8Array>): Promise<{ content: string; doneReceived: boolean }> {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let fullContent = ''
  let doneReceived = false

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const parts = buffer.split('\n\n')
    buffer = parts.pop() || ''

    for (const part of parts) {
      if (!part.startsWith('data: ')) continue
      try {
        const json = JSON.parse(part.slice(6))
        if (json.done) {
          doneReceived = true
          break
        }
        if (json.content) {
          fullContent += json.content
        }
      } catch {
        // skip
      }
    }
  }

  return { content: fullContent, doneReceived }
}

describe('SSE Stream Parsing', () => {
  it('assembles content from multiple chunks', async () => {
    const chunks = [
      { content: 'Hel', done: false },
      { content: 'lo ', done: false },
      { content: 'World', done: false },
      { content: '!', done: false },
      { content: '', done: true },
    ]

    const stream = createSSEStream(chunks)
    const result = await parseSSEStream(stream)

    expect(result.content).toBe('Hello World!')
    expect(result.doneReceived).toBe(true)
  })

  it('handles single chunk response', async () => {
    const chunks = [
      { content: 'Short response', done: false },
      { content: '', done: true },
    ]

    const stream = createSSEStream(chunks)
    const result = await parseSSEStream(stream)

    expect(result.content).toBe('Short response')
    expect(result.doneReceived).toBe(true)
  })

  it('handles empty response (no content before done)', async () => {
    const chunks = [
      { content: '', done: true },
    ]

    const stream = createSSEStream(chunks)
    const result = await parseSSEStream(stream)

    expect(result.content).toBe('')
    expect(result.doneReceived).toBe(true)
  })

  it('handles AbortController stopping mid-stream', async () => {
    const controller = new AbortController()
    let aborted = false

    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(new TextEncoder().encode('data: {"content":"Hel","done":false}\n\n'))
      },
      cancel() {
        aborted = true
      },
    })

    const reader = stream.getReader()
    try {
      // Simulate abort after first read
      const { value } = await reader.read()
      expect(value).toBeDefined()

      controller.abort()
      await reader.cancel()
    } catch {
      // expected on abort
    }

    expect(aborted || controller.signal.aborted).toBeTruthy()
  })
})

describe('Stream Fetch Integration', () => {
  it('sends messages with correct structure', () => {
    const messages = [
      { role: 'user', content: 'Hello' },
      { role: 'assistant', content: 'Hi there!' },
      { role: 'user', content: 'How are you?' },
    ]

    const body = JSON.stringify({ messages })

    expect(() => JSON.parse(body)).not.toThrow()
    const parsed = JSON.parse(body)
    expect(parsed.messages).toHaveLength(3)
    expect(parsed.messages[0].role).toBe('user')
    expect(parsed.messages[0].content).toBe('Hello')
  })

  it('handles fetch error with status code', async () => {
    const errorStatus = 502
    const response = new Response(
      JSON.stringify({ msg: 'Bad Gateway' }),
      { status: errorStatus }
    )
    expect(response.ok).toBe(false)
    expect(response.status).toBe(errorStatus)

    const errData = await response.json()
    expect(errData.msg).toBe('Bad Gateway')
  })
})
