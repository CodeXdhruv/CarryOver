<script lang="ts">
  import { onMount } from 'svelte';
  import { currentScreen, theme } from './lib/stores';
  import Header from './lib/components/Header.svelte';
  import HomeScreen from './lib/screens/HomeScreen.svelte';
  import BackupScreen from './lib/screens/BackupScreen.svelte';
  import RestoreScreen from './lib/screens/RestoreScreen.svelte';
  import ToolsScreen from './lib/screens/ToolsScreen.svelte';
  import SettingsScreen from './lib/screens/SettingsScreen.svelte';

  onMount(() => {
    theme.subscribe((val) => {
      if (val === 'dark' || (val === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
    });
  });
</script>

<div class="min-h-screen bg-[#F8FAFC] dark:bg-[#0B0F17] text-gray-900 dark:text-gray-100 flex flex-col font-sans transition-colors duration-200">
  <Header />

  <main class="flex-1 px-4 sm:px-6 py-5 flex flex-col justify-start max-w-5xl w-full mx-auto">
    {#if $currentScreen === 'home'}
      <HomeScreen />
    {:else if $currentScreen === 'backup'}
      <BackupScreen />
    {:else if $currentScreen === 'restore'}
      <RestoreScreen />
    {:else if $currentScreen === 'tools'}
      <ToolsScreen />
    {:else if $currentScreen === 'settings'}
      <SettingsScreen />
    {/if}
  </main>

  <footer class="py-3 px-6 text-center text-[11px] text-gray-400 dark:text-gray-600 border-t border-gray-200/60 dark:border-gray-800/60">
    Carryover Migration Engine v1.0.0 · Debian Package Built for Local Linux Architecture
  </footer>
</div>
