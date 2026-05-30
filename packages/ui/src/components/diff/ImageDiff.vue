<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from "vue";
import { ImageOff, Loader2 } from "lucide-vue-next";
import type { ChangedFile } from "@git-diff/domain";

type Props = {
    file: ChangedFile;
    repoRoot: string;
    commitRef?: string;
};

type Side = {
    label: string;
    objectUrl: string | null;
    error: string | null;
    bytes: number;
    width: number;
    height: number;
    missing: boolean;
    loading: boolean;
};

const props = defineProps<Props>();

const oldSide = ref<Side>(createSide("Before"));
const newSide = ref<Side>(createSide("After"));

function createSide(label: string): Side {
    return {
        label,
        objectUrl: null,
        error: null,
        bytes: 0,
        width: 0,
        height: 0,
        missing: false,
        loading: false,
    };
}

function revoke(side: Side): void {
    if (side.objectUrl) {
        URL.revokeObjectURL(side.objectUrl);
        side.objectUrl = null;
    }
}

function resetSides(): void {
    revoke(oldSide.value);
    revoke(newSide.value);
    oldSide.value = createSide("Before");
    newSide.value = createSide("After");
}

// Working-tree mode: old = HEAD, new = worktree (empty ref).
// Commit mode: old = parent commit, new = the commit itself.
function refsForFile(): { oldRef: string | null; newRef: string | null } {
    const status = props.file.status;

    if (props.commitRef) {
        const parent = `${props.commitRef}^`;

        if (status === "added") {
            return { oldRef: null, newRef: props.commitRef };
        }

        if (status === "deleted") {
            return { oldRef: parent, newRef: null };
        }

        return { oldRef: parent, newRef: props.commitRef };
    }

    if (status === "added" || status === "untracked") {
        return { oldRef: null, newRef: "" };
    }

    if (status === "deleted") {
        return { oldRef: "HEAD", newRef: null };
    }

    return { oldRef: "HEAD", newRef: "" };
}

async function fetchSide(side: Side, ref: string, path: string): Promise<void> {
    side.loading = true;
    side.error = null;

    try {
        const { data, mime } = await window.diffApp.readRepositoryFileBytes({
            root: props.repoRoot,
            path,
            ref,
        });

        side.bytes = data.byteLength;
        // Cast: IPC always backs Uint8Array with an ArrayBuffer (not Shared),
        // but TS infers ArrayBufferLike which Blob's typing rejects.
        const blob = new Blob([data as BlobPart], { type: mime || guessMime(path) });

        revoke(side);
        side.objectUrl = URL.createObjectURL(blob);
    } catch (error) {
        // notFound is expected for added-at-ref or deleted-at-worktree cases —
        // the caller decides which sides to fetch, so any 404 here surfaces a
        // real problem (e.g. the repo moved).
        const message =
            typeof error === "object" && error !== null && "kind" in error
                ? (error as { kind: string }).kind === "notFound"
                    ? "Blob not available at this ref"
                    : ((error as { message?: string }).message ?? "Failed to load")
                : ((error as { message?: string }).message ?? "Failed to load");

        side.error = message;
    } finally {
        side.loading = false;
    }
}

function guessMime(path: string): string {
    const dot = path.lastIndexOf(".");
    const ext = dot < 0 ? "" : path.slice(dot).toLowerCase();

    switch (ext) {
        case ".png":
            return "image/png";
        case ".jpg":
        case ".jpeg":
            return "image/jpeg";
        case ".gif":
            return "image/gif";
        case ".webp":
            return "image/webp";
        case ".svg":
            return "image/svg+xml";
        case ".avif":
            return "image/avif";
        case ".bmp":
            return "image/bmp";
        case ".ico":
            return "image/x-icon";
        default:
            return "application/octet-stream";
    }
}

function load(): void {
    resetSides();

    const { oldRef, newRef } = refsForFile();
    const oldPath = props.file.oldPath ?? props.file.path;

    if (oldRef === null) {
        oldSide.value.missing = true;
    } else {
        void fetchSide(oldSide.value, oldRef, oldPath);
    }

    if (newRef === null) {
        newSide.value.missing = true;
    } else {
        void fetchSide(newSide.value, newRef, props.file.path);
    }
}

function onImgLoaded(side: Side, event: Event): void {
    const img = event.target as HTMLImageElement;

    side.width = img.naturalWidth;
    side.height = img.naturalHeight;
}

function sizeLabel(bytes: number): string {
    if (bytes < 1024) {
        return `${bytes} B`;
    }

    if (bytes < 1024 * 1024) {
        return `${(bytes / 1024).toFixed(1)} KB`;
    }

    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

watch(
    () => [props.file.path, props.file.fingerprint, props.repoRoot, props.commitRef].join("|"),
    () => load(),
    { immediate: true },
);

onBeforeUnmount(() => {
    revoke(oldSide.value);
    revoke(newSide.value);
});
</script>

<template>
    <div
        :style="{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: '12px',
            padding: '16px',
            background: 'var(--gd-bg-code, var(--gd-bg))',
        }"
    >
        <div
            v-for="side in [oldSide, newSide]"
            :key="side.label"
            :style="{
                display: 'flex',
                flexDirection: 'column',
                gap: '8px',
                border: '1px solid var(--gd-border-soft)',
                borderRadius: '8px',
                padding: '12px',
                background: 'var(--gd-panel)',
                minHeight: '160px',
            }"
        >
            <header
                class="flex items-center justify-between"
                :style="{
                    fontSize: '12px',
                    color: 'var(--gd-text-3)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.04em',
                }"
            >
                <span>{{ side.label }}</span>
                <span v-if="!side.missing && !side.loading && !side.error">
                    {{ sizeLabel(side.bytes) }}
                    <template v-if="side.width > 0">· {{ side.width }}×{{ side.height }}</template>
                </span>
            </header>

            <div
                class="grid flex-1 place-items-center"
                :style="{ minHeight: '120px', overflow: 'hidden' }"
            >
                <div
                    v-if="side.missing"
                    class="flex flex-col items-center gap-2 text-sm"
                    :style="{ color: 'var(--gd-text-3)' }"
                >
                    <ImageOff :size="20" />
                    <span>Not present</span>
                </div>
                <div
                    v-else-if="side.loading"
                    class="flex items-center gap-2 text-sm"
                    :style="{ color: 'var(--gd-text-3)' }"
                >
                    <Loader2 :size="14" class="animate-spin" />
                    Loading…
                </div>
                <div v-else-if="side.error" class="text-xs" :style="{ color: 'var(--gd-text-3)' }">
                    {{ side.error }}
                </div>
                <img
                    v-else-if="side.objectUrl"
                    :src="side.objectUrl"
                    :alt="side.label"
                    :style="{
                        maxWidth: '100%',
                        maxHeight: '480px',
                        objectFit: 'contain',
                    }"
                    @load="(event) => onImgLoaded(side, event)"
                />
            </div>
        </div>
    </div>
</template>
