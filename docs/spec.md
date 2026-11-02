# SPEC: Carryover Technical Specification

**Version:** 1.0  Companion docs: `prd.md`, `workflow.md`, `master_prompt.md`

---

## 1. Architecture

```
┌──────────────────────────────┐
│  UI (Tauri + Svelte/React)   │   thin: buttons, wizard, progress
└──────────────┬───────────────┘
               │ JSON over stdout/stdin (events) + CLI args
┌──────────────▼───────────────┐
│  Engine (Go, single binary)  │   all logic lives here
│  scanner | planner | backup  │
│  mapper  | restore | report  │
└───┬──────────┬──────────┬────┘
    │          │          │
  restic     rclone    native pkg managers
 (bundled)  (optional)  (apt/dnf/pacman/...)
```

**Rule:** the engine is fully usable without the UI (`carryover scan`, `backup`, `restore`). The UI only spawns the engine and renders its JSON events.

## 2. Tech stack
| Layer | Choice | Why |
|---|---|---|
| Engine | Go 1.22+ | Static single binary, easy cross-compile, good OS/process libs |
| Backup core | restic (bundled, pinned version) | Encryption, dedup, incremental, resumable |
| Remote targets | rclone (optional, P1) | Any cloud via restic's rclone backend |
| UI | Tauri 2 + Svelte + TypeScript + Tailwind | Small footprint, native feel |
| Manifest | YAML (`manifest.yaml`), JSON Schema validated | Human readable |
| Package maps | YAML data files in `engine/mapper/data/` | Community-editable |
| Packaging | AppImage, .deb, .rpm, Flatpak | Runs on any distro |
| CI/CD | GitHub Actions | Build, test, release |
| Test lab | Docker | Ubuntu, Debian, Fedora, Arch, openSUSE containers |
| Provisioning | Ansible (integration tests + optional playbook export) | Idempotent installs |

## 2.1 DevOps tooling (3 tools, each with one job)
1. **GitHub Actions:** lint, unit tests, matrix integration tests, build artifacts, publish GitHub Release.
2. **Docker:** per-distro containers simulating "old OS" and "new OS" to test backup then restore end to end.
3. **Ansible:** consumes generated playbook in integration tests and offers "Export as Ansible playbook" in the app.

---

## 3. Repository layout
```
carryover/
├── engine/
│   ├── cmd/carryover/main.go
│   ├── internal/
│   │   ├── scanner/        # os, apps, folders, dotfiles, projects, secrets
│   │   ├── manifest/       # structs, schema, read/write, version migration
│   │   ├── backup/         # restic wrapper, excludes, progress parsing
│   │   ├── restore/        # file placement, dotfile merge, conflict handling
│   │   ├── mapper/         # package name mapping + data/*.yaml
│   │   ├── installer/      # per package-manager executors
│   │   ├── rebuild/        # project rebuild allowlist + runner
│   │   ├── privilege/      # pkexec/sudo handling
│   │   ├── events/         # JSON event stream
│   │   └── report/         # final report generation
│   └── go.mod
├── ui/                      # Tauri app
│   ├── src-tauri/
│   └── src/ (routes, components, stores, styles)
├── ansible/                 # playbook templates + test playbooks
├── docker/                  # Dockerfile.ubuntu / .fedora / .arch / .debian / .opensuse
├── packaging/               # appimage/, deb/, rpm/, flatpak/
├── schemas/manifest.schema.json
├── tests/ (unit, integration, fixtures)
├── .github/workflows/ (ci.yml, release.yml)
├── docs/ (prd.md, spec.md, workflow.md)
└── README.md, LICENSE (MIT or Apache-2.0)
```

---

## 4. CLI contract (engine)
```
carryover scan      [--json] [--out scan.json]
carryover backup    --dest <path|rclone:remote:path> [--select selection.yaml] [--password-file f] [--dry-run] [--json]
carryover list      --dest <...>                      # list snapshots
carryover restore   --dest <...> [--snapshot id|latest] [--plan-only] [--yes] [--only files,dotfiles,apps,projects] [--json]
carryover verify    --dest <...>                      # restic check + manifest checksum
carryover map       <pkg> --from apt --to pacman      # debug mapping
carryover export    --format ansible --out playbook.yml
carryover doctor                                      # check restic, privileges, disk space
```
Exit codes: `0` ok, `1` general error, `2` partial success (warnings), `3` auth/password error, `4` destination error.

