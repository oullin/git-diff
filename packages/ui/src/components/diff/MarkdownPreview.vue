<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import DOMPurify from "dompurify";
import { Marked } from "marked";

import { getAddedLineNumbers, renderMarkdownWithLineAnchors } from "@lib/markdownPreview";
import type { ChangedFile } from "@git-diff/domain";

// Standalone Marked instance keeps preview-specific options (line-by-line
// inline rendering) isolated from any other markdown rendered in the app.
const marked = new Marked();

const props = defineProps<{
    file: ChangedFile;
    repoRoot: string;
    commitRef?: string;
}>();

const content = ref<string>("");
const loading = ref(true);
const error = ref<string>("");

const addedLines = computed(() => getAddedLineNumbers(props.file));

const rendered = computed(() =>
    renderMarkdownWithLineAnchors(
        content.value,
        addedLines.value,
        (line) => marked.parseInline(line) as string,
        (html) => DOMPurify.sanitize(html),
    ),
);

async function load(): Promise<void> {
    loading.value = true;
    error.value = "";

    try {
        if (props.commitRef) {
            // Commit-mode preview needs the file as it existed at this ref.
            // The renderer bridge doesn't accept ref on readRepositoryFile yet,
            // so use the file-range endpoint with a large window. The 500-line
            // cap is the API's hard ceiling — preview is best-effort for huge
            // markdown files.
            const range = await window.diffApp.readRepositoryFileRange({
                root: props.repoRoot,
                path: props.file.path,
                ref: props.commitRef,
                startLine: 1,
                endLine: 500,
            });

            content.value = range.lines.join("\n");
        } else {
            const file = await window.diffApp.readRepositoryFile(props.repoRoot, props.file.path);

            content.value = file.binary ? "" : file.content;
        }
    } catch (err) {
        error.value = err instanceof Error ? err.message : String(err);
        content.value = "";
    } finally {
        loading.value = false;
    }
}

onMounted(load);
watch(() => [props.file.path, props.repoRoot, props.commitRef], load);
</script>

<template>
    <div
        class="markdown-preview"
        :style="{
            padding: '16px 24px',
            fontFamily:
                '-apple-system, BlinkMacSystemFont, \'Segoe UI\', Roboto, Helvetica, Arial, sans-serif',
            fontSize: '14px',
            lineHeight: '1.6',
            color: 'var(--gd-text)',
            background: 'var(--gd-bg)',
        }"
    >
        <div v-if="loading" :style="{ color: 'var(--gd-text-3)' }">Loading preview…</div>
        <div v-else-if="error" :style="{ color: 'var(--gd-error)' }">
            Failed to load preview: {{ error }}
        </div>
        <div v-else v-html="rendered" />
    </div>
</template>

<style scoped>
.markdown-preview :deep(.md-line) {
    padding: 1px 8px;
    margin: 0 -8px;
    border-radius: 4px;
}

.markdown-preview :deep(.md-line-added) {
    background: rgb(34 197 94 / 0.12);
    box-shadow: inset 2px 0 0 rgb(34 197 94 / 0.5);
}

.markdown-preview :deep(code) {
    font-family: var(--font-mono);
    font-size: 12.5px;
    padding: 1px 4px;
    background: var(--gd-panel-2);
    border-radius: 3px;
}

.markdown-preview :deep(a) {
    color: var(--gd-accent);
    text-decoration: underline;
}

.markdown-preview :deep(strong) {
    font-weight: 600;
}

.markdown-preview :deep(em) {
    font-style: italic;
}
</style>
