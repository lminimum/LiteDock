<template>
  <PagePlaceholder
    :title="t('pages.orchestration.title')"
    :description="t('pages.orchestration.description')"
    :features-title="t('pages.orchestration.features')"
    :features="features"
    :icon="GitBranch"
  />
</template>

<script setup lang="ts">
import { GitBranch } from 'lucide-vue-next'
import { t } from '@/i18n'
import PagePlaceholder from '@/components/ui/PagePlaceholder.vue'

const features = [
  t('pages.orchestration.compose'),
  t('pages.orchestration.multiContainer'),
  t('pages.orchestration.kubernetes'),
  t('pages.orchestration.discovery'),
  t('pages.orchestration.config'),
]
</script>
