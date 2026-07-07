<script lang="ts">
  import { cn } from '$lib/utils';
  import type { Snippet } from 'svelte';

  type Variant = 'default' | 'destructive' | 'outline' | 'ghost';
  type Size = 'default' | 'sm' | 'lg';

  let {
    variant = 'default',
    size = 'default',
    class: className = '',
    disabled = false,
    type = 'button',
    onclick,
    children
  }: {
    variant?: Variant;
    size?: Size;
    class?: string;
    disabled?: boolean;
    type?: 'button' | 'submit';
    onclick?: (e: MouseEvent) => void;
    children: Snippet;
  } = $props();

  const variants: Record<Variant, string> = {
    default: 'bg-monti-blue/80 hover:bg-monti-blue text-white border border-cyan-400/30 shadow-neon',
    destructive: 'bg-red-600/80 hover:bg-red-600 text-white border border-red-400/40',
    outline: 'border border-monti-line/50 bg-monti-panel/60 hover:bg-monti-panel text-slate-100',
    ghost: 'hover:bg-white/5 text-slate-200'
  };

  const sizes: Record<Size, string> = {
    default: 'h-10 px-4 py-2',
    sm: 'h-8 rounded-md px-3 text-xs',
    lg: 'h-12 rounded-lg px-8 text-base'
  };
</script>

<button
  {type}
  {disabled}
  {onclick}
  class={cn(
    'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-400 disabled:pointer-events-none disabled:opacity-50',
    variants[variant],
    sizes[size],
    className
  )}
>
  {@render children()}
</button>