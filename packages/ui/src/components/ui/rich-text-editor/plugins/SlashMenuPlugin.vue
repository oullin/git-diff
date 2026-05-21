<script setup lang="ts">
import { computed, h, ref, type Component } from "vue";
import { Teleport } from "vue";
import type { ResolvedRichTextFeatures } from "@rich-text-editor/features";
import { useLexicalComposer, TypeaheadMenuPlugin, MenuOption } from "lexical-vue";
import {
    $createParagraphNode,
    $getSelection,
    $isRangeSelection,
    type LexicalEditor,
    type TextNode,
} from "lexical";
import { $setBlocksType } from "@lexical/selection";
import { $createHeadingNode, $createQuoteNode } from "@lexical/rich-text";
import { $createCodeNode } from "@lexical/code";
import {
    INSERT_CHECK_LIST_COMMAND,
    INSERT_ORDERED_LIST_COMMAND,
    INSERT_UNORDERED_LIST_COMMAND,
} from "@lexical/list";
import { INSERT_TABLE_COMMAND } from "@lexical/table";
import {
    Code,
    Heading1,
    Heading2,
    Heading3,
    Image as ImageIcon,
    List,
    ListChecks,
    ListOrdered,
    Pilcrow,
    Quote,
    Table as TableIcon,
} from "lucide-vue-next";

type Props = {
    features?: ResolvedRichTextFeatures;
};
const props = defineProps<Props>();

type Emits = {
    "request-image": [];
};
const emit = defineEmits<Emits>();

class SlashOption extends MenuOption {
    label: string;
    icon: Component;
    keywords: string[];
    perform: (editor: LexicalEditor) => void;

    constructor(
        label: string,
        icon: Component,
        keywords: string[],
        perform: (editor: LexicalEditor) => void,
    ) {
        super(label);
        this.label = label;
        this.icon = icon;
        this.keywords = keywords;
        this.perform = perform;
    }
}

const editor = useLexicalComposer();

function clearSelectionAndRun(
    fn: (editor: LexicalEditor) => void,
): (editor: LexicalEditor) => void {
    return (ed) => fn(ed);
}

const baseOptions: SlashOption[] = [
    new SlashOption("Paragraph", Pilcrow, ["paragraph", "p", "text"], (ed) => {
        ed.update(() => {
            const selection = $getSelection();

            if ($isRangeSelection(selection)) {
                $setBlocksType(selection, () => $createParagraphNode());
            }
        });
    }),
    new SlashOption("Heading 1", Heading1, ["h1", "title"], (ed) => {
        ed.update(() => {
            const selection = $getSelection();

            if ($isRangeSelection(selection)) {
                $setBlocksType(selection, () => $createHeadingNode("h1"));
            }
        });
    }),
    new SlashOption("Heading 2", Heading2, ["h2"], (ed) => {
        ed.update(() => {
            const selection = $getSelection();

            if ($isRangeSelection(selection)) {
                $setBlocksType(selection, () => $createHeadingNode("h2"));
            }
        });
    }),
    new SlashOption("Heading 3", Heading3, ["h3"], (ed) => {
        ed.update(() => {
            const selection = $getSelection();

            if ($isRangeSelection(selection)) {
                $setBlocksType(selection, () => $createHeadingNode("h3"));
            }
        });
    }),
    new SlashOption("Bulleted list", List, ["ul", "list", "bullet"], (ed) => {
        ed.dispatchCommand(INSERT_UNORDERED_LIST_COMMAND, undefined);
    }),
    new SlashOption("Numbered list", ListOrdered, ["ol", "ordered", "numbered"], (ed) => {
        ed.dispatchCommand(INSERT_ORDERED_LIST_COMMAND, undefined);
    }),
    new SlashOption("Checklist", ListChecks, ["check", "todo"], (ed) => {
        ed.dispatchCommand(INSERT_CHECK_LIST_COMMAND, undefined);
    }),
    new SlashOption("Quote", Quote, ["quote", "blockquote"], (ed) => {
        ed.update(() => {
            const selection = $getSelection();

            if ($isRangeSelection(selection)) {
                $setBlocksType(selection, () => $createQuoteNode());
            }
        });
    }),
    new SlashOption("Code block", Code, ["code", "pre"], (ed) => {
        ed.update(() => {
            const selection = $getSelection();

            if ($isRangeSelection(selection)) {
                $setBlocksType(selection, () => $createCodeNode());
            }
        });
    }),
    new SlashOption("Table", TableIcon, ["table"], (ed) => {
        ed.dispatchCommand(INSERT_TABLE_COMMAND, { rows: "3", columns: "3", includeHeaders: true });
    }),
    new SlashOption("Image", ImageIcon, ["image", "img", "picture"], () => {
        emit("request-image");
    }),
].map((opt) => {
    opt.perform = clearSelectionAndRun(opt.perform);

    return opt;
});

