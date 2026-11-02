<script lang="ts">
  import { currentScreen } from '../stores';
</script>

<header class="h-14 border-b border-gray-200 dark:border-gray-800 bg-white/70 dark:bg-gray-900/70 backdrop-blur-md px-6 flex items-center justify-between sticky top-0 z-20">
  <div class="flex items-center space-x-3 cursor-pointer" on:click={() => ($currentScreen = 'home')}>
    <div class="w-7 h-7 rounded-lg bg-accent flex items-center justify-center text-white font-bold text-sm shadow-sm shadow-accent/40">
      C
    </div>
    <span class="font-semibold tracking-tight text-gray-900 dark:text-gray-100 text-base">Carryover</span>
  </div>

  <nav class="flex items-center space-x-1">
    <button
      class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {$currentScreen === 'home' ? 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100' : 'text-gray-500 hover:text-gray-800 dark:hover:text-gray-300'}"
      on:click={() => ($currentScreen = 'home')}
    >
      Home
    </button>
    <button
      class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {$currentScreen === 'settings' ? 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100' : 'text-gray-500 hover:text-gray-800 dark:hover:text-gray-300'}"
      on:click={() => ($currentScreen = 'settings')}
    >
      Settings
    </button>
  </nav>
</header>
