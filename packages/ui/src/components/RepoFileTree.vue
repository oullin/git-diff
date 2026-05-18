<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { FileTree, type FileTreeItemHandle } from "@pierre/trees";
import "@pierre/trees/web-components";

type Props = {
  paths: readonly string[];
  selectedPath?: string;
  changedPaths?: ReadonlySet<string>;
  initialExpansion?: "open" | "closed";
};

const props = withDefaults(defineProps<Props>(), {
  changedPaths: () => new Set<string>(),
  initialExpansion: "open",
});
const emit = defineEmits<{ select: [path: string] }>();

const mountEl = ref<HTMLDivElement | null>(null);
let tree: FileTree | null = null;

function applyChangedDecorations(root: HTMLElement, changed: ReadonlySet<string>) {
  const rows = root.querySelectorAll<HTMLElement>("[data-item-path]");
  rows.forEach((row) => {
    const path = row.dataset.itemPath ?? "";
    if (changed.has(path)) {
      row.setAttribute("data-changed", "true");
    } else {
      row.removeAttribute("data-changed");
    }
  });
  const flattened = root.querySelectorAll<HTMLElement>("[data-item-flattened-subitem]");
  flattened.forEach((seg) => {
    const path = seg.getAttribute("data-item-flattened-subitem") ?? "";
    if (changed.has(path)) {
      seg.setAttribute("data-changed", "true");
    } else {
      seg.removeAttribute("data-changed");
    }
  });
}

onMounted(() => {
  if (!mountEl.value) return;
  tree = new FileTree({
    paths: [...props.paths],
    flattenEmptyDirectories: true,
    initialExpansion: props.initialExpansion,
    search: false,
  });
  tree.subscribe(() => {
    const focused = tree?.getFocusedItem();
    if (focused && !focused.isDirectory()) {
      emit("select", focused.getPath());
    }
    if (mountEl.value) {
      applyChangedDecorations(mountEl.value, props.changedPaths);
    }
  });
  tree.render({ containerWrapper: mountEl.value });
  applyChangedDecorations(mountEl.value, props.changedPaths);
});

watch(
  () => props.paths,
  (next) => {
    if (!tree) return;
    tree.resetPaths([...next]);
    if (mountEl.value) {
      applyChangedDecorations(mountEl.value, props.changedPaths);
    }
  },
);

watch(
  () => props.changedPaths,
  (next) => {
    if (!mountEl.value) return;
    applyChangedDecorations(mountEl.value, next);
  },
);

watch(
  () => props.selectedPath,
  (path) => {
    if (!tree || !path) return;
    const item: FileTreeItemHandle | null = tree.getItem(path);
    if (item && !item.isSelected()) {
      item.select();
    }
  },
);

onBeforeUnmount(() => {
  tree?.cleanUp();
  tree = null;
});
</script>

<template>
  <div ref="mountEl" class="repo-file-tree h-full w-full" />
</template>

<style scoped>
.repo-file-tree {
  --trees-row-height: 24px;
  min-height: 120px;
}

.repo-file-tree :deep([data-changed="true"]) {
  position: relative;
  color: hsl(var(--primary, 220 90% 56%));
  font-weight: 500;
}

.repo-file-tree :deep([data-changed="true"])::before {
  content: "";
  position: absolute;
  left: 2px;
  top: 50%;
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  background: hsl(var(--primary, 220 90% 56%));
  transform: translateY(-50%);
}
</style>
