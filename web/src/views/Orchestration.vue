<template>
  <div class="page-placeholder">
    <div class="placeholder-content">
      <div class="placeholder-icon">🔀</div>
      <h2>编排管理</h2>
      <p>Docker Compose 和 Kubernetes 编排管理功能正在开发中...</p>
      <div class="placeholder-features">
        <h3>即将支持的功能:</h3>
        <ul>
          <li>Docker Compose 项目管理</li>
          <li>多容器服务编排</li>
          <li>Kubernetes 集群集成</li>
          <li>服务发现和负载均衡</li>
          <li>配置管理和密钥管理</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// 编排管理页面 - 占位符
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
</style>