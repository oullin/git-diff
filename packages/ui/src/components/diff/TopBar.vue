<script setup lang="ts">
import { RefreshCw, Settings2 } from "lucide-vue-next";
import { Popover, PopoverContent, PopoverTrigger } from "@ui/popover";
import BranchPicker from "@components/diff/BranchPicker.vue";
import DiffStat from "@components/diff/DiffStat.vue";
import FileSearchPopover from "@components/diff/FileSearchPopover.vue";
import Kbd from "@components/diff/Kbd.vue";
import TweaksPanel from "@components/diff/TweaksPanel.vue";
import UserMenu from "@components/diff/UserMenu.vue";
import type { Tweaks } from "@composables/useTweaks";
import type { AuthUser, FileSearchResult, RepositoryState } from "@git-diff/contracts";

defineProps<{
    state: RepositoryState | null;
    currentUser: AuthUser | null;
    userInitials: string;
    tweaks: Tweaks;
    creatingBranch: boolean;
    branchCreateError: string;
}>();

const emit = defineEmits<{
    "update:tweak": [key: keyof Tweaks, value: Tweaks[keyof Tweaks]];
    refresh: [];
    "log-out": [];
    "switch-branch": [branch: string];
    "create-branch": [name: string];
    "select-result": [result: FileSearchResult];
}>();
</script>

<template>
    <div
        class="flex items-center"
        :style="{
            height: '56px',
            flexShrink: 0,
            gap: '10px',
            padding: '0 14px',
            borderBottom: '1px solid var(--gd-border)',
            background: 'var(--gd-bg)',
        }"
    >
        <BranchPicker
            :state="state"
            :creating-branch="creatingBranch"
            :branch-create-error="branchCreateError"
            @switch-branch="(branch) => emit('switch-branch', branch)"
            @create-branch="(name) => emit('create-branch', name)"
        />

        <div class="flex items-center" :style="{ gap: '8px', paddingLeft: '4px' }">
            <span
                :style="{
                    fontSize: '11.5px',
                    fontFamily: 'var(--font-mono)',
                    color: 'var(--gd-text-3)',
                    padding: '3px 7px',
                    background: 'var(--gd-panel-2)',
                    borderRadius: '5px',
                    border: '1px solid var(--gd-border)',
                }"
                >{{ state?.headSha?.slice(0, 8) || "—" }}</span
            >
            <span
                v-if="state"
                :style="{
                    fontSize: '13px',
                    color: 'var(--gd-text-2)',
                    maxWidth: '320px',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                }"
                >{{ state.files.length }} changed file{{
                    state.files.length === 1 ? "" : "s"
                }}</span
            >
            <DiffStat v-if="state" :add="state.additions" :del="state.deletions" />
        </div>

        <div
            :style="{
                width: '1px',
                height: '22px',
                background: 'var(--gd-border)',
                margin: '0 6px',
            }"
        />

        <FileSearchPopover @select-result="(result) => emit('select-result', result)" />

        <button
            type="button"
            class="inline-flex items-center"
            :style="{
                gap: '6px',
                height: '30px',
                padding: '0 10px',
                borderRadius: '8px',
                border: '1px solid transparent',
                background: 'transparent',
                color: 'var(--gd-text-2)',
                fontSize: '13.5px',
                fontWeight: 500,
                cursor: 'pointer',
            }"
            :disabled="!state"
            @click="emit('refresh')"
        >
            <RefreshCw :size="14" />
            Refresh
            <Kbd>R</Kbd>
        </button>

        <Popover>
            <PopoverTrigger as-child>
                <button
                    type="button"
                    class="inline-flex items-center"
                    :style="{
                        gap: '6px',
                        height: '30px',
                        padding: '0 10px',
                        borderRadius: '8px',
                        border: '1px solid transparent',
                        background: 'transparent',
                        color: 'var(--gd-text-2)',
                        fontSize: '13.5px',
                        fontWeight: 500,
                        whiteSpace: 'nowrap',
                        cursor: 'pointer',
                    }"
                >
                    <Settings2 :size="14" />
                    Tweaks
                </button>
            </PopoverTrigger>
            <PopoverContent
                align="end"
                :side-offset="8"
                class="w-[280px] p-0 border-0 shadow-none bg-transparent"
            >
                <div
                    :style="{
                        width: '280px',
                        background: 'var(--gd-panel)',
                        border: '1px solid var(--gd-border)',
                        borderRadius: '10px',
                        boxShadow: 'var(--gd-shadow-lg)',
                        color: 'var(--gd-text)',
                        fontFamily: 'var(--font-sans)',
                        fontSize: '13px',
                        overflow: 'hidden',
                    }"
                >
                    <div
                        :style="{
                            padding: '10px 12px',
                            borderBottom: '1px solid var(--gd-border)',
                            fontSize: '13px',
                            fontWeight: 600,
                            color: 'var(--gd-text)',
                        }"
                    >
                        Tweaks
                    </div>
                    <div :style="{ padding: '12px' }">
                        <TweaksPanel
                            :tweaks="tweaks"
                            @update:tweak="(key, value) => emit('update:tweak', key, value)"
                        />
                    </div>
                </div>
            </PopoverContent>
        </Popover>

        <div
            :style="{
                width: '1px',
                height: '22px',
                background: 'var(--gd-border)',
                margin: '0 2px',
            }"
        />

        <UserMenu
            :current-user="currentUser"
            :user-initials="userInitials"
            @log-out="emit('log-out')"
        />
    </div>
</template>
