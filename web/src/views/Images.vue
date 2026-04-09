<template>
  <div class="page-placeholder">
    <div class="placeholder-content">
      <div class="placeholder-icon">🖼️</div>
      <h2>镜像管理</h2>
      <p>Docker 镜像管理功能正在开发中...</p>
      <div class="placeholder-features">
        <h3>即将支持的功能:</h3>
        <ul>
          <li>镜像列表和搜索</li>
          <li>拉取和推送镜像</li>
          <li>镜像构建管理</li>
          <li>镜像仓库配置</li>
          <li>镜像清理和优化</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// 镜像管理页面 - 占位符
</script>

<style scoped>
.page-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  text-align: center;
}

.placeholder-content {
  max-width: 600px;
}

.placeholder-icon {
  font-size: 4rem;
  margin-bottom: 20px;
}

.placeholder-content h2 {
  color: #1e293b;
  margin-bottom: 12px;
}

.placeholder-content p {
  color: #64748b;
  margin-bottom: 32px;
}

.placeholder-features {
  text-align: left;
  background: white;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.placeholder-features h3 {
  color: #1e293b;
  margin-bottom: 16px;
}

.placeholder-features ul {
  list-style: none;
  padding: 0;
}

.placeholder-features li {
  padding: 8px 0;
  color: #374151;
  position: relative;
  padding-left: 20px;
}

.placeholder-features li::before {
  content: '✓';
  position: absolute;
  left: 0;
  color: #10b981;
  font-weight: bold;
}

@media (max-width: 767px) {
  .page-placeholder {
    padding: 16px;
    min-height: 50vh;
  }

  .placeholder-icon {
    font-size: 3rem;
    margin-bottom: 16px;
  }

  .placeholder-content h2 {
    font-size: 1.5rem;
  }

  .placeholder-content p {
    font-size: 0.875rem;
  }

  .placeholder-features {
    padding: 20px;
  }

  .placeholder-features h3 {
    font-size: 1.125rem;
  }

  .placeholder-features li {
    font-size: 0.875rem;
  }
}
</style>