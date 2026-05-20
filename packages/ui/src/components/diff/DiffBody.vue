<script setup lang="ts">
import { computed, ref } from "vue";
import { ChevronDown, ChevronUp, Plus } from "lucide-vue-next";
import { highlighterRev, highlightLine, languageFor } from "@lib/highlight";
import { diffBgs, type DiffStyleColors } from "@lib/accent";
import { parsePatch, splitPatch, type PatchLine, type SplitRow } from "@lib/patch";
import { computeWordHi, type Range } from "@lib/wordHi";
import CommentThread from "./CommentThread.vue";
import SplitHandle from "./SplitHandle.vue";
import type { ChangedFile, DiffHunkStyle, DiffSection, DiffViewMode, ReviewComment } from "@api";
import type { RichTextFeatures } from "@ui/rich-text-editor";

const props = withDefaults(
  defineProps<{
    file: ChangedFile;
    viewMode: DiffViewMode;
    diffStyle: DiffHunkStyle;
    density: "comfortable" | "compact";
    wordHighlight: boolean;
    hideWhitespace: boolean;
    comments: ReviewComment[];
    replyFeatures: RichTextFeatures;
    splitRatio?: number;
  }>(),
  { splitRatio: 0.5 },
);

const emit = defineEmits<{
  "add-comment": [section: DiffSection, line: PatchLine];
  "delete-comment": [comment: ReviewComment];
  "reply-comment": [parent: ReviewComment, bodyHtml: string];
  "update:splitRatio": [value: number];
}>();

const wrapperStyle = computed(() => {
  const left = Math.min(80, Math.max(20, props.splitRatio * 100));
  const right = 100 - left;
  return {
    minWidth: props.viewMode === "split" ? "1800px" : "1200px",
    position: "relative" as const,
    "--gd-split-cols": `${left.toFixed(2)}% ${right.toFixed(2)}%`,
  };
});

const sectionRefs = ref<Record<string, HTMLElement | null>>({});

function setSectionRef(id: string) {
  return (el: unknown) => {
    sectionRefs.value[id] = el instanceof HTMLElement ? el : null;
  };
}

const lineH = computed(() => (props.density === "compact" ? 22 : 24));
const colors = computed<DiffStyleColors>(() => diffBgs(props.diffStyle));
const lang = computed(() => languageFor(props.file.path));

function commentsForLine(section: DiffSection, line: PatchLine | undefined): ReviewComment[] {
  if (!line) return [];
  return props.comments.filter((c) => {
    if (c.filePath !== props.file.path || c.diffSection !== section.kind) return false;
    const lineNumberForSide = c.side === "left" ? line.oldLine : line.newLine;
    return lineNumberForSide === c.lineNumber;
  });
}

function highlightHtml(text: string): string {
  // Read the rev so Vue re-runs this when a language finishes loading.
  void highlighterRev.value;
  return highlightLine(text || " ", lang.value);
}

function withRanges(text: string, ranges: Range[], cls: "wh-add" | "wh-rem"): string {
  if (!ranges.length) {
    return escapeHtml(text);
  }
  const out: string[] = [];
  let cursor = 0;
  for (const [start, end] of ranges) {
    if (start > cursor) out.push(escapeHtml(text.slice(cursor, start)));
    out.push(`<span class="${cls}">${escapeHtml(text.slice(start, end))}</span>`);
    cursor = end;
  }
  if (cursor < text.length) out.push(escapeHtml(text.slice(cursor)));
  return out.join("");
}

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (ch) => {
    switch (ch) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      default:
        return "&#39;";
    }
  });
}

interface CellRender {
  kind: "ctx" | "add" | "rem" | "empty";
  num: number | "";
  side: "left" | "right";
  html: string;
}

function renderPair(row: Extract<SplitRow, { kind: "pair" }>): {
  left: CellRender;
  right: CellRender;
} {
  const left = row.left;
  const right = row.right;
  let leftHtml = "";
  let rightHtml = "";
  let leftKind: CellRender["kind"] = left ? "rem" : "empty";
  let rightKind: CellRender["kind"] = right ? "add" : "empty";
  if (left && right && props.wordHighlight) {
    const { hiL, hiR } = computeWordHi(left.text, right.text);
    leftHtml = withRanges(left.text, hiL, "wh-rem");
    rightHtml = withRanges(right.text, hiR, "wh-add");
  } else {
    leftHtml = left ? highlightHtml(left.text) : "";
    rightHtml = right ? highlightHtml(right.text) : "";
  }
  return {
    left: { kind: leftKind, num: left?.oldLine ?? "", side: "left", html: leftHtml },
    right: { kind: rightKind, num: right?.newLine ?? "", side: "right", html: rightHtml },
  };
}

