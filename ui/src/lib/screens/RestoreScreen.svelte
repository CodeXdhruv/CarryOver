<script lang="ts">
  import { currentScreen } from '../stores';
  import Stepper from '../components/Stepper.svelte';
  import ProgressBar from '../components/ProgressBar.svelte';
  import Chip from '../components/Chip.svelte';
  import Modal from '../components/Modal.svelte';

  const steps = ['Find', 'Unlock', 'Preview', 'Restore'];
  let currentStep = 0;

  // Step 0: Find State
  let backupPath = '/media/usb/carryover-backup';
  let backupMeta = {
    sourceOS: 'Ubuntu 24.04 (apt)',
    hostname: 'thinkpad-x1',
    created: '2026-11-02 10:15 UTC',
    totalFiles: 24,
    totalApps: 48,
    totalProjects: 3,
  };

  // Step 1: Unlock
  let password = '';
  let unlockError = '';
  let unlocking = false;

  // Step 2: Preview
  let conflictPolicy: 'keep_both' | 'overwrite' | 'skip' = 'keep_both';
  let migrationScore = 92;
  let activeTab: 'apps' | 'files' | 'projects' = 'apps';

  // Preview Mappings
  let appMappings = [
    { source: 'build-essential', target: 'base-devel', status: 'matched', reason: 'Curated mapping' },
    { source: 'git', target: 'git', status: 'matched', reason: 'Exact name match' },
    { source: 'curl', target: 'curl', status: 'matched', reason: 'Exact name match' },
    { source: 'libssl-dev', target: 'openssl', status: 'matched', reason: 'Curated library map' },
    { source: 'python3-pip', target: 'python-pip', status: 'matched', reason: 'Curated mapping' },
    { source: 'libyaml-dev', target: 'libyaml', status: 'suggested', reason: 'Transformed heuristic' },
    { source: 'proprietary-driver-tool', target: 'N/A', status: 'unmapped', reason: 'Not in repository' },
  ];

  // Step 3: Run & Report
  let restoring = false;
  let restoreProgress = 0;
  let restorePhase = 'Preparing...';
  let restoreDone = false;
  let privilegeModalOpen = false;

  let phases = [
    { name: 'Base toolchain verification', done: false },
    { name: 'Package manager installation', done: false },
    { name: 'Shell frameworks restoration', done: false },
    { name: 'Files & personal folders placement', done: false },
    { name: 'Dotfiles safe merge (.bashrc, .zshrc)', done: false },
    { name: 'SSH & file permissions hardening', done: false },
    { name: 'Developer projects rebuild (npm ci, etc.)', done: false },
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
    privilegeModalOpen = true;
  }

  function confirmElevationAndRestore() {
    privilegeModalOpen = false;
    currentStep = 3;
    restoring = true;

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
    }, 600);
  }
</script>

