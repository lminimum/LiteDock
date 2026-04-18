import { defineStore } from "pinia";
import { ref, computed } from "vue";
import api from "@/utils/api";

export interface User {
  id: string;
  username: string;
  email?: string;
  role: string;
}

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(localStorage.getItem("litedock-token"));

  const isAuthenticated = computed(() => !!token.value && !!user.value);

  const login = async (credentials: { username: string; password: string }) => {
    // Check for default admin credentials
    if (
      credentials.username === "admin" &&
      credentials.password === "admin123"
    ) {
      // Set up default admin user
      user.value = {
        id: "1",
        username: "admin",
        role: "admin",
      };
      token.value = "default-admin-token";
      localStorage.setItem("litedock-token", "default-admin-token");
      localStorage.setItem("litedock-user", JSON.stringify(user.value));

      return { success: true };
    }

    try {
      const response = await api.post("/auth/login", credentials);
      const { token: authToken, user: userData } = response.data;

      token.value = authToken;
      user.value = userData;

      localStorage.setItem("litedock-token", authToken);

      return { success: true };
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || "登录失败",
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
      // Check if we have stored default admin credentials
      const storedToken = localStorage.getItem("litedock-token");
      const storedUser = localStorage.getItem("litedock-user");

      if (storedToken === "default-admin-token" && storedUser) {
        token.value = storedToken;
        user.value = JSON.parse(storedUser);
        return true;
      }
      return false;
    }

    try {
      const response = await api.get("/auth/me");
      user.value = response.data;
      return true;
    } catch (error) {
      // If using default token, restore it
      if (token.value === "default-admin-token") {
        const storedUser = localStorage.getItem("litedock-user");
        if (storedUser) {
          user.value = JSON.parse(storedUser);
          return true;
        }
      }
      logout();
      return false;
    }
  };

  return {
    user,
    token,
    isAuthenticated,
    login,
    logout,
    checkAuth,
  };
});