function bgFor(kind: CellRender["kind"]): string {
  if (kind === "add") return colors.value.addBg;
  if (kind === "rem") return colors.value.remBg;
  return "transparent";
}

function numBgFor(kind: CellRender["kind"]): string {
  if (kind === "add") return colors.value.addNum;
  if (kind === "rem") return colors.value.remNum;
  return "transparent";
}

function barFor(kind: CellRender["kind"]): string {
  if (kind === "add") return colors.value.addBar;
  if (kind === "rem") return colors.value.remBar;
  return "transparent";
}

function sign(kind: CellRender["kind"]): string {
  if (kind === "add") return "+";
  if (kind === "rem") return "−";
  return " ";
}

function hunkHeaderText(text: string): { range: string; trailer: string } {
  const parts = text.split("@@");
  if (parts.length < 3) return { range: text, trailer: "" };
  return {
    range: `@@${parts[1]}@@`,
    trailer: parts.slice(2).join("@@").trim(),
  };
}
</script>

<template>
  <div
    :style="{
      fontFamily: 'var(--font-mono)',
      fontSize: '13.5px',
      lineHeight: `${lineH}px`,
      background: 'var(--gd-bg-code, var(--gd-bg))',
    }"
  >
    <section v-for="section in file.sections" :key="section.id">
      <div data-diff-scroller="true" :style="{ overflowX: 'auto', overflowY: 'visible' }">
        <div :ref="setSectionRef(section.id)" :style="wrapperStyle">
          <SplitHandle
            v-if="viewMode === 'split'"
            :ratio="splitRatio"
            :container-el="sectionRefs[section.id] ?? null"
            @update:ratio="(value) => emit('update:splitRatio', value)"
          />
          <template v-if="viewMode === 'split'">
            <template v-for="row in splitPatch(section, hideWhitespace)" :key="row.id">
              <template v-if="row.kind === 'meta'">
                <div
                  v-if="row.line.text.startsWith('@@')"
                  class="flex items-center"
                  data-hunk-anchor
                  :style="{
                    gap: '10px',
                    padding: '8px 16px',
                    background: 'var(--gd-bg-gutter, var(--gd-panel))',
                    color: 'var(--gd-text-3)',
                    fontSize: '13px',
                    borderTop: '1px solid var(--gd-border-soft)',
                    borderBottom: '1px solid var(--gd-border-soft)',
                    boxShadow: '0 1px 0 var(--gd-edge-hi-2) inset',
                  }"
                >
                  <ChevronDown :size="12" />
                  <span :style="{ fontFamily: 'var(--font-mono)', color: 'var(--gd-accent)' }">{{
                    hunkHeaderText(row.line.text).range
                  }}</span>
                  <span :style="{ color: 'var(--gd-text-3)' }">{{
                    hunkHeaderText(row.line.text).trailer
                  }}</span>
                  <div class="flex-1" />
                  <button
                    type="button"
                    title="Expand context up"
                    :style="{
                      width: '26px',
                      height: '26px',
                      borderRadius: '6px',
                      border: '1px solid transparent',
                      background: 'transparent',
                      color: 'var(--gd-text-3)',
                      display: 'inline-flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      padding: 0,
                      cursor: 'pointer',
                    }"
                  >
                    <ChevronUp :size="12" />
                  </button>
                  <button
                    type="button"
                    title="Expand context down"
                    :style="{
                      width: '26px',
                      height: '26px',
                      borderRadius: '6px',
                      border: '1px solid transparent',
                      background: 'transparent',
                      color: 'var(--gd-text-3)',
                      display: 'inline-flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      padding: 0,
                      cursor: 'pointer',
                    }"
                  >
                    <ChevronDown :size="12" />
                  </button>
                </div>
              </template>
              <template v-else-if="row.kind === 'context'">
                <div
                  class="gd-row relative"
                  :style="{ display: 'grid', gridTemplateColumns: 'var(--gd-split-cols, 1fr 1fr)' }"
                >
                  <div
                    class="flex"
                    :style="{
                      background: 'transparent',
                      minHeight: `${lineH}px`,
                      minWidth: 0,
                      overflow: 'hidden',
                    }"
                  >
                    <div
                      :style="{
                        width: '56px',
                        flexShrink: 0,
                        textAlign: 'right',
                        padding: '0 10px 0 4px',
                        color: 'var(--gd-text-muted)',
                        fontSize: '12px',
                        fontVariantNumeric: 'tabular-nums',
                        userSelect: 'none',
                      }"
                    >
                      {{ row.line.oldLine ?? "" }}
                    </div>
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
                        >&nbsp;</span
                      ><code data-diff-line-text v-html="highlightHtml(row.line.text)" />
                    </div>
                  </div>
                  <div
                    class="flex"
                    :style="{
                      background: 'transparent',
                      borderLeft: '1px solid var(--gd-border-soft)',
                      minHeight: `${lineH}px`,
                      minWidth: 0,
                      overflow: 'hidden',
                    }"
                  >
                    <div
                      :style="{
                        width: '56px',
                        flexShrink: 0,
                        textAlign: 'right',
                        padding: '0 10px 0 4px',
                        color: 'var(--gd-text-muted)',
                        fontSize: '12px',
                        fontVariantNumeric: 'tabular-nums',
                        userSelect: 'none',
                      }"
                    >
                      {{ row.line.newLine ?? "" }}
                    </div>
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
                        >&nbsp;</span
                      ><code data-diff-line-text v-html="highlightHtml(row.line.text)" />
                    </div>
                  </div>
                  <button
                    type="button"
                    class="gd-add-comment"
                    title="Add comment"
                    @click="emit('add-comment', section, row.line)"
                  >
                    <Plus :size="12" :stroke-width="2.5" />
                  </button>
                </div>
                <template v-for="c in commentsForLine(section, row.line)" :key="c.id">
                  <CommentThread
                    :comment="c"
                    :reply-features="replyFeatures"
                    @delete="emit('delete-comment', c)"
                    @reply="(body) => emit('reply-comment', c, body)"
                  />
                </template>
              </template>
              <template v-else>
                <div
                  class="gd-row relative"
                  :style="{ display: 'grid', gridTemplateColumns: 'var(--gd-split-cols, 1fr 1fr)' }"
                >
                  <div
                    class="flex"
                    :style="{
                      background: bgFor(renderPair(row).left.kind),
                      minHeight: `${lineH}px`,
                      minWidth: 0,
                      overflow: 'hidden',
                    }"
                  >
                    <div
                      :style="{
                        width: '56px',
                        flexShrink: 0,
                        textAlign: 'right',
                        padding: '0 10px 0 4px',
                        background: numBgFor(renderPair(row).left.kind),
                        color: 'var(--gd-text-muted)',
                        fontSize: '12px',
                        fontVariantNumeric: 'tabular-nums',
                        userSelect: 'none',
                        position: 'relative',
                      }"
                    >
                      <div
                        v-if="renderPair(row).left.kind === 'rem'"
                        :style="{
                          position: 'absolute',
                          left: 0,
                          top: 0,
                          bottom: 0,
                          width: '2px',
                          background: barFor('rem'),
                        }"
                      />
                      {{ renderPair(row).left.num }}
                    </div>
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
                        >{{ sign(renderPair(row).left.kind) }}</span
                      ><code data-diff-line-text v-html="renderPair(row).left.html" />
                    </div>
                  </div>
                  <div
                    class="flex"
                    :style="{
                      background: bgFor(renderPair(row).right.kind),
                      borderLeft: '1px solid var(--gd-border-soft)',
                      minHeight: `${lineH}px`,
                      minWidth: 0,
                      overflow: 'hidden',
                    }"
                  >
                    <div
                      :style="{
                        width: '56px',
                        flexShrink: 0,
                        textAlign: 'right',
                        padding: '0 10px 0 4px',
                        background: numBgFor(renderPair(row).right.kind),
                        color: 'var(--gd-text-muted)',
                        fontSize: '12px',
                        fontVariantNumeric: 'tabular-nums',
                        userSelect: 'none',
                        position: 'relative',
                      }"
                    >
                      <div
                        v-if="renderPair(row).right.kind === 'add'"
                        :style="{
                          position: 'absolute',
                          left: 0,
                          top: 0,
                          bottom: 0,
                          width: '2px',
                          background: barFor('add'),
                        }"
                      />
                      {{ renderPair(row).right.num }}
                    </div>
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
                        >{{ sign(renderPair(row).right.kind) }}</span
                      ><code data-diff-line-text v-html="renderPair(row).right.html" />
                    </div>
                  </div>
                  <button
                    v-if="row.left || row.right"
                    type="button"
                    class="gd-add-comment"
                    title="Add comment"
                    @click="emit('add-comment', section, (row.right ?? row.left) as PatchLine)"
                  >
                    <Plus :size="12" :stroke-width="2.5" />
                  </button>
                </div>
                <template
                  v-for="line in [row.left, row.right].filter(Boolean) as PatchLine[]"
                  :key="`${row.id}:${line.id}`"
                >
                  <CommentThread
                    v-for="c in commentsForLine(section, line)"
                    :key="c.id"
                    :comment="c"
                    :reply-features="replyFeatures"
                    @delete="emit('delete-comment', c)"
                    @reply="(body) => emit('reply-comment', c, body)"
                  />
                </template>
              </template>
            </template>
          </template>
          <template v-else>
            <template v-for="line in parsePatch(section, hideWhitespace)" :key="line.id">
              <div
                v-if="line.type === 'meta' && line.text.startsWith('@@')"
                class="flex items-center"
                data-hunk-anchor
                :style="{
                  gap: '10px',
                  padding: '8px 16px',
                  background: 'var(--gd-bg-gutter, var(--gd-panel))',
                  color: 'var(--gd-text-3)',
                  fontSize: '13px',
                  borderTop: '1px solid var(--gd-border-soft)',
                  borderBottom: '1px solid var(--gd-border-soft)',
                  boxShadow: '0 1px 0 var(--gd-edge-hi-2) inset',
                }"
              >
                <ChevronDown :size="12" />
                <span :style="{ fontFamily: 'var(--font-mono)', color: 'var(--gd-accent)' }">{{
                  hunkHeaderText(line.text).range
                }}</span>
                <span :style="{ color: 'var(--gd-text-3)' }">{{
                  hunkHeaderText(line.text).trailer
                }}</span>
              </div>
              <template v-else-if="line.type !== 'meta'">
                <div
                  class="gd-row relative flex"
                  :style="{
                    background: bgFor(
                      line.type === 'add' ? 'add' : line.type === 'del' ? 'rem' : 'ctx',
                    ),
                    minHeight: `${lineH}px`,
                  }"
                >
                  <div
                    :style="{
                      width: '56px',
                      flexShrink: 0,
                      textAlign: 'right',
                      padding: '0 10px 0 4px',
                      background: numBgFor(
                        line.type === 'add' ? 'add' : line.type === 'del' ? 'rem' : 'ctx',
                      ),
                      color: 'var(--gd-text-muted)',
                      fontSize: '12px',
                      fontVariantNumeric: 'tabular-nums',
                      userSelect: 'none',
                      position: 'relative',
                    }"
                  >
                    {{ line.oldLine ?? "" }}
                  </div>
                  <div
                    :style="{
                      width: '56px',
                      flexShrink: 0,
                      textAlign: 'right',
                      padding: '0 10px 0 4px',
                      background: numBgFor(
                        line.type === 'add' ? 'add' : line.type === 'del' ? 'rem' : 'ctx',
                      ),
                      color: 'var(--gd-text-muted)',
                      fontSize: '12px',
                      fontVariantNumeric: 'tabular-nums',
                      userSelect: 'none',
                      position: 'relative',
                    }"
                  >
                    <div
                      v-if="line.type === 'add' || line.type === 'del'"
                      :style="{
                        position: 'absolute',
                        left: 0,
                        top: 0,
                        bottom: 0,
                        width: '2px',
                        background: barFor(line.type === 'add' ? 'add' : 'rem'),
                      }"
                    />
                    {{ line.newLine ?? "" }}
                  </div>
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
                      >{{
                        sign(line.type === "add" ? "add" : line.type === "del" ? "rem" : "ctx")
                      }}</span
                    ><code data-diff-line-text v-html="highlightHtml(line.text)" />
                  </div>
                  <button
                    type="button"
                    class="gd-add-comment"
                    title="Add comment"
                    @click="emit('add-comment', section, line)"
                  >
                    <Plus :size="12" :stroke-width="2.5" />
                  </button>
                </div>
                <template v-for="c in commentsForLine(section, line)" :key="c.id">
                  <CommentThread
                    :comment="c"
                    :reply-features="replyFeatures"
                    @delete="emit('delete-comment', c)"
                    @reply="(body) => emit('reply-comment', c, body)"
                  />
                </template>
              </template>
            </template>
          </template>
        </div>
      </div>
    </section>
  </div>
</template>
