import { ref, type Ref } from 'vue';

export interface UseDiffLayout {
	collapsed: Ref<Record<string, boolean>>;
	previewing: Ref<Record<string, boolean>>;
	toggleCollapsed: (path: string) => void;
	togglePreview: (path: string) => void;
	reset: () => void;
}

export function useDiffLayout(): UseDiffLayout {
	const collapsed = ref<Record<string, boolean>>({});

	const previewing = ref<Record<string, boolean>>({});

	function toggleCollapsed(path: string): void {
		collapsed.value[path] = !collapsed.value[path];
	}

	function togglePreview(path: string): void {
		previewing.value[path] = !previewing.value[path];
	}

	function reset(): void {
		collapsed.value = {};
		previewing.value = {};
	}

	return {
		collapsed,
		previewing,
		toggleCollapsed,
		togglePreview,
		reset,
	};
}
