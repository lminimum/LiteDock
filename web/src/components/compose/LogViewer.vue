<template>
  <div class="log-viewer">
    <div class="log-toolbar flex items-center justify-between p-2">
      <span class="text-sm font-mono">Logs</span>
      <div class="flex gap-2">
        <input
          v-model="searchFilter"
          class="input input-sm"
          placeholder="Filter..."
        />
        <button class="btn btn-sm btn-ghost" @click="$emit('clear')">
          Clear
        </button>
      </div>
    </div>
    <pre ref="logContainerRef" class="log-content"><code>{{ filteredLogs || 'No logs available.' }}</code></pre>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'

const props = withDefaults(defineProps<{
  logs?: string
  streaming?: boolean
}>(), {
  logs: '',
  streaming: false,
})

defineEmits<{
  clear: []
}>()

const searchFilter = ref('')
const logContainerRef = ref<HTMLPreElement | null>(null)

const filteredLogs = computed(() => {
  if (!props.logs) return ''
  if (!searchFilter.value.trim()) return props.logs
  return props.logs
    .split('\n')
    .filter((line) =>
      line.toLowerCase().includes(searchFilter.value.toLowerCase())
    )
    .join('\n')
})

watch(
  () => props.logs,
  async () => {
    await nextTick()
    if (logContainerRef.value) {
      logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
    }
  }
)
</script>

<style scoped>
.log-viewer {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.log-toolbar {
  background: var(--color-background-weak);
  border-bottom: 1px solid var(--color-border-weak);
  color: var(--color-text);
}

.log-content {
  background: #1a1b1e;
  color: #e0e0e0;
  font-family: var(--font-mono);
  padding: var(--space-3);
  border-radius: 0;
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
  margin: 0;
}

.log-content code {
  color: inherit;
  font-family: inherit;
}
</style>
