/**
 * Svelte package entry — prefer importing the component:
 *   import MontiEmbed from '@monti/embed-svelte/MontiEmbed.svelte'
 * or:
 *   import MontiEmbed from '@monti/embed-svelte'  (via "svelte" export field)
 *
 * This module re-exports core types and a programmatic mount helper for
 * hosts that do not compile .svelte files from node_modules.
 */
export {
  MontiEmbed,
  createMontiEmbed,
  resolveEmbed,
  buildEmbedIframeUrl,
  type EmbedOptions,
  type EmbedError,
  type EmbedPosition,
  type EmbedResolveResult,
} from "@monti/embed-core";

export { mountMontiEmbedSvelte } from "./mount.js";
