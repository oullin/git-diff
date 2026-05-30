<script setup lang="ts">
import { ref } from "vue";
import { ArrowRight, Eye, EyeOff, Lock, RotateCcw } from "lucide-vue-next";
import type { AuthLoginResponse } from "@git-diff/contracts";
import DiffLogo from "@components/diff/DiffLogo.vue";
import { useAuthForm } from "@composables/useAuthForm";
import { services } from "@lib/services";
import { Input } from "@ui/input";
import { Label } from "@ui/label";

const props = defineProps<{
    osUsername: string;
}>();

const emit = defineEmits<{
    (event: "logged-in", response: AuthLoginResponse): void;
    (event: "wiped"): void;
}>();

const password = ref("");
const remember = ref(true);
const showPassword = ref(false);
const { submitting, error, submit: runSubmit, fail } = useAuthForm();

async function submit() {
    if (!password.value) {
        fail("Enter your password.");

        return;
    }

    const response = await runSubmit(() => services().auth.login(password.value, remember.value));

    if (response) {
        emit("logged-in", response);
    }
}

async function forgotPassword() {
    const confirmed = window.confirm(
        `Reset ${props.osUsername}? This deletes the saved password, preferences, and review data for this user. This cannot be undone.`,
    );

    if (!confirmed) {
        return;
    }

    const wiped = await runSubmit(async () => {
        await services().auth.wipe(props.osUsername);

        return true;
    });

    if (wiped) {
        emit("wiped");
    }
}
</script>

