<script setup lang="ts">
import { ref } from "vue";
import { Settings2, X } from "lucide-vue-next";
import { ACCENTS, type Accent } from "@lib/accent";
import type { Tweaks } from "@composables/useTweaks";

defineProps<{ tweaks: Tweaks }>();

const emit = defineEmits<{
  "update:tweak": [key: keyof Tweaks, value: Tweaks[keyof Tweaks]];
}>();

const open = ref(false);

function setKey<K extends keyof Tweaks>(key: K, value: Tweaks[K]) {
  emit("update:tweak", key, value);
}

const accentList = Object.values(ACCENTS) as Accent[];
</script>

<template>
  <button v-if="!open" class="gd-tweaks-toggle" type="button" title="Tweaks" @click="open = true">
    <Settings2 :size="13" />
    Tweaks
  </button>
  <div v-else class="gd-tweaks-panel">
    <div
      class="flex items-center justify-between"
      :style="{
        padding: '10px 12px',
        borderBottom: '1px solid var(--gd-border)',
      }"
    >
      <span :style="{ fontSize: '13px', fontWeight: 600, color: 'var(--gd-text)' }">Tweaks</span>
      <button
        type="button"
        :style="{
          width: '24px',
          height: '24px',
          borderRadius: '6px',
          background: 'transparent',
          border: 0,
          color: 'var(--gd-text-3)',
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: 'pointer',
        }"
        @click="open = false"
      >
        <X :size="14" />
      </button>
    </div>
    <div :style="{ padding: '10px 12px', display: 'flex', flexDirection: 'column', gap: '14px' }">
      <section>
        <div
          :style="{
            fontSize: '11px',
            fontWeight: 600,
            letterSpacing: '0.4px',
            textTransform: 'uppercase',
            color: 'var(--gd-text-muted)',
            marginBottom: '8px',
          }"
        >
          Layout
        </div>
        <div class="flex flex-col" :style="{ gap: '8px' }">
          <label class="flex items-center justify-between" :style="{ fontSize: '12.5px' }">
            <span :style="{ color: 'var(--gd-text-2)' }">View mode</span>
            <select
              :value="tweaks.viewMode"
              :style="{
                background: 'var(--gd-panel-2)',
                color: 'var(--gd-text)',
                border: '1px solid var(--gd-border)',
                borderRadius: '6px',
                padding: '2px 6px',
              }"
              @change="
                setKey('viewMode', ($event.target as HTMLSelectElement).value as Tweaks['viewMode'])
              "
            >
              <option value="split">Split</option>
              <option value="unified">Unified</option>
            </select>
          </label>
          <label class="flex items-center justify-between" :style="{ fontSize: '12.5px' }">
            <span :style="{ color: 'var(--gd-text-2)' }">Density</span>
            <select
              :value="tweaks.density"
              :style="{
                background: 'var(--gd-panel-2)',
                color: 'var(--gd-text)',
                border: '1px solid var(--gd-border)',
                borderRadius: '6px',
                padding: '2px 6px',
              }"
              @change="
                setKey('density', ($event.target as HTMLSelectElement).value as Tweaks['density'])
              "
            >
              <option value="comfortable">Comfortable</option>
              <option value="compact">Compact</option>
            </select>
          </label>
          <label class="flex items-center justify-between" :style="{ fontSize: '12.5px' }">
            <span :style="{ color: 'var(--gd-text-2)' }">Show minimap</span>
            <input
              type="checkbox"
              :checked="tweaks.showMinimap"
              @change="setKey('showMinimap', ($event.target as HTMLInputElement).checked)"
            />
          </label>
          <label class="flex items-center justify-between" :style="{ fontSize: '12.5px' }">
            <span :style="{ color: 'var(--gd-text-2)' }">Status bar</span>
            <input
              type="checkbox"
              :checked="tweaks.showStatusBar"
              @change="setKey('showStatusBar', ($event.target as HTMLInputElement).checked)"
            />
          </label>
        </div>
      </section>

      <section>
        <div
          :style="{
            fontSize: '11px',
            fontWeight: 600,
            letterSpacing: '0.4px',
            textTransform: 'uppercase',
            color: 'var(--gd-text-muted)',
            marginBottom: '8px',
          }"
        >
          Diff
        </div>
        <div class="flex flex-col" :style="{ gap: '8px' }">
          <label class="flex items-center justify-between" :style="{ fontSize: '12.5px' }">
            <span :style="{ color: 'var(--gd-text-2)' }">Hunk style</span>
            <select
              :value="tweaks.diffStyle"
              :style="{
                background: 'var(--gd-panel-2)',
                color: 'var(--gd-text)',
                border: '1px solid var(--gd-border)',
                borderRadius: '6px',
                padding: '2px 6px',
              }"
              @change="
                setKey(
                  'diffStyle',
                  ($event.target as HTMLSelectElement).value as Tweaks['diffStyle'],
                )
              "
            >
              <option value="soft">Soft</option>
              <option value="punchy">Punchy</option>
              <option value="bar">Bar</option>
            </select>
          </label>
          <label class="flex items-center justify-between" :style="{ fontSize: '12.5px' }">
            <span :style="{ color: 'var(--gd-text-2)' }">Word-level highlight</span>
            <input
              type="checkbox"
              :checked="tweaks.wordHighlight"
              @change="setKey('wordHighlight', ($event.target as HTMLInputElement).checked)"
            />
          </label>
        </div>
      </section>

      <section>
        <div
          :style="{
            fontSize: '11px',
            fontWeight: 600,
            letterSpacing: '0.4px',
            textTransform: 'uppercase',
            color: 'var(--gd-text-muted)',
            marginBottom: '8px',
          }"
        >
          Theme
        </div>
        <div class="flex items-center" :style="{ gap: '8px' }">
          <button
            v-for="a in accentList"
            :key="a.key"
            type="button"
            :title="a.name"
            :style="{
              width: '22px',
              height: '22px',
              borderRadius: '6px',
              background: `linear-gradient(135deg, ${a.strong}, ${a.hex})`,
              border:
                tweaks.accent === a.key ? '2px solid var(--gd-text)' : '2px solid transparent',
              cursor: 'pointer',
            }"
            @click="setKey('accent', a.key)"
          />
        </div>
      </section>
    </div>
  </div>
</template>
