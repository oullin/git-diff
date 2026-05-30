<script setup lang="ts">
import { Check } from "lucide-vue-next";
import Kbd from "@components/diff/Kbd.vue";

const emit = defineEmits<{
    "start-review": [];
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
                cursor: 'pointer',
            }"
            @click="emit('start-review')"
        >
            <Check :size="13" />
            Submit review
            <Kbd tone="on-accent">⌘↵</Kbd>
        </button>
    </div>
</template>
