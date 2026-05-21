import { onBeforeUnmount, reactive } from "vue";
import { useLexicalComposer } from "lexical-vue";
import {
    $getSelection,
    $isRangeSelection,
    SELECTION_CHANGE_COMMAND,
    COMMAND_PRIORITY_LOW,
    type LexicalEditor,
} from "lexical";
import { $isLinkNode } from "@lexical/link";
import { $isListNode, ListNode } from "@lexical/list";
import { $isHeadingNode, $isQuoteNode } from "@lexical/rich-text";
import { $isCodeNode } from "@lexical/code";
import { $findMatchingParent, $getNearestNodeOfType, mergeRegister } from "@lexical/utils";

export type BlockType =
    | "paragraph"
    | "h1"
    | "h2"
    | "h3"
    | "h4"
    | "h5"
    | "h6"
    | "quote"
    | "code"
    | "bullet"
    | "number"
    | "check";

export type ToolbarState = {
    isBold: boolean;
    isItalic: boolean;
    isUnderline: boolean;
    isStrikethrough: boolean;
    isCode: boolean;
    isLink: boolean;
    blockType: BlockType;
    canUndo: boolean;
    canRedo: boolean;
};

function defaultState(): ToolbarState {
    return {
        isBold: false,
        isItalic: false,
        isUnderline: false,
        isStrikethrough: false,
        isCode: false,
        isLink: false,
        blockType: "paragraph",
        canUndo: false,
        canRedo: false,
    };
}

function readSelection(editor: LexicalEditor, state: ToolbarState): void {
    const selection = $getSelection();

    if (!$isRangeSelection(selection)) {
        Object.assign(state, defaultState());

        return;
    }

    state.isBold = selection.hasFormat("bold");
    state.isItalic = selection.hasFormat("italic");
    state.isUnderline = selection.hasFormat("underline");
    state.isStrikethrough = selection.hasFormat("strikethrough");
    state.isCode = selection.hasFormat("code");

    const anchorNode = selection.anchor.getNode();
    const element =
        anchorNode.getKey() === "root"
            ? anchorNode
            : ($findMatchingParent(anchorNode, (node) => {
                  const parent = node.getParent();

                  return parent !== null && parent.getKey() === "root";
              }) ?? anchorNode.getTopLevelElementOrThrow());

    const elementKey = element.getKey();
    const elementDOM = editor.getElementByKey(elementKey);

    const linkParent = $findMatchingParent(anchorNode, $isLinkNode);

    state.isLink = linkParent !== null;

    if (elementDOM === null) {
        state.blockType = "paragraph";

        return;
    }

    if ($isListNode(element)) {
        const parentList = $getNearestNodeOfType<ListNode>(anchorNode, ListNode);
        const type = parentList ? parentList.getListType() : element.getListType();

        if (type === "check") {
            state.blockType = "check";
        } else if (type === "number") {
            state.blockType = "number";
        } else {
            state.blockType = "bullet";
        }

        return;
    }

    if ($isHeadingNode(element)) {
        state.blockType = element.getTag() as BlockType;

        return;
    }

    if ($isQuoteNode(element)) {
        state.blockType = "quote";

        return;
    }

    if ($isCodeNode(element)) {
        state.blockType = "code";

        return;
    }

    state.blockType = "paragraph";
}

export function useToolbarState(): ToolbarState {
    const editor = useLexicalComposer();
    const state = reactive(defaultState());

    const sync = (): void => {
        editor.getEditorState().read(() => readSelection(editor, state));
    };

    sync();

    const unregister = mergeRegister(
        editor.registerUpdateListener(({ editorState }) => {
            editorState.read(() => readSelection(editor, state));
        }),
        editor.registerCommand(
            SELECTION_CHANGE_COMMAND,
            () => {
                editor.getEditorState().read(() => readSelection(editor, state));

                return false;
            },
            COMMAND_PRIORITY_LOW,
        ),
        editor.registerEditableListener(() => {
            sync();
        }),
    );

    onBeforeUnmount(() => unregister());

    return state;
}
