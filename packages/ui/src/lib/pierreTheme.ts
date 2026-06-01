import { registerCustomTheme } from '@pierre/diffs';
import type { ThemeRegistration } from 'shiki';

// Bridges the app's existing Shiki theme JSONs (Primer-based Licht/Dunkel,
// already used by the legacy highlighter) into @pierre/diffs so the library's
// FileDiff renderer tokenizes code with the same palette the app has always
// shipped. The library keys themes by name; we expose them as 'licht'/'dunkel'
// and select between them per app theme via FileDiff's `theme`/`themeType`.

let registered = false;

/** Idempotent — safe to call on every app boot. */
export function registerPierreThemes(): void {
	if (registered) {
		return;
	}

	registered = true;

	// The library matches the registered name against the theme JSON's `name`
	// field, so these must equal "Licht"/"Dunkel" exactly.
	registerCustomTheme('Licht', () => import('@themes/licht.json').then((m) => m.default as unknown as ThemeRegistration));
	registerCustomTheme('Dunkel', () => import('@themes/dunkel.json').then((m) => m.default as unknown as ThemeRegistration));
}

/** Theme map passed to FileDiff options; pairs with `themeType` from useTheme. */
export const PIERRE_THEMES = { light: 'Licht', dark: 'Dunkel' } as const;
