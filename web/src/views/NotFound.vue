<template>
  <div class="not-found">
    <div class="not-found-content">
      <div class="error-code">404</div>
      <h1>页面未找到</h1>
      <p>抱歉，您访问的页面不存在或已被移动。</p>
      <div class="not-found-actions">
        <button @click="goBack" class="btn-secondary">
          ← 返回上页
        </button>
        <router-link to="/" class="btn-primary">
          🏠 返回首页
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

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
  font-size: 6rem;
  font-weight: 700;
  color: #cbd5e1;
  margin-bottom: 20px;
}

.not-found-content h1 {
  color: #1e293b;
  margin-bottom: 12px;
  font-size: 2rem;
}

.not-found-content p {
  color: #64748b;
  margin-bottom: 32px;
}

.not-found-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.btn-primary,
.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5a67d8;
}

.btn-secondary {
  background: #f8fafc;
  color: #64748b;
  border: 1px solid #d1d5db;
}

.btn-secondary:hover {
  background: #f1f5f9;
}

@media (max-width: 767px) {
  .not-found {
    padding: 16px;
    min-height: 50vh;
  }

  .not-found-content {
    max-width: 100%;
  }

  .error-code {
    font-size: 4rem;
    margin-bottom: 16px;
  }

  .not-found-content h1 {
    font-size: 1.5rem;
  }

  .not-found-content p {
    font-size: 0.875rem;
    margin-bottom: 24px;
  }

  .not-found-actions {
    flex-direction: column;
  }

  .btn-primary,
  .btn-secondary {
    width: 100%;
    justify-content: center;
  }
}
</style>