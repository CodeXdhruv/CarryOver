<script lang="ts">
  export let steps: string[] = [];
  export let currentStep: number = 0;
</script>

<div class="flex items-center justify-between w-full max-w-lg mx-auto my-4">
  {#each steps as step, i}
    <div class="flex items-center {i < steps.length - 1 ? 'flex-1' : ''}">
      <div class="flex flex-col items-center">
        <div
          class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold transition-all {i === currentStep ? 'bg-accent text-white shadow-sm ring-4 ring-accent/20' : i < currentStep ? 'bg-emerald-500 text-white' : 'bg-gray-200 dark:bg-gray-800 text-gray-500'}"
        >
          {#if i < currentStep}
            ✓
          {:else}
            {i + 1}
          {/if}
        </div>
        <span class="text-[11px] font-medium mt-1 {i === currentStep ? 'text-accent dark:text-accent font-semibold' : 'text-gray-400 dark:text-gray-500'}">
          {step}
        </span>
      </div>

      {#if i < steps.length - 1}
        <div class="h-0.5 flex-1 mx-2 mb-4 transition-colors {i < currentStep ? 'bg-emerald-500' : 'bg-gray-200 dark:bg-gray-800'}"></div>
      {/if}
    </div>
  {/each}
</div>
