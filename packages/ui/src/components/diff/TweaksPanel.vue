<script setup lang="ts">
import { Sun, Moon, Monitor } from 'lucide-vue-next';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@ui/select';
import { Switch } from '@ui/switch';
import { Label } from '@ui/label';
import { ACCENTS, type Accent } from '@lib/accent';
import type { Tweaks } from '@composables/useTweaks';
import type { ThemeChoice } from '@composables/useTheme';

defineProps<{ tweaks: Tweaks }>();

const emit = defineEmits<{
	'update:tweak': [key: keyof Tweaks, value: Tweaks[keyof Tweaks]];
}>();

function setKey<K extends keyof Tweaks>(key: K, value: Tweaks[K]) {
	emit('update:tweak', key, value);
}

const accentList = Object.values(ACCENTS) as Accent[];

const themeOptions: { value: ThemeChoice; label: string; icon: typeof Sun }[] = [
	{ value: 'light', label: 'Light', icon: Sun },
	{ value: 'dark', label: 'Dark', icon: Moon },
	{ value: 'system', label: 'System', icon: Monitor },
];
</script>

<template>
	<div class="flex flex-col gap-4">
		<section>
			<div class="mb-2 text-[10.5px] font-semibold uppercase tracking-wider" :style="{ color: 'var(--gd-text-muted)' }">Layout</div>
			<div class="flex flex-col gap-2.5">
				<div class="flex items-center justify-between gap-3">
					<Label for="tweaks-view-mode" class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> View mode </Label>
					<Select :model-value="tweaks.viewMode" @update:model-value="(value) => setKey('viewMode', value as Tweaks['viewMode'])">
						<SelectTrigger id="tweaks-view-mode" class="h-7 w-[120px] text-[12.5px]">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="split">Split</SelectItem>
							<SelectItem value="unified">Unified</SelectItem>
						</SelectContent>
					</Select>
				</div>

				<div class="flex items-center justify-between gap-3">
					<Label for="tweaks-density" class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Density </Label>
					<Select :model-value="tweaks.density" @update:model-value="(value) => setKey('density', value as Tweaks['density'])">
						<SelectTrigger id="tweaks-density" class="h-7 w-[120px] text-[12.5px]">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="comfortable">Comfortable</SelectItem>
							<SelectItem value="compact">Compact</SelectItem>
						</SelectContent>
					</Select>
				</div>

				<div class="flex items-center justify-between gap-3">
					<Label for="tweaks-minimap" class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Show minimap </Label>
					<Switch id="tweaks-minimap" :model-value="tweaks.showMinimap" @update:model-value="(value) => setKey('showMinimap', value)" />
				</div>

				<div class="flex items-center justify-between gap-3">
					<Label for="tweaks-status-bar" class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Status bar </Label>
					<Switch id="tweaks-status-bar" :model-value="tweaks.showStatusBar" @update:model-value="(value) => setKey('showStatusBar', value)" />
				</div>
			</div>
		</section>

		<div class="h-px" :style="{ background: 'var(--gd-border)' }" />

		<section>
			<div class="mb-2 text-[10.5px] font-semibold uppercase tracking-wider" :style="{ color: 'var(--gd-text-muted)' }">Diff</div>
			<div class="flex flex-col gap-2.5">
				<div class="flex items-center justify-between gap-3">
					<Label for="tweaks-hunk-style" class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Hunk style </Label>
					<Select :model-value="tweaks.diffStyle" @update:model-value="(value) => setKey('diffStyle', value as Tweaks['diffStyle'])">
						<SelectTrigger id="tweaks-hunk-style" class="h-7 w-[120px] text-[12.5px]">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="soft">Soft</SelectItem>
							<SelectItem value="punchy">Punchy</SelectItem>
							<SelectItem value="bar">Bar</SelectItem>
						</SelectContent>
					</Select>
				</div>

				<div class="flex items-center justify-between gap-3">
					<Label for="tweaks-word-highlight" class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Word-level highlight </Label>
					<Switch id="tweaks-word-highlight" :model-value="tweaks.wordHighlight" @update:model-value="(value) => setKey('wordHighlight', value)" />
				</div>

				<div class="flex items-center justify-between gap-3">
					<Label for="tweaks-hide-resolved" class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Hide resolved &amp; outdated comments </Label>
					<Switch id="tweaks-hide-resolved" :model-value="tweaks.hideResolved" @update:model-value="(value) => setKey('hideResolved', value)" />
				</div>
			</div>
		</section>

		<div class="h-px" :style="{ background: 'var(--gd-border)' }" />

		<section>
			<div class="mb-2 text-[10.5px] font-semibold uppercase tracking-wider" :style="{ color: 'var(--gd-text-muted)' }">Theme</div>
			<div class="flex flex-col gap-3">
				<div class="flex items-center justify-between gap-3">
					<Label class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Mode </Label>
					<div
						class="inline-flex items-center"
						role="radiogroup"
						aria-label="Theme mode"
						:style="{
							gap: '2px',
							padding: '2px',
							borderRadius: '8px',
							background: 'var(--gd-panel-2)',
							border: '1px solid var(--gd-border)',
						}"
					>
						<button
							v-for="opt in themeOptions"
							:key="opt.value"
							type="button"
							role="radio"
							:aria-checked="tweaks.theme === opt.value"
							:aria-label="opt.label"
							:title="opt.label"
							class="inline-flex items-center justify-center"
							:style="{
								width: '28px',
								height: '24px',
								borderRadius: '6px',
								border: 0,
								cursor: 'pointer',
								background: tweaks.theme === opt.value ? 'var(--gd-bg)' : 'transparent',
								color: tweaks.theme === opt.value ? 'var(--gd-text)' : 'var(--gd-text-3)',
								boxShadow: tweaks.theme === opt.value ? '0 1px 2px rgb(0 0 0 / 0.08), 0 0 0 1px var(--gd-border)' : 'none',
								transition: 'background 140ms ease, color 140ms ease',
							}"
							@click="setKey('theme', opt.value)"
						>
							<component :is="opt.icon" :size="13" />
						</button>
					</div>
				</div>

				<div class="flex items-center justify-between gap-3">
					<Label class="text-[12.5px] font-normal" :style="{ color: 'var(--gd-text-2)' }"> Accent </Label>
					<div class="flex items-center gap-2">
						<button
							v-for="a in accentList"
							:key="a.key"
							type="button"
							:title="a.name"
							:aria-label="a.name"
							:aria-pressed="tweaks.accent === a.key"
							class="cursor-pointer p-0"
							:style="{
								width: '22px',
								height: '22px',
								borderRadius: '6px',
								background: `linear-gradient(135deg, ${a.strong}, ${a.hex})`,
								border: tweaks.accent === a.key ? '2px solid var(--gd-text)' : '2px solid var(--gd-border)',
								boxShadow: tweaks.accent === a.key ? '0 0 0 2px var(--gd-panel)' : 'none',
							}"
							@click="setKey('accent', a.key)"
						/>
					</div>
				</div>
			</div>
		</section>
	</div>
</template>
