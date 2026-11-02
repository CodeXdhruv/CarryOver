<script lang="ts">
  import { onMount } from 'svelte';
  import { systemInfo } from '../stores';
  import { EngineBridge, type DoctorReport, type PackageResolution, type SnapshotItem } from '../engine';

  let activeTab: 'doctor' | 'mapper' | 'snapshots' | 'export' = 'doctor';

  // Doctor state
  let doctorReport: DoctorReport | null = null;
  let runningDoctor = false;

  // Mapper state
  let sourcePkg = 'build-essential';
  let fromPM = 'apt';
  let toPM = 'pacman';
  let mappingResult: PackageResolution | null = null;
  let mappingLoading = false;
  let popularPackages = ['build-essential', 'git', 'neovim', 'ripgrep', 'libssl-dev', 'python3-pip', 'docker.io', 'curl', 'htop', 'vlc'];

  // Snapshots & Verification state
  let repoPath = '/tmp/carryover-backup';
  let repoPassword = '';
  let snapshots: SnapshotItem[] = [];
  let loadingSnapshots = false;
  let verificationStatus = '';
  let verifying = false;

  // Ansible Export state
  let exportTargetPM = 'apt';
  let exportedPlaybook = '';
  let exporting = false;
  let copied = false;

  onMount(async () => {
    if ($systemInfo?.package_manager) {
      fromPM = $systemInfo.package_manager;
      exportTargetPM = $systemInfo.package_manager;
    }
    runDoctor();
    translatePackage();
  });

  async function runDoctor() {
    runningDoctor = true;
    doctorReport = await EngineBridge.getDoctorReport();
    runningDoctor = false;
  }

  async function translatePackage() {
    if (!sourcePkg.trim()) return;
    mappingLoading = true;
    mappingResult = await EngineBridge.mapPackage(sourcePkg.trim(), fromPM, toPM);
    mappingLoading = false;
  }

  async function loadSnapshots() {
    loadingSnapshots = true;
    snapshots = await EngineBridge.listSnapshots(repoPath, repoPassword);
    loadingSnapshots = false;
  }

  async function verifyRepo() {
    verifying = true;
    verificationStatus = 'Verifying cryptographic repository signatures and snapshots...';
    const res = await EngineBridge.verifyBackup(repoPath, repoPassword);
    verificationStatus = res.message;
    verifying = false;
  }

  async function generatePlaybook() {
    exporting = true;
    exportedPlaybook = await EngineBridge.exportPlaybook(null, exportTargetPM);
    exporting = false;
  }

  function copyPlaybook() {
    navigator.clipboard.writeText(exportedPlaybook);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }
</script>

