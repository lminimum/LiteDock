<template>
  <div class="page-placeholder">
    <div class="placeholder-content">
      <div class="placeholder-card">
        <div class="placeholder-icon">
          <GitBranch :size="48" :stroke-width="1" />
        </div>
        <h2>{{ t('pages.orchestration.title') }}</h2>
        <p>{{ t('pages.orchestration.description') }}</p>
        <div class="placeholder-features">
          <h3>{{ t('pages.orchestration.features') }}</h3>
          <ul>
            <li v-for="feature in features" :key="feature">
              <Check :size="14" :stroke-width="2" class="check-icon" />
              <span>{{ feature }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GitBranch, Check } from 'lucide-vue-next'
import { t } from '@/i18n'

const features = [
  t('pages.orchestration.compose'),
  t('pages.orchestration.multiContainer'),
  t('pages.orchestration.kubernetes'),
  t('pages.orchestration.discovery'),
  t('pages.orchestration.config'),
]
</script>

<style scoped>
.page-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  padding: var(--space-6);
}

.placeholder-content {
  max-width: 600px;
  width: 100%;
}

.placeholder-card {
  text-align: center;
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  padding: var(--space-10) var(--space-8);
}

.placeholder-icon {
  color: var(--color-text-weaker);
  margin-bottom: var(--space-6);
}

.placeholder-card h2 {
  font-family: var(--font-mono);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  margin-bottom: var(--space-3);
}

.placeholder-card p {
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text-weak);
  margin-bottom: var(--space-8);
}

.placeholder-features {
  text-align: left;
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  padding: var(--space-6);
}

.placeholder-features h3 {
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-strong);
  margin-bottom: var(--space-4);
}

.placeholder-features ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.placeholder-features li {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text);
}

.check-icon {
  color: var(--color-success);
  flex-shrink: 0;
}

@media (max-width: 767px) {
  .page-placeholder {
    padding: var(--space-4);
    min-height: 50vh;
  }

  .placeholder-card {
    padding: var(--space-6) var(--space-4);
  }

  .placeholder-card h2 {
    font-size: var(--font-size-xl);
  }

  .placeholder-features {
    padding: var(--space-4);
  }
}
</style>
