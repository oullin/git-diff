<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useLexicalComposer } from "lexical-vue";
import {
    $getSelection,
    $isRangeSelection,
    COMMAND_PRIORITY_LOW,
    SELECTION_CHANGE_COMMAND,
} from "lexical";
import { $isLinkNode, TOGGLE_LINK_COMMAND } from "@lexical/link";
import { $findMatchingParent, mergeRegister } from "@lexical/utils";
import { Check, Link2Off, Pencil, X } from "lucide-vue-next";

const editor = useLexicalComposer();

const open = ref(false);
const editing = ref(false);
const url = ref("");
const draft = ref("");
const position = ref<{ top: number; left: number } | null>(null);
const inputEl = ref<HTMLInputElement | null>(null);

function getRect(): DOMRect | null {
    const nativeSelection = window.getSelection();

    if (!nativeSelection || nativeSelection.rangeCount === 0) {
        return null;
    }

    const range = nativeSelection.getRangeAt(0);

    return range.getBoundingClientRect();
}

function close(): void {
    open.value = false;
    editing.value = false;
    position.value = null;
}

function refresh(): void {
    editor.getEditorState().read(() => {
        const selection = $getSelection();

        if (!$isRangeSelection(selection)) {
            close();

            return;
        }

        const node = selection.anchor.getNode();
        const linkNode = $findMatchingParent(node, $isLinkNode);

        if (!linkNode) {
            close();

            return;
        }

        const rect = getRect();

        if (!rect) {
            close();

            return;
        }

        url.value = linkNode.getURL();
        draft.value = linkNode.getURL();
        position.value = {
            top: rect.bottom + window.scrollY + 6,
            left: rect.left + window.scrollX,
        };
        open.value = true;
    });
}

function startEdit(): void {
    editing.value = true;
    draft.value = url.value;
    nextTick(() => inputEl.value?.focus());
}

function applyEdit(): void {
    const value = draft.value.trim();

    if (!value) {
        return;
    }

    editor.dispatchCommand(TOGGLE_LINK_COMMAND, value);
    editing.value = false;
}

function removeLink(): void {
    editor.dispatchCommand(TOGGLE_LINK_COMMAND, null);
    close();
}

const unregister = mergeRegister(
    editor.registerUpdateListener(() => refresh()),
    editor.registerCommand(
        SELECTION_CHANGE_COMMAND,
        () => {
            refresh();

            return false;
        },
        COMMAND_PRIORITY_LOW,
    ),
);

onMounted(() => {
    refresh();
});

onBeforeUnmount(() => unregister());

watch(open, (val) => {
    if (!val) {
        editing.value = false;
    }
});
</script>

<template>
    <div
        v-if="open && position"
        class="absolute z-50 flex items-center gap-1 rounded-md border border-border bg-popover px-2 py-1 text-sm shadow-md"
        :style="{ top: `${position.top}px`, left: `${position.left}px` }"
        @mousedown.prevent
    >
        <template v-if="!editing">
            <a
                :href="url"
                target="_blank"
                rel="noopener noreferrer"
                class="max-w-[18rem] truncate text-primary underline-offset-2 hover:underline"
            >
                {{ url }}
            </a>
            <button
                type="button"
                class="ml-1 rounded p-1 hover:bg-accent"
                :title="'Edit link'"
                @click="startEdit"
            >
                <Pencil class="h-3.5 w-3.5" />
            </button>
            <button
                type="button"
                class="rounded p-1 hover:bg-accent"
                :title="'Remove link'"
                @click="removeLink"
            >
                <Link2Off class="h-3.5 w-3.5" />
            </button>
        </template>
        <template v-else>
            <input
                ref="inputEl"
                v-model="draft"
                type="url"
                class="w-64 rounded border border-input bg-background px-2 py-1 text-sm focus:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                placeholder="https://"
                @keydown.enter.prevent="applyEdit"
                @keydown.escape.prevent="close"
            />
            <button
                type="button"
                class="rounded p-1 hover:bg-accent"
                :title="'Apply'"
                @click="applyEdit"
            >
                <Check class="h-3.5 w-3.5" />
            </button>
            <button
                type="button"
                class="rounded p-1 hover:bg-accent"
                :title="'Cancel'"
                @click="close"
            >
                <X class="h-3.5 w-3.5" />
            </button>
        </template>
    </div>
</template>