### 4.1 Event stream (for UI)
Newline-delimited JSON on stdout when `--json`:
```json
{"type":"phase","name":"scan","status":"start"}
{"type":"progress","phase":"backup","percent":42,"bytes_done":123456,"bytes_total":293847,"current":"~/Documents/x.pdf"}
{"type":"warning","code":"PKG_UNMAPPED","message":"No match for 'foo' on pacman","item":"foo"}
{"type":"need_privilege","reason":"Install packages","command_preview":"pacman -S --needed git curl"}
{"type":"done","status":"partial","report":"/path/report.json"}
```

---

## 5. Scanner

### 5.1 OS detection
Read `/etc/os-release` (`ID`, `ID_LIKE`, `VERSION_ID`). Map to package manager:
| ID family | Manager |
|---|---|
| debian, ubuntu, linuxmint, pop | apt |
| fedora, rhel, centos, rocky | dnf |
| arch, manjaro, endeavouros | pacman |
| opensuse*, sles | zypper |

Also detect: flatpak, snap, pip (user), pipx, npm global, cargo, gem (optional), go install binaries (optional).

### 5.2 App inventory (user-installed only)
- **apt:** `apt-mark showmanual`, minus packages in the base install (use `/var/log/installer/initial-status.gz` when present). Fall back to manual-marked list.
- **dnf:** `dnf repoquery --userinstalled`
- **pacman:** `pacman -Qqe` (explicit), optionally `-Qqm` for AUR (flag as "AUR")
- **zypper:** `zypper se -i` filtered by user-installed
- **flatpak:** `flatpak list --app --columns=application,origin`
- **snap:** `snap list` (skip core/base/snapd)
- **pip:** `pip list --user --format=json`, `pipx list --json`
- **npm:** `npm ls -g --depth=0 --json`
- **cargo:** `cargo install --list`
Each item records name, version, source, and whether it is a GUI app (desktop file present).

### 5.3 Folder discovery
Default candidates (XDG-aware, via `xdg-user-dirs`): Documents, Pictures, Desktop, Music, Videos, Downloads (unchecked by default), plus `~/Projects`, `~/dev`, `~/code`, `~/workspace`, `~/src`. Show size estimates (after exclusions) using a fast walker with goroutines.

### 5.4 Dotfiles and config (default allowlist)
```
~/.bashrc ~/.bash_profile ~/.bash_aliases ~/.profile ~/.zshrc ~/.zprofile ~/.zshenv
~/.config/fish/ ~/.gitconfig ~/.gitignore_global ~/.vimrc ~/.config/nvim/ ~/.tmux.conf
~/.ssh/ (config, known_hosts, keys: ENCRYPTED and flagged sensitive)
~/.gnupg/ (sensitive) ~/.config/Code/User/{settings.json,keybindings.json,snippets/}
~/.config/starship.toml ~/.config/alacritty ~/.config/kitty ~/.inputrc ~/.npmrc (sensitive) ~/.aws, ~/.kube (sensitive, opt-in)
~/.local/share/fonts ~/.fonts ~/.themes ~/.icons
```
Plus: VS Code extension list (`code --list-extensions`), `crontab -l`, `dconf dump /` (P1), `~/.local/share/applications` custom launchers.

Sensitive items are unchecked-by-default except SSH config; user opts in with a warning.

### 5.5 Shell configuration capture
- Backup the files themselves.
- Also record: `$SHELL`, exported env vars from rc files (parsed, not executed), PATH additions, aliases, functions list.
- Detect frameworks: oh-my-zsh, oh-my-bash, starship, nvm, pyenv, rbenv, sdkman, asdf, conda. Record install method and re-install in restore (these need re-installing; rc files reference them).
- On restore, warn when rc files source paths that do not exist yet; reinstall tooling first, then place rc files.

### 5.6 Project detection and rebuild recipes
Walk project roots (depth ≤ 4). Detect manifests and map to rebuild commands (allowlist only):

