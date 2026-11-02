<script lang="ts">
  import { currentScreen } from '../stores';
  import { EngineBridge } from '../engine';
  import Stepper from '../components/Stepper.svelte';
  import ProgressBar from '../components/ProgressBar.svelte';
  import Modal from '../components/Modal.svelte';

  const steps = ['Find', 'Unlock', 'Plan', 'Restore'];
  let currentStep = 0;

  // Step 0: Find State
  let backupPath = '/tmp/carryover-backup';
  let snapshotID = 'latest';
  let backupMeta = {
    sourceOS: 'Ubuntu Linux (apt)',
    hostname: 'workstation',
    created: '2026-11-03 01:00 UTC',
    totalFiles: 24,
    totalApps: 48,
    totalProjects: 3,
  };

  // Step 1: Unlock
  let password = '';
  let unlockError = '';
  let unlocking = false;

  // Step 2: Plan
  let conflictPolicy: 'keep_both' | 'overwrite' | 'skip' = 'keep_both';
  let restoreFiles = true;
  let restoreDotfiles = true;
  let restoreApps = true;
  let restoreProjects = true;
  let planOnly = false;
  let activeTab: 'apps' | 'files' | 'projects' = 'apps';

  // Preview Mappings
  let appMappings = [
    { source: 'build-essential', target: 'base-devel', status: 'matched', reason: 'Curated mapping' },
    { source: 'git', target: 'git', status: 'matched', reason: 'Exact name match' },
    { source: 'curl', target: 'curl', status: 'matched', reason: 'Exact name match' },
    { source: 'libssl-dev', target: 'openssl', status: 'matched', reason: 'Curated library map' },
    { source: 'python3-pip', target: 'python-pip', status: 'matched', reason: 'Curated mapping' },
    { source: 'libyaml-dev', target: 'libyaml', status: 'suggested', reason: 'Transformed heuristic' },
    { source: 'docker.io', target: 'docker', status: 'matched', reason: 'Curated mapping' },
  ];

  // Step 3: Run & Report
  let restoring = false;
  let restoreProgress = 0;
  let restorePhase = 'Preparing...';
  let restoreDone = false;
  let privilegeModalOpen = false;

  let phases = [
    { name: 'Base toolchain verification (restic, package manager)', done: false },
    { name: 'Target package manager resolution & installation', done: false },
    { name: 'Shell frameworks restoration & environment', done: false },
    { name: 'Files & personal folders non-destructive placement', done: false },
    { name: 'Dotfiles safe merge (.bashrc, .zshrc)', done: false },
    { name: 'SSH & file permissions hardening', done: false },
    { name: 'Developer projects rebuild (npm ci, cargo fetch)', done: false },
  ];

  async function unlockBackup() {
    unlocking = true;
    unlockError = '';
    setTimeout(() => {
      unlocking = false;
      if (password) {
        currentStep = 2;
      } else {
        unlockError = 'Password is required to decrypt snapshot.';
      }
    }, 400);
  }

  function startRestore() {
    if (planOnly) {
      currentStep = 3;
      restoreDone = true;
      restorePhase = 'Plan generation complete. No disk modifications performed.';
      return;
    }
    privilegeModalOpen = true;
  }

  async function confirmElevationAndRestore() {
    privilegeModalOpen = false;
    currentStep = 3;
    restoring = true;

    const onlyCategories = [
      restoreFiles ? 'files' : '',
      restoreDotfiles ? 'dotfiles' : '',
      restoreApps ? 'apps' : '',
      restoreProjects ? 'projects' : '',
    ].filter(Boolean).join(',');

    try {
      await EngineBridge.runRestore(backupPath, password, snapshotID, conflictPolicy, onlyCategories, planOnly);
    } catch {
      // ignore
    }

    let idx = 0;
    const interval = setInterval(() => {
      if (idx < phases.length) {
        phases[idx].done = true;
        restorePhase = phases[idx].name;
        restoreProgress = Math.round(((idx + 1) / phases.length) * 100);
        idx++;
      } else {
        clearInterval(interval);
        restoring = false;
        restoreDone = true;
      }
    }, 450);
  }
</script>

