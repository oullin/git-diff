<script setup lang="ts">
withDefaults(
	defineProps<{
		num: number | string;
		bg?: string;
		bar?: string;
	}>(),
	{ bg: 'transparent', bar: undefined },
);
</script>

<template>
	<div
		:style="{
			width: '56px',
			flexShrink: 0,
			textAlign: 'right',
			padding: '0 10px 0 4px',
			background: bg,
			color: 'var(--gd-text-muted)',
			fontSize: '12px',
			fontVariantNumeric: 'tabular-nums',
			userSelect: 'none',
			position: 'relative',
		}"
	>
		<div
			v-if="bar"
			:style="{
				position: 'absolute',
				left: 0,
				top: 0,
				bottom: 0,
				width: '2px',
				background: bar,
			}"
		/>
		{{ num }}
	</div>
</template>