<div class="max-w-xl mx-auto py-2 space-y-4">
  <Stepper {steps} {currentStep} />

  <!-- STEP 0: FIND BACKUP -->
  {#if currentStep === 0}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 space-y-5 shadow-sm">
      <div class="space-y-1">
        <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Select Backup Destination</h2>
        <p class="text-xs text-gray-500">Pick the directory or external drive containing your Carryover backup repository.</p>
      </div>

      <div class="space-y-3">
        <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Backup Repository Path</label>
        <input
          type="text"
          bind:value={backupPath}
          class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 font-mono"
        />

        <!-- Public Metadata Preview Card -->
        <div class="bg-gray-50 dark:bg-gray-800/40 p-4 rounded-xl border border-gray-200/60 dark:border-gray-700/60 space-y-2 text-xs">
          <div class="flex items-center justify-between pb-1 border-b border-gray-200/40 dark:border-gray-700/40">
            <span class="font-semibold text-gray-900 dark:text-gray-100">Public Metadata (manifest.public.json)</span>
            <span class="text-emerald-600 font-medium">Valid Snapshot</span>
          </div>
          <div class="grid grid-cols-2 gap-2 text-gray-600 dark:text-gray-300 pt-1">
            <div>Source Distro: <strong class="text-gray-900 dark:text-gray-100">{backupMeta.sourceOS}</strong></div>
            <div>Hostname: <strong class="text-gray-900 dark:text-gray-100">{backupMeta.hostname}</strong></div>
            <div>Created: <strong class="text-gray-900 dark:text-gray-100">{backupMeta.created}</strong></div>
            <div>Content: <strong class="text-gray-900 dark:text-gray-100">{backupMeta.totalApps} apps · {backupMeta.totalProjects} projects</strong></div>
          </div>
        </div>
      </div>

      <div class="flex justify-end pt-2">
        <button
          class="bg-accent hover:bg-accent-hover text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all"
          on:click={() => (currentStep = 1)}
        >
          Next: Unlock Repository →
        </button>
      </div>
    </div>

  <!-- STEP 1: UNLOCK -->
  {:else if currentStep === 1}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 space-y-5 shadow-sm max-w-md mx-auto">
      <div class="space-y-1 text-center">
        <div class="w-10 h-10 rounded-full bg-blue-50 dark:bg-blue-950/60 text-accent flex items-center justify-center mx-auto text-lg mb-2">
          🔒
        </div>
        <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Unlock Backup</h2>
        <p class="text-xs text-gray-500">Enter the encryption password you created when backing up.</p>
      </div>

      <div class="space-y-3">
        <input
          type="password"
          bind:value={password}
          placeholder="Enter backup password"
          class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg text-xs bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          on:keydown={(e) => e.key === 'Enter' && unlockBackup()}
        />
        {#if unlockError}
          <p class="text-xs text-rose-500">{unlockError}</p>
        {/if}
      </div>

      <div class="flex justify-between items-center pt-2">
        <button on:click={() => (currentStep = 0)} class="text-xs text-gray-500 hover:text-gray-700">← Back</button>
        <button
          disabled={unlocking}
          class="bg-accent hover:bg-accent-hover text-white px-5 py-2 rounded-lg font-semibold text-xs shadow-sm transition-all disabled:opacity-50"
          on:click={unlockBackup}
        >
          {unlocking ? 'Verifying...' : 'Unlock & Preview'}
        </button>
      </div>
    </div>

  <!-- STEP 2: PREVIEW -->
  {:else if currentStep === 2}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm space-y-0">
      <div class="p-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
        <div>
          <div class="flex items-center space-x-2">
            <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Migration Preview</h2>
            <span class="bg-emerald-100 dark:bg-emerald-950 text-emerald-600 text-[10px] font-bold px-2 py-0.5 rounded-full">
              Score: {migrationScore}%
            </span>
          </div>
          <p class="text-xs text-gray-400">Review package mappings and conflict handling before making changes</p>
        </div>

        <div class="text-xs">
          <label class="text-gray-500 mr-2">Conflict Policy:</label>
          <select
            bind:value={conflictPolicy}
            class="px-2 py-1 border border-gray-200 dark:border-gray-700 rounded-md text-xs bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            <option value="keep_both">Keep both (save old copy)</option>
            <option value="overwrite">Overwrite existing</option>
            <option value="skip">Skip conflicts</option>
          </select>
        </div>
      </div>

      <!-- Mapping items list -->
      <div class="p-4 max-h-72 overflow-y-auto space-y-2">
        {#each appMappings as map}
          <div class="p-2.5 rounded-lg border border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs">
            <div class="flex items-center space-x-2">
              <span class="font-mono font-medium text-gray-800 dark:text-gray-200">{map.source}</span>
              <span class="text-gray-400">→</span>
              <span class="font-mono font-semibold {map.status === 'unmapped' ? 'text-gray-400 line-through' : 'text-accent'}">
                {map.target}
              </span>
            </div>
            <div class="flex items-center space-x-2">
              <span class="text-[11px] text-gray-400">{map.reason}</span>
              <Chip status={map.status === 'matched' ? 'matched' : map.status === 'suggested' ? 'suggested' : 'unmapped'}>
                {map.status}
              </Chip>
            </div>
          </div>
        {/each}
      </div>

      <div class="p-4 bg-gray-50 dark:bg-gray-800/40 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs">
        <span class="text-gray-500">Dotfiles like <code>.bashrc</code> will be safely merged below marker blocks.</span>
        <button
          class="bg-accent hover:bg-accent-hover text-white px-5 py-2 rounded-lg font-semibold shadow-sm transition-all text-xs"
          on:click={startRestore}
        >
          Start Restore
        </button>
      </div>
    </div>

  <!-- STEP 3: RUN & REPORT -->
  {:else if currentStep === 3}
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-6 space-y-6 shadow-sm">
      {#if !restoreDone}
        <div class="space-y-4">
          <div class="flex items-center space-x-3">
            <div class="w-8 h-8 rounded-full border-3 border-accent border-t-transparent animate-spin"></div>
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Restoring system...</h2>
              <p class="text-xs text-gray-400">{restorePhase}</p>
            </div>
          </div>
          <ProgressBar percent={restoreProgress} label="Phase Progress" />
        </div>

        <div class="space-y-2 pt-2">
          {#each phases as phase}
            <div class="flex items-center space-x-3 text-xs">
              <span class="w-4 h-4 rounded-full flex items-center justify-center font-bold {phase.done ? 'text-emerald-500' : 'text-gray-300'}">
                {phase.done ? '✓' : '○'}
              </span>
              <span class="{phase.done ? 'text-gray-800 dark:text-gray-200 font-medium' : 'text-gray-400'}">{phase.name}</span>
            </div>
          {/each}
        </div>
      {:else}
        <!-- Report State -->
        <div class="space-y-5">
          <div class="text-center space-y-1">
            <div class="w-12 h-12 bg-emerald-100 dark:bg-emerald-950 text-emerald-600 rounded-full flex items-center justify-center mx-auto text-2xl font-bold mb-2">
              ✓
            </div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">Restore Complete</h2>
            <p class="text-xs text-gray-500">Your personal files, shell configs, and applications have been restored.</p>
          </div>

          <!-- Needs Attention Section First -->
          <div class="bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50 rounded-xl p-4 space-y-2 text-xs">
            <h3 class="font-semibold text-amber-800 dark:text-amber-300">Needs Attention (1 manual step)</h3>
            <ul class="list-disc list-inside text-amber-700 dark:text-amber-400 space-y-1">
              <li>Package <code>proprietary-driver-tool</code> was not found on this distro repository. Install manually if required.</li>
            </ul>
          </div>

          <div class="bg-gray-50 dark:bg-gray-800/40 rounded-xl p-4 space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-gray-500">Files placed:</span>
              <span class="font-semibold text-gray-800 dark:text-gray-200">24 items (0 errors)</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">Shell configurations merged:</span>
              <span class="font-semibold text-gray-800 dark:text-gray-200">.bashrc, .profile</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">Projects rebuilt:</span>
              <span class="font-semibold text-gray-800 dark:text-gray-200">3 projects (npm ci, cargo fetch)</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500">Previous file conflicts:</span>
              <span class="font-mono text-gray-500">Preserved in ~/.carryover-old/</span>
            </div>
          </div>

          <div class="flex justify-center space-x-3 pt-2">
            <button
              class="bg-accent hover:bg-accent-hover text-white px-5 py-2 rounded-lg text-xs font-semibold shadow-sm transition-all"
              on:click={() => ($currentScreen = 'home')}
            >
              Finish & Return Home
            </button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Privilege Explanation Modal -->
<Modal
  title="Administrator Privileges Required"
  open={privilegeModalOpen}
  on:close={() => (privilegeModalOpen = false)}
>
  <div class="space-y-3">
    <p>Carryover needs administrative permission to install selected applications using your package manager.</p>
    <div class="bg-gray-100 dark:bg-gray-800 p-2.5 rounded-lg font-mono text-xs text-gray-800 dark:text-gray-200">
      pacman -S --needed --noconfirm base-devel git curl openssl python-pip
    </div>
    <p class="text-xs text-gray-400">You will see your operating system's authentication prompt next.</p>
  </div>

  <svelte:fragment slot="actions">
    <button
      class="px-3 py-1.5 text-xs text-gray-500 hover:text-gray-700"
      on:click={() => (privilegeModalOpen = false)}
    >
      Cancel
    </button>
    <button
      class="bg-accent hover:bg-accent-hover text-white px-4 py-1.5 rounded-lg text-xs font-semibold"
      on:click={confirmElevationAndRestore}
    >
      Authorize & Proceed
    </button>
  </svelte:fragment>
</Modal>
