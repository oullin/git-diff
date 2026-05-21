<script setup lang="ts" generic="T extends string">
defineProps<{
    modelValue: T;
    options: ReadonlyArray<{ value: T; label: string }>;
}>();
defineEmits<{ "update:modelValue": [value: T] }>();
</script>

<template>
    <div
        class="inline-flex items-center gap-px rounded-lg border p-[2px]"
        :style="{
            background: 'var(--gd-panel-2)',
            borderColor: 'var(--gd-border)',
        }"
    >
        <button
            v-for="opt in options"
            :key="opt.value"
            type="button"
            class="rounded-md px-[10px] text-[13px] font-medium transition-colors"
            :style="{
                height: '24px',
                background: modelValue === opt.value ? 'var(--gd-panel-3)' : 'transparent',
                color: modelValue === opt.value ? 'var(--gd-text)' : 'var(--gd-text-3)',
                boxShadow:
                    modelValue === opt.value
                        ? '0 1px 0 rgba(255,255,255,0.04), inset 0 0 0 1px var(--gd-border)'
                        : 'none',
            }"
            @click="$emit('update:modelValue', opt.value)"
        >
            {{ opt.label }}
        </button>
    </div>
</template>
