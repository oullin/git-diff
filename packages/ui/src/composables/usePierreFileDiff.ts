import { parseDiffFromFile, processFile, type FileContents, type FileDiffMetadata } from '@pierre/diffs';
import type { ChangedFile, DiffSection, DiffSectionKind } from '@git-diff/domain';

// Data layer for the @pierre/diffs renderer. Resolves a FileDiffMetadata for
// one DiffSection, preferring "full-file mode" (fetch the whole old + new file
// contents at the right git refs and let the library compute the diff, which
// unlocks native context expansion and honours ignore-whitespace). When a side
// can't be fetched (blob > 2 MB cap, missing ref — e.g. a PR base that isn't
// surfaced yet), it falls back to the pre-computed git patch via processFile()
// so the diff always renders, just without native expansion.

export interface DiffSourceContext {
	repoRoot: string;
	/** Head SHA when reviewing a commit/PR; empty/undefined = working tree. */
	commitRef?: string;
	/** PR base SHA when reviewing a pull request; absent for plain commits. */
	baseRef?: string;
	/** Honoured only on the full-file path (the patch is already computed). */
	ignoreWhitespace: boolean;
}

export interface SectionRefs {
	/** git ref for the OLD side; null when there is no old side (untracked). */
	oldRef: string | null;
	/** git ref for the NEW side; '' = working tree, ':0' = index. */
	newRef: string | null;
}

export interface ResolvedFileDiff {
	meta: FileDiffMetadata | undefined;
	/** true = patch fallback (no native expansion); false = full-file. */
	partial: boolean;
}

/**
 * Maps a section kind + diff context to the (oldRef, newRef) pair to read whole
 * file contents from. Pure + unit-tested.
 *   working/staged   : HEAD ↔ :0 (index)
 *   working/unstaged : :0   ↔ '' (working tree)
 *   working/untracked: —    ↔ '' (new file only)
 *   commit           : sha^ ↔ sha
 *   pull request     : baseRef ↔ sha
 */
export function resolveSectionRefs(kind: DiffSectionKind, ctx: DiffSourceContext): SectionRefs {
	switch (kind) {
		case 'staged':
			return { oldRef: 'HEAD', newRef: ':0' };

		case 'unstaged':
			return { oldRef: ':0', newRef: '' };

		case 'untracked':
			return { oldRef: null, newRef: '' };

		case 'commit': {
			const head = ctx.commitRef && ctx.commitRef !== '' ? ctx.commitRef : 'HEAD';
			const old = ctx.baseRef && ctx.baseRef !== '' ? ctx.baseRef : `${head}^`;

			return { oldRef: old, newRef: head };
		}
	}
}

async function readText(root: string, path: string, ref: string): Promise<string | null> {
	try {
		const { data } = await window.diffApp.readRepositoryFileBytes({ root, path, ref });

		return new TextDecoder().decode(data);
	} catch {
		// Missing ref, blob too large (413), or binary — caller falls back.
		return null;
	}
}

export async function resolveFileDiff(file: ChangedFile, section: DiffSection, ctx: DiffSourceContext): Promise<ResolvedFileDiff> {
	const patchFallback = (): ResolvedFileDiff => ({
		meta: processFile(section.patch, { isGitDiff: true, cacheKey: section.id }),
		partial: true,
	});

	if (section.binary || file.binary) {
		return { meta: undefined, partial: true };
	}

	// PR diffs are computed by the backend against the merge-base, but the only
	// base ref surfaced here is a (possibly-moved) branch name — recomputing a
	// full-file diff from its current tip could drift from the real PR diff. The
	// pre-computed patch is the accurate source, so PRs use patch mode. (Native
	// expansion for PRs needs the exact base SHA surfaced — a future change.)
	if (section.kind === 'commit' && ctx.baseRef) {
		return patchFallback();
	}

	const refs = resolveSectionRefs(section.kind, ctx);
	const oldPath = file.oldPath ?? file.path;

	// Added files have no old side; deleted files have no new side. An empty
	// string is a valid FileContents (pure addition / deletion).
	const oldText = file.status === 'added' || refs.oldRef === null ? '' : await readText(ctx.repoRoot, oldPath, refs.oldRef);

	const newText = file.status === 'deleted' || refs.newRef === null ? '' : await readText(ctx.repoRoot, file.path, refs.newRef);

	if (oldText === null || newText === null) {
		return patchFallback();
	}

	try {
		const oldFile: FileContents = { name: oldPath, contents: oldText, cacheKey: `${section.id}:old` };
		const newFile: FileContents = { name: file.path, contents: newText, cacheKey: `${section.id}:new` };
		const meta = parseDiffFromFile(oldFile, newFile, { ignoreWhitespace: ctx.ignoreWhitespace }, false);

		return { meta, partial: false };
	} catch {
		return patchFallback();
	}
}
