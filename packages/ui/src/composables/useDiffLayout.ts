import { ref, type Ref } from "vue";

// useDiffLayout holds per-file UI layout state that lives only in the
// renderer: which file sections are collapsed and the split-pane ratio
// each file remembers for its split-view. Both maps are keyed by file
// path; entries are created lazily on first interaction.
export interface UseDiffLayout {
  collapsed: Ref<Record<string, boolean>>;
  splitRatios: Ref<Record<string, number>>;
  toggleCollapsed: (path: string) => void;
  setSplitRatio: (path: string, ratio: number) => void;
  reset: () => void;
}

export function useDiffLayout(): UseDiffLayout {
  const collapsed = ref<Record<string, boolean>>({});
  const splitRatios = ref<Record<string, number>>({});

  function toggleCollapsed(path: string): void {
    collapsed.value[path] = !collapsed.value[path];
  }

  function setSplitRatio(path: string, ratio: number): void {
    splitRatios.value[path] = ratio;
  }

  function reset(): void {
    collapsed.value = {};
    splitRatios.value = {};
  }

  return { collapsed, splitRatios, toggleCollapsed, setSplitRatio, reset };
}
