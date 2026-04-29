<template>
  <PagePlaceholder
    :title="t('pages.images.title')"
    :description="t('pages.images.description')"
    :features-title="t('pages.images.features')"
    :features="features"
    :icon="ImageIcon"
  />
</template>

<script setup lang="ts">
import { Image as ImageIcon } from 'lucide-vue-next'
import { t } from '@/i18n'
import PagePlaceholder from '@/components/ui/PagePlaceholder.vue'

const features = [
  t('pages.images.list'),
  t('pages.images.pull'),
  t('pages.images.build'),
  t('pages.images.config'),
  t('pages.images.cleanup'),
]
</script>
