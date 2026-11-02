<script lang="ts">
  import { onMount } from 'svelte';
  import { currentScreen, systemInfo } from '../stores';
  import { EngineBridge, type DoctorReport } from '../engine';

  let doctorReport: DoctorReport | null = null;
  let loadingDoctor = true;

  onMount(async () => {
    if (!$systemInfo) {
      $systemInfo = await EngineBridge.getSystemInfo();
    }
    doctorReport = await EngineBridge.getDoctorReport();
    loadingDoctor = false;
  });
</script>

<div class="space-y-6 max-w-4xl mx-auto py-2">
  <!-- System Status Strip -->
  <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-4 shadow-sm flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
    <div class="flex items-center space-x-3.5">
      <div class="w-10 h-10 rounded-xl bg-indigo-50 dark:bg-indigo-950/60 border border-indigo-100 dark:border-indigo-900/40 flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-mono text-lg font-bold">
        🐧
      </div>
      <div>
        <div class="flex items-center space-x-2">
          <span class="font-semibold text-gray-900 dark:text-gray-100 text-sm">
            {$systemInfo?.os_name || 'Ubuntu Linux'}
          </span>
          <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 border border-gray-200 dark:border-gray-700">
            {$systemInfo?.arch || 'x86_64'}
          </span>
        </div>
        <p class="text-xs text-gray-500 mt-0.5 font-mono">
          Host: {$systemInfo?.hostname || 'localhost'} · User: {$systemInfo?.username || 'user'} · PM: {$systemInfo?.package_manager || 'apt'}
        </p>
      </div>
    </div>

    <div class="flex items-center space-x-2 text-xs">
      {#if loadingDoctor}
        <span class="text-gray-400">Auditing system...</span>
      {:else if doctorReport?.overall_status === 'pass'}
        <span class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-medium bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border border-emerald-200/60 dark:border-emerald-800/60">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5 animate-pulse"></span>
          Preflight Checks Passed
        </span>
      {:else}
        <span class="inline-flex items-center px-2.5 py-1 rounded-md text-[11px] font-medium bg-amber-50 dark:bg-amber-950/50 text-amber-700 dark:text-amber-300 border border-amber-200/60 dark:border-amber-800/60">
          <span class="w-1.5 h-1.5 rounded-full bg-amber-500 mr-1.5"></span>
          Preflight Notice
        </span>
      {/if}

      <button
        type="button"
        on:click={() => ($currentScreen = 'tools')}
        class="text-indigo-600 dark:text-indigo-400 hover:underline font-medium text-xs px-2 py-1"
      >
        Doctor Details →
      </button>
    </div>
  </div>

  <!-- Hero Header -->
  <div class="text-center space-y-2 py-2">
    <h1 class="text-2xl sm:text-3xl font-extrabold tracking-tight text-gray-900 dark:text-gray-50">
      Minimalist Linux OS Migration & Backup
    </h1>
    <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 max-w-xl mx-auto leading-relaxed">
      Carry over your files, shell environment, dotfiles, installed packages, and developer workspace recipes across distributions seamlessly.
    </p>
  </div>

  <!-- 4 Main Action Cards -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <!-- Card 1: Create Backup -->
    <button
      type="button"
      class="group text-left relative bg-white dark:bg-[#131926] border border-gray-200/90 dark:border-gray-800/90 hover:border-indigo-500 dark:hover:border-indigo-500 p-5 rounded-xl shadow-sm hover:shadow-md transition-all flex flex-col justify-between h-48 focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
      on:click={() => ($currentScreen = 'backup')}
    >
      <div class="space-y-2.5">
        <div class="w-9 h-9 rounded-lg bg-indigo-50 dark:bg-indigo-950/70 border border-indigo-100 dark:border-indigo-900/50 flex items-center justify-center text-indigo-600 dark:text-indigo-400 text-lg">
          📦
        </div>
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-gray-100 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
            Create System Backup
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 leading-relaxed">
            Scans personal directories, installed native/snap/flatpak apps, dotfiles, and git dev repositories. Excludes regenerable build artifacts.
          </p>
        </div>
      </div>

      <div class="flex items-center text-xs font-semibold text-indigo-600 dark:text-indigo-400 group-hover:translate-x-1 transition-transform">
        Launch backup wizard →
      </div>
    </button>

    <!-- Card 2: Restore Backup -->
    <button
      type="button"
      class="group text-left relative bg-white dark:bg-[#131926] border border-gray-200/90 dark:border-gray-800/90 hover:border-emerald-500 dark:hover:border-emerald-500 p-5 rounded-xl shadow-sm hover:shadow-md transition-all flex flex-col justify-between h-48 focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
      on:click={() => ($currentScreen = 'restore')}
    >
      <div class="space-y-2.5">
        <div class="w-9 h-9 rounded-lg bg-emerald-50 dark:bg-emerald-950/70 border border-emerald-100 dark:border-emerald-900/50 flex items-center justify-center text-emerald-600 dark:text-emerald-400 text-lg">
          ⚡
        </div>
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-gray-100 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
            Restore onto this OS
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 leading-relaxed">
            Unpacks snapshots, translates package names across distros, merges dotfiles non-destructively, and triggers dev project rebuilds.
          </p>
        </div>
      </div>

      <div class="flex items-center text-xs font-semibold text-emerald-600 dark:text-emerald-400 group-hover:translate-x-1 transition-transform">
        Launch restore wizard →
      </div>
    </button>

    <!-- Card 3: Cross-Distro Package Mapper -->
    <button
      type="button"
      class="group text-left relative bg-white dark:bg-[#131926] border border-gray-200/90 dark:border-gray-800/90 hover:border-purple-500 dark:hover:border-purple-500 p-5 rounded-xl shadow-sm hover:shadow-md transition-all flex flex-col justify-between h-48 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
      on:click={() => ($currentScreen = 'tools')}
    >
      <div class="space-y-2.5">
        <div class="w-9 h-9 rounded-lg bg-purple-50 dark:bg-purple-950/70 border border-purple-100 dark:border-purple-900/50 flex items-center justify-center text-purple-600 dark:text-purple-400 text-lg">
          🔄
        </div>
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-gray-100 group-hover:text-purple-600 dark:group-hover:text-purple-400 transition-colors">
            Package Translation Matrix
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 leading-relaxed">
            Translate package names between Debian/Ubuntu (apt), Fedora (dnf), Arch (pacman), and openSUSE (zypper) with confidence ranking.
          </p>
        </div>
      </div>

      <div class="flex items-center text-xs font-semibold text-purple-600 dark:text-purple-400 group-hover:translate-x-1 transition-transform">
        Open package translator →
      </div>
    </button>

    <!-- Card 4: System Doctor & Tools -->
    <button
      type="button"
      class="group text-left relative bg-white dark:bg-[#131926] border border-gray-200/90 dark:border-gray-800/90 hover:border-blue-500 dark:hover:border-blue-500 p-5 rounded-xl shadow-sm hover:shadow-md transition-all flex flex-col justify-between h-48 focus:outline-none focus:ring-2 focus:ring-blue-500/50"
      on:click={() => ($currentScreen = 'tools')}
    >
      <div class="space-y-2.5">
        <div class="w-9 h-9 rounded-lg bg-blue-50 dark:bg-blue-950/70 border border-blue-100 dark:border-blue-900/50 flex items-center justify-center text-blue-600 dark:text-blue-400 text-lg">
          🩺
        </div>
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-gray-100 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
            Diagnostics, Snapshots & Export
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 leading-relaxed">
            Verify backup repository integrity, inspect previous snapshots, and export headless Ansible migration playbooks.
          </p>
        </div>
      </div>

      <div class="flex items-center text-xs font-semibold text-blue-600 dark:text-blue-400 group-hover:translate-x-1 transition-transform">
        Explore utilities →
      </div>
    </button>
  </div>

  <!-- Features Footer Strip -->
  <div class="border-t border-gray-200/80 dark:border-gray-800/80 pt-5 grid grid-cols-2 sm:grid-cols-4 gap-3 text-center">
    <div class="p-2">
      <span class="block text-xs font-semibold text-gray-800 dark:text-gray-200">Non-Destructive</span>
      <span class="text-[11px] text-gray-400">Zero file overwrite without prompt</span>
    </div>
    <div class="p-2">
      <span class="block text-xs font-semibold text-gray-800 dark:text-gray-200">Encrypted</span>
      <span class="text-[11px] text-gray-400">AES-256 restic repository</span>
    </div>
    <div class="p-2">
      <span class="block text-xs font-semibold text-gray-800 dark:text-gray-200">Cross-Distro</span>
      <span class="text-[11px] text-gray-400">apt, dnf, pacman, zypper</span>
    </div>
    <div class="p-2">
      <span class="block text-xs font-semibold text-gray-800 dark:text-gray-200">Lightweight</span>
      <span class="text-[11px] text-gray-400">Minimal memory footprint</span>
    </div>
  </div>
</div>
