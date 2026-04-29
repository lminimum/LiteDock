<template>
  <PagePlaceholder
    :title="t('pages.networks.title')"
    :description="t('pages.networks.description')"
    :features-title="t('pages.networks.features')"
    :features="features"
    :icon="Globe"
  />
</template>

<script setup lang="ts">
import { Globe } from 'lucide-vue-next'
import { t } from '@/i18n'
import PagePlaceholder from '@/components/ui/PagePlaceholder.vue'

const features = [
  t('pages.networks.list'),
  t('pages.networks.config'),
  t('pages.networks.connect'),
  t('pages.networks.monitor'),
  t('pages.networks.policy'),
]
</script>
