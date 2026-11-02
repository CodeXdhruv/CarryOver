# WORKFLOW: Carryover

Companion docs: `prd.md`, `spec.md`, `master_prompt.md`

This file describes (A) the end-user workflow, (B) the internal engine workflow, (C) the developer workflow, and (D) the CI/CD workflow.

---

## A. End-user workflow

### A1. Overview
```
OLD OS                                      NEW OS
──────                                      ──────
Install Carryover                           Install Carryover (AppImage/.deb/.rpm/Flatpak)
      │                                           │
 Scan → Choose → Protect → Backup           Find backup → Unlock → Preview → Restore → Report
      │                                           ▲
      └──────── USB / external drive / cloud ─────┘
```

### A2. Backup (old OS)
1. **Open Carryover → Create Backup.**
2. **Scan (auto, ~30 s).** Detects distro, package manager, installed apps, common folders, dotfiles, dev projects.
3. **Choose.** Four tabs: Files, Apps, Settings, Projects. Defaults are pre-ticked. User can untick anything. Bottom bar shows total size and savings from skipped rebuildable folders.
4. **Protect.** Pick destination (USB/folder/cloud) and set a password.
5. **Run.** Progress bar, current file, ETA. Can be cancelled and resumed.
6. **Done.** Summary + optional printable "recovery sheet". User safely removes the drive.

### A3. Restore (new OS)
1. **Install Carryover** and open it. If a backup is found on a mounted drive, a banner offers Restore.
2. **Find backup.** Shows source OS, date, size (public metadata only).
3. **Unlock** with the password.
4. **Preview (dry run).** Shows:
   - Files and where they will go
   - Apps with mapping status (✓ matched, ⚠ suggested, ✕ not found)
   - Projects and rebuild commands
   - Conflicts and chosen policy (default: keep both)
   - Migration score (e.g. 92%)
5. **Start restore.** If admin rights are needed, an explanation modal appears, then the system's authentication prompt.
6. **Report.** "Needs attention" first (unmapped apps, failed rebuilds, manual steps), then successes.

### A4. Edge cases (UX behaviour)
| Situation | Behaviour |
|---|---|
| Wrong password | Inline message, retry; nothing touched |
| Drive unplugged mid-backup | Pause, explain, resume when reconnected |
| Destination full | Pre-check blocks start; shows needed vs available |
| No network on restore | Files/dotfiles still restored; apps/projects queued with "retry when online" |
| App not found on new distro | Marked ✕ with suggestion or Flatpak alternative; listed in manual steps |
| File already exists | Existing moved to `~/.carryover-old/<timestamp>/`, restored copy placed |
| Different username | Paths are `$HOME`-relative, so it works automatically |
| Interrupted restore | Re-run safely (idempotent) |

---

## B. Engine workflow (internal)

### B1. Backup pipeline
```
scan
 ├─ detect_os()                → os_id, version, package_manager
 ├─ list_apps()                → per manager lists
 ├─ find_folders()             → XDG dirs + dev roots + sizes
 ├─ find_dotfiles()            → allowlist ∩ existing
 ├─ capture_shell()            → rc files, env exports, frameworks (nvm, omz…)
 ├─ find_projects()            → marker files → rebuild recipes + exclude dirs
 └─ secret_scan()   (P1)       → warnings
        ↓
selection (UI or selection.yaml)
        ↓
build_manifest() → validate against schema
        ↓
build_file_list() + build_excludes()
        ↓
restic init/backup (password via file descriptor) → events → progress
        ↓
write manifest into snapshot + manifest.public.json beside repo
        ↓
verify (restic check subset + manifest hash)
        ↓
done event + report
```

### B2. Restore pipeline
```
open repo (password) → read manifest
        ↓
detect new OS + package manager
        ↓
plan():
   files        → target paths, conflicts
   apps         → mapper.resolve() with confidence
   projects     → rebuild commands re-derived from allowlist
   privileges   → grouped commands
        ↓
preview (user confirms)        [--plan-only stops here]
        ↓
preflight: network, space, privilege
        ↓
1 install base toolchain
2 install apps (one privileged call per manager)
3 install dev tooling (nvm, pyenv, starship, oh-my-zsh…)
4 restore files
5 restore dotfiles (merge or replace per policy)
6 fix permissions (~/.ssh 700, keys 600)
7 rebuild projects (npm ci, flutter pub get, pip install …)
8 restore VS Code extensions / desktop settings
        ↓
report.json + UI report
```
Each step: independent, logs result, continues on failure, retried twice on network errors.

