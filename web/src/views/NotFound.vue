<template>
  <div class="not-found">
    <div class="not-found-content">
      <div class="error-code">404</div>
      <h1>{{ t('pages.notFound.title') }}</h1>
      <p>{{ t('pages.notFound.description') }}</p>
      <div class="not-found-actions">
        <button @click="goBack" class="btn btn-secondary">
          <ArrowLeft :size="18" />
          {{ t('pages.notFound.back') }}
        </button>
        <router-link to="/" class="btn btn-primary">
          <Home :size="18" />
          {{ t('pages.notFound.home') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ArrowLeft, Home } from 'lucide-vue-next'
import { t } from '@/i18n'

const router = useRouter()

const goBack = () => {
  router.go(-1)
}
</script>

<style scoped>
.not-found {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  text-align: center;
}

.not-found-content {
  max-width: 500px;
}

.error-code {
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-weaker);
  margin-bottom: var(--space-4);
}

.not-found-content h1 {
  color: var(--color-text-strong);
  margin-bottom: var(--space-3);
  font-size: var(--font-size-2xl);
}

.not-found-content p {
  color: var(--color-text);
  margin-bottom: var(--space-8);
}

.not-found-actions {
  display: flex;
  gap: var(--space-4);
  justify-content: center;
}

@media (max-width: 767px) {
  .not-found {
    padding: var(--space-4);
    min-height: 50vh;
  }

  .not-found-content {
    max-width: 100%;
  }

  .error-code {
    font-size: var(--font-size-3xl);
    margin-bottom: var(--space-4);
  }

  .not-found-actions {
    flex-direction: column;
  }
}
</style>