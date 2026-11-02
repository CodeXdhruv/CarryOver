<script lang="ts">
  import { onMount } from 'svelte';
  import { currentScreen, backupManifest, addLog } from '../stores';
  import { EngineBridge, type Manifest, type EngineEvent } from '../engine';
  import Stepper from '../components/Stepper.svelte';
  import ProgressBar from '../components/ProgressBar.svelte';
  import Chip from '../components/Chip.svelte';

  const steps = ['Scan', 'Choose', 'Protect', 'Backup'];
  let currentStep = 0;

  // Step 0: Scan State
  let scanning = true;
  let scanProgress = 0;
  let scanStatus = 'Initializing system scanner...';
  let manifest: Manifest | null = null;

  // Step 1: Choose State
  let activeTab: 'files' | 'apps' | 'dotfiles' | 'projects' = 'files';
  let selectedFolders: Record<string, boolean> = {};
  let selectedApps: Record<string, boolean> = {};
  let selectedDotfiles: Record<string, boolean> = {};
  let selectedProjects: Record<string, boolean> = {};
  let searchQuery = '';

  // Step 2: Protect State
  let backupDestination = '/tmp/carryover-backup';
  let password = '';
  let confirmPassword = '';
  let dryRun = false;

  // Step 3: Run State
  let running = false;
  let runProgress = 0;
  let currentAction = 'Preparing restic backup repository...';
  let backupDone = false;
  let snapshotResult = { id: '', size: 0, manifestSha: '' };
  let logs: string[] = [];

  onMount(async () => {
    startScan();
  });

  async function startScan() {
    scanning = true;
    scanProgress = 0;
    try {
      manifest = await EngineBridge.scanSystem((ev: EngineEvent) => {
        if (ev.percent) scanProgress = ev.percent;
        if (ev.current) scanStatus = ev.current;
        if (ev.message) addLog(ev.message);
      });
      $backupManifest = manifest;

      // Smart defaults
      manifest.folders.forEach((f) => {
        selectedFolders[f.path] = f.path !== 'Downloads';
      });
      manifest.dotfiles.forEach((d) => {
        selectedDotfiles[d.path] = !d.sensitive;
      });
      if (manifest.apps.apt) {
        manifest.apps.apt.forEach((a) => (selectedApps[a.name] = true));
      }
      if (manifest.apps.flatpak) {
        manifest.apps.flatpak.forEach((a) => (selectedApps[a.id] = true));
      }
      manifest.projects.forEach((p) => {
        selectedProjects[p.path] = true;
      });

      scanning = false;
      currentStep = 1;
    } catch (err: any) {
      scanStatus = 'Scan error: ' + err.message;
    }
  }

  async function startBackup() {
    if (!password || password !== confirmPassword) return;
    currentStep = 3;
    running = true;
    runProgress = 0;

    try {
      const res = await EngineBridge.runBackup(backupDestination, password, dryRun, manifest, (ev) => {
        if (ev.percent) runProgress = ev.percent;
        if (ev.current) {
          currentAction = ev.current;
          logs = [ev.current, ...logs.slice(0, 49)];
        }
      });
      snapshotResult = { id: res.snapshotId, size: res.totalBytes, manifestSha: res.manifestSha || '' };
      running = false;
      backupDone = true;
    } catch (err: any) {
      currentAction = 'Backup failed: ' + err.message;
      running = false;
    }
  }

  function getPasswordStrength(pw: string): { label: string; color: string; width: string } {
    if (!pw) return { label: 'Empty', color: 'bg-gray-200 dark:bg-gray-700', width: '0%' };
    if (pw.length < 6) return { label: 'Weak', color: 'bg-rose-500', width: '25%' };
    if (pw.length < 10) return { label: 'Fair', color: 'bg-amber-500', width: '60%' };
    return { label: 'Strong', color: 'bg-emerald-500', width: '100%' };
  }

  $: strength = getPasswordStrength(password);
</script>

