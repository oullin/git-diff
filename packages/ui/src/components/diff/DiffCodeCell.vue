<script setup lang="ts">
withDefaults(
	defineProps<{
		html: string;
		sign?: string;
	}>(),
	{ sign: ' ' },
);
</script>

<template>
	<div
		:style="{
			flex: 1,
			minWidth: 0,
			padding: '0 12px',
			whiteSpace: 'pre',
			color: 'var(--gd-text)',
		}"
	>
		<span
			:style="{
				display: 'inline-block',
				width: '12px',
				color: 'var(--gd-text-muted)',
			}"
			>{{ sign }}</span
		><code data-diff-line-text v-html="html" />
	</div>
</template>
