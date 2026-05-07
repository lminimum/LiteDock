<template>
  <div class="editor-container">
    <div class="editor-toolbar flex items-center justify-between p-2">
      <span class="text-sm">Compose File</span>
      <button
        class="btn btn-sm btn-primary"
        :disabled="readonly"
        @click="$emit('save', currentContent)"
      >
        Save
      </button>
    </div>
    <div ref="editorContainerRef" class="editor-wrapper"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import * as monaco from 'monaco-editor'
import type { ComposeTemplate } from '@/types'

// Monaco workers via Vite's ?worker import (handles bundling automatically)
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker.js?worker'

self.MonacoEnvironment = {
  getWorker(_workerId: string, _label: string) {
    return new EditorWorker()
  },
}

const props = withDefaults(defineProps<{
  modelValue: string
  readonly?: boolean
  template?: ComposeTemplate | null
}>(), {
  readonly: false,
  template: null,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  save: [content: string]
}>()

const editorContainerRef = ref<HTMLDivElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null
const currentContent = ref(props.modelValue)

onMounted(async () => {
  await nextTick()
  if (!editorContainerRef.value) return

  editor = monaco.editor.create(editorContainerRef.value, {
    value: props.modelValue,
    language: 'yaml',
    theme: 'vs-dark',
    readOnly: props.readonly,
    minimap: { enabled: false },
    automaticLayout: true,
    fontSize: 14,
    tabSize: 2,
  })

  editor.onDidChangeModelContent(() => {
    const value = editor?.getValue() ?? ''
    currentContent.value = value
    emit('update:modelValue', value)
  })

  // If template is provided and editor is empty, insert template content
  if (props.template && !props.modelValue) {
    editor.setValue(props.template.content)
  }
})

onUnmounted(() => {
  editor?.dispose()
  editor = null
})

watch(() => props.modelValue, (newVal) => {
  if (editor && editor.getValue() !== newVal) {
    editor.setValue(newVal)
  }
})

watch(() => props.readonly, (val) => {
  editor?.updateOptions({ readOnly: val })
})

watch(() => props.template, (tpl) => {
  if (tpl && editor && !editor.getValue().trim()) {
    editor.setValue(tpl.content)
  }
})
</script>

<style scoped>
.editor-container {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.editor-toolbar {
  background: var(--color-background-weak);
  border-bottom: 1px solid var(--color-border-weak);
  color: var(--color-text);
}

.editor-wrapper {
  min-height: 300px;
  flex: 1;
}
</style>
