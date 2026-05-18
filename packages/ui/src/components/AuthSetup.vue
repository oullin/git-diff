<script setup lang="ts">
import { ref } from "vue";
import { GitPullRequest } from "lucide-vue-next";
import type { AuthLoginResponse } from "@api";

const props = defineProps<{
  osUsername: string;
}>();

const emit = defineEmits<{
  (event: "completed", response: AuthLoginResponse): void;
}>();

const password = ref("");
const confirm = ref("");
const submitting = ref(false);
const error = ref("");

async function submit() {
  error.value = "";

  if (password.value.length < 6) {
    error.value = "Password must be at least 6 characters.";

    return;
  }

  if (password.value !== confirm.value) {
    error.value = "Passwords do not match.";

    return;
  }

  submitting.value = true;

  try {
    const response = await window.diffApp.authSetup(password.value);
    emit("completed", response);
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
        <h1 class="text-lg font-semibold">Welcome, {{ props.osUsername }}</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Set a password to protect your local preferences and reviews.
        </p>
      </div>

      <label class="block">
        <span class="text-xs font-medium text-muted-foreground">Password</span>
        <input
          v-model="password"
          type="password"
          autocomplete="new-password"
          class="mt-1 h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring"
          :disabled="submitting"
        />
      </label>

      <label class="block">
        <span class="text-xs font-medium text-muted-foreground">Confirm password</span>
        <input
          v-model="confirm"
          type="password"
          autocomplete="new-password"
          class="mt-1 h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring"
          :disabled="submitting"
        />
      </label>

      <div v-if="error" class="text-sm text-destructive">{{ error }}</div>

      <button type="submit" class="toolbar-btn w-full justify-center" :disabled="submitting">
        {{ submitting ? "Setting password…" : "Set password & continue" }}
      </button>
    </form>
  </div>
</template>
