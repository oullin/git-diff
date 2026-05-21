<script setup lang="ts">
import { Dialog, DialogBody, DialogFooter, DialogHeader, DialogTitle } from "@ui/dialog";
import { LazyRichTextEditor, type RichTextFeatures } from "@ui/rich-text-editor";

defineProps<{
    open: boolean;
    filePath?: string;
    lineNumber?: number;
    modelValue: string;
    features: RichTextFeatures;
}>();

defineEmits<{
    "update:modelValue": [value: string];
    save: [];
    cancel: [];
}>();
</script>

<template>
    <Dialog :show="open" max-width="2xl" @close="$emit('cancel')">
        <DialogHeader>
            <DialogTitle>
                <span v-if="filePath">Comment on {{ filePath }}:{{ lineNumber }}</span>
                <span v-else>Add comment</span>
            </DialogTitle>
        </DialogHeader>
        <DialogBody>
            <LazyRichTextEditor
                :model-value="modelValue"
                min-height="8rem"
                placeholder="Leave a comment…"
                aria-label="Line comment"
                autofocus
                :features="features"
                @update:model-value="(value: string) => $emit('update:modelValue', value)"
            />
        </DialogBody>
        <DialogFooter>
            <button
                type="button"
                :style="{
                    height: '30px',
                    padding: '0 12px',
                    borderRadius: '8px',
                    border: '1px solid var(--gd-border)',
                    background: 'var(--gd-panel-2)',
                    color: 'var(--gd-text-2)',
                    fontSize: '12.5px',
                    fontWeight: 500,
                    cursor: 'pointer',
                }"
                @click="$emit('cancel')"
            >
                Cancel
            </button>
            <button
                type="button"
                :style="{
                    height: '30px',
                    padding: '0 12px',
                    borderRadius: '8px',
                    border: '1px solid var(--gd-border)',
                    background: 'var(--gd-accent-strong)',
                    color: '#fff',
                    fontSize: '12.5px',
                    fontWeight: 600,
                    cursor: 'pointer',
                }"
                @click="$emit('save')"
            >
                Save comment
            </button>
        </DialogFooter>
    </Dialog>
</template>