<template>
    <div
        class="grid min-h-screen lg:grid-cols-2"
        :style="{ background: 'var(--gd-bg-canvas)', color: 'var(--gd-text)' }"
    >
        <div class="flex flex-col gap-6 p-6 md:p-10">
            <div class="flex items-center gap-2.5">
                <DiffLogo :size="28" />
                <span
                    class="text-[14px] font-semibold tracking-tight"
                    :style="{ color: 'var(--gd-text)' }"
                >
                    Git Diff
                </span>
            </div>

            <div class="flex flex-1 items-center justify-center">
                <div
                    class="w-full max-w-sm rounded-2xl border p-7 md:p-8"
                    :style="{
                        background: 'var(--gd-panel)',
                        borderColor: 'var(--gd-border)',
                        boxShadow: 'var(--gd-shadow-lg)',
                    }"
                >
                    <form class="flex flex-col gap-6" @submit.prevent="submit">
                        <div class="flex flex-col gap-1.5">
                            <h1
                                class="text-[22px] font-semibold tracking-tight"
                                :style="{ color: 'var(--gd-text)' }"
                            >
                                Welcome back
                            </h1>
                            <p
                                class="text-[13.5px] leading-relaxed"
                                :style="{ color: 'var(--gd-text-3)' }"
                            >
                                Sign in as
                                <span class="font-medium" :style="{ color: 'var(--gd-text-2)' }">{{
                                    props.osUsername
                                }}</span>
                                to unlock your local data.
                            </p>
                        </div>

                        <div class="flex flex-col gap-4">
                            <div class="grid gap-2">
                                <Label for="password" :style="{ color: 'var(--gd-text-2)' }"
                                    >Password</Label
                                >
                                <div class="relative">
                                    <Lock
                                        class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
                                        :style="{ color: 'var(--gd-text-3)' }"
                                    />
                                    <Input
                                        id="password"
                                        v-model="password"
                                        :type="showPassword ? 'text' : 'password'"
                                        autocomplete="current-password"
                                        class="pl-9 pr-9"
                                        :disabled="submitting"
                                        required
                                    />
                                    <button
                                        type="button"
                                        class="absolute right-2 top-1/2 inline-flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-md transition-colors hover:bg-[color:var(--gd-hover)] disabled:opacity-50"
                                        :style="{ color: 'var(--gd-text-3)' }"
                                        :aria-label="
                                            showPassword ? 'Hide password' : 'Show password'
                                        "
                                        :disabled="submitting"
                                        @click="showPassword = !showPassword"
                                    >
                                        <EyeOff v-if="showPassword" class="h-4 w-4" />
                                        <Eye v-else class="h-4 w-4" />
                                    </button>
                                </div>
                            </div>

                            <label
                                class="flex items-center gap-2 text-[13px]"
                                :style="{ color: 'var(--gd-text-3)' }"
                            >
                                <input
                                    v-model="remember"
                                    type="checkbox"
                                    class="h-3.5 w-3.5 rounded"
                                    :style="{
                                        accentColor: 'var(--gd-accent)',
                                        borderColor: 'var(--gd-border-strong)',
                                    }"
                                    :disabled="submitting"
                                />
                                Remember me on this device
                            </label>

                            <p
                                v-if="error"
                                class="text-[13px] leading-relaxed"
                                :style="{ color: 'var(--gd-removed)' }"
                            >
                                {{ error }}
                            </p>

                            <button
                                type="submit"
                                class="mt-1 inline-flex h-10 w-full items-center justify-center gap-2 rounded-md text-[13.5px] font-semibold tracking-tight transition-opacity duration-150 hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
                                :style="{
                                    background:
                                        'linear-gradient(135deg, var(--gd-accent-strong), var(--gd-accent))',
                                    color: '#fff',
                                    boxShadow:
                                        '0 0 0 1px rgba(255,255,255,0.06) inset, 0 6px 16px -4px var(--gd-accent-soft)',
                                }"
                                :disabled="submitting"
                            >
                                <span>{{ submitting ? "Signing in…" : "Sign in" }}</span>
                                <ArrowRight v-if="!submitting" class="h-4 w-4" />
                            </button>
                        </div>

                        <div class="text-center">
                            <button
                                type="button"
                                class="inline-flex items-center gap-1.5 text-[12px] underline-offset-4 transition-colors hover:underline disabled:opacity-50"
                                :style="{ color: 'var(--gd-text-muted)' }"
                                :disabled="submitting"
                                @click="forgotPassword"
                            >
                                <RotateCcw class="h-3 w-3" />
                                Forgot password? Reset this user
                            </button>
                        </div>
                    </form>
                </div>
            </div>

            <p class="text-[11.5px]" :style="{ color: 'var(--gd-text-muted)' }">
                Local-first. Your data never leaves this device.
            </p>
        </div>

        <div
            class="relative hidden overflow-hidden lg:block"
            :style="{ background: 'var(--gd-bg)' }"
        >
            <div
                class="absolute inset-0"
                :style="{
                    backgroundImage:
                        'radial-gradient(circle at center, var(--gd-border-strong) 1px, transparent 1px)',
                    backgroundSize: '24px 24px',
                    opacity: 0.4,
                }"
            />
            <div
                class="absolute inset-0"
                :style="{
                    background:
                        'radial-gradient(ellipse at 70% 30%, var(--gd-accent-soft), transparent 60%)',
                }"
            />

            <div class="absolute inset-0 flex items-center justify-center p-12">
                <div
                    class="w-full max-w-md rounded-xl border font-mono text-[12px]"
                    :style="{
                        background: 'var(--gd-panel-2)',
                        borderColor: 'var(--gd-border)',
                        boxShadow: 'var(--gd-shadow-lg)',
                    }"
                >
                    <div
                        class="flex items-center gap-2 border-b px-4 py-2.5"
                        :style="{ borderColor: 'var(--gd-border-soft)' }"
                    >
                        <span class="window-dots">
                            <span />
                            <span />
                            <span />
                        </span>
                        <span class="ml-1" :style="{ color: 'var(--gd-text-3)' }">repo.go</span>
                    </div>
                    <pre
                        class="px-4 py-3 leading-relaxed"
                        :style="{ color: 'var(--gd-text-2)' }"
                    ><span :style="{ color: 'var(--gd-text-muted)' }">  1</span>  func (r *<span :style="{ color: 'var(--gd-text)' }">Repo</span>) <span :style="{ color: 'var(--gd-text)' }">Open</span>(path string) error {
<span :style="{ color: 'var(--gd-text-muted)' }">  2</span>    <span class="rounded px-1" :style="{ background: 'var(--diff-removed-bg)', color: 'var(--gd-removed)' }">- return exec.Run(&quot;git&quot;, &quot;open&quot;, path)</span>
<span :style="{ color: 'var(--gd-text-muted)' }">  3</span>    <span class="rounded px-1" :style="{ background: 'var(--diff-added-bg)', color: 'var(--gd-added)' }">+ return r.git.PlainOpen(path)</span>
<span :style="{ color: 'var(--gd-text-muted)' }">  4</span>  }</pre>
                </div>
            </div>
        </div>
    </div>
</template>