const enabledOptions = computed<SlashOption[]>(() => {
    const features = props.features;

    return baseOptions.filter((opt) => {
        if (opt.label === "Table") {
            return features?.tables ?? true;
        }

        if (opt.label === "Image") {
            return features?.images ?? true;
        }

        if (opt.label === "Checklist") {
            return features?.checklist ?? true;
        }

        return true;
    });
});

const queryString = ref<string | null>("");

function filterOptions(query: string | null): SlashOption[] {
    const source = enabledOptions.value;

    if (!query) {
        return source;
    }

    const lower = query.toLowerCase();

    return source.filter(
        (opt) =>
            opt.label.toLowerCase().includes(lower) ||
            opt.keywords.some((kw) => kw.includes(lower)),
    );
}

function triggerFn(
    text: string,
): { leadOffset: number; matchingString: string; replaceableString: string } | null {
    const match = /(?:^|\s)\/(\w*)$/.exec(text);

    if (match === null) {
        return null;
    }

    const matchingString = match[1] ?? "";

    return {
        leadOffset: match.index + (match[0].startsWith("/") ? 0 : 1),
        matchingString,
        replaceableString: `/${matchingString}`,
    };
}

function onSelectOption({
    option,
    closeMenu,
    textNodeContainingQuery,
}: {
    option: SlashOption;
    closeMenu: () => void;
    textNodeContainingQuery: TextNode | null;
    matchingString: string;
}): void {
    editor.update(() => {
        if (textNodeContainingQuery !== null) {
            textNodeContainingQuery.remove();
        }
    });
    option.perform(editor);
    closeMenu();
}

function onQueryChange(value: string | null): void {
    queryString.value = value;
}
</script>

<template>
    <TypeaheadMenuPlugin
        :options="filterOptions(queryString)"
        :trigger-fn="triggerFn"
        @query-change="onQueryChange"
        @select-option="onSelectOption"
    >
        <template #default="{ anchorElementRef, itemProps, matchingString }">
            <Teleport v-if="anchorElementRef" :to="anchorElementRef">
                <div
                    class="z-50 max-h-72 w-60 overflow-y-auto rounded-md border border-border bg-popover p-1 text-sm shadow-md"
                >
                    <button
                        v-for="(option, index) in itemProps.options"
                        :key="option.key"
                        type="button"
                        :ref="(el) => option.setRefElement(el as Element | null)"
                        class="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left hover:bg-accent"
                        :class="
                            itemProps.selectedIndex === index
                                ? 'bg-accent text-accent-foreground'
                                : ''
                        "
                        @mousedown.prevent
                        @click="itemProps.selectOptionAndCleanUp(option as SlashOption)"
                        @mouseenter="itemProps.setHighlightedIndex(index)"
                    >
                        <component :is="(option as SlashOption).icon" class="h-4 w-4" />
                        <span>{{ (option as SlashOption).label }}</span>
                    </button>
                    <div
                        v-if="itemProps.options.length === 0"
                        class="px-2 py-1.5 text-xs text-muted-foreground"
                    >
                        No commands matching "{{ matchingString }}"
                    </div>
                </div>
            </Teleport>
        </template>
    </TypeaheadMenuPlugin>
</template>
