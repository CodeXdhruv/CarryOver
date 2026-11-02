<script lang="ts">
  export let status: 'matched' | 'suggested' | 'unmapped' | 'sensitive' | 'neutral' = 'neutral';
</script>

<span
  class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium transition-colors
  {status === 'matched' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800' : ''}
  {status === 'suggested' ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 border border-amber-200 dark:border-amber-800' : ''}
  {status === 'unmapped' ? 'bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-400 border border-rose-200 dark:border-rose-800' : ''}
  {status === 'sensitive' ? 'bg-purple-50 text-purple-700 dark:bg-purple-950/40 dark:text-purple-400 border border-purple-200 dark:border-purple-800' : ''}
  {status === 'neutral' ? 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300 border border-gray-200 dark:border-gray-700' : ''}"
>
  {#if status === 'matched'}✓{:else if status === 'suggested'}⚠{:else if status === 'unmapped'}✕{:else if status === 'sensitive'}🔒{/if}
  <span class="ml-1"><slot /></span>
</span>
