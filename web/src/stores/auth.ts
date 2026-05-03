import { defineStore } from "pinia";
import { ref, computed } from "vue";
import api from "@/utils/api";
import { t } from "@/i18n";

export interface User {
  id: string;
  username: string;
  email?: string;
  role: string;
}

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(localStorage.getItem("litedock-token"));
  const setupComplete = ref(false);

  const isAuthenticated = computed(() => !!token.value && !!user.value);

  const checkSetupStatus = async (): Promise<boolean> => {
    try {
      const data: any = await api.get("/auth/setup-status");
      setupComplete.value = data?.setup_complete ?? false;
      return setupComplete.value;
    } catch (error) {
      console.error("Failed to check setup status:", error);
      return false;
    }
  };

  const login = async (credentials: { username: string; password: string }) => {
    try {
      const data: any = await api.post("/auth/login", credentials);
      token.value = data.token;
      user.value = data.user;
      localStorage.setItem("litedock-token", data.token);
      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.message || t("errors.loginFailed"),
      };
    }
  };

  const logout = () => {
    token.value = null;
    user.value = null;
    localStorage.removeItem("litedock-token");
    localStorage.removeItem("litedock-user");
  };

  const checkAuth = async () => {
    if (!token.value) {
      return false;
    }

    try {
      const data: any = await api.get("/auth/me");
      user.value = data;
      return true;
    } catch (error) {
      logout();
      return false;
    }
  };

  return {
    user,
    token,
    isAuthenticated,
    setupComplete,
    checkSetupStatus,
    login,
    logout,
    checkAuth,
  };
});
