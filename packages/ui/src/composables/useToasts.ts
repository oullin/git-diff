import { ref, type Ref } from 'vue';
import type { ToastItem } from '@ui/toast';

export interface UseToasts {
	toasts: Ref<ToastItem[]>;
	show: (toast: Omit<ToastItem, 'id'>, ttlMs?: number) => string;
	dismiss: (id: string) => void;
}

export function useToasts(): UseToasts {
	const toasts: Ref<ToastItem[]> = ref([]);

	function show(toast: Omit<ToastItem, 'id'>, ttlMs = 6000): string {
		const id = `toast-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;

		toasts.value = [...toasts.value, { id, ...toast }];

		if (ttlMs > 0) {
			window.setTimeout(() => dismiss(id), ttlMs);
		}

		return id;
	}

	function dismiss(id: string): void {
		toasts.value = toasts.value.filter((toast) => toast.id !== id);
	}

	return { toasts, show, dismiss };
}
