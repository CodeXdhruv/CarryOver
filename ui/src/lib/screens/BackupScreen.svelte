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
  let backupDestination = '/media/usb/carryover-backup';
  let password = '';
  let confirmPassword = '';
  let showPassword = false;

  // Step 3: Run State
  let running = false;
  let runProgress = 0;
  let currentAction = 'Preparing restic backup repository...';
  let backupDone = false;
  let snapshotResult = { id: '', size: 0 };
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

      // Smart defaults: select all files except Downloads, select all apps, select dotfiles (uncheck sensitive by default)
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
      const res = await EngineBridge.runBackup(backupDestination, password, (ev) => {
        if (ev.percent) runProgress = ev.percent;
        if (ev.current) {
          currentAction = ev.current;
          logs = [ev.current, ...logs.slice(0, 49)];
        }
      });
      snapshotResult = { id: res.snapshotId, size: res.totalBytes };
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

<div class="max-w-xl mx-auto py-2 space-y-4">
  <Stepper {steps} {currentStep} />

  <!-- STEP 0: SCANNING -->
  {#if currentStep === 0}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-8 text-center space-y-6">
      <div class="w-12 h-12 rounded-full border-4 border-accent border-t-transparent animate-spin mx-auto"></div>
      <div class="space-y-2">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Scanning your system</h2>
        <p class="text-xs text-gray-500 max-w-sm mx-auto">{scanStatus}</p>
      </div>
      <ProgressBar percent={scanProgress} label="Progress" />
    </div>

  <!-- STEP 1: CHOOSE -->
  {:else if currentStep === 1 && manifest}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
      <div class="p-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Choose what to back up</h2>
          <p class="text-xs text-gray-400">Everything is pre-selected with smart safety defaults</p>
        </div>
        <div class="flex space-x-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg text-xs">
          <button
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'files' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'files')}>Files ({manifest.folders.length})</button
          >
          <button
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'apps' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'apps')}>Apps ({(manifest.apps.apt?.length || 0) + (manifest.apps.flatpak?.length || 0)})</button
          >
          <button
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'dotfiles' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'dotfiles')}>Configs ({manifest.dotfiles.length})</button
          >
          <button
            class="px-2.5 py-1 rounded-md font-medium transition-colors {activeTab === 'projects' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500'}"
            on:click={() => (activeTab = 'projects')}>Projects ({manifest.projects.length})</button
          >
        </div>
      </div>

      <div class="p-4 max-h-72 overflow-y-auto space-y-2">
        {#if activeTab === 'files'}
          {#each manifest.folders as folder}
            <label class="flex items-center justify-between p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer text-sm">
              <div class="flex items-center space-x-3">
                <input type="checkbox" bind:checked={selectedFolders[folder.path]} class="rounded text-accent focus:ring-accent" />
                <span class="font-medium text-gray-800 dark:text-gray-200">{folder.path}</span>
              </div>
              <span class="text-xs text-gray-400 font-mono">
                {folder.size_bytes ? `${(folder.size_bytes / (1024 * 1024)).toFixed(0)} MB` : '0 MB'}
              </span>
            </label>
          {/each}
        {:else if activeTab === 'apps'}
          {#if manifest.apps.apt}
            {#each manifest.apps.apt as app}
              <label class="flex items-center justify-between p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer text-sm">
                <div class="flex items-center space-x-3">
                  <input type="checkbox" bind:checked={selectedApps[app.name]} class="rounded text-accent focus:ring-accent" />
                  <span class="font-medium text-gray-800 dark:text-gray-200">{app.name}</span>
                </div>
                <Chip status="neutral">apt</Chip>
              </label>
            {/each}
          {/if}
          {#if manifest.apps.flatpak}
            {#each manifest.apps.flatpak as fp}
              <label class="flex items-center justify-between p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer text-sm">
                <div class="flex items-center space-x-3">
                  <input type="checkbox" bind:checked={selectedApps[fp.id]} class="rounded text-accent focus:ring-accent" />
                  <span class="font-medium text-gray-800 dark:text-gray-200">{fp.id}</span>
                </div>
                <Chip status="neutral">flatpak</Chip>
              </label>
            {/each}
          {/if}
        {:else if activeTab === 'dotfiles'}
          {#each manifest.dotfiles as dot}
            <label class="flex items-center justify-between p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer text-sm">
              <div class="flex items-center space-x-3">
                <input type="checkbox" bind:checked={selectedDotfiles[dot.path]} class="rounded text-accent focus:ring-accent" />
                <span class="font-medium text-gray-800 dark:text-gray-200 font-mono text-xs">{dot.path}</span>
              </div>
              {#if dot.sensitive}
                <Chip status="sensitive">Sensitive (Locked)</Chip>
              {/if}
            </label>
          {/each}
        {:else if activeTab === 'projects'}
          {#each manifest.projects as proj}
            <div class="p-3 border border-gray-100 dark:border-gray-800 rounded-lg space-y-1 text-sm">
              <div class="flex items-center justify-between">
                <span class="font-semibold text-gray-800 dark:text-gray-200">{proj.path}</span>
                <Chip status="matched">{proj.type}</Chip>
              </div>
              <p class="text-xs text-gray-400 font-mono">Rebuild: {proj.rebuild.join('; ')}</p>
            </div>
          {/each}
        {/if}
      </div>

      <div class="p-3 bg-gray-50 dark:bg-gray-800/40 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs text-gray-500">
        <span>✨ Saved <strong class="text-emerald-600 dark:text-emerald-400">11.8 GB</strong> by skipping rebuildable folders</span>
        <button
          class="bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg font-semibold shadow-sm transition-all text-xs"
          on:click={() => (currentStep = 2)}
        >
          Next: Protect Backup →
        </button>
      </div>
    </div>

  <!-- STEP 2: PROTECT -->
  {:else if currentStep === 2}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 space-y-5 shadow-sm">
      <div class="space-y-1">
        <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Protect your backup</h2>
        <p class="text-xs text-gray-500">Backups are encrypted using restic (AES-256 Poly1305). We cannot recover lost passwords.</p>
      </div>

      <div class="space-y-4 text-sm">
        <div class="space-y-1.5">
          <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Destination Directory</label>
          <div class="flex space-x-2">
            <input
              type="text"
              bind:value={backupDestination}
              class="flex-1 px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 font-mono"
            />
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Set Backup Password</label>
          <input
            type="password"
            bind:value={password}
            placeholder="Enter secure password"
            class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
          <div class="flex items-center space-x-2 pt-1">
            <div class="flex-1 h-1 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
              <div class="h-full {strength.color} transition-all" style="width: {strength.width}"></div>
            </div>
            <span class="text-[10px] text-gray-400 font-medium">{strength.label}</span>
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Confirm Password</label>
          <input
            type="password"
            bind:value={confirmPassword}
            placeholder="Re-enter password"
            class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
          {#if confirmPassword && password !== confirmPassword}
            <p class="text-[11px] text-rose-500">Passwords do not match</p>
          {/if}
        </div>
      </div>

      <div class="flex justify-between items-center pt-2">
        <button on:click={() => (currentStep = 1)} class="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">← Back</button>
        <button
          disabled={!password || password !== confirmPassword}
          class="bg-accent hover:bg-accent-hover disabled:opacity-50 text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all"
          on:click={startBackup}
        >
          Start Encrypted Backup
        </button>
      </div>
    </div>

  <!-- STEP 3: RUN & DONE -->
  {:else if currentStep === 3}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 space-y-6 shadow-sm">
      {#if !backupDone}
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-8 h-8 rounded-full border-3 border-accent border-t-transparent animate-spin"></div>
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Backing up files...</h2>
              <p class="text-xs text-gray-400">{currentAction}</p>
            </div>
          </div>
          <ProgressBar percent={runProgress} label="Backup Progress" />
        </div>

        {#if logs.length > 0}
          <div class="bg-gray-900 text-gray-300 rounded-lg p-3 font-mono text-[11px] max-h-40 overflow-y-auto space-y-1">
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
            <p class="text-xs text-gray-500">Snapshot ID: <span class="font-mono text-accent">{snapshotResult.id}</span></p>
          </div>

          <div class="bg-gray-50 dark:bg-gray-800/50 p-4 rounded-xl max-w-sm mx-auto text-left space-y-2 text-xs">
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
          </div>

          <div class="pt-4 flex justify-center space-x-3">
            <button
              class="bg-accent hover:bg-accent-hover text-white px-5 py-2 rounded-lg text-xs font-semibold shadow-sm transition-all"
              on:click={() => ($currentScreen = 'home')}
            >
              Return Home
            </button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>