<div class="space-y-5 max-w-4xl mx-auto py-2">
  <!-- Tabs Header -->
  <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-2">
    <div>
      <h1 class="text-xl font-bold text-gray-900 dark:text-gray-100">System Tools & Utilities</h1>
      <p class="text-xs text-gray-500 mt-0.5">Diagnostics, cross-distro package mapping, snapshot verification, and headless export.</p>
    </div>

    <div class="flex space-x-1 bg-gray-100 dark:bg-[#131926] p-1 rounded-lg border border-gray-200 dark:border-gray-800 text-xs">
      <button
        type="button"
        class="px-3 py-1.5 rounded-md font-medium transition-all {activeTab === 'doctor' ? 'bg-white dark:bg-gray-800 text-indigo-600 dark:text-indigo-400 shadow-sm font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900'}"
        on:click={() => (activeTab = 'doctor')}
      >
        🩺 Doctor
      </button>

      <button
        type="button"
        class="px-3 py-1.5 rounded-md font-medium transition-all {activeTab === 'mapper' ? 'bg-white dark:bg-gray-800 text-indigo-600 dark:text-indigo-400 shadow-sm font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900'}"
        on:click={() => (activeTab = 'mapper')}
      >
        🔄 Package Mapper
      </button>

      <button
        type="button"
        class="px-3 py-1.5 rounded-md font-medium transition-all {activeTab === 'snapshots' ? 'bg-white dark:bg-gray-800 text-indigo-600 dark:text-indigo-400 shadow-sm font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900'}"
        on:click={() => (activeTab = 'snapshots')}
      >
        💾 Snapshots & Verify
      </button>

      <button
        type="button"
        class="px-3 py-1.5 rounded-md font-medium transition-all {activeTab === 'export' ? 'bg-white dark:bg-gray-800 text-indigo-600 dark:text-indigo-400 shadow-sm font-semibold' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900'}"
        on:click={() => { activeTab = 'export'; if (!exportedPlaybook) generatePlaybook(); }}
      >
        📄 Ansible Export
      </button>
    </div>
  </div>

  <!-- TAB 1: DOCTOR -->
  {#if activeTab === 'doctor'}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-5 shadow-sm space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-sm font-bold text-gray-900 dark:text-gray-100">System Pre-flight Diagnostics</h2>
          <p class="text-xs text-gray-500">Audits local restic engine, permissions, disk space, and package manager availability.</p>
        </div>
        <button
          type="button"
          on:click={runDoctor}
          disabled={runningDoctor}
          class="px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 dark:bg-indigo-950/60 dark:hover:bg-indigo-900/60 text-indigo-600 dark:text-indigo-400 text-xs font-semibold rounded-lg transition-colors border border-indigo-200/60 dark:border-indigo-800/60"
        >
          {runningDoctor ? 'Auditing...' : 'Re-run Checks'}
        </button>
      </div>

      {#if doctorReport}
        <div class="space-y-2.5 pt-2">
          {#each doctorReport.checks as check}
            <div class="flex items-start justify-between p-3 rounded-lg border border-gray-100 dark:border-gray-800/80 bg-gray-50/50 dark:bg-gray-900/40 text-xs">
              <div class="space-y-0.5">
                <div class="flex items-center space-x-2">
                  <span class="font-bold text-gray-900 dark:text-gray-200 capitalize">{check.name}</span>
                  {#if check.status === 'pass'}
                    <span class="px-1.5 py-0.2 rounded text-[10px] font-semibold bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">PASSED</span>
                  {:else if check.status === 'warn'}
                    <span class="px-1.5 py-0.2 rounded text-[10px] font-semibold bg-amber-100 dark:bg-amber-950/60 text-amber-700 dark:text-amber-400">WARNING</span>
                  {:else}
                    <span class="px-1.5 py-0.2 rounded text-[10px] font-semibold bg-red-100 dark:bg-red-950/60 text-red-700 dark:text-red-400">FAILED</span>
                  {/if}
                </div>
                <p class="text-gray-600 dark:text-gray-400">{check.message}</p>
                {#if check.details}
                  <p class="text-[11px] font-mono text-gray-400">{check.details}</p>
                {/if}
              </div>

              <div>
                {#if check.status === 'pass'}
                  <span class="text-emerald-500 font-bold">✓</span>
                {:else if check.status === 'warn'}
                  <span class="text-amber-500 font-bold">!</span>
                {:else}
                  <span class="text-red-500 font-bold">✕</span>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

  <!-- TAB 2: PACKAGE MAPPER -->
  {:else if activeTab === 'mapper'}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-5 shadow-sm space-y-5">
      <div>
        <h2 class="text-sm font-bold text-gray-900 dark:text-gray-100">Cross-Distribution Package Translation</h2>
        <p class="text-xs text-gray-500">Inspect how Debian/Ubuntu apt packages map to Fedora (dnf), Arch (pacman), and openSUSE (zypper).</p>
      </div>

      <!-- Controls -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div class="space-y-1">
          <label for="pkgInput" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Package Name</label>
          <input
            id="pkgInput"
            type="text"
            bind:value={sourcePkg}
            on:input={translatePackage}
            placeholder="e.g. build-essential"
            class="w-full bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-xs focus:ring-2 focus:ring-indigo-500/40 focus:outline-none"
          />
        </div>

        <div class="space-y-1">
          <label for="fromSelect" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Source Manager</label>
          <select
            id="fromSelect"
            bind:value={fromPM}
            on:change={translatePackage}
            class="w-full bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-xs focus:ring-2 focus:ring-indigo-500/40 focus:outline-none"
          >
            <option value="apt">Debian / Ubuntu (apt)</option>
            <option value="pacman">Arch Linux (pacman)</option>
            <option value="dnf">Fedora / RHEL (dnf)</option>
            <option value="zypper">openSUSE (zypper)</option>
          </select>
        </div>

        <div class="space-y-1">
          <label for="toSelect" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Target Manager</label>
          <select
            id="toSelect"
            bind:value={toPM}
            on:change={translatePackage}
            class="w-full bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-xs focus:ring-2 focus:ring-indigo-500/40 focus:outline-none"
          >
            <option value="pacman">Arch Linux (pacman)</option>
            <option value="dnf">Fedora / RHEL (dnf)</option>
            <option value="apt">Debian / Ubuntu (apt)</option>
            <option value="zypper">openSUSE (zypper)</option>
          </select>
        </div>
      </div>

      <!-- Quick presets -->
      <div class="space-y-1.5">
        <span class="text-[11px] text-gray-400">Quick Test Candidates:</span>
        <div class="flex flex-wrap gap-1.5">
          {#each popularPackages as pkg}
            <button
              type="button"
              class="px-2 py-0.5 rounded text-[11px] bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 transition-colors font-mono"
              on:click={() => { sourcePkg = pkg; translatePackage(); }}
            >
              {pkg}
            </button>
          {/each}
        </div>
      </div>

      <!-- Result Card -->
      {#if mappingResult}
        <div class="p-4 rounded-xl border border-indigo-100 dark:border-indigo-900/40 bg-indigo-50/40 dark:bg-indigo-950/20 space-y-3">
          <div class="flex items-center justify-between text-xs">
            <span class="font-bold text-gray-900 dark:text-gray-100">Translation Result</span>
            <span class="px-2 py-0.5 rounded-full text-[10px] font-semibold font-mono uppercase bg-indigo-100 dark:bg-indigo-900/60 text-indigo-700 dark:text-indigo-300">
              Confidence: {mappingResult.confidence}
            </span>
          </div>

          <div class="flex items-center justify-center space-x-6 py-2 text-sm font-mono">
            <div class="text-center">
              <span class="text-[11px] text-gray-400 block mb-0.5">{mappingResult.from_pm}</span>
              <span class="font-bold text-gray-900 dark:text-gray-100 px-3 py-1.5 rounded-lg bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 inline-block shadow-sm">
                {mappingResult.original_name}
              </span>
            </div>

            <div class="text-indigo-500 font-bold text-lg">➔</div>

            <div class="text-center">
              <span class="text-[11px] text-gray-400 block mb-0.5">{mappingResult.to_pm}</span>
              <span class="font-bold text-emerald-600 dark:text-emerald-400 px-3 py-1.5 rounded-lg bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 inline-block shadow-sm">
                {mappingResult.target_name}
              </span>
            </div>
          </div>

          <p class="text-xs text-gray-500 dark:text-gray-400 text-center font-sans">
            Resolution logic: <span class="text-gray-700 dark:text-gray-300">{mappingResult.reason}</span>
          </p>
        </div>
      {/if}
    </div>

  <!-- TAB 3: SNAPSHOTS & VERIFY -->
  {:else if activeTab === 'snapshots'}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-5 shadow-sm space-y-4">
      <div>
        <h2 class="text-sm font-bold text-gray-900 dark:text-gray-100">Backup Repository & Snapshot Verification</h2>
        <p class="text-xs text-gray-500">Inspect existing snapshots on disk or USB, and test repository cryptographic integrity.</p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div class="space-y-1">
          <label for="repoPath" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Repository Path</label>
          <input
            id="repoPath"
            type="text"
            bind:value={repoPath}
            class="w-full bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-xs focus:ring-2 focus:ring-indigo-500/40 focus:outline-none"
          />
        </div>

        <div class="space-y-1">
          <label for="repoPass" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">Encryption Password</label>
          <input
            id="repoPass"
            type="password"
            bind:value={repoPassword}
            placeholder="Enter repository password"
            class="w-full bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-xs focus:ring-2 focus:ring-indigo-500/40 focus:outline-none"
          />
        </div>
      </div>

      <div class="flex items-center space-x-2 pt-1">
        <button
          type="button"
          on:click={loadSnapshots}
          disabled={loadingSnapshots}
          class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-semibold shadow-sm transition-all"
        >
          {loadingSnapshots ? 'Reading repository...' : 'List Snapshots'}
        </button>

        <button
          type="button"
          on:click={verifyRepo}
          disabled={verifying}
          class="px-3 py-1.5 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-lg text-xs font-semibold transition-all border border-gray-200 dark:border-gray-700"
        >
          {verifying ? 'Checking...' : 'Run Cryptographic Check'}
        </button>
      </div>

      {#if verificationStatus}
        <div class="p-3 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200/60 dark:border-emerald-800/60 text-xs text-emerald-800 dark:text-emerald-300">
          {verificationStatus}
        </div>
      {/if}

      {#if snapshots.length > 0}
        <div class="space-y-2 pt-2">
          <span class="text-xs font-bold text-gray-700 dark:text-gray-300">Detected Snapshots ({snapshots.length}):</span>
          {#each snapshots as snap}
            <div class="p-3 rounded-lg border border-gray-200/60 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/50 text-xs flex items-center justify-between">
              <div>
                <span class="font-mono font-bold text-indigo-600 dark:text-indigo-400">{snap.short_id}</span>
                <span class="text-gray-400 mx-1">·</span>
                <span class="text-gray-600 dark:text-gray-300">{new Date(snap.time).toLocaleString()}</span>
                <p class="text-[11px] text-gray-400 font-mono mt-0.5">Host: {snap.hostname} · User: {snap.username}</p>
              </div>
              <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-gray-200 dark:bg-gray-800 text-gray-600 dark:text-gray-400">
                Encrypted
              </span>
            </div>
          {/each}
        </div>
      {/if}
    </div>

  <!-- TAB 4: ANSIBLE EXPORT -->
  {:else if activeTab === 'export'}
    <div class="bg-white dark:bg-[#131926] border border-gray-200/80 dark:border-gray-800/80 rounded-xl p-5 shadow-sm space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-sm font-bold text-gray-900 dark:text-gray-100">Export Ansible Migration Playbook</h2>
          <p class="text-xs text-gray-500">Generate an idempotent Ansible YAML playbook for automated, headless provisioning on fresh systems.</p>
        </div>

        <div class="flex items-center space-x-2">
          <select
            bind:value={exportTargetPM}
            on:change={generatePlaybook}
            class="bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-2.5 py-1 text-xs focus:outline-none"
          >
            <option value="apt">Target: apt</option>
            <option value="dnf">Target: dnf</option>
            <option value="pacman">Target: pacman</option>
            <option value="zypper">Target: zypper</option>
          </select>

          <button
            type="button"
            on:click={copyPlaybook}
            class="px-3 py-1 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-semibold shadow-sm transition-all"
          >
            {copied ? 'Copied!' : 'Copy YAML'}
          </button>
        </div>
      </div>

      <div class="relative">
        <pre class="bg-gray-900 text-gray-200 font-mono text-[11px] p-4 rounded-xl overflow-x-auto max-h-96 leading-relaxed border border-gray-800">
{exportedPlaybook || 'Generating playbook...'}
        </pre>
      </div>
    </div>
  {/if}
</div>