| Marker file | Excluded dirs | Rebuild command |
|---|---|---|
| `package.json` + `package-lock.json` | node_modules, .next, .nuxt, dist, .turbo, .cache | `npm ci` |
| + `yarn.lock` | same | `yarn install --frozen-lockfile` |
| + `pnpm-lock.yaml` | same | `pnpm install --frozen-lockfile` |
| `package.json` only | same | `npm install` |
| `pubspec.yaml` | .dart_tool, build, .pub-cache | `flutter pub get` (or `dart pub get`) |
| android project (`build.gradle*`) | .gradle, build, app/build, .idea (opt) | `./gradlew build` (optional, default: only dependency sync `./gradlew --refresh-dependencies`) |
| `requirements.txt` | venv, .venv, env, __pycache__, .mypy_cache, .pytest_cache | `python -m venv .venv && .venv/bin/pip install -r requirements.txt` |
| `pyproject.toml` (poetry) | .venv, __pycache__ | `poetry install` |
| `Pipfile` | .venv | `pipenv install` |
| `Cargo.toml` | target | `cargo fetch` (build optional) |
| `go.mod` | vendor (if untracked) | `go mod download` |
| `composer.json` | vendor | `composer install` |
| `Gemfile` | vendor/bundle, .bundle | `bundle install` |
| `pom.xml` | target | `mvn -q dependency:resolve` |
| `Podfile` | Pods | skip on Linux (note only) |
| `*.csproj` | bin, obj | `dotnet restore` |
| `docker-compose.yml` | none | note: `docker compose pull` (optional) |

**Global default excludes:** `node_modules`, `.gradle`, `build`, `dist`, `target`, `venv`, `.venv`, `__pycache__`, `.pytest_cache`, `.mypy_cache`, `.next`, `.nuxt`, `.turbo`, `.dart_tool`, `Pods`, `.cache`, `.tox`, `*.pyc`, `*.log`, `.DS_Store`, `Thumbs.db`, `.Trash*`, `*.tmp`, browser caches, `~/.cache`, `~/.local/share/Trash`, `~/.npm/_cacache`.

Safety: do NOT exclude `build/` or `dist/` unless a recognized project marker sits in the same directory (avoid deleting real user data named "build").

If a project is a git repo: record remote URL and current branch; warn about uncommitted/unpushed work (P1).

### 5.7 Secret scanner (P1)
Check for `.env*`, `*.pem`, `id_rsa*`, `credentials.json`, `.npmrc` tokens, AWS keys by regex. Show a warning list, never print secret values.

---

## 6. Manifest (v1)
```yaml
schema_version: 1
backup_meta:
  id: 2026-11-02T10-15-00Z
  created: 2026-11-02T10:15:00Z
  carryover_version: 1.0.0
  source:
    hostname: laptop
    os_id: ubuntu
    os_version: "22.04"
    arch: x86_64
    package_manager: apt
    username: sam
    home: /home/sam
    shell: /bin/bash
    desktop: GNOME
folders:                          # paths relative to $HOME
  - path: Documents
    size_bytes: 1203948
  - path: Pictures
  - path: Projects
exclude:
  global: [node_modules, .gradle, .venv, __pycache__]
  per_project:
    Projects/web-app: [node_modules, .next]
dotfiles:
  - path: .bashrc
  - path: .gitconfig
  - path: .ssh/config
    sensitive: true
shell:
  default: /bin/bash
  frameworks: [nvm, starship]
  env_exports: {EDITOR: nvim}
apps:
  apt:       [{name: git}, {name: curl}, {name: build-essential}]
  snap:      [{name: code, classic: true}]
  flatpak:   [{id: org.telegram.desktop, origin: flathub}]
  npm_global:[typescript, nodemon]
  pip_user:  [black]
  cargo:     [ripgrep]
  vscode_extensions: [esbenp.prettier-vscode]
projects:
  - path: Projects/web-app
    type: node
    package_manager: npm
    rebuild: ["npm ci"]
    git: {remote: git@github.com:me/web-app.git, branch: main}
  - path: Projects/mobile
    type: flutter
    rebuild: ["flutter pub get"]
checksums:
  manifest_sha256: "..."
```
The manifest is stored both inside the restic snapshot and as a small unencrypted `manifest.public.json` containing only non-sensitive metadata (OS, date, snapshot id) so the restore screen can show info before the password is entered.

---

## 7. Package mapper

### 7.1 Data files
`engine/mapper/data/packages.yaml`:
```yaml
build-essential:
  apt: build-essential
  dnf: "@development-tools"
  pacman: base-devel
  zypper: "patterns-devel-base-devel_basis"
python3-pip:
  apt: python3-pip
  dnf: python3-pip
  pacman: python-pip
  zypper: python3-pip
git: {apt: git, dnf: git, pacman: git, zypper: git}
curl: {apt: curl, dnf: curl, pacman: curl, zypper: curl}
vlc: {apt: vlc, dnf: vlc, pacman: vlc, zypper: vlc}
libssl-dev: {apt: libssl-dev, dnf: openssl-devel, pacman: openssl, zypper: libopenssl-devel}
nodejs: {apt: nodejs, dnf: nodejs, pacman: nodejs, zypper: nodejs}
```
Seed with the top ~300 packages. Store canonical (neutral) names as keys.

### 7.2 Resolution order
1. Exact match in curated table
2. Same name exists in the target repo (`apt-cache policy`, `dnf info`, `pacman -Si`, `zypper info`)
3. Known suffix transforms (`-dev` → `-devel`, `lib*-dev`, `python3-x` → `python-x`)
4. Flatpak/Snap equivalent for GUI apps (same app ID or name)
5. Fuzzy search in target repo (shown as "suggested", never auto-installed)
6. Unmapped → listed in report with manual-action hint

Each result has confidence: `exact | repo-same | transformed | suggested | none`. Only `exact`, `repo-same`, `transformed` are auto-selected.

Sources preference: if app was Flatpak/Snap in old OS, prefer Flatpak on new OS (distro-independent).

---

## 8. Backup engine

- Wrapper around `restic`. Repository password from user, passed via `RESTIC_PASSWORD_FILE` (memfd/tmpfs, never logged, never on command line).
- Command: `restic backup --files-from list.txt --exclude-file excludes.txt --json --tag carryover`.
- Parse restic JSON progress into events.
- Progress, ETA, throttled to 4 events/second.
- Resumable: restic is natively resumable; keep a `state.json` in a temp dir.
- Verify at end: `restic check --read-data-subset=5%` and manifest hash.
- Metadata: preserve permissions, mtimes, symlinks; extended attributes where supported.
- Large file warning: files > 2 GB flagged in selection screen.
- Disk space check on destination before start.
- Destination types: local path/USB (P0), SFTP, S3-compatible, rclone remotes (P1).

---

## 9. Restore engine

### 9.1 Planning (always first)
Produces a `plan.json`:
- files to place (source snapshot path → target path under new `$HOME`)
- conflicts (target exists)
- apps with mapping + confidence
- projects with rebuild commands
- required privilege steps
- estimated time and download size

### 9.2 Path translation
Stored paths are `$HOME`-relative. On restore, resolve to the new user's `$HOME`. If the XDG dir name differs (e.g. localized "Documentos"), map via `xdg-user-dirs` keys.

### 9.3 Conflict policy
Default: **keep both**. Existing file moved to `~/.carryover-old/<timestamp>/<relpath>`, then restored file placed. Alternatives per run: skip, overwrite (with backup), ask per file (UI list).
Dotfiles `.bashrc`/`.zshrc`/`.profile`: offer **merge** mode that appends restored content below a marker block `# >>> carryover restored >>>` rather than replacing distro defaults.

### 9.4 Ordering (important)
1. Pre-flight: network, disk space, privilege check
2. Install base toolchain: `git`, `curl`, `build-essential` equivalent, Flatpak setup if needed
3. Install apps (grouped per manager, one privileged call each)
4. Install shell/dev tooling (nvm, pyenv, etc.) per recorded frameworks
5. Restore files
6. Restore dotfiles (after tooling exists)
7. Fix permissions (`~/.ssh` 700, keys 600)
8. Rebuild projects (sequential or limited parallelism, with per-project timeout)
9. Restore VS Code extensions, desktop settings
10. Report

