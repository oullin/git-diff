<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

const props = withDefaults(
	defineProps<{
		show?: boolean;
		maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl';
		closeable?: boolean;
	}>(),
	{
		show: false,
		maxWidth: '2xl',
		closeable: true,
	},
);

const emit = defineEmits<{
	(e: 'close'): void;
}>();

const dialog = ref<HTMLDialogElement | null>(null);

const content = ref<HTMLElement | null>(null);

const showSlot = ref(props.show);

const previouslyFocusedElement = ref<HTMLElement | null>(null);

const previousBodyOverflow = ref<string | null>(null);

let closeTimeout: ReturnType<typeof setTimeout> | null = null;

const tabbableSelector = ['button:not([disabled])', '[href]', 'input:not([disabled])', 'select:not([disabled])', 'textarea:not([disabled])', "[tabindex]:not([tabindex='-1'])"].join(', ');

const maxWidthClass = computed(() => {
	return {
		sm: 'sm:max-w-sm',
		md: 'sm:max-w-md',
		lg: 'sm:max-w-lg',
		xl: 'sm:max-w-xl',
		'2xl': 'sm:max-w-2xl',
	}[props.maxWidth];
});

const isFocusableElement = (element: HTMLElement | null): element is HTMLElement => {
	if (element === null) {
		return false;
	}

	if (!element.isConnected || element.hasAttribute('disabled')) {
		return false;
	}

	return typeof element.focus === 'function';
};

const restoreFocus = (): void => {
	if (isFocusableElement(previouslyFocusedElement.value)) {
		previouslyFocusedElement.value.focus();
	}

	previouslyFocusedElement.value = null;
};

const focusInitialElement = (): void => {
	const dialogContent = content.value;

	if (dialogContent === null) {
		return;
	}

	const autofocusElement = dialogContent.querySelector<HTMLElement>('[autofocus]');

	if (isFocusableElement(autofocusElement)) {
		autofocusElement.focus();

		return;
	}

	const firstTabbableElement = Array.from(dialogContent.querySelectorAll<HTMLElement>(tabbableSelector)).find((element) => isFocusableElement(element));

	if (firstTabbableElement !== undefined) {
		firstTabbableElement.focus();

		return;
	}

	dialogContent.focus();
};

const clearPendingClose = (): void => {
	if (closeTimeout !== null) {
		clearTimeout(closeTimeout);
		closeTimeout = null;
	}
};

const lockBodyScroll = (): void => {
	if (previousBodyOverflow.value === null) {
		previousBodyOverflow.value = document.body.style.overflow;
	}

	document.body.style.overflow = 'hidden';
};

const restoreBodyScroll = (): void => {
	if (previousBodyOverflow.value === null) {
		return;
	}

	document.body.style.overflow = previousBodyOverflow.value;
	previousBodyOverflow.value = null;
};

const closeDialog = (): void => {
	clearPendingClose();

	closeTimeout = setTimeout(() => {
		if (dialog.value?.open) {
			dialog.value.close();
		}

		showSlot.value = false;
		restoreFocus();
		restoreBodyScroll();
		closeTimeout = null;
	}, 200);
};

watch(
	() => props.show,
	async (show, wasShowing) => {
		if (show) {
			clearPendingClose();
			previouslyFocusedElement.value = document.activeElement instanceof HTMLElement ? document.activeElement : null;
			lockBodyScroll();
			showSlot.value = true;

			await nextTick();

			if (!dialog.value?.open) {
				dialog.value?.showModal();
			}

			await nextTick();

			focusInitialElement();

			return;
		}

		if (wasShowing) {
			closeDialog();
		}
	},
	{ immediate: true },
);

const close = (): void => {
	if (props.closeable) {
		emit('close');
	}
};

const closeOnEscape = (event: KeyboardEvent) => {
	if (event.key === 'Escape' && props.show) {
		event.preventDefault();
		close();
	}
};

onMounted(() => document.addEventListener('keydown', closeOnEscape));

onUnmounted(() => {
	clearPendingClose();

	if (dialog.value?.open) {
		dialog.value.close();
	}

	document.removeEventListener('keydown', closeOnEscape);
	restoreBodyScroll();
	restoreFocus();
});
</script>

<template>
	<dialog ref="dialog" class="m-0 min-h-full min-w-full overflow-y-auto bg-transparent backdrop:bg-transparent">
		<div class="fixed inset-0 z-50 flex min-h-full items-start justify-center overflow-y-auto px-4 py-6 sm:items-center sm:px-0" scroll-region>
			<transition
				enter-active-class="ease-out duration-300"
				enter-from-class="opacity-0"
				enter-to-class="opacity-100"
				leave-active-class="ease-in duration-200"
				leave-from-class="opacity-100"
				leave-to-class="opacity-0"
			>
				<div v-show="show" class="fixed inset-0 transition-all" @click="close">
					<div class="absolute inset-0 bg-background/80 backdrop-blur-sm" />
				</div>
			</transition>

			<transition
				enter-active-class="ease-out duration-300"
				enter-from-class="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
				enter-to-class="opacity-100 translate-y-0 sm:scale-100"
				leave-active-class="ease-in duration-200"
				leave-from-class="opacity-100 translate-y-0 sm:scale-100"
				leave-to-class="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
			>
				<div
					ref="content"
					v-show="show"
					data-slot="dialog-content"
					:class="maxWidthClass"
					tabindex="-1"
					class="relative z-50 mb-6 overflow-hidden rounded-2xl border border-border bg-popover text-popover-foreground shadow-xl transition-all sm:mx-auto sm:w-full"
				>
					<slot v-if="showSlot" />
				</div>
			</transition>
		</div>
	</dialog>
</template>
