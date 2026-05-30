import { ref, type Ref } from "vue";

export interface UseDiffLayout {
    collapsed: Ref<Record<string, boolean>>;
    splitRatios: Ref<Record<string, number>>;
    previewing: Ref<Record<string, boolean>>;
    toggleCollapsed: (path: string) => void;
    setSplitRatio: (path: string, ratio: number) => void;
    togglePreview: (path: string) => void;
    reset: () => void;
}

export function useDiffLayout(): UseDiffLayout {
    const collapsed = ref<Record<string, boolean>>({});
    const splitRatios = ref<Record<string, number>>({});
    const previewing = ref<Record<string, boolean>>({});

    function toggleCollapsed(path: string): void {
        collapsed.value[path] = !collapsed.value[path];
    }

    function setSplitRatio(path: string, ratio: number): void {
        splitRatios.value[path] = ratio;
    }

    function togglePreview(path: string): void {
        previewing.value[path] = !previewing.value[path];
    }

    function reset(): void {
        collapsed.value = {};
        splitRatios.value = {};
        previewing.value = {};
    }

    return {
        collapsed,
        splitRatios,
        previewing,
        toggleCollapsed,
        setSplitRatio,
        togglePreview,
        reset,
    };
}