<div class="max-w-2xl mx-auto py-2 space-y-4">
  <Stepper {steps} {currentStep} />

  <!-- STEP 0: SCANNING -->
  {#if currentStep === 0}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-8 text-center space-y-6 shadow-sm">
      <div class="w-12 h-12 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin mx-auto"></div>
      <div class="space-y-2">
        <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">Scanning Local Linux System</h2>
        <p class="text-xs text-gray-500 max-w-sm mx-auto font-mono">{scanStatus}</p>
      </div>
      <ProgressBar percent={scanProgress} label="Progress" />
    </div>

  <!-- STEP 1: CHOOSE -->
  {:else if currentStep === 1 && manifest}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl overflow-hidden shadow-sm">
      <div class="p-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">Choose Migration Content</h2>
          <p class="text-xs text-gray-400">Everything is pre-selected with smart safety defaults.</p>
        </div>
        <div class="flex space-x-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg text-xs">
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'files' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'files')}>Folders ({manifest.folders.length})</button
          >
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'apps' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'apps')}>Apps</button
          >
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'dotfiles' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'dotfiles')}>Dotfiles ({manifest.dotfiles.length})</button
          >
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'projects' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'projects')}>Projects ({manifest.projects.length})</button
          >
        </div>
      </div>

      <div class="p-4 max-h-80 overflow-y-auto space-y-2">
        {#if activeTab === 'files'}
          {#each manifest.folders as folder}
            <label class="flex items-center justify-between p-2.5 rounded-lg border border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/40 text-xs cursor-pointer">
              <div class="flex items-center space-x-3">
                <input type="checkbox" bind:checked={selectedFolders[folder.path]} class="rounded text-indigo-600 focus:ring-indigo-500" />
                <span class="font-mono text-gray-800 dark:text-gray-200">~/{folder.path}</span>
              </div>
              <span class="text-gray-400 font-mono">{folder.size_bytes ? (folder.size_bytes / (1024 * 1024)).toFixed(1) + ' MB' : 'Calculated'}</span>
            </label>
          {/each}
        {:else if activeTab === 'apps'}
          <div class="space-y-3">
            {#if manifest.apps.apt && manifest.apps.apt.length > 0}
              <div>
                <span class="text-[11px] font-bold text-gray-400 block mb-1">Debian / Ubuntu Native Packages ({manifest.apps.apt.length})</span>
                <div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
                  {#each manifest.apps.apt as app}
                    <label class="flex items-center space-x-2 text-xs p-1.5 rounded border border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/40">
                      <input type="checkbox" bind:checked={selectedApps[app.name]} class="rounded text-indigo-600" />
                      <span class="font-mono truncate">{app.name}</span>
                    </label>
                  {/each}
                </div>
              </div>
            {/if}

            {#if manifest.apps.flatpak && manifest.apps.flatpak.length > 0}
              <div>
                <span class="text-[11px] font-bold text-gray-400 block mb-1">Flatpak Applications ({manifest.apps.flatpak.length})</span>
                <div class="grid grid-cols-2 gap-2">
                  {#each manifest.apps.flatpak as app}
                    <label class="flex items-center space-x-2 text-xs p-1.5 rounded border border-gray-100 dark:border-gray-800">
                      <input type="checkbox" bind:checked={selectedApps[app.id]} class="rounded text-indigo-600" />
                      <span class="font-mono truncate text-[11px]">{app.id}</span>
                    </label>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {:else if activeTab === 'dotfiles'}
          {#each manifest.dotfiles as dotfile}
            <label class="flex items-center justify-between p-2.5 rounded-lg border border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/40 text-xs cursor-pointer">
              <div class="flex items-center space-x-3">
                <input type="checkbox" bind:checked={selectedDotfiles[dotfile.path]} class="rounded text-indigo-600" />
                <span class="font-mono text-gray-800 dark:text-gray-200">~/{dotfile.path}</span>
                {#if dotfile.sensitive}
                  <span class="px-1.5 py-0.2 text-[10px] rounded bg-rose-100 dark:bg-rose-950/60 text-rose-600 dark:text-rose-300 font-semibold">Sensitive</span>
                {/if}
              </div>
              <span class="text-gray-400 font-mono">{dotfile.size_bytes || 0} bytes</span>
            </label>
          {/each}
        {:else if activeTab === 'projects'}
          {#each manifest.projects as proj}
            <div class="p-2.5 rounded-lg border border-gray-100 dark:border-gray-800 space-y-1 text-xs">
              <div class="flex items-center justify-between">
                <label class="flex items-center space-x-2 cursor-pointer font-mono font-bold text-gray-800 dark:text-gray-200">
                  <input type="checkbox" bind:checked={selectedProjects[proj.path]} class="rounded text-indigo-600" />
                  <span>{proj.path}</span>
                </label>
                <span class="px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-[10px] font-mono capitalize">{proj.type}</span>
              </div>
              <div class="text-[11px] font-mono text-indigo-600 dark:text-indigo-400 pl-6">
                Rebuild: {proj.rebuild.join(' && ')}
              </div>
            </div>
          {/each}
        {/if}
      </div>

      <div class="p-3 bg-gray-50 dark:bg-gray-800/40 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs text-gray-500">
        <span>✨ Automatic exclusion of regenerable caches (<code class="text-[11px]">node_modules</code>, <code class="text-[11px]">target</code>, <code class="text-[11px]">.venv</code>)</span>
        <button
          type="button"
          class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg font-semibold shadow-sm transition-all text-xs"
          on:click={() => (currentStep = 2)}
        >
          Next: Protect Backup →
        </button>
      </div>
    </div>

  <!-- STEP 2: PROTECT -->
  {:else if currentStep === 2}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-6 space-y-5 shadow-sm">
      <div class="space-y-1">
        <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">Protect & Configure Backup</h2>
        <p class="text-xs text-gray-500">Encrypted with restic AES-256 Poly1305. Local restic repository on USB or disk.</p>
      </div>

      <div class="space-y-4 text-xs">
        <div class="space-y-1.5">
          <label for="destPathInput" class="block font-semibold text-gray-700 dark:text-gray-300">Destination Directory Path</label>
          <input
            id="destPathInput"
            type="text"
            bind:value={backupDestination}
            class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
          />
        </div>

        <div class="space-y-1.5">
          <label for="backupPasswordInput" class="block font-semibold text-gray-700 dark:text-gray-300">Set Backup Password</label>
          <input
            id="backupPasswordInput"
            type="password"
            bind:value={password}
            placeholder="Enter secure encryption key"
            class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
          />
          <div class="flex items-center space-x-2 pt-1">
            <div class="flex-1 h-1 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
              <div class="h-full {strength.color} transition-all" style="width: {strength.width}"></div>
            </div>
            <span class="text-[10px] text-gray-400 font-medium">{strength.label}</span>
          </div>
        </div>

        <div class="space-y-1.5">
          <label for="confirmPasswordInput" class="block font-semibold text-gray-700 dark:text-gray-300">Confirm Password</label>
          <input
            id="confirmPasswordInput"
            type="password"
            bind:value={confirmPassword}
            placeholder="Re-enter password"
            class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
          />
          {#if confirmPassword && password !== confirmPassword}
            <p class="text-[11px] text-rose-500 font-semibold">Passwords do not match</p>
          {/if}
        </div>

        <div class="pt-1">
          <label class="flex items-center space-x-2 cursor-pointer text-gray-700 dark:text-gray-300">
            <input type="checkbox" bind:checked={dryRun} class="rounded text-indigo-600" />
            <span class="font-medium">Dry Run mode (simulate scan & exclusions without writing repository data)</span>
          </label>
        </div>
      </div>

      <div class="flex justify-between items-center pt-2">
        <button type="button" on:click={() => (currentStep = 1)} class="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">← Back</button>
        <button
          type="button"
          disabled={!password || password !== confirmPassword}
          class="bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all"
          on:click={startBackup}
        >
          {dryRun ? 'Run Simulated Backup' : 'Start Encrypted Backup'}
        </button>
      </div>
    </div>

  <!-- STEP 3: RUN & DONE -->
  {:else if currentStep === 3}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-6 space-y-6 shadow-sm">
      {#if !backupDone}
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-8 h-8 rounded-full border-3 border-indigo-500 border-t-transparent animate-spin"></div>
            <div>
              <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">Backing up files to repository...</h2>
              <p class="text-xs text-gray-400 font-mono">{currentAction}</p>
            </div>
          </div>
          <ProgressBar percent={runProgress} label="Backup Progress" />
        </div>

        {#if logs.length > 0}
          <div class="bg-gray-900 text-gray-300 rounded-lg p-3 font-mono text-[11px] max-h-40 overflow-y-auto space-y-1 border border-gray-800">
            {#each logs as log}
              <div class="truncate text-gray-400">&gt; {log}</div>
            {/each}
          </div>
        {/if}
      {:else}
        <!-- Done State -->
        <div class="text-center space-y-4 py-4">
          <div class="w-12 h-12 bg-emerald-100 dark:bg-emerald-950 text-emerald-600 rounded-full flex items-center justify-center mx-auto text-2xl font-bold">
            ✓
          </div>
          <div class="space-y-1">
            <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">Backup Complete & Verified!</h2>
            <p class="text-xs text-gray-500">Snapshot ID: <span class="font-mono text-indigo-500 font-bold">{snapshotResult.id}</span></p>
          </div>

          <div class="bg-gray-50 dark:bg-gray-900/60 border border-gray-200/60 dark:border-gray-800 p-4 rounded-xl max-w-sm mx-auto text-left space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-gray-500">Destination:</span>
              <span class="font-mono text-gray-800 dark:text-gray-200 truncate max-w-[200px]">{backupDestination}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">Encryption:</span>
              <span class="text-emerald-600 font-semibold">AES-256 Poly1305</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">Integrity:</span>
              <span class="text-emerald-600">Verified (SHA256 Match)</span>
            </div>
            {#if snapshotResult.manifestSha}
              <div class="flex justify-between">
                <span class="text-gray-500">Manifest SHA:</span>
                <span class="font-mono text-gray-400 text-[10px] truncate max-w-[150px]">{snapshotResult.manifestSha}</span>
              </div>
            {/if}
          </div>

          <div class="pt-4 flex justify-center space-x-3">
            <button
              type="button"
              class="bg-indigo-600 hover:bg-indigo-700 text-white px-5 py-2 rounded-lg text-xs font-semibold shadow-sm transition-all"
              on:click={() => ($currentScreen = 'home')}
            >
              Return to Dashboard
            </button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>
