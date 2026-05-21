<template>
  <div class="images-page">
    <PageHeader :title="t('pages.images.title')">
      <button @click="fetchImages" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('common.refresh') }}
      </button>
      <button @click="handlePrune" class="btn btn-secondary" :disabled="pruning || machines.length === 0">
        <Trash2 :size="16" />
        {{ pruning ? 'Pruning...' : t('pages.images.pruneImages') }}
      </button>
      <button @click="showPullModal = true" class="btn btn-primary" :disabled="machines.length === 0">
        <Download :size="16" />
        {{ t('pages.images.pullImage') }}
      </button>
    </PageHeader>

    <CollapsibleFilters
      v-if="!loading && !error && images.length > 0"
      v-model="searchQuery"
      :search-placeholder="t('common.searchPlaceholder') || 'Search images...'"
      search-label="Search"
      filter-label="Filters"
      :has-filters="true"
    >
      <template #filters>
        <select v-model="machineFilter" class="input">
          <option value="">{{ t('common.allMachines') }}</option>
          <option v-for="opt in machineOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>
      </template>
      <template #right>
        <ViewToggle v-model="viewMode" />
      </template>
    </CollapsibleFilters>

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

    <Transition name="view-fade" mode="out-in">
      <div v-if="viewMode === 'card'" key="card">
        <template v-for="group in groupedItems" :key="group.machineId">
          <div class="machine-section-header">
            <Server :size="16" class="icon" />
            {{ group.machineName }}
            <span class="count">{{ group.items.length }} {{ t('common.images') }}</span>
          </div>
          <div class="card-grid">
            <ImageCard
              v-for="img in group.items"
              :key="`${img.machineId}:${img.id}`"
              :image="img"
              @inspect="handleInspect"
              @delete="confirmDelete"
            />
          </div>
        </template>
      </div>

      <div v-else key="list">
        <template v-for="group in groupedItems" :key="group.machineId">
          <div class="machine-section-header">
            <Server :size="16" class="icon" />
            {{ group.machineName }}
            <span class="count">{{ group.items.length }} {{ t('common.images') }}</span>
          </div>
          <div class="item-list">
            <div v-for="img in group.items" :key="`${img.machineId}:${img.id}`" class="item-list-row">
              <div class="item-list-info">
                <div class="item-list-title">{{ img.repoTags?.[0] || 'untagged' }}</div>
                <div class="item-list-meta">
                  <span class="text-muted">ID: {{ img.id.replace('sha256:', '').slice(0, 12) }}</span>
                  <span class="badge badge-info">{{ img.repoTags?.length || 0 }} tags</span>
                  <span>{{ (img.size / 1048576).toFixed(1) }} MB</span>
                </div>
              </div>
              <div class="item-list-actions">
                <button @click="handleInspect(img)" class="btn btn-sm btn-ghost">
                  <Eye :size="14" /> {{ t('common.inspect') }}
                </button>
                <button @click="confirmDelete(img)" class="btn btn-sm btn-ghost btn-danger-text">
                  <Trash2 :size="14" /> {{ t('common.delete') }}
                </button>
              </div>
            </div>
          </div>
        </template>
      </div>
    </Transition>

    <InspectModal
      :visible="showInspect"
      :title="inspectImage?.repoTags?.[0] || inspectImage?.id || ''"
      :fields="inspectFields"
      @close="showInspect = false"
    />

    <ImagePullModal
      v-if="machines.length > 0"
      :show="showPullModal"
      :machine-id="defaultCreateMachineId"
      :machines="machines"
      @close="showPullModal = false"
      @pulled="onImagePulled"
    />

    <ConfirmModal
      :visible="confirmState !== null"
      :title="confirmState?.title || ''"
      :message="confirmState?.message || ''"
      :confirm-text="confirmState?.confirmText"
      :danger="confirmState?.danger ?? false"
      :disabled="confirmBusy"
      @confirm="confirmAction"
      @cancel="cancelConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Download, Trash2, RefreshCw, Eye, Server } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import { imageService } from '@/services/imageService'
