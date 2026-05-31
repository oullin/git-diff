<script setup lang="ts">
import { FolderOpen, GitPullRequest, Plus } from 'lucide-vue-next';

defineProps<{
	kind: 'hero' | 'loading' | 'error';
	lastRepoRoot: string;
	error: string;
}>();

const emit = defineEmits<{
	'add-repository': [];
	'open-last-repo': [];
}>();
</script>

<template>
	<div v-if="kind === 'hero'" class="flex flex-1 flex-col overflow-auto" :style="{ background: 'var(--gd-bg)' }">
		<div class="mx-auto w-full max-w-5xl px-8 pt-16 pb-10">
			<div class="hero">
				<div class="flex h-10 w-10 items-center justify-center rounded-md border" :style="{ borderColor: 'var(--gd-border)', background: 'var(--gd-panel)' }">
					<GitPullRequest class="h-5 w-5" :style="{ color: 'var(--gd-text-3)' }" />
				</div>
				<h1 class="hero-title">A local git diff viewer</h1>
				<p class="hero-subtitle">Inspect changes across your repositories with a fast file tree, side-by-side diffs, and lightweight local reviews.</p>
				<div class="hero-cta-row">
					<button class="toolbar-btn" type="button" @click="emit('add-repository')"><Plus class="h-4 w-4" />Add repository</button>
					<button v-if="lastRepoRoot" class="toolbar-btn" type="button" @click="emit('open-last-repo')"><FolderOpen class="h-4 w-4" />Open last repo</button>
				</div>
				<div class="hero-version">No repository selected.</div>
			</div>
		</div>
	</div>
	<div v-else-if="kind === 'loading'" class="grid flex-1 place-items-center text-sm" :style="{ color: 'var(--gd-text-3)' }">Loading repository…</div>
	<div v-else class="grid flex-1 place-items-center p-8">
		<div class="max-w-xl rounded-md border p-5" :style="{ borderColor: 'var(--gd-border)', background: 'var(--gd-panel)' }">
			<div class="font-semibold">Unable to read repository</div>
			<p class="mt-2 text-sm" :style="{ color: 'var(--gd-text-3)' }">{{ error }}</p>
			<button class="mt-4 toolbar-btn" type="button" @click="emit('add-repository')">Choose another repository</button>
		</div>
	</div>
</template>
