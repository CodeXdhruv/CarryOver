<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  export let title: string = '';
  export let open: boolean = false;

  const dispatch = createEventDispatcher();
  function close() {
    dispatch('close');
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in">
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xl max-w-md w-full p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100">{title}</h3>
        <button on:click={close} class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-sm">✕</button>
      </div>

      <div class="text-sm text-gray-600 dark:text-gray-300">
        <slot />
      </div>

      <div class="flex justify-end space-x-2 pt-2">
        <slot name="actions" />
      </div>
    </div>
  </div>
{/if}