### 9.5 Privilege handling
Never run the whole app as root. Batch privileged operations and call `pkexec` (fallback `sudo` in terminal) once per package manager. UI shows the command preview before elevation (`need_privilege` event).

### 9.6 Failure handling
Each step is independent. A failure marks the item `failed` in the report and continues. Retries: 2 for network steps. Idempotent: running restore twice is safe (skip identical files by checksum, `--needed` flags for installers).

### 9.7 Rebuild runner
- Runs only commands generated from the allowlist table (§5.6), never arbitrary strings from the manifest. When loading a manifest, re-derive commands from the detected `type` and ignore free-text commands unless the user enables "custom commands" and confirms.
- Working directory is the restored project path; timeout default 15 min; output streamed to log.

---

## 10. Report
`report.json` + human view in the UI. Sections: Files (restored/skipped/conflicted), Dotfiles, Apps (installed / already present / unmapped / failed), Projects (rebuilt / failed), Warnings, Next manual steps (e.g. "Log in to Google Drive", "Re-add GPG key trust", "Reinstall proprietary drivers"). Exportable as Markdown.

---

## 11. UI specification

### 11.1 Design system
- **Style:** minimal, calm, generous whitespace, single accent color.
- **Type:** Inter (bundled), 14/16 px base; mono for paths.
- **Color tokens:** `--bg`, `--surface`, `--text`, `--muted`, `--accent` (#4F7CFF), `--success`, `--warn`, `--danger`; automatic light/dark via `prefers-color-scheme`.
- **Radius** 12 px, **spacing scale** 4/8/12/16/24/32, subtle borders over shadows.
- **Window:** 880×620 default, min 720×520, single-column content max-width 640px.
- **Motion:** 150ms ease, respect reduced-motion.
- **Components:** Button (primary/ghost), Card, Checkbox tree, Stepper, ProgressBar, Toast, Modal, LogDrawer, StatusPill, EmptyState.

### 11.2 Screens
1. **Home:** App name, two large cards: *Create Backup* and *Restore*. Footer: Settings, Help. If a backup is detected on a mounted drive, show a subtle banner "Backup found on USB. Restore?".
2. **Backup 1/4, Scan:** progress spinner with "Looking at your apps, folders and projects…". Auto-advances when complete (≤30s).
3. **Backup 2/4, Choose:** tabs: *Files · Apps · Settings · Projects*. Each tab: checkbox list with size or count, "Select all/none", search. A summary bar at the bottom: "4.2 GB (saved 11.8 GB by skipping rebuildable folders)". Sensitive items marked with a lock icon.
4. **Backup 3/4, Protect:** choose destination (folder picker/USB list/cloud later), set password (strength meter, confirm). Info text: "We can't recover this password."
5. **Backup 4/4, Run:** large progress bar, current item, ETA, collapsible log, Cancel (safe). Then **Done**: success state, summary, "Open backup folder", "Save recovery sheet" (printable PDF listing what was saved).
6. **Restore 1/4, Find:** pick destination or auto-detected backups; show public metadata (source OS, date, size).
7. **Restore 2/4, Unlock:** password.
8. **Restore 3/4, Preview:** three sections with counts (Files · Apps · Projects). Apps show mapping chips: ✓ matched, ⚠ suggested, ✕ not found. Expandable per item. Conflict policy dropdown. Migration score badge. Primary button "Start restore".
9. **Restore 4/4, Run → Report:** progress per phase (checklist with live status), privilege prompt explanation modal, final report with *Needs attention* section first.
10. **Settings:** theme, default destination, exclusion rules editor, package mapping overrides, advanced (custom rebuild commands, parallelism), About/Logs.

### 11.3 UX details
- Smart defaults: everything checked except Downloads and sensitive items.
- Never use jargon: "Rebuildable folders" with a tooltip "e.g. node_modules. We'll reinstall these for you."
- Inline errors with action buttons (Retry, Skip, Open log).
- Keyboard: Tab order, Enter = primary, Esc = back/cancel.
- First-run: a 3-line explanation, no multi-screen onboarding.
- Empty/error states designed for: no backup found, wrong password, destination full, drive unplugged, no network.
- Closing mid-run asks for confirmation.

---

## 12. Packaging and distribution
| Format | Tool | Notes |
|---|---|---|
| AppImage | `appimagetool` / Tauri bundler | Primary universal artifact; bundles WebKitGTK deps, restic, engine |
| .deb | Tauri bundler / `dpkg-deb` | Debian, Ubuntu, Mint, Pop!_OS |
| .rpm | Tauri bundler / `rpmbuild` | Fedora, openSUSE |
| Flatpak | `flatpak-builder` | Needs `--filesystem=home` and `--talk-name=org.freedesktop.Flatpak` (flatpak-spawn) permission to run host package managers; document limitations |
| Arch | PKGBUILD (AUR) | community |
| Install script | `curl … | sh` optional | verifies checksum/signature |

Sign release artifacts (minisign/GPG) and publish SHA256SUMS. Desktop entry + icon (256 px SVG/PNG) included.

---

## 13. Security
- Encryption by restic (AES-256, Poly1305). No password storage; no telemetry by default.
- Passwords handled in memory only, zeroized; never in logs, args or env of child processes beyond file descriptors.
- Sensitive paths flagged; restored with strict permissions.
- Manifest commands re-derived from allowlist (prevents malicious manifest execution).
- Privilege escalation only via polkit/sudo with shown command preview.
- Pin and verify bundled restic checksum; dependency scanning in CI (`govulncheck`, `npm audit`, `cargo audit`).
- Logs redact home paths optionally ("Share logs" mode).

## 14. Performance targets
Scan ≤30s (typical dev machine), UI idle memory ≤150 MB, backup throughput limited by disk (restic), engine startup ≤200 ms.

## 15. Testing strategy
- **Unit:** scanner parsers (fixtures for each package manager output), mapper, manifest, excludes.
- **Integration (Docker):** for each distro image: seed a fake home (documents, dotfiles, node project, python project), run `scan`, `backup` into a volume, then in a *different* distro container run `restore`, then assert: files byte-identical (sha256), dotfiles merged, packages installed, `node_modules` regenerated, report status.
- **Matrix:** old OS × new OS = {ubuntu, debian, fedora, arch, opensuse} (sample matrix in PRs, full on release).
- **Ansible check:** run exported playbook in `--check` and real mode in container.
- **UI:** Playwright against Tauri/web build with a mocked engine event stream; visual snapshots for the main screens.
- **Manual QA:** real VMs for GNOME and KDE.

## 16. CI/CD (GitHub Actions)
`ci.yml` (PR/push): lint (golangci-lint, eslint, prettier), `go test ./...`, UI tests, Docker integration matrix, security scans.
`release.yml` (tag `v*`): build engine for x86_64/aarch64, build Tauri bundles (AppImage/.deb/.rpm), build Flatpak, sign, checksums, create GitHub Release with changelog.
Caching: Go modules, cargo, npm. Use `docker buildx` for images.

## 17. Observability and logs
Logs to `~/.local/state/carryover/logs/` (rotating). Levels: info/warn/error/debug. "Copy diagnostics" button in Settings (redacted).

## 18. Acceptance criteria (MVP)
- [ ] Backup of a seeded home with `node_modules` produces snapshot excluding it and manifest lists rebuild command.
- [ ] Restore on a different distro container reproduces files byte-identically.
- [ ] ≥85% of seeded apps map automatically between Ubuntu, Fedora and Arch.
- [ ] `.bashrc` merge preserves new-OS defaults and adds restored content.
- [ ] Wrong password yields exit code 3 and friendly UI message.
- [ ] Interrupted backup resumes without re-uploading completed data.
- [ ] Dry run changes nothing on disk.
- [ ] AppImage launches on Ubuntu 22.04, Fedora 40 and Arch (clean installs).
- [ ] No privileged action happens without a visible prompt.

## 19. Open questions
1. Include proprietary driver/NVIDIA hints in the report?
2. Should P1 support incremental "keep in sync" mode rather than one-shot migration?
3. License choice (MIT vs Apache-2.0)?
