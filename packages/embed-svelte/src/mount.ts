import {
  MontiEmbed,
  type EmbedError,
  type EmbedOptions,
  type EmbedResolveResult,
} from "@monti/embed-core";

export interface MountMontiEmbedSvelteOptions extends EmbedOptions {
  /** Target element; floating mode still appends launcher to document.body. */
  target?: HTMLElement;
  inline?: boolean;
}

export interface MontiEmbedSvelteHandle {
  open: () => void;
  close: () => void;
  toggle: () => void;
  destroy: () => void;
  getInstance: () => MontiEmbed;
}

/**
 * Imperative mount for SvelteKit/Vite apps that want a tiny API without importing the .svelte file.
 * When `inline: true`, pass `target` as the container.
 */
export async function mountMontiEmbedSvelte(
  options: MountMontiEmbedSvelteOptions,
): Promise<MontiEmbedSvelteHandle> {
  const { target, inline, ...rest } = options;
  const embed = new MontiEmbed({
    ...rest,
    container: inline ? target ?? rest.container ?? null : rest.container ?? null,
  });
  await embed.mount();
  return {
    open: () => embed.open(),
    close: () => embed.close(),
    toggle: () => embed.toggle(),
    destroy: () => embed.destroy(),
    getInstance: () => embed,
  };
}

export type { EmbedError, EmbedResolveResult };
