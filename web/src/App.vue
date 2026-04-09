<template>
  <div id="app">
    <router-view />
    <SetupModal 
      :show="showSetupModal"
      @close="handleSetupClose"
      @complete="handleSetupComplete"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import SetupModal from '@/components/SetupModal.vue'

const authStore = useAuthStore()
const router = useRouter()

const showSetupModal = ref(false)

// Check if the system has been configured
const checkConfiguration = () => {
  const isConfigured = localStorage.getItem('litedock-configured') === 'true'
  showSetupModal.value = !isConfigured
}

// Handle when setup is closed
const handleSetupClose = () => {
  showSetupModal.value = false
}

// Handle when setup is completed successfully
const handleSetupComplete = () => {
  showSetupModal.value = false
  // The SetupModal already handles routing to login
}

onMounted(() => {
  // Check configuration status on mount
  checkConfiguration()
  
  // Check authentication status
  authStore.checkAuth()
})

// Also watch for route changes that might affect configuration status
watch(
  () => router.currentRoute.value,
  () => {
    checkConfiguration()
  }
)
</script>

<style>
#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  height: 100vh;
  margin: 0;
  padding: 0;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
}
</style>