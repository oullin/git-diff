<script setup lang="ts">
import { ref } from "vue";
import { GitPullRequest } from "lucide-vue-next";
import type { AuthLoginResponse } from "@api";

const props = defineProps<{
  osUsername: string;
}>();

const emit = defineEmits<{
  (event: "logged-in", response: AuthLoginResponse): void;
  (event: "wiped"): void;
}>();

const password = ref("");
const remember = ref(true);
const submitting = ref(false);
const error = ref("");

async function submit() {
  error.value = "";

  if (!password.value) {
    error.value = "Enter your password.";

    return;
  }

  submitting.value = true;

  try {
    const response = await window.diffApp.authLogin(password.value, remember.value);
    emit("logged-in", response);
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
  } finally {
    submitting.value = false;
  }
}

async function forgotPassword() {
  const confirmed = window.confirm(
    `Reset ${props.osUsername}? This deletes the saved password, preferences, and review data for this user. This cannot be undone.`,
  );

  if (!confirmed) {
    return;
  }

  submitting.value = true;

  try {
    await window.diffApp.authWipe(props.osUsername);
    emit("wiped");
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="flex h-screen items-center justify-center bg-background p-6 text-foreground">
    <form
      class="w-full max-w-sm space-y-5 rounded-md border border-border bg-section p-6"
      @submit.prevent="submit"
    >
      <div class="flex items-center gap-2">
        <GitPullRequest class="h-5 w-5 text-muted-foreground" />
        <span class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Git Diff
        </span>
      </div>
      <div>
        <h1 class="text-lg font-semibold">Sign in as {{ props.osUsername }}</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Enter your password to unlock your local data.
        </p>
      </div>

      <label class="block">
        <span class="text-xs font-medium text-muted-foreground">Password</span>
        <input
          v-model="password"
          type="password"
          autocomplete="current-password"
          class="mt-1 h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring"
          :disabled="submitting"
        />
      </label>

      <label class="flex items-center gap-2 text-sm text-muted-foreground">
        <input v-model="remember" type="checkbox" :disabled="submitting" />
        Remember me on this device
      </label>

      <div v-if="error" class="text-sm text-destructive">{{ error }}</div>

      <button type="submit" class="toolbar-btn w-full justify-center" :disabled="submitting">
        {{ submitting ? "Signing in…" : "Sign in" }}
      </button>

      <button
        type="button"
        class="block w-full text-center text-xs text-muted-foreground hover:text-foreground"
        :disabled="submitting"
        @click="forgotPassword"
      >
        Forgot password? Reset this user
      </button>
    </form>
  </div>
</template>
