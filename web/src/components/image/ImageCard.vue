<template>
  <div class="card card-hover">
    <div class="card-header">
      <div class="image-header-left">
        <Box :size="18" class="text-muted flex-shrink-0" />
        <div class="image-header-titles">
          <div class="card-title">{{ displayName }}</div>
          <div class="text-xs text-muted">{{ formattedSize }}</div>
        </div>
      </div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <span class="badge badge-info">{{ tagCount }} tags</span>
        <span v-if="labelCount > 0" class="badge">{{ labelCount }} labels</span>
      </div>
    </div>

    <div class="card-body">
      <div class="card-info-row">
        <span class="label">ID</span>
        <span class="value truncate" :title="image.id">{{ shortId }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Tags</span>
        <span class="value">
          <span v-if="(image.repoTags?.length || 0) === 0" class="text-muted">untagged</span>
          <span v-else class="flex flex-wrap gap-1 justify-end">
            <span
              v-for="tag in (image.repoTags || []).slice(0, 3)"
              :key="tag"
              class="badge badge-info"
            >{{ tag }}</span>
            <span v-if="(image.repoTags?.length || 0) > 3" class="text-xs text-muted">+{{ (image.repoTags?.length || 0) - 3 }} more</span>
          </span>
        </span>
      </div>
      <div class="card-info-row">
        <span class="label">Created</span>
        <span class="value">{{ formattedDate }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Labels</span>
        <span class="value">{{ labelCount }}</span>
      </div>
    </div>

    <div class="card-actions">
      <button @click="emit('inspect', image)" class="btn btn-sm btn-ghost">
        <Eye :size="14" />
        Inspect
      </button>
      <button @click="emit('delete', image)" class="btn btn-sm btn-danger">
        <Trash2 :size="14" />
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Box, Eye, Trash2 } from 'lucide-vue-next'
import type { Image } from '@/types'
import { formatSize, formatDate } from '@/utils/format'

const props = defineProps<{
  image: Image
}>()

const emit = defineEmits<{
  inspect: [image: Image]
  delete: [image: Image]
}>()

const displayName = computed(() => props.image.repoTags?.[0] || 'untagged')

const shortId = computed(() => {
  const id = props.image.id.replace('sha256:', '')
  return id.slice(0, 12)
})



const tagCount = computed(() => props.image.repoTags?.length || 0)

const labelCount = computed(() => Object.keys(props.image.labels || {}).length)

const formattedSize = computed(() => formatSize(props.image.size))

const formattedDate = computed(() => formatDate(props.image.createdAt))
</script>

<style scoped>
.image-header-left {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  min-width: 0;
}

.image-header-titles {
  min-width: 0;
}
</style>
