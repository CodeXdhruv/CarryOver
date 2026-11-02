<script lang="ts">
  export let percent: number = 0;
  export let label: string = '';
  export let detail: string = '';
</script>

<div class="w-full space-y-1.5">
  <div class="flex justify-between items-center text-xs">
    <span class="font-medium text-gray-700 dark:text-gray-300 truncate max-w-xs">{label}</span>
    <span class="font-semibold text-accent">{percent}%</span>
  </div>
  <div class="w-full h-2.5 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
    <div
      class="h-full bg-accent transition-all duration-300 rounded-full ease-out"
      style="width: {Math.min(100, Math.max(0, percent))}%"
    ></div>
  </div>
  {#if detail}
    <p class="text-[11px] text-gray-400 dark:text-gray-500 font-mono truncate">{detail}</p>
  {/if}
</div>
