<script lang="ts">
  import { currentScreen } from '../stores';
  let backupDetected = true;
</script>

<div class="space-y-6 max-w-2xl mx-auto py-4">
  {#if backupDetected}
    <div class="bg-accent/10 border border-accent/30 rounded-xl p-4 flex items-center justify-between text-sm">
      <div class="flex items-center space-x-3">
        <span class="text-xl">💾</span>
        <div>
          <p class="font-medium text-gray-900 dark:text-gray-100">Existing backup found on USB Drive</p>
          <p class="text-xs text-gray-500">Ubuntu 24.04 · 4.2 GB · Created 2 hours ago</p>
        </div>
      </div>
      <button
        class="bg-accent hover:bg-accent-hover text-white px-3 py-1.5 rounded-lg text-xs font-semibold shadow-sm transition-all"
        on:click={() => ($currentScreen = 'restore')}
      >
        Restore Now
      </button>
    </div>
  {/if}

  <div class="text-center space-y-2 pt-4">
    <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-gray-100">Switch Linux OS without losing anything</h1>
    <p class="text-sm text-gray-500 max-w-md mx-auto">
      Back up files, dotfiles, app lists, and developer project recipes. Restore seamlessly onto any distribution.
    </p>
  </div>

  <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2">
    <!-- Card 1: Create Backup -->
    <div
      class="group relative bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 hover:border-accent dark:hover:border-accent p-6 rounded-xl shadow-sm hover:shadow-md transition-all cursor-pointer flex flex-col justify-between h-56"
      on:click={() => ($currentScreen = 'backup')}
    >
      <div class="space-y-3">
        <div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-950/60 flex items-center justify-center text-accent text-xl">
          ⬆
        </div>
        <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Create Backup</h2>
        <p class="text-xs text-gray-500 leading-relaxed">
          Scans your applications, dotfiles, and dev projects. Automatically skips regenerable folders like node_modules.
        </p>
      </div>

      <div class="flex items-center text-xs font-medium text-accent group-hover:translate-x-0.5 transition-transform">
        Start backup wizard →
      </div>
    </div>

    <!-- Card 2: Restore Backup -->
    <div
      class="group relative bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 hover:border-emerald-500 dark:hover:border-emerald-500 p-6 rounded-xl shadow-sm hover:shadow-md transition-all cursor-pointer flex flex-col justify-between h-56"
      on:click={() => ($currentScreen = 'restore')}
    >
      <div class="space-y-3">
        <div class="w-10 h-10 rounded-xl bg-emerald-50 dark:bg-emerald-950/60 flex items-center justify-center text-emerald-600 text-xl">
          ⬇
        </div>
        <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Restore on this OS</h2>
        <p class="text-xs text-gray-500 leading-relaxed">
          Maps apps to the new distro's package manager, safely merges shell configuration, and rebuilds projects.
        </p>
      </div>

      <div class="flex items-center text-xs font-medium text-emerald-600 dark:text-emerald-400 group-hover:translate-x-0.5 transition-transform">
        Start restore wizard →
      </div>
    </div>
  </div>

  <div class="border-t border-gray-100 dark:border-gray-800 pt-6 grid grid-cols-3 gap-2 text-center text-xs text-gray-400">
    <div>
      <span class="block font-semibold text-gray-700 dark:text-gray-300">Cross-Distro</span>
      apt, dnf, pacman, zypper
    </div>
    <div>
      <span class="block font-semibold text-gray-700 dark:text-gray-300">Zero Data Loss</span>
      Non-destructive merge & safety copies
    </div>
    <div>
      <span class="block font-semibold text-gray-700 dark:text-gray-300">Encrypted</span>
      restic core with AES-256
    </div>
  </div>
</div>