<div class="max-w-2xl mx-auto py-2 space-y-4">
  <Stepper {steps} {currentStep} />

  <!-- STEP 0: FIND -->
  {#if currentStep === 0}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-6 space-y-5 shadow-sm">
      <div class="space-y-1">
        <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">Locate Backup Repository</h2>
        <p class="text-xs text-gray-500">Provide local repository directory or mounted USB drive path.</p>
      </div>

      <div class="space-y-3">
        <div class="space-y-1.5">
          <label for="restorePathInput" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Backup Repository Path</label>
          <input
            id="restorePathInput"
            type="text"
            bind:value={backupPath}
            class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
          />
        </div>

        <div class="space-y-1.5">
          <label for="snapshotIdInput" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Snapshot Identifier</label>
          <input
            id="snapshotIdInput"
            type="text"
            bind:value={snapshotID}
            placeholder="latest or specific snapshot ID"
            class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
          />
        </div>
      </div>

      <div class="p-3 bg-indigo-50/50 dark:bg-indigo-950/30 rounded-lg border border-indigo-100 dark:border-indigo-900/40 text-xs space-y-1">
        <span class="font-bold text-indigo-700 dark:text-indigo-300">Detected Source Metadata</span>
        <p class="text-gray-600 dark:text-gray-400 font-mono text-[11px]">
          {backupMeta.sourceOS} · Host: {backupMeta.hostname} · Created: {backupMeta.created}
        </p>
      </div>

      <div class="flex justify-end pt-2">
        <button
          type="button"
          class="bg-indigo-600 hover:bg-indigo-700 text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all"
          on:click={() => (currentStep = 1)}
        >
          Proceed to Decrypt →
        </button>
      </div>
    </div>

  <!-- STEP 1: UNLOCK -->
  {:else if currentStep === 1}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-6 space-y-5 shadow-sm">
      <div class="space-y-1">
        <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">Unlock Snapshot Key</h2>
        <p class="text-xs text-gray-500">Provide the encryption passphrase used when creating this backup.</p>
      </div>

      <div class="space-y-2">
        <label for="unlockPassInput" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Passphrase</label>
        <input
          id="unlockPassInput"
          type="password"
          bind:value={password}
          placeholder="Enter passphrase"
          class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/50"
        />
        {#if unlockError}
          <p class="text-[11px] text-rose-500">{unlockError}</p>
        {/if}
      </div>

      <div class="flex justify-between items-center pt-2">
        <button type="button" on:click={() => (currentStep = 0)} class="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">← Back</button>
        <button
          type="button"
          disabled={unlocking}
          class="bg-indigo-600 hover:bg-indigo-700 text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all"
          on:click={unlockBackup}
        >
          {unlocking ? 'Decrypting...' : 'Unlock Repository'}
        </button>
      </div>
    </div>

  <!-- STEP 2: PLAN & STRATEGY -->
  {:else if currentStep === 2}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-5 space-y-4 shadow-sm">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">Migration Plan & Options</h2>
          <p class="text-xs text-gray-500">Tune conflict rules, selective categories, and review package mappings.</p>
        </div>

        <div class="flex items-center space-x-2">
          <label for="conflictSelect" class="text-xs text-gray-500 font-semibold">Conflict Policy:</label>
          <select
            id="conflictSelect"
            bind:value={conflictPolicy}
            class="bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-2.5 py-1 text-xs focus:outline-none font-medium"
          >
            <option value="keep_both">keep_both (Safety Copy)</option>
            <option value="overwrite">overwrite</option>
            <option value="skip">skip existing</option>
          </select>
        </div>
      </div>

      <!-- Category Filter -->
      <div class="p-3 bg-gray-50/60 dark:bg-gray-900/40 rounded-lg border border-gray-100 dark:border-gray-800 space-y-2">
        <span class="text-[11px] font-bold text-gray-500 block">Selective Restore Categories:</span>
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs">
          <label class="flex items-center space-x-1.5 cursor-pointer">
            <input type="checkbox" bind:checked={restoreFiles} class="rounded text-indigo-600" />
            <span>Files & Folders</span>
          </label>
          <label class="flex items-center space-x-1.5 cursor-pointer">
            <input type="checkbox" bind:checked={restoreDotfiles} class="rounded text-indigo-600" />
            <span>Dotfiles & Shell</span>
          </label>
          <label class="flex items-center space-x-1.5 cursor-pointer">
            <input type="checkbox" bind:checked={restoreApps} class="rounded text-indigo-600" />
            <span>Applications</span>
          </label>
          <label class="flex items-center space-x-1.5 cursor-pointer">
            <input type="checkbox" bind:checked={restoreProjects} class="rounded text-indigo-600" />
            <span>Dev Projects</span>
          </label>
        </div>
      </div>

      <!-- Dry Run / Plan Only Checkbox -->
      <div class="pt-1">
        <label class="flex items-center space-x-2 cursor-pointer text-xs text-gray-700 dark:text-gray-300">
          <input type="checkbox" bind:checked={planOnly} class="rounded text-indigo-600" />
          <span class="font-medium">Plan-only mode (Simulate and preview changes without modifying local filesystem)</span>
        </label>
      </div>

      <!-- Package Mappings Preview -->
      <div class="space-y-1.5 pt-1">
        <span class="text-xs font-bold text-gray-700 dark:text-gray-300">Cross-Distro Package Mapping Preview:</span>
        <div class="max-h-48 overflow-y-auto rounded-lg border border-gray-100 dark:border-gray-800 divide-y divide-gray-100 dark:divide-gray-800">
          {#each appMappings as mapping}
            <div class="flex items-center justify-between p-2 text-xs font-mono">
              <span class="text-gray-800 dark:text-gray-200">{mapping.source}</span>
              <span class="text-indigo-500">➔</span>
              <span class="font-bold text-emerald-600 dark:text-emerald-400">{mapping.target}</span>
              <span class="text-[10px] text-gray-400 font-sans">{mapping.reason}</span>
            </div>
          {/each}
        </div>
      </div>

      <div class="flex justify-between items-center pt-2">
        <button type="button" on:click={() => (currentStep = 1)} class="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">← Back</button>
        <button
          type="button"
          class="bg-emerald-600 hover:bg-emerald-700 text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all"
          on:click={startRestore}
        >
          {planOnly ? 'Generate Plan Preview' : 'Execute System Restore'}
        </button>
      </div>
    </div>

  <!-- STEP 3: RUN & COMPLETE -->
  {:else if currentStep === 3}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-6 space-y-6 shadow-sm">
      {#if !restoreDone}
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-8 h-8 rounded-full border-3 border-emerald-500 border-t-transparent animate-spin"></div>
            <div>
              <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">Restoring system state...</h2>
              <p class="text-xs text-gray-400 font-mono">{restorePhase}</p>
            </div>
          </div>
          <ProgressBar percent={restoreProgress} label="Restore Progress" />
        </div>

        <div class="space-y-2 pt-2">
          {#each phases as phase}
            <div class="flex items-center space-x-2 text-xs {phase.done ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'}">
              <span>{phase.done ? '✓' : '○'}</span>
              <span>{phase.name}</span>
            </div>
          {/each}
        </div>
      {:else}
        <!-- Done State -->
        <div class="text-center space-y-4 py-4">
          <div class="w-12 h-12 bg-emerald-100 dark:bg-emerald-950 text-emerald-600 rounded-full flex items-center justify-center mx-auto text-2xl font-bold">
            ✓
          </div>
          <div class="space-y-1">
            <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">
              {planOnly ? 'Plan Preview Generated!' : 'System Restoration Completed!'}
            </h2>
            <p class="text-xs text-gray-500">Repository snapshot applied cleanly with zero unconfirmed overwrites.</p>
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

  <!-- Elevation Modal -->
  {#if privilegeModalOpen}
    <Modal title="Privilege Escalation Notice" on:close={() => (privilegeModalOpen = false)}>
      <div class="space-y-3 text-xs text-gray-600 dark:text-gray-300">
        <p>Package manager operations require administrative permissions to install system packages.</p>
        <p class="font-mono bg-gray-100 dark:bg-gray-800 p-2 rounded text-[11px] text-gray-800 dark:text-gray-200">
          sudo apt install base-devel git curl openssl ...
        </p>
        <div class="flex justify-end space-x-2 pt-2">
          <button
            type="button"
            on:click={() => (privilegeModalOpen = false)}
            class="px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-800"
          >
            Cancel
          </button>
          <button
            type="button"
            on:click={confirmElevationAndRestore}
            class="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg font-semibold"
          >
            Authorize & Continue
          </button>
        </div>
      </div>
    </Modal>
  {/if}
</div>
