<script setup lang="ts">
import { onBeforeUnmount } from "vue";
import { useLexicalComposer } from "lexical-vue";
import { $wrapNodeInElement, mergeRegister } from "@lexical/utils";
import { $createImageNode, type ImageNodeProps } from "@rich-text-editor/nodes/ImageNode";
import { INSERT_IMAGE_COMMAND } from "@rich-text-editor/plugins/imageCommand";

import {
  $createParagraphNode,
  $insertNodes,
  $isRootOrShadowRoot,
  COMMAND_PRIORITY_EDITOR,
} from "lexical";

const editor = useLexicalComposer();

const unregister = mergeRegister(
  editor.registerCommand<ImageNodeProps>(
    INSERT_IMAGE_COMMAND,
    (payload) => {
      const node = $createImageNode(payload);

      $insertNodes([node]);
      if ($isRootOrShadowRoot(node.getParentOrThrow())) {
        $wrapNodeInElement(node, $createParagraphNode).selectEnd();
      }

      return true;
    },
    COMMAND_PRIORITY_EDITOR,
  ),
);

onBeforeUnmount(() => unregister());
</script>

<template>
  <span class="hidden" />
</template>
