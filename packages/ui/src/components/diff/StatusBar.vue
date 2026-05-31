<script setup lang="ts">
import Kbd from '@diff/Kbd.vue';

defineProps<{
	filePath: string;
	viewedCount: number;
	total: number;
}>();
</script>

<template>
	<div class="gd-statusbar">
		<span class="inline-flex items-center" :style="{ gap: '5px' }">
			<span
				:style="{
					width: '7px',
					height: '7px',
					borderRadius: '50%',
					background: 'var(--gd-added)',
				}"
			/>
			Synced
		</span>
		<span :style="{ color: 'var(--gd-text-muted)' }">·</span>
		<span :style="{ fontFamily: 'var(--font-mono)' }">{{ filePath || '—' }}</span>

		<div class="flex-1" />

		<span class="inline-flex items-center" :style="{ gap: '4px' }">
			<Kbd>J / K</Kbd>
			<span>Navigate</span>
		</span>
		<span class="inline-flex items-center" :style="{ gap: '4px' }">
			<Kbd>V</Kbd>
			<span>Mark viewed</span>
		</span>

		<div class="flex-1" />

		<span>{{ viewedCount }}/{{ total }} viewed</span>
	</div>
</template>
