<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { FileText, Loader2 } from "lucide-vue-next";
import { isImagePath } from "@git-diff/domain";
import type { RepositoryFile } from "@git-diff/domain";

type Props = {
  file: RepositoryFile | null;
  path: string;
  repoRoot?: string;
  loading?: boolean;
  error?: string;
};

const props = defineProps<Props>();

const imageUrl = ref<string | null>(null);

const imageError = ref<string | null>(null);

const imageLoading = ref(false);

const sizeLabel = computed(() => {
  const size = props.file?.size ?? 0;

  if (size < 1024) {
    return `${size} B`;
  }

  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }

  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
});

const isImage = computed(() => isImagePath(props.path));

function revoke(): void {
  if (imageUrl.value) {
    URL.revokeObjectURL(imageUrl.value);
    imageUrl.value = null;
  }
}

async function loadImage(): Promise<void> {
  revoke();
  imageError.value = null;

  if (!isImage.value || !props.repoRoot) {
    return;
  }

  imageLoading.value = true;

  try {
    const { data, mime } = await window.diffApp.readRepositoryFileBytes({
      root: props.repoRoot,
      path: props.path,
    });

    const blob = new Blob([data as BlobPart], { type: mime || "application/octet-stream" });

    imageUrl.value = URL.createObjectURL(blob);
  } catch (error) {
    imageError.value = error instanceof Error ? error.message : String(error);
  } finally {
    imageLoading.value = false;
  }
}

watch(
  () => [props.path, props.repoRoot, isImage.value].join("|"),
  () => void loadImage(),
  { immediate: true },
);

onBeforeUnmount(() => revoke());
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <div v-if="loading" class="grid flex-1 place-items-center text-sm text-muted-foreground">
      <div class="flex items-center gap-2">
        <Loader2 class="h-4 w-4 animate-spin" />
        Loading file…
      </div>
    </div>

    <div v-else-if="error" class="grid flex-1 place-items-center p-8 text-sm text-destructive">
      {{ error }}
    </div>

    <div v-else-if="!file" class="grid flex-1 place-items-center text-sm text-muted-foreground">
      <div class="text-center">
        <FileText class="mx-auto h-8 w-8 text-muted-foreground" />
        <p class="mt-3">Select a file to preview.</p>
      </div>
    </div>

    <template v-else>
      <header
        class="flex items-center justify-between border-b border-border px-4 py-2 text-xs text-muted-foreground"
      >
        <span class="truncate font-mono">{{ path }}</span>
        <span class="ml-3 shrink-0">{{ sizeLabel }}</span>
      </header>

      <div
        v-if="isImage && imageLoading"
        class="grid flex-1 place-items-center text-sm text-muted-foreground"
      >
        <div class="flex items-center gap-2">
          <Loader2 class="h-4 w-4 animate-spin" />
          Loading image…
        </div>
      </div>

      <div
        v-else-if="isImage && imageError"
        class="grid flex-1 place-items-center p-8 text-sm text-destructive"
      >
        {{ imageError }}
      </div>

      <div v-else-if="isImage && imageUrl" class="grid flex-1 place-items-center overflow-auto p-4">
        <img
          :src="imageUrl"
          :alt="path"
          :style="{ maxWidth: '100%', maxHeight: '100%', objectFit: 'contain' }"
        />
      </div>

      <div
        v-else-if="file.binary"
        class="grid flex-1 place-items-center p-8 text-sm text-muted-foreground"
      >
        Binary file ({{ sizeLabel }}) — preview unavailable.
      </div>

      <div
        v-else-if="file.truncated"
        class="grid flex-1 place-items-center p-8 text-sm text-muted-foreground"
      >
        File is larger than 2 MB ({{ sizeLabel }}) — preview unavailable.
      </div>

      <pre
        v-else
        class="min-h-0 flex-1 overflow-auto bg-panel p-4 font-mono text-xs leading-relaxed"
      ><code>{{ file.content }}</code></pre>
    </template>
  </div>
</template>