import type { Image, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import ImageCard from '@/components/image/ImageCard.vue'
import ImagePullModal from '@/components/image/ImagePullModal.vue'
import InspectModal from '@/components/ui/InspectModal.vue'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'
import { formatSize, formatDate } from '@/utils/format'
import ViewToggle from '@/components/ui/ViewToggle.vue'
import CollapsibleFilters from '@/components/ui/CollapsibleFilters.vue'
import { useViewMode } from '@/composables/useViewMode'
import { useMachineFilter } from '@/composables/useMachineFilter'

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
const viewMode = useViewMode('images')

const confirmState = ref<{
  title: string
  message: string
  confirmText?: string
  danger?: boolean
  action: 'delete' | 'prune'
  id?: string
} | null>(null)
const confirmBusy = ref(false)

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

const { machineFilter, machineOptions, groupedItems } = useMachineFilter(
  filteredImages,
  machines,
  (img) => img.machineId,
  (img) => img.machine,
)

const defaultCreateMachineId = computed(() => {
  if (machineFilter.value) return machineFilter.value
  return machines.value[0]?.id ?? ''
})

const showInspect = ref(false)
const selectedImage = ref<ImageWithMachine | null>(null)

const formatDateTime = formatDate

const inspectImage = computed(() => selectedImage.value)

const inspectFields = computed(() => {
  const img = selectedImage.value
  if (!img) return []
  return [
    { label: t('pages.images.inspect.id'), value: img.id },
    { label: t('pages.images.inspect.tags'), value: img.repoTags?.join(', ') || '-' },
    { label: t('pages.images.inspect.digests'), value: img.repoDigests?.join(', ') || '-' },
    { label: t('pages.images.inspect.size'), value: formatSize(img.size) },
    { label: t('pages.images.inspect.created'), value: formatDateTime(img.createdAt) },
    { label: t('pages.images.inspect.cachedAt'), value: formatDateTime(img.cachedAt) },
    { label: 'Labels', value: Object.keys(img.labels || {}).join(', ') || '-' },
    { label: 'Machine', value: img.machine },
  ]
})

const handleInspect = (image: Image) => {
  selectedImage.value = image as ImageWithMachine
  showInspect.value = true
}

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

const cancelConfirm = () => {
  if (confirmBusy.value) return
  confirmState.value = null
}

const openDeleteConfirm = (image: Image) => {
  confirmState.value = {
    title: t('pages.images.delete.title') || t('pages.images.delete.confirm'),
    message: t('pages.images.delete.confirm'),
    confirmText: t('pages.images.delete.delete') || t('pages.images.delete.confirm'),
    danger: true,
    action: 'delete',
    id: image.id,
  }
}

const openPruneConfirm = () => {
  confirmState.value = {
    title: t('pages.images.prune.title') || 'Prune Images',
    message: t('pages.images.prune.confirm'),
    confirmText: t('pages.images.prune.confirm'),
    danger: true,
    action: 'prune',
  }
}

const confirmAction = async () => {
  const state = confirmState.value
  if (!state || confirmBusy.value) return
  confirmBusy.value = true
  confirmState.value = null

  try {
    if (state.action === 'delete' && state.id) {
      await performDeleteImage(state.id)
    } else if (state.action === 'prune') {
      await performPruneImages()
    }
  } finally {
    confirmBusy.value = false
  }
}

const performDeleteImage = async (id: string) => {
  try {
    const image = images.value.find(img => img.id === id)
    if (!image) return
    await imageService.delete(image.machineId, id)
    await fetchImages()
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  }
}

const performPruneImages = async () => {
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

const confirmDelete = async (image: Image) => {
  openDeleteConfirm(image)
}

const handlePrune = async () => {
  openPruneConfirm()
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

  .item-list-row {
    grid-template-columns: 1fr;
    gap: var(--space-2);
  }
  .item-list-actions {
    justify-content: flex-end;
  }
}

</style>
