<script setup lang="ts">
import { AlertCircle, Check, MessageSquare } from "lucide-vue-next";
import Kbd from "@components/diff/Kbd.vue";

const emit = defineEmits<{
    "start-review": [];
    "open-review-panel": [intent: "comment" | "request"];
}>();
</script>

<template>
    <div
        :style="{
            padding: '6px 6px 14px',
            borderTop: '1px solid var(--gd-border)',
            background: 'var(--gd-panel)',
            display: 'flex',
            flexDirection: 'column',
            gap: '14px',
        }"
    >
        <button
            type="button"
            :style="{
                height: '36px',
                borderRadius: '8px',
                border: '1px solid transparent',
                background: 'var(--gd-accent-strong)',
                color: '#fff',
                fontSize: '13.5px',
                fontWeight: 600,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                whiteSpace: 'nowrap',
                boxShadow:
                    '0 0 0 1px var(--gd-accent-strong), 0 1px 0 rgba(255,255,255,0.08) inset, 0 6px 14px -4px var(--gd-accent-soft)',
                cursor: 'pointer',
            }"
            @click="emit('start-review')"
        >
            <Check :size="13" />
            Submit review
            <Kbd tone="on-accent">⌘↵</Kbd>
        </button>
        <div :style="{ display: 'flex', gap: '6px' }">
            <button
                type="button"
                :style="{
                    flex: 1,
                    height: '28px',
                    borderRadius: '7px',
                    border: '1px solid var(--gd-border)',
                    background: 'var(--gd-panel-2)',
                    color: 'var(--gd-text-2)',
                    fontSize: '12px',
                    fontWeight: 500,
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '5px',
                    cursor: 'pointer',
                }"
                @click="emit('open-review-panel', 'comment')"
            >
                <MessageSquare :size="12" />
                Comment
            </button>
            <button
                type="button"
                :style="{
                    flex: 1,
                    height: '28px',
                    borderRadius: '7px',
                    border: '1px solid var(--gd-border)',
                    background: 'var(--gd-panel-2)',
                    color: 'var(--gd-text-2)',
                    fontSize: '12px',
                    fontWeight: 500,
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '5px',
                    cursor: 'pointer',
                }"
                @click="emit('open-review-panel', 'request')"
            >
                <AlertCircle :size="12" />
                Request
            </button>
        </div>
    </div>
</template>
