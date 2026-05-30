<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{ add: number; del: number; mini?: boolean; squares?: boolean }>();

// GitHub Primer 5-square diffstat — same proportioning the design bundle uses.
const squareKinds = computed<("g" | "r" | "n")[]>(() => {
    const total = props.add + props.del;
    let g = total === 0 ? 0 : Math.round((props.add / total) * 5);

    if (props.add > 0 && g === 0) {g = 1;}

    let r = total === 0 ? 0 : Math.min(5 - g, Math.round((props.del / total) * 5));

    if (props.del > 0 && r === 0 && g < 5) {r = 1;}

    const n = Math.max(0, 5 - g - r);

    return [
        ...Array(g).fill("g" as const),
        ...Array(r).fill("r" as const),
        ...Array(n).fill("n" as const),
    ];
});

const squareColor: Record<"g" | "r" | "n", string> = {
    g: "var(--success-emphasis)",
    r: "var(--danger-fg)",
    n: "var(--gd-panel-3)",
};
</script>

<template>
    <span
        v-if="squares"
        class="inline-flex items-center"
        :title="`+${add} −${del}`"
        :style="{ gap: '2px' }"
    >
        <span
            v-for="(kind, i) in squareKinds"
            :key="i"
            :style="{
                width: '8px',
                height: '8px',
                borderRadius: '2px',
                background: squareColor[kind],
            }"
        />
    </span>
    <span
        v-else
        class="inline-flex items-center gap-1 font-mono font-semibold"
        :style="{ fontSize: mini ? '10.5px' : '11px' }"
    >
        <span :style="{ color: 'var(--gd-added)' }">+{{ add }}</span>
        <span :style="{ color: 'var(--gd-removed)' }">−{{ del }}</span>
    </span>
</template>
