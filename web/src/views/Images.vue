<template>
  <div class="images-page">
    <PageHeader :title="t('pages.images.title')">
      <template #actions>
        <button @click="handlePrune" class="btn btn-secondary" :disabled="pruning || machines.length === 0">
          <Trash2 :size="16" />
          {{ pruning ? 'Pruning...' : t('pages.images.pruneImages') }}
        </button>
        <button @click="showPullModal = true" class="btn btn-primary" :disabled="machines.length === 0">
          <Download :size="16" />
          {{ t('pages.images.pullImage') }}
        </button>
      </template>
    </PageHeader>

    <div v-if="!loading && !error && images.length > 0" class="filters">
      <div class="search-box">
        <input
          v-model="searchQuery"
          :placeholder="t('common.searchPlaceholder') || 'Search images...'"
          type="text"
          class="input"
        />
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <RefreshCw :size="24" class="spinning" />
      <span>{{ t('common.refresh') }}...</span>
    </div>

    <div v-else-if="error" class="error-state card text-center">
      <p class="mb-4">{{ error }}</p>
      <button @click="fetchImages" class="btn btn-secondary">{{ t('common.refresh') }}</button>
    </div>

    <div v-else-if="images.length === 0" class="empty-state card text-center">
      <p class="mb-4">{{ t('pages.images.noImages') }}</p>
      <button
        v-if="machines.length > 0"
        @click="showPullModal = true"
        class="btn btn-primary"
      >
        {{ t('pages.images.pullImage') }}
      </button>
    </div>

    <div v-else-if="filteredImages.length === 0" class="empty-state card text-center">
      <p>{{ t('pages.images.noImages') }}</p>
    </div>

    <div v-else class="card-grid">
      <ImageCard
        v-for="img in filteredImages"
        :key="`${img.machineId}:${img.id}`"
        :image="img"
        @inspect="handleInspect"
        @delete="confirmDelete"
      />
    </div>

    <InspectModal
      :visible="showInspect"
      :title="inspectImage?.repoTags?.[0] || inspectImage?.id || ''"
      :fields="inspectFields"
      @close="showInspect = false"
    />

    <ImagePullModal
      v-if="machines.length > 0"
      :show="showPullModal"
      :machine-id="machines[0].id"
      @close="showPullModal = false"
      @pulled="onImagePulled"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Download, Trash2, RefreshCw } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import { imageService } from '@/services/imageService'
import type { Image, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import ImageCard from '@/components/image/ImageCard.vue'
import ImagePullModal from '@/components/image/ImagePullModal.vue'
import InspectModal from '@/components/ui/InspectModal.vue'

interface ImageWithMachine extends Image {
  machine: string
}

const loading = ref(false)
const error = ref('')
const pruning = ref(false)
const searchQuery = ref('')
const showPullModal = ref(false)
const machines = ref<RemoteMachine[]>([])
const images = ref<ImageWithMachine[]>([])

const showInspect = ref(false)
const selectedImage = ref<ImageWithMachine | null>(null)

function formatSize(bytes?: number): string {
  if (bytes === undefined || bytes === null) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

const inspectImage = computed(() => selectedImage.value)

const inspectFields = computed(() => {
  const img = selectedImage.value
  if (!img) return []
  return [
    { label: t('pages.images.inspect.id'), value: img.id },
    { label: t('pages.images.inspect.tags'), value: img.repoTags?.join(', ') || '-' },
    { label: t('pages.images.inspect.digests'), value: img.repoDigests?.join(', ') || '-' },
    { label: t('pages.images.inspect.size'), value: formatSize(img.size) },
    { label: t('pages.images.inspect.created'), value: img.createdAt || '-' },
    { label: 'Labels', value: Object.keys(img.labels || {}).join(', ') || '-' },
    { label: 'Machine', value: img.machine },
  ]
})

const handleInspect = (image: Image) => {
  selectedImage.value = image as ImageWithMachine
  showInspect.value = true
}

const filteredImages = computed(() => {
  let filtered = images.value

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(
      img =>
        img.repoTags.some(tag => tag.toLowerCase().includes(q)) ||
        img.id.toLowerCase().includes(q)
    )
  }

  return filtered
})

const fetchImages = async () => {
  loading.value = true
  error.value = ''
  try {
    const allMachines = await remoteMachineService.list()
    machines.value = allMachines

    const results = await Promise.all(
      allMachines.map(async (m) => {
        try {
          const imgs = await imageService.list(m.id)
          return imgs.map(img => ({ ...img, machineId: m.id, machine: m.name }))
        } catch {
          return [] as ImageWithMachine[]
        }
      })
    )

    const allImages: ImageWithMachine[] = []
    for (const imgs of results) {
      allImages.push(...imgs)
    }

    allImages.sort((a, b) => {
      if (a.machine !== b.machine) return a.machine.localeCompare(b.machine)
      const aName = a.repoTags?.[0] || ''
      const bName = b.repoTags?.[0] || ''
      return aName.localeCompare(bName)
    })

    images.value = allImages
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  } finally {
    loading.value = false
  }
}

const confirmDelete = async (image: Image) => {
  if (!confirm(t('pages.images.delete.confirm'))) return

  try {
    await imageService.delete(image.machineId, image.id)
    await fetchImages()
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  }
}

const handlePrune = async () => {
  if (!confirm(t('pages.images.prune.confirm'))) return

  pruning.value = true
  try {
    let totalReclaimed = 0
    let totalDeleted = 0

    for (const m of machines.value) {
      try {
        const result = await imageService.prune(m.id)
        totalReclaimed += result.spaceReclaimed
        totalDeleted += result.imagesDeleted
      } catch {
        // skip offline machines
      }
    }

    if (totalDeleted > 0) {
      alert(t('pages.images.prune.success', { count: totalDeleted, space: formatSize(totalReclaimed) }))
    } else {
      alert(t('pages.images.prune.noImages'))
    }

    await fetchImages()
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  } finally {
    pruning.value = false
  }
}

const onImagePulled = () => {
  showPullModal.value = false
  fetchImages()
}

onMounted(() => fetchImages())
</script>

<style scoped>
.images-page {
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

@media (max-width: 768px) {
  .card-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .filters {
    flex-direction: column;
  }
}
</style>
