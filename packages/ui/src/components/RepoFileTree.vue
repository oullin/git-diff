<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { FileTree, type FileTreeItemHandle } from "@pierre/trees";
import "@pierre/trees/web-components";

type Props = {
  paths: readonly string[];
  selectedPath?: string;
};

const props = defineProps<Props>();
const emit = defineEmits<{ select: [path: string] }>();

const mountEl = ref<HTMLDivElement | null>(null);
let tree: FileTree | null = null;

onMounted(() => {
  if (!mountEl.value) return;
  tree = new FileTree({
    paths: [...props.paths],
    flattenEmptyDirectories: true,
    initialExpansion: "open",
    search: false,
  });
  tree.subscribe(() => {
    const focused = tree?.getFocusedItem();
    if (focused && !focused.isDirectory()) {
      emit("select", focused.getPath());
    }
  });
  tree.render({ containerWrapper: mountEl.value });
});

watch(
  () => props.paths,
  (next) => {
    if (!tree) return;
    tree.resetPaths([...next]);
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
</style>
