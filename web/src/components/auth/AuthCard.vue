<template>
  <div class="auth-container">
    <div class="auth-card">
      <div v-if="$slots.header" class="auth-header">
        <slot name="header" />
      </div>

      <div v-if="$slots.default" class="auth-content">
        <slot />
      </div>

      <div v-if="$slots.footer" class="auth-footer">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.auth-container {
  min-height: 100vh;
  background: var(--color-background);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
}

.auth-card {
  background: var(--color-background);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  max-width: 400px;
  width: 100%;
  padding: var(--space-8);
}

.auth-header {
  text-align: center;
  margin-bottom: var(--space-6);
}

.auth-content {
  display: flex;
  flex-direction: column;
}

.auth-footer {
  margin-top: var(--space-6);
  text-align: center;
  padding-top: var(--space-6);
  border-top: 1px solid var(--color-border-subtle);
}

@media (max-width: 480px) {
  .auth-card {
    padding: var(--space-6);
    max-width: 100%;
  }
}
</style>
