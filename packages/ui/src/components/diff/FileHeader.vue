<script setup lang="ts">
import { computed } from "vue";
import { Check, ChevronDown, ChevronRight, Copy, Eye } from "lucide-vue-next";
import StatusBadge from "@diff/StatusBadge.vue";
import DiffStat from "@diff/DiffStat.vue";
import { isMarkdownPath } from "@lib/markdownPreview";
import type { ChangedFile, DiffSection } from "@git-diff/contracts";

const props = defineProps<{
    file: ChangedFile;
    collapsed: boolean;
    viewed: boolean;
    previewing?: boolean;
}>();

const emit = defineEmits<{
    "toggle-collapsed": [];
    "toggle-viewed": [];
    "toggle-preview": [];
    copy: [path: string];
}>();

const parts = computed(() => props.file.path.split("/"));
const hasUnstaged = computed(() =>
    props.file.sections.some((s: DiffSection) => s.kind === "unstaged" || s.kind === "untracked"),
);
const canPreview = computed(
    () => isMarkdownPath(props.file.path) && props.file.status !== "deleted",
);
</script>

<template>
    <div
        class="flex items-center sticky top-0 z-[5]"
        :style="{
            gap: '10px',
            padding: '10px 16px',
            height: '52px',
            backgroundColor: 'var(--gd-panel-2)',
            backgroundImage:
                'linear-gradient(180deg, rgb(255 255 255 / 0.025), rgb(255 255 255 / 0) 40%)',
            borderBottom: '1px solid var(--gd-border)',
            boxShadow: '0 1px 0 var(--gd-edge-hi) inset',
        }"
    >
        <button
            type="button"
            :style="{
                width: '26px',
                height: '26px',
                borderRadius: '6px',
                background: 'transparent',
                border: '1px solid transparent',
                color: 'var(--gd-text-3)',
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                padding: 0,
                cursor: 'pointer',
            }"
            @click="emit('toggle-collapsed')"
        >
            <ChevronRight v-if="collapsed" :size="14" />
            <ChevronDown v-else :size="14" />
        </button>
        <StatusBadge :status="file.status" />
        <div
            class="flex items-center min-w-0"
            :style="{ gap: '2px', fontFamily: 'var(--font-mono)', fontSize: '13.5px' }"
        >
            <template v-for="(part, i) in parts" :key="i">
                <span v-if="i > 0" :style="{ color: 'var(--gd-text-muted)', padding: '0 2px' }"
                    >/</span
                >
                <span
                    :style="{
                        color: i === parts.length - 1 ? 'var(--gd-text)' : 'var(--gd-text-3)',
                        fontWeight: i === parts.length - 1 ? 600 : 400,
                        whiteSpace: 'nowrap',
                    }"
                    >{{ part }}</span
                >
            </template>
        </div>
        <span
            v-if="hasUnstaged"
            :style="{
                fontSize: '10.5px',
                fontWeight: 600,
                letterSpacing: '0.4px',
                textTransform: 'uppercase',
                color: 'var(--attention-fg)',
                padding: '2px 6px',
                background: 'var(--attention-subtle)',
                border: '1px solid transparent',
                borderRadius: '999px',
            }"
            >Unstaged</span
        >

        <div class="flex-1" />

        <DiffStat :add="file.additions" :del="file.deletions" squares />
        <button
            v-if="canPreview"
            type="button"
            :style="{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                height: '28px',
                padding: '0 10px',
                borderRadius: '7px',
                background: previewing ? 'var(--gd-accent-soft)' : 'var(--gd-panel-2)',
                border: '1px solid',
                borderColor: previewing ? 'var(--gd-accent-strong)' : 'var(--gd-border)',
                color: previewing ? 'var(--gd-accent)' : 'var(--gd-text-2)',
                fontSize: '13px',
                fontWeight: 500,
                cursor: 'pointer',
            }"
            :title="previewing ? 'Show diff' : 'Preview rendered markdown'"
            @click="emit('toggle-preview')"
        >
            <Eye :size="13" />
            {{ previewing ? "Diff" : "Preview" }}
        </button>
        <button
            type="button"
            :style="{
                height: '26px',
                padding: '0 8px',
                borderRadius: '8px',
                border: '1px solid transparent',
                background: 'transparent',
                color: 'var(--gd-text-2)',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '4px',
                cursor: 'pointer',
            }"
            title="Copy path"
            @click="emit('copy', file.path)"
        >
            <Copy :size="14" />
        </button>
        <button
            type="button"
            :style="{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                height: '28px',
                padding: '0 10px',
                borderRadius: '7px',
                background: 'var(--gd-panel-2)',
                border: '1px solid var(--gd-border)',
                color: viewed ? 'var(--gd-text-3)' : 'var(--gd-text-2)',
                fontSize: '13px',
                fontWeight: 500,
                cursor: 'pointer',
            }"
            @click="emit('toggle-viewed')"
        >
            <span
                :style="{
                    width: '14px',
                    height: '14px',
                    borderRadius: '3px',
                    background: viewed ? 'var(--success-emphasis)' : 'transparent',
                    border: viewed ? 'none' : '1.5px solid var(--gd-border)',
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: '#ffffff',
                }"
            >
                <Check v-if="viewed" :size="10" :stroke-width="3.5" />
            </span>
            {{ viewed ? "Viewed" : "Mark viewed" }}
        </button>
    </div>
</template>
