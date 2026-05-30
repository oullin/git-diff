import { computed, type ComputedRef, type Ref } from "vue";
import type { ChangedFile } from "@git-diff/domain";

export interface UseFileListFilterOptions {
    files: Ref<ChangedFile[]> | ComputedRef<ChangedFile[]>;
    allPaths: Ref<string[]> | ComputedRef<string[]>;
    searchQuery: Ref<string> | ComputedRef<string>;
    isViewed: (file: ChangedFile) => boolean;
}

export function useFileListFilter(opts: UseFileListFilterOptions) {
    const totalCount = computed(() => opts.files.value.length);
    const viewedCount = computed(() => opts.files.value.filter((f) => opts.isViewed(f)).length);

    const filteredFiles = computed(() => {
        const q = opts.searchQuery.value.trim().toLowerCase();

        return q
            ? opts.files.value.filter((f) => f.path.toLowerCase().includes(q))
            : opts.files.value;
    });

    const filteredAllPaths = computed(() => {
        const q = opts.searchQuery.value.trim().toLowerCase();

        return q
            ? opts.allPaths.value.filter((path) => path.toLowerCase().includes(q))
            : opts.allPaths.value;
    });

    const progressPct = computed(() =>
        totalCount.value === 0 ? 0 : (viewedCount.value / totalCount.value) * 100,
    );

    return { totalCount, viewedCount, filteredFiles, filteredAllPaths, progressPct };
}
