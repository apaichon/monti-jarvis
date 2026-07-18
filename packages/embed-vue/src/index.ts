import {
  defineComponent,
  h,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  type App,
  type PropType,
} from "vue";
import {
  MontiEmbed,
  type EmbedError,
  type EmbedPosition,
  type EmbedResolveResult,
} from "@monti/embed-core";

export type { EmbedError, EmbedPosition, EmbedResolveResult };
export { MontiEmbed };

export const MontiEmbedVue = defineComponent({
  name: "MontiEmbed",
  props: {
    embedKey: { type: String, required: true },
    apiBase: { type: String, required: true },
    parentOrigin: { type: String, default: undefined },
    position: { type: String as PropType<EmbedPosition>, default: "bottom-right" },
    agentId: { type: String, default: undefined },
    theme: { type: String, default: undefined },
    locale: { type: String, default: undefined },
    open: { type: Boolean, default: false },
    inline: { type: Boolean, default: false },
    skipResolve: { type: Boolean, default: false },
  },
  emits: {
    open: () => true,
    close: () => true,
    ready: (_result?: EmbedResolveResult) => true,
    error: (_error: EmbedError) => true,
    destroy: () => true,
    "update:open": (_value: boolean) => true,
  },
  setup(props, { emit, expose }) {
    const host = ref<HTMLElement | null>(null);
    let embed: MontiEmbed | null = null;

    async function mountEmbed() {
      embed?.destroy();
      embed = null;
      if (!props.embedKey || !props.apiBase) return;

      embed = new MontiEmbed({
        embedKey: props.embedKey,
        apiBase: props.apiBase,
        parentOrigin: props.parentOrigin,
        position: props.position,
        agentId: props.agentId,
        theme: props.theme,
        locale: props.locale,
        open: props.open,
        skipResolve: props.skipResolve,
        container: props.inline ? host.value : null,
        onOpen: () => {
          emit("open");
          emit("update:open", true);
        },
        onClose: () => {
          emit("close");
          emit("update:open", false);
        },
        onReady: (r) => emit("ready", r),
        onError: (e) => emit("error", e),
        onDestroy: () => emit("destroy"),
      });
      await embed.mount();
    }

    onMounted(() => {
      void mountEmbed();
    });

    onBeforeUnmount(() => {
      embed?.destroy();
      embed = null;
    });

    watch(
      () =>
        [
          props.embedKey,
          props.apiBase,
          props.parentOrigin,
          props.position,
          props.agentId,
          props.theme,
          props.locale,
          props.inline,
          props.skipResolve,
        ] as const,
      () => {
        void mountEmbed();
      },
    );

    watch(
      () => props.open,
      (next) => {
        if (!embed) return;
        if (next) embed.open();
        else embed.close();
      },
    );

    expose({
      open: () => embed?.open(),
      close: () => embed?.close(),
      toggle: () => embed?.toggle(),
      destroy: () => {
        embed?.destroy();
        embed = null;
      },
      getInstance: () => embed,
    });

    return () =>
      h("div", {
        ref: host,
        "data-monti-embed-vue": "1",
        style: props.inline
          ? { width: "100%", minHeight: "480px", height: "100%" }
          : { display: "contents" },
      });
  },
});

export interface MontiEmbedPluginOptions {
  /** Default apiBase applied when component prop is omitted (still required on component). */
  apiBase?: string;
}

export function createMontiEmbedPlugin(defaults: MontiEmbedPluginOptions = {}) {
  return {
    install(app: App) {
      app.component("MontiEmbed", MontiEmbedVue);
      if (defaults.apiBase) {
        app.provide("montiEmbedApiBase", defaults.apiBase);
      }
    },
  };
}

export default MontiEmbedVue;
