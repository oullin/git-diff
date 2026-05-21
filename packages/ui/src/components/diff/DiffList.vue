<script setup lang="ts">
import type { ChangedFile, DiffSection, RepositoryFile, ReviewComment } from "@git-diff/contracts";
import type { PatchLine } from "@lib/patch";
import type { RichTextFeatures } from "@ui/rich-text-editor";
import DiffBody from "@entry/components/diff/DiffBody.vue";
import FileHeader from "@entry/components/diff/FileHeader.vue";
import FileContentViewer from "@entry/components/FileContentViewer.vue";
import MarkdownPreview from "@entry/components/diff/MarkdownPreview.vue";
import type { LineSelectionRange } from "@composables/useLineSelection";
import type { DiffViewMode } from "@git-diff/contracts";
import type { Tweaks } from "@composables/useTweaks";

// DiffList renders the right-hand column of the diff view: either a
// FileContentViewer fallback (when the selected path isn't part of the
// diff), an empty-state message (when the repo has no changes), or a
// scrolling list of FileHeader + DiffBody cards for each changed file.
//
// The component is intentionally thin -- nothing it owns survives a
// route change. State (selection, collapse map, split ratios, viewed
// flags, comments) is read from props; mutating actions are emitted up
// so App.vue keeps a single source of truth.
defineProps<{
    files: ChangedFile[];
    selectedPath: string;
    selectedIsChanged: boolean;
    selectedRepoFile: RepositoryFile | null;
    selectedFileLoading: boolean;
    selectedFileError: string;
    collapsed: Record<string, boolean>;
    splitRatios: Record<string, number>;
    tweaks: Tweaks;
    diffViewMode: DiffViewMode;
    hideWhitespace: boolean;
    reviewComments: ReviewComment[];
    commentFeatures: RichTextFeatures;
    repoRoot: string;
    commitRef?: string;
    previewing: Record<string, boolean>;
    isViewedFn: (file: ChangedFile) => boolean;
    fileElementID: (path: string) => string;
}>();

const emit = defineEmits<{
    "toggle-collapsed": [path: string];
    "toggle-viewed": [file: ChangedFile];
    "toggle-preview": [path: string];
    "copy-path": [path: string];
    "open-comment-for-line": [
        file: ChangedFile,
        section: DiffSection,
        line: PatchLine,
        range?: LineSelectionRange,
    ];
    "delete-comment": [comment: ReviewComment];
    "reply-comment": [parent: ReviewComment, bodyHtml: string];
    "update:split-ratio": [path: string, ratio: number];
}>();
</script>

<template>
    <FileContentViewer
        v-if="selectedPath && !selectedIsChanged"
        :file="selectedRepoFile"
        :path="selectedPath"
        :loading="selectedFileLoading"
        :error="selectedFileError"
    />
    <div
        v-else-if="files.length === 0"
        class="grid h-full place-items-center text-sm"
        :style="{ color: 'var(--gd-text-3)' }"
    >
        No changes detected. Edit some files and refresh.
    </div>
    <template v-else>
        <article
            v-for="file in files"
            :id="fileElementID(file.path)"
            :key="file.path"
            :style="{
                background: 'var(--gd-panel)',
                overflow: 'clip',
                borderBottom: '1px solid var(--gd-border)',
            }"
        >
            <FileHeader
                :file="file"
                :collapsed="!!collapsed[file.path]"
                :viewed="isViewedFn(file)"
                :previewing="!!previewing[file.path]"
                @toggle-collapsed="emit('toggle-collapsed', file.path)"
                @toggle-viewed="emit('toggle-viewed', file)"
                @toggle-preview="emit('toggle-preview', file.path)"
                @copy="(path: string) => emit('copy-path', path)"
            />
            <MarkdownPreview
                v-if="!collapsed[file.path] && previewing[file.path]"
                :file="file"
                :repo-root="repoRoot"
                :commit-ref="commitRef"
            />
            <DiffBody
                v-else-if="!collapsed[file.path]"
                :file="file"
                :view-mode="diffViewMode"
                :diff-style="tweaks.diffStyle"
                :density="tweaks.density"
                :word-highlight="tweaks.wordHighlight"
                :hide-whitespace="hideWhitespace"
                :comments="reviewComments"
                :reply-features="commentFeatures"
                :repo-root="repoRoot"
                :commit-ref="commitRef"
                :split-ratio="splitRatios[file.path] ?? 0.5"
                @add-comment="
                    (section, line, range) =>
                        emit('open-comment-for-line', file, section, line, range)
                "
                @delete-comment="(comment) => emit('delete-comment', comment)"
                @reply-comment="(parent, bodyHtml) => emit('reply-comment', parent, bodyHtml)"
                @update:split-ratio="
                    (value: number) => emit('update:split-ratio', file.path, value)
                "
            />
        </article>
    </template>
</template>
