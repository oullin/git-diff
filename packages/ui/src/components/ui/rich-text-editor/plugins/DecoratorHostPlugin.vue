<script lang="ts">
import {
  Fragment,
  Teleport,
  defineComponent,
  h,
  onBeforeUnmount,
  shallowRef,
  type VNode,
} from "vue";
import { useLexicalComposer } from "lexical-vue";

export default defineComponent({
  name: "DecoratorHostPlugin",
  setup() {
    const editor = useLexicalComposer();
    const decorators = shallowRef(editor.getDecorators<VNode>());

    const unregister = editor.registerDecoratorListener<VNode>((next) => {
      decorators.value = next;
    });

    onBeforeUnmount(() => unregister());

    return () => {
      const map = decorators.value;
      const teleports: VNode[] = [];
      for (const nodeKey of Object.keys(map)) {
        const target = editor.getElementByKey(nodeKey);
        if (target !== null) {
          teleports.push(
            h(
              Teleport as unknown as object,
              { to: target, key: nodeKey },
              () => map[nodeKey],
            ) as VNode,
          );
        }
      }
      return h(Fragment, null, teleports);
    };
  },
});
</script>
