<template>
  <PagePlaceholder
    :title="t('pages.volumes.title')"
    :description="t('pages.volumes.description')"
    :features-title="t('pages.volumes.features')"
    :features="features"
    :icon="HardDrive"
  />
</template>

<script setup lang="ts">
import { HardDrive } from 'lucide-vue-next'
import { t } from '@/i18n'
import PagePlaceholder from '@/components/ui/PagePlaceholder.vue'

const features = [
  t('pages.volumes.list'),
  t('pages.volumes.mount'),
  t('pages.volumes.backup'),
  t('pages.volumes.monitor'),
  t('pages.volumes.driver'),
]
</script>
