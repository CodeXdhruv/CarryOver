<script lang="ts">
  import { onMount } from 'svelte';
  import { currentScreen, theme } from './lib/stores';
  import Header from './lib/components/Header.svelte';
  import HomeScreen from './lib/screens/HomeScreen.svelte';
  import BackupScreen from './lib/screens/BackupScreen.svelte';
  import RestoreScreen from './lib/screens/RestoreScreen.svelte';
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

<div class="min-h-screen bg-gray-50 dark:bg-[#0E1117] text-gray-900 dark:text-gray-100 flex flex-col">
  <Header />

  <main class="flex-1 px-4 sm:px-6 py-4 flex flex-col justify-start">
    {#if $currentScreen === 'home'}
      <HomeScreen />
    {:else if $currentScreen === 'backup'}
      <BackupScreen />
    {:else if $currentScreen === 'restore'}
      <RestoreScreen />
    {:else if $currentScreen === 'settings'}
      <SettingsScreen />
    {/if}
  </main>
</div>
