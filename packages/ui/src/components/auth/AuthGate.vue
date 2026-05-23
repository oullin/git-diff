<script setup lang="ts">
import { storeToRefs } from "pinia";
import type { AuthLoginResponse } from "@git-diff/contracts";
import AuthLogin from "@entry/components/AuthLogin.vue";
import AuthSetup from "@entry/components/AuthSetup.vue";
import { useAuthStore } from "@/stores/auth.store";

// Cross-domain follow-up actions (hydrating preferences, opening the last
// repo, replaying the launch intent) live on the host because they span
// multiple stores -- the host listens to the @entered / @wiped events and
// runs its own orchestration.
const emit = defineEmits<{
    entered: [response: AuthLoginResponse];
    wiped: [];
}>();

const authStore = useAuthStore();
const { mode, osUsername } = storeToRefs(authStore);

function onCompleted(response: AuthLoginResponse): void {
    authStore.complete(response);
    emit("entered", response);
}

function onWiped(): void {
    authStore.markWiped();
    emit("wiped");
}
</script>

<template>
    <div
        v-if="mode === 'loading'"
        class="grid h-screen place-items-center bg-background text-sm text-muted-foreground"
    >
        Loading…
    </div>
    <AuthSetup v-else-if="mode === 'setup'" :os-username="osUsername" @completed="onCompleted" />
    <AuthLogin
        v-else-if="mode === 'login'"
        :os-username="osUsername"
        @logged-in="onCompleted"
        @wiped="onWiped"
    />
    <slot v-else />
</template>
