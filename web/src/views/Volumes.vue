<template>
  <div class="page-placeholder">
    <div class="placeholder-content">
      <div class="placeholder-card">
        <div class="placeholder-icon">
          <HardDrive :size="48" :stroke-width="1" />
        </div>
        <h2>存储卷管理</h2>
        <p>Docker 存储卷管理功能正在开发中...</p>
        <div class="placeholder-features">
          <h3>即将支持的功能:</h3>
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
import { HardDrive, Check } from 'lucide-vue-next'

const features = [
  '存储卷列表和创建',
  '卷挂载管理',
  '数据备份和恢复',
  '存储监控',
  '存储驱动配置'
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
