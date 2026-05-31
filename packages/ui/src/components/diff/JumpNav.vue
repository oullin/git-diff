<script setup lang="ts">
import { ArrowDown, ArrowUp } from 'lucide-vue-next';

defineProps<{ index: number; total: number }>();
defineEmits<{ prev: []; next: [] }>();
</script>

<template>
	<div class="gd-jump-nav">
		<button class="gd-jump-btn" type="button" title="Previous change (K)" :disabled="index <= 0" @click="$emit('prev')">
			<ArrowUp :size="13" />
		</button>
		<div :style="{ height: '1px', background: 'var(--gd-border)', margin: '2px 4px' }" />
		<span
			:style="{
				fontSize: '10.5px',
				color: 'var(--gd-text-3)',
				textAlign: 'center',
				padding: '2px 0',
				fontFamily: 'var(--font-mono)',
				fontWeight: 600,
			}"
			>{{ Math.max(index + 1, 0) }}/{{ Math.max(total, 1) }}</span
		>
		<div :style="{ height: '1px', background: 'var(--gd-border)', margin: '2px 4px' }" />
		<button class="gd-jump-btn" type="button" title="Next change (J)" :disabled="index >= total - 1" @click="$emit('next')">
			<ArrowDown :size="13" />
		</button>
	</div>
</template>
