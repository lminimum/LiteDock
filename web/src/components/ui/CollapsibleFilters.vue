<template>
  <div class="collapsible-filters">
    <div class="filter-toolbar">
      <button
        class="btn btn-ghost filter-toggle"
        :class="{ active: searchOpen }"
        @click="searchOpen = !searchOpen; if (searchOpen) $nextTick(() => searchInputRef?.focus())"
        :title="searchLabel"
      >
        <Search :size="16" />
      </button>
      <button
        v-if="hasFilters"
        class="btn btn-ghost filter-toggle"
        :class="{ active: filtersOpen }"
        @click="filtersOpen = !filtersOpen"
        :title="filterLabel"
      >
        <SlidersHorizontal :size="16" />
      </button>
      <div class="filter-slot-right">
        <slot name="right" />
      </div>
    </div>
    <div v-if="searchOpen" class="filter-search-row">
      <input
        ref="searchInputRef"
        :value="modelValue"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        :placeholder="searchPlaceholder"
        type="text"
        class="input"
      />
    </div>
    <div v-if="filtersOpen" class="filter-options-row">
      <slot name="filters" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Search, SlidersHorizontal } from 'lucide-vue-next'

defineProps<{
  modelValue: string
  searchPlaceholder: string
  searchLabel?: string
  filterLabel?: string
  hasFilters?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const searchOpen = ref(false)
const filtersOpen = ref(false)
const searchInputRef = ref<HTMLInputElement | null>(null)
</script>

<style scoped>
.collapsible-filters {
  margin-bottom: var(--space-4);
}

.filter-toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.filter-toggle {
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
}

.filter-toggle.active {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
  border-color: var(--color-border);
}

.filter-slot-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.filter-search-row {
  margin-top: var(--space-2);
}

.filter-options-row {
  display: flex;
  gap: var(--space-3);
  margin-top: var(--space-2);
  flex-wrap: wrap;
}

.filter-options-row :deep(select) {
  min-width: 130px;
}
</style>