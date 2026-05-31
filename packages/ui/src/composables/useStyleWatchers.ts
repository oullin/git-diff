import { watch, type Ref } from 'vue';
import { applyAccent, diffBgs, type Accent } from '@lib/accent';
import { setTheme } from '@composables/useTheme';
import type { Tweaks } from '@composables/useTweaks';

// Wires the three side-effecting watchers that connect preferences to the DOM:
//   * accent ref  -> CSS custom properties
//   * theme name  -> data-theme on <html>
//   * diff style  -> --gd-word-add/rem-bg CSS variables
// Caller passes the accent ref and tweaks ref directly; the composable runs the
// watchers eagerly so styles are applied on mount.
export function useStyleWatchers(accent: Ref<Accent>, tweaks: Ref<Tweaks>): void {
	watch(accent, applyAccent, { immediate: true });

	watch(
		() => tweaks.value.theme,
		(choice) => setTheme(choice),
		{ immediate: true },
	);

	watch(
		() => tweaks.value.diffStyle,
		(style) => {
			const colors = diffBgs(style);

			document.documentElement.style.setProperty('--gd-word-add-bg', colors.addStrong);
			document.documentElement.style.setProperty('--gd-word-rem-bg', colors.remStrong);
		},
		{ immediate: true },
	);
}
