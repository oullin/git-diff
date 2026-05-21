<script setup lang="ts">
import { Dialog, DialogBody, DialogHeader, DialogTitle } from "@ui/dialog";
import { LazyRichTextEditor, type RichTextFeatures } from "@ui/rich-text-editor";

defineProps<{
    open: boolean;
    summaryDraft: string;
    features: RichTextFeatures;
}>();

defineEmits<{
    close: [];
    "update:summaryDraft": [value: string];
    "start-review": [];
}>();
</script>

<template>
    <Dialog :show="open" max-width="2xl" @close="$emit('close')">
        <DialogHeader>
            <DialogTitle>Conversation</DialogTitle>
        </DialogHeader>
        <DialogBody class="max-h-[70vh] overflow-auto">
            <div>
                <div :style="{ fontSize: '13px', fontWeight: 600, color: 'var(--gd-text)' }">
                    Review notes
                </div>
                <p :style="{ marginTop: '4px', fontSize: '12px', color: 'var(--gd-text-3)' }">
                    Local-only review timeline and rich comments.
                </p>
                <LazyRichTextEditor
                    class="mt-3"
                    :model-value="summaryDraft"
                    min-height="8rem"
                    placeholder="Write a review summary…"
                    aria-label="Review summary"
                    :features="features"
                    @update:model-value="(value: string) => $emit('update:summaryDraft', value)"
                />
                <button
                    type="button"
                    class="mt-3 w-full"
                    :style="{
                        height: '32px',
                        borderRadius: '8px',
                        border: 0,
                        background: 'var(--gd-accent-strong)',
                        color: '#fff',
                        fontSize: '13px',
                        fontWeight: 600,
                        cursor: 'pointer',
                    }"
                    @click="$emit('start-review')"
                >
                    Start review
                </button>
            </div>
        </DialogBody>
    </Dialog>
</template>