### B3. Package mapping flow
```
for each app in manifest:
   exact in curated table?           → target name (exact)
   same name in target repo?         → (repo-same)
   suffix transform works?           → (transformed)
   GUI app with Flatpak/Snap match?  → flatpak id (alt)
   fuzzy search hit?                 → (suggested, not auto-selected)
   else                              → (none) → manual list
```

### B4. Privilege flow
```
engine builds command → emits need_privilege(reason, preview)
UI shows modal explaining → user accepts
engine runs: pkexec <helper> (fallback sudo in terminal)
```

---

## C. Developer workflow (for building the project)

### C1. Milestones
| # | Milestone | Deliverable | Done when |
|---|---|---|---|
| M1 | Scanner CLI | `carryover scan --json` | Outputs valid manifest draft on Ubuntu/Fedora/Arch |
| M2 | Backup | `carryover backup` with restic + excludes | Seeded home backs up, node_modules skipped |
| M3 | Restore files/dotfiles | `carryover restore --only files,dotfiles` | Byte-identical, merge works |
| M4 | Mapper + installer | apps reinstall on another distro | ≥85% auto-mapped in test matrix |
| M5 | Rebuild runner | projects rebuilt from allowlist | `npm ci`, `pip install`, `flutter pub get` verified |
| M6 | GUI | Tauri wizard on top of CLI | All screens in spec §11 work with real engine |
| M7 | Packaging | AppImage + .deb + .rpm | Launch on clean VMs |
| M8 | CI/CD + Docker lab | Full matrix green, release automation | Tagged release publishes artifacts |
| M9 | Polish (P1) | rclone, Flatpak, GNOME/KDE settings, secret scan | Beta feedback addressed |

### C2. Daily loop
```
branch → code → unit tests → docker integration test (make test-matrix) → PR → CI green → review → merge
```

### C3. Make targets
```
make build        # engine + ui
make test         # unit tests
make lab          # build docker images for all distros
make test-matrix  # old→new OS integration tests
make package      # AppImage, deb, rpm
make lint
```

### C4. Branching and versioning
Trunk-based with short-lived feature branches. Conventional Commits. SemVer. Changelog generated from commits.

---

## D. CI/CD workflow (GitHub Actions)

```
push / PR
  ├─ lint (golangci-lint, eslint, prettier)
  ├─ unit tests (Go, UI)
  ├─ build engine
  ├─ docker integration matrix (ubuntu, debian, fedora, arch, opensuse)
  │     seed fake home → backup (old) → restore (new) → assert checksums/apps/rebuilds
  ├─ ansible check on exported playbook
  └─ security scan (govulncheck, npm audit, cargo audit)

tag v*
  ├─ build engine (x86_64, aarch64)
  ├─ build Tauri bundles: AppImage, .deb, .rpm
  ├─ build Flatpak
  ├─ sign + SHA256SUMS
  └─ GitHub Release (notes from changelog)
```

### D1. Docker integration test detail
1. Container "old": Ubuntu image with seeded `/home/test` (docs, `.bashrc`, `.gitconfig`, Node project with `node_modules`, Python project with venv, installed apt packages).
2. Run `carryover scan && carryover backup --dest /vol/repo --password-file pw`.
3. Container "new": Fedora (or Arch, etc.) mounting the same `/vol`.
4. Run `carryover restore --dest /vol/repo --yes --json`.
5. Assertions: sha256 equality for files, dotfile merge markers, packages present (`rpm -q`/`pacman -Q`), `node_modules` exists after rebuild, `report.json.status` in {success, partial}, no unexpected failures.

---

## E. Release checklist
- [ ] All acceptance criteria in `spec.md §18` pass
- [ ] Matrix green on 5 distros
- [ ] AppImage tested on clean VMs (GNOME + KDE)
- [ ] Artifacts signed, checksums published
- [ ] README with install instructions per distro
- [ ] Changelog and version tag
