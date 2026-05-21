<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { ChevronDown, ChevronUp, X } from "lucide-vue-next";
import { applyHighlights, clearHighlights, searchDiff, setActiveHighlight } from "@lib/diffSearch";

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: [] }>();

const query = ref("");
const caseSensitive = ref(false);
const inputRef = ref<HTMLInputElement | null>(null);

let wrappers: HTMLSpanElement[] = [];
const activeIndex = ref(0);

const matchCount = computed(() => wrappers.length);

watch(
    () => props.open,
    async (next) => {
        if (next) {
            await nextTick();
            inputRef.value?.focus();
            inputRef.value?.select();
        } else {
            clear();
        }
    },
);

watch([query, caseSensitive], () => {
    refresh();
});

function clear() {
    clearHighlights();
    wrappers = [];
    activeIndex.value = 0;
}

function refresh() {
    clear();
    const result = searchDiff(query.value, document, {
        caseSensitive: caseSensitive.value,
    });

    if (result.matches.length === 0) {
        return;
    }

    wrappers = applyHighlights(result.matches);
    activeIndex.value = 0;
    setActiveHighlight(wrappers, 0);
}

function step(delta: number) {
    if (wrappers.length === 0) {
        return;
    }

    activeIndex.value = (activeIndex.value + delta + wrappers.length) % wrappers.length;
    setActiveHighlight(wrappers, activeIndex.value);
}

function onKey(event: KeyboardEvent) {
    if (event.key === "Enter") {
        event.preventDefault();
        step(event.shiftKey ? -1 : 1);
    } else if (event.key === "Escape") {
        event.preventDefault();
        emit("close");
    }
}

onMounted(() => {
    if (props.open) {
        refresh();
    }
});
</script>

<template>
    <div
        v-if="props.open"
        class="fixed right-6 top-20 z-50 flex items-center gap-2 rounded-md border border-border bg-background/95 px-2 py-1.5 shadow-lg backdrop-blur"
        style="min-width: 22rem"
    >
        <input
            ref="inputRef"
            v-model="query"
            class="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            type="text"
            placeholder="Find in diff…"
            @keydown="onKey"
        />
        <span class="text-xs tabular-nums text-muted-foreground">
            {{ matchCount === 0 ? "0/0" : `${activeIndex + 1}/${matchCount}` }}
        </span>
        <label class="flex items-center gap-1 text-[11px] text-muted-foreground">
            <input v-model="caseSensitive" type="checkbox" class="h-3 w-3" />
            Aa
        </label>
        <button
            type="button"
            class="rounded p-1 hover:bg-muted"
            :disabled="matchCount === 0"
            @click="step(-1)"
            title="Previous (Shift+Enter)"
        >
            <ChevronUp class="h-3.5 w-3.5" />
        </button>
        <button
            type="button"
            class="rounded p-1 hover:bg-muted"
            :disabled="matchCount === 0"
            @click="step(1)"
            title="Next (Enter)"
        >
            <ChevronDown class="h-3.5 w-3.5" />
        </button>
        <button
            type="button"
            class="rounded p-1 hover:bg-muted"
            @click="emit('close')"
            title="Close (Esc)"
        >
            <X class="h-3.5 w-3.5" />
        </button>
    </div>
</template>
