<script lang="ts">
  import { theme, currentScreen, systemInfo } from '../stores';

  let defaultDestination = '/tmp/carryover-backup';
  let allowCustomRebuild = false;
  let defaultExcludes = 'node_modules, .gradle, build, dist, target, venv, .venv, __pycache__, .pytest_cache, .next, .nuxt, .turbo, .cache';
  let mappingOverrides = `# Custom mapping overrides (YAML)\n# my-pkg: {pacman: my-arch-pkg, dnf: my-fedora-pkg}`;
  let saveMessage = '';

  function saveSettings() {
    saveMessage = 'Preferences saved successfully.';
    setTimeout(() => (saveMessage = ''), 2000);
  }
</script>

<div class="max-w-2xl mx-auto py-2 space-y-5">
  <div class="flex items-center justify-between pb-2 border-b border-gray-200 dark:border-gray-800">
    <div>
      <h1 class="text-xl font-bold text-gray-900 dark:text-gray-100">Preferences & Configuration</h1>
      <p class="text-xs text-gray-500">Configure Carryover local engine parameters, appearance, and exclusions.</p>
    </div>
    <button
      type="button"
      class="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 font-semibold"
      on:click={() => ($currentScreen = 'home')}
    >
      ✕ Close
    </button>
  </div>

  <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-5 space-y-4 shadow-sm text-xs">
    <!-- Appearance -->
    <div class="space-y-1.5">
      <span class="block font-semibold text-gray-800 dark:text-gray-200">Appearance Theme</span>
      <div class="flex space-x-2">
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg border {$theme === 'auto' ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300'}"
          on:click={() => ($theme = 'auto')}>System</button
        >
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg border {$theme === 'light' ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300'}"
          on:click={() => ($theme = 'light')}>Light</button
        >
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg border {$theme === 'dark' ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 font-semibold' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300'}"
          on:click={() => ($theme = 'dark')}>Dark</button
        >
      </div>
    </div>

    <!-- Default Destination -->
    <div class="space-y-1.5">
      <label for="defaultDestInput" class="block font-semibold text-gray-800 dark:text-gray-200">Default Backup Destination</label>
      <input
        id="defaultDestInput"
        type="text"
        bind:value={defaultDestination}
        class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
      />
    </div>

    <!-- Global Exclusion Rules -->
    <div class="space-y-1.5">
      <label for="globalExcludesInput" class="block font-semibold text-gray-800 dark:text-gray-200">Global Exclusion Rules (comma-separated)</label>
      <textarea
        id="globalExcludesInput"
        rows="2"
        bind:value={defaultExcludes}
        class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
      ></textarea>
      <p class="text-[10px] text-gray-400">Rebuildable folders are skipped during backup to save bandwidth and storage.</p>
    </div>

    <!-- Package Mapping Overrides -->
    <div class="space-y-1.5">
      <label for="mappingOverridesInput" class="block font-semibold text-gray-800 dark:text-gray-200">Package Mapping Overrides (~/.config/carryover/mappings.yaml)</label>
      <textarea
        id="mappingOverridesInput"
        rows="3"
        bind:value={mappingOverrides}
        class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
      ></textarea>
    </div>

    <!-- Security & Rebuild Options -->
    <div class="space-y-2 pt-2 border-t border-gray-100 dark:border-gray-800">
      <label class="flex items-center space-x-2 cursor-pointer">
        <input type="checkbox" bind:checked={allowCustomRebuild} class="rounded text-indigo-600 focus:ring-indigo-500" />
        <span class="text-gray-700 dark:text-gray-300 font-medium">Allow custom project rebuild commands (disabled by default)</span>
      </label>
      <p class="text-[10px] text-gray-400 ml-6">
        When disabled, only strict security-allowlisted commands (<code class="font-mono">npm ci</code>, <code class="font-mono">cargo fetch</code>, etc.) are executed.
      </p>
    </div>

    <!-- Local Machine Status -->
    {#if $systemInfo}
      <div class="pt-3 border-t border-gray-100 dark:border-gray-800 space-y-1 text-gray-500 font-mono text-[11px]">
        <div>Host: {$systemInfo.hostname} ({$systemInfo.os_name})</div>
        <div>User: {$systemInfo.username} (Home: {$systemInfo.home})</div>
        <div>Engine Package Manager: {$systemInfo.package_manager}</div>
      </div>
    {/if}
  </div>

  <div class="flex items-center justify-between pt-1">
    <span class="text-xs text-emerald-600 font-medium">{saveMessage}</span>
    <button
      type="button"
      class="bg-indigo-600 hover:bg-indigo-700 text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all"
      on:click={saveSettings}
    >
      Save Preferences
    </button>
  </div>
</div>
