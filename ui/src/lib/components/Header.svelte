<script lang="ts">
  import { onMount } from 'svelte';
  import { currentScreen, theme, systemInfo } from '../stores';
  import { EngineBridge } from '../engine';

  onMount(async () => {
    if (!$systemInfo) {
      const info = await EngineBridge.getSystemInfo();
      $systemInfo = info;
    }
  });

  function toggleTheme() {
    $theme = $theme === 'dark' ? 'light' : 'dark';
  }
</script>

<header class="h-14 border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-[#0f131a]/80 backdrop-blur-md px-4 sm:px-6 flex items-center justify-between sticky top-0 z-30 select-none">
  <div class="flex items-center space-x-3">
    <button
      type="button"
      class="flex items-center space-x-2.5 focus:outline-none group text-left"
      on:click={() => ($currentScreen = 'home')}
      aria-label="Go to Home"
    >
      <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-indigo-500 to-blue-600 flex items-center justify-center text-white font-bold text-sm shadow-sm shadow-indigo-500/30 group-hover:scale-105 transition-transform">
        C
      </div>
      <div>
        <span class="font-bold tracking-tight text-gray-900 dark:text-gray-100 text-sm">Carryover</span>
        <span class="hidden sm:inline-block text-[10px] uppercase font-semibold tracking-wider text-indigo-500 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/70 px-1.5 py-0.5 rounded ml-1.5">Debian</span>
      </div>
    </button>

    {#if $systemInfo}
      <div class="hidden md:flex items-center space-x-1.5 pl-3 border-l border-gray-200 dark:border-gray-800 text-[11px] text-gray-500">
        <span class="inline-block w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
        <span class="font-medium text-gray-700 dark:text-gray-300">{$systemInfo.os_name || $systemInfo.os_id}</span>
        <span>•</span>
        <span>{$systemInfo.package_manager || 'apt'}</span>
      </div>
    {/if}
  </div>

  <div class="flex items-center space-x-1">
    <nav class="flex items-center space-x-1" aria-label="Main Navigation">
      <button
        type="button"
        class="px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {$currentScreen === 'home' ? 'bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800/60'}"
        on:click={() => ($currentScreen = 'home')}
      >
        Dashboard
      </button>

      <button
        type="button"
        class="px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {$currentScreen === 'backup' ? 'bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800/60'}"
        on:click={() => ($currentScreen = 'backup')}
      >
        Backup
      </button>

      <button
        type="button"
        class="px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {$currentScreen === 'restore' ? 'bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800/60'}"
        on:click={() => ($currentScreen = 'restore')}
      >
        Restore
      </button>

      <button
        type="button"
        class="px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {$currentScreen === 'tools' ? 'bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800/60'}"
        on:click={() => ($currentScreen = 'tools')}
      >
        Tools
      </button>

      <button
        type="button"
        class="px-2.5 py-1.5 text-xs font-medium rounded-lg transition-colors {$currentScreen === 'settings' ? 'bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800/60'}"
        on:click={() => ($currentScreen = 'settings')}
      >
        Settings
      </button>
    </nav>

    <div class="h-4 w-[1px] bg-gray-200 dark:bg-gray-800 mx-1"></div>

    <button
      type="button"
      on:click={toggleTheme}
      class="p-1.5 rounded-lg text-gray-500 hover:text-gray-800 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
      title="Toggle Dark/Light Mode"
      aria-label="Toggle Dark/Light Mode"
    >
      {#if $theme === 'dark'}
        <span class="text-xs">☀️</span>
      {:else}
        <span class="text-xs">🌙</span>
      {/if}
    </button>
  </div>
</header>
