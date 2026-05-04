<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
    <div class="card modal-card" @click.stop>
      <h3 class="modal-title">Pull Image</h3>

      <form @submit.prevent="handlePull">
        <div class="form-field">
          <label class="form-label">Repository</label>
          <input
            v-model="repository"
            class="input"
            placeholder="nginx"
            :disabled="submitting"
          />
        </div>

        <div class="form-field">
          <label class="form-label">Tag</label>
          <input
            v-model="tag"
            class="input"
            placeholder="latest"
            :disabled="submitting"
          />
        </div>

        <div v-if="error" class="form-error">{{ error }}</div>

        <div class="modal-actions">
          <button
            type="button"
            class="btn btn-ghost"
            @click="$emit('close')"
            :disabled="submitting"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="submitting || !repository.trim()"
          >
            {{ submitting ? 'Pulling...' : 'Pull' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { imageService } from '@/services/imageService'
import type { Image } from '@/types'

const props = defineProps<{
  machineId: string
  show: boolean
}>()

const emit = defineEmits<{
  pulled: [image: Image]
  close: []
}>()

const repository = ref('')
const tag = ref('latest')
const error = ref('')
const submitting = ref(false)

watch(() => props.show, (val) => {
  if (val) {
    repository.value = ''
    tag.value = 'latest'
    error.value = ''
    submitting.value = false
  }
})

const handlePull = async () => {
  if (!repository.value.trim()) {
    error.value = 'Repository is required'
    return
  }

  error.value = ''
  submitting.value = true

  try {
    const image = await imageService.pull(props.machineId, {
      repository: repository.value.trim(),
      tag: tag.value.trim() || 'latest',
    })
    emit('pulled', image)
    emit('close')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Failed to pull image'
    error.value = msg
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--color-background-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: var(--space-4);
}

.modal-card {
  max-width: 420px;
  width: 100%;
  padding: var(--space-6);
}

.modal-title {
  margin-bottom: var(--space-6);
}

.form-field {
  margin-bottom: var(--space-4);
}

.form-label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
}

.form-error {
  margin-top: var(--space-3);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--color-error);
  background: var(--color-error-bg);
  border-radius: var(--radius-sm);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-6);
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border-weak);
}
</style>
