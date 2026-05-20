import { defineStore } from "pinia";
import { ref } from "vue";
import type { AuthLoginResponse, AuthUser } from "@git-diff/contracts";

// useAuthStore owns the rendered-side auth session: the four-state mode
// machine, the OS-derived username the backend pins to, and the current
// authenticated user. Bootstrap drives initial discovery; the
// login/logout/wipe transitions just update local state -- the heavy
// lifting (preferences hydration, repo refresh, launch-intent replay) is
// orchestrated by App.vue because it spans multiple domains.
export type AuthMode = "loading" | "setup" | "login" | "ready";

export const useAuthStore = defineStore("auth", () => {
  const mode = ref<AuthMode>("loading");
  const osUsername = ref("");
  const currentUser = ref<AuthUser | null>(null);
  const bootstrapError = ref("");

  async function bootstrap(): Promise<{ entered: boolean }> {
    mode.value = "loading";

    try {
      const result = await window.diffApp.authBootstrap();
      osUsername.value = result.state.osUsername;

      if (result.user) {
        currentUser.value = result.user;
        mode.value = "ready";

        return { entered: true };
      }

      mode.value = result.state.needsSetup ? "setup" : "login";

      return { entered: false };
    } catch (cause) {
      bootstrapError.value = cause instanceof Error ? cause.message : String(cause);
      mode.value = "login";

      return { entered: false };
    }
  }

  function complete(response: AuthLoginResponse): void {
    currentUser.value = response.user;
    mode.value = "ready";
  }

  function markWiped(): void {
    currentUser.value = null;
    mode.value = "setup";
  }

  async function logout(): Promise<void> {
    try {
      await window.diffApp.authLogout();
    } catch {
      // ignore: we reset locally either way.
    }

    currentUser.value = null;
  }

  return {
    mode,
    osUsername,
    currentUser,
    bootstrapError,
    bootstrap,
    complete,
    markWiped,
    logout,
  };
});
