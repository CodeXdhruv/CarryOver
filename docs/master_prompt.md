# MASTER PROMPT: Build "Carryover", a Linux OS Migration Application

> Paste this whole file into your AI coding agent (Claude Code, Cursor, etc.). Put `prd.md`, `spec.md` and `workflow.md` in `/docs` first and tell the agent to treat them as the source of truth. Build in the milestone order given below. Do not skip ahead.

---

## 0. ROLE
You are a senior staff engineer and product designer. You will build **Carryover**, a production-quality, minimal desktop application for Linux that helps users switch operating systems without losing anything. You write clean, tested, documented code and keep the solution as simple as the requirements allow.

## 1. PRODUCT IN ONE PARAGRAPH
Carryover backs up a user's files, dotfiles and shell configuration, the list of installed applications, and "rebuild recipes" for developer projects (it skips `node_modules`, `.gradle`, `venv`, `build`, etc.). The backup is encrypted and stored on a USB/drive/folder (cloud later). On a new Linux distro, the user installs Carryover, selects the backup, previews what will happen, and Carryover places files in the correct folders, merges shell config, reinstalls apps using the **new distro's package manager** (with name mapping), and rebuilds projects. The app must be installable on any mainstream Linux distro.

## 2. SOURCE OF TRUTH
Read `docs/prd.md` (what and why), `docs/spec.md` (how), `docs/workflow.md` (flows and milestones) fully before writing code. If anything conflicts, `spec.md` wins for technical choices, `prd.md` for scope. If something is ambiguous, choose the simplest option, document the decision in `docs/DECISIONS.md`, and continue. Do not stop to ask unless blocked.

## 3. NON-NEGOTIABLE REQUIREMENTS
1. **Runs on any Linux distro:** distributed as AppImage (primary), .deb, .rpm, Flatpak. A bare `.deb` alone is not enough.
2. **Engine-first architecture:** all logic in a Go single-binary engine with a CLI. The UI (Tauri + Svelte + TypeScript + Tailwind) is a thin layer that spawns the engine and renders its newline-delimited JSON events.
3. **Backup core = restic** (bundled, version-pinned, checksum-verified). Do not write your own encryption or dedup.
4. **Always preview before changing anything** (dry run / plan), and never overwrite without moving the old file to `~/.carryover-old/<timestamp>/`.
5. **Rebuildable folders excluded from backup** and a rebuild recipe stored instead (table in spec §5.6). Only exclude `build/` or `dist/` when a recognized project marker file is in the same directory.
6. **Rebuild commands come from an allowlist keyed by project type**, never executed as free text from the manifest.
7. **Cross-distro package mapping** (apt, dnf, pacman, zypper, flatpak, snap, pip, npm -g, cargo) with confidence levels; only high-confidence mappings are auto-selected.
8. **Shell configuration** (`.bashrc`, `.zshrc`, `.profile`, aliases, `.gitconfig`, etc.) is backed up and restored with a safe merge mode; shell frameworks (nvm, oh-my-zsh, starship, pyenv, sdkman, asdf, conda) are detected and re-installed **before** rc files are placed.
9. **Encrypted by default.** Sensitive items (SSH keys, GPG, tokens) flagged with a lock icon; passwords never stored, never logged, never passed on a command line.
10. **Privilege handling:** the app never runs as root. Batch privileged operations and use `pkexec` (fallback `sudo` in terminal) with a visible command preview.
11. **Idempotent and resumable:** running restore twice is safe; interrupted backups resume.
12. **Paths are `$HOME`-relative** so a different username on the new OS works.
13. **No telemetry.** Offline-capable except when installing packages or rebuilding projects.

## 4. TECH STACK (fixed; do not substitute without writing a decision record)
- **Engine:** Go 1.22+, cobra CLI, `gopkg.in/yaml.v3`, `jsonschema` validation
- **Backup:** restic (bundled); rclone optional (P1)
- **UI:** Tauri 2, Svelte, TypeScript, Tailwind CSS, Inter font bundled
- **Data:** YAML manifest (`schema_version: 1`), JSON Schema in `schemas/manifest.schema.json`, package map YAML in `engine/internal/mapper/data/`
- **DevOps tools (3):** GitHub Actions (CI/CD), Docker (distro test lab), Ansible (integration testing and optional playbook export)
- **Packaging:** Tauri bundler for AppImage/.deb/.rpm, `flatpak-builder` for Flatpak, PKGBUILD for Arch

## 5. REPOSITORY LAYOUT (create exactly)
```
carryover/
├── engine/ (cmd/carryover, internal/{scanner,manifest,backup,restore,mapper,installer,rebuild,privilege,events,report})
├── ui/ (src-tauri, src)
├── ansible/
├── docker/ (Dockerfile.ubuntu .debian .fedora .arch .opensuse)
├── packaging/ (appimage, deb, rpm, flatpak, arch)
├── schemas/manifest.schema.json
├── tests/ (unit, integration, fixtures)
├── .github/workflows/ (ci.yml, release.yml)
├── docs/ (prd.md, spec.md, workflow.md, DECISIONS.md)
├── Makefile, README.md, LICENSE, CHANGELOG.md
```

## 6. FEATURE DETAIL TO IMPLEMENT

### 6.1 Scanner
- Detect OS from `/etc/os-release` and map to package manager (debian/ubuntu/mint/pop → apt; fedora/rhel → dnf; arch/manjaro → pacman; opensuse → zypper).
- List **user-installed** apps per manager: `apt-mark showmanual` (minus base install), `dnf repoquery --userinstalled`, `pacman -Qqe`, `flatpak list --app`, `snap list` (skip core/base), `pip list --user`, `pipx list`, `npm ls -g --depth=0 --json`, `cargo install --list`, VS Code extensions.
- Find folders via XDG user dirs + common dev roots (`Projects`, `dev`, `code`, `workspace`, `src`) with size estimates after exclusions.
- Find dotfiles from the allowlist in spec §5.4; parse (do not execute) rc files to record env exports, aliases, PATH changes, frameworks.
- Detect projects (depth ≤ 4) via marker files and record rebuild commands + git remote/branch.
- (P1) secret scan and git dirty/unpushed warnings.

### 6.2 Backup
- Build `manifest.yaml`, file list and exclude list. Run `restic backup --json`, parse progress into events. Verify with `restic check --read-data-subset=5%`. Write `manifest.public.json` (non-sensitive metadata) next to the repo.
- Pre-checks: destination writable, disk space, large-file warnings.

### 6.3 Restore
- Plan → preview → execute in the order defined in spec §9.4. Keep-both conflict policy by default; `.bashrc`/`.zshrc`/`.profile` merge mode using marker block `# >>> carryover restored >>>`.
- Fix permissions: `~/.ssh` 700, private keys 600.
- Each step independent: log failure, continue, retry network steps twice. Produce `report.json` with a "needs attention" list and manual next steps.

### 6.4 Package mapper
- Seed `packages.yaml` with the ~300 most common packages (build tools, languages, libs `-dev`/`-devel`, editors, browsers, media apps, docker, git, etc.).
- Resolution order: curated → same name in target repo → suffix transforms → Flatpak/Snap equivalent → fuzzy suggestion → unmapped. Users can override mappings in Settings (stored in `~/.config/carryover/mappings.yaml`).

### 6.5 Export
- `carryover export --format ansible` generates an idempotent playbook for the app installs and dotfile placement.

## 7. CLI CONTRACT
```
carryover scan|backup|list|restore|verify|map|export|doctor
```
with flags and exit codes exactly as specified in spec §4. All commands support `--json` event output.

## 8. UI REQUIREMENTS (minimal, great UX)
**Design principles:** one primary action per screen; calm, spacious; single accent color (#4F7CFF); light/dark automatic; Inter font; 12px radius; subtle borders; 150 ms motion (respect reduced motion); window 880×620 (min 720×520); content column max 640px.

**Screens (build all):**
1. Home: two large cards (Create Backup, Restore), Settings, Help; banner if a backup is detected on a mounted drive.
2. Backup: Scan → Choose (tabs Files/Apps/Settings/Projects, checkboxes, search, size summary with "saved X GB by skipping rebuildable folders") → Protect (destination + password with strength meter, "we can't recover this password") → Run (progress, ETA, collapsible log, safe cancel) → Done (summary, open folder, printable recovery sheet).
3. Restore: Find backup (shows public metadata) → Unlock → Preview (Files · Apps · Projects, mapping chips ✓ ⚠ ✕, conflict policy, migration score) → Run (phase checklist, privilege explanation modal) → Report ("Needs attention" first, exportable as Markdown).
4. Settings: theme, default destination, exclusion rules editor, mapping overrides, advanced (custom rebuild commands off by default, parallelism), logs/diagnostics.

**UX rules:** smart defaults (everything checked except Downloads and sensitive items); plain language ("Rebuildable folders: we'll reinstall these for you"); inline errors with Retry/Skip/Open log; full keyboard navigation (Tab, Enter, Esc); designed states for wrong password, full disk, unplugged drive, no network, no backup found; confirm before closing during a run.

## 9. SECURITY CHECKLIST
Encrypted by restic; passwords via file descriptor/tmpfs only and zeroized; no secrets in logs (redact values and optionally home paths); manifest commands re-derived from allowlist; verify bundled binary checksums; sign release artifacts and publish SHA256SUMS; run `govulncheck`, `npm audit`, `cargo audit` in CI.

## 10. TESTING REQUIREMENTS
- Unit tests for parsers (fixture outputs for every package manager), mapper, manifest, excludes, conflict handling.
- **Docker integration matrix:** seed a fake home in an "old" container (docs, dotfiles, Node project with `node_modules`, Python project with venv, installed packages), run scan and backup, then restore in a *different distro* container and assert: sha256 equality, dotfile merge markers, packages installed, `node_modules` rebuilt, report status. Distros: ubuntu, debian, fedora, arch, opensuse.
- UI tests with Playwright against a mocked engine event stream.
- Acceptance criteria in spec §18 must all pass.

## 11. CI/CD
- `ci.yml`: lint, unit tests, build, docker matrix, security scans, on every PR/push.
- `release.yml` on tag `v*`: build engine (x86_64, aarch64), Tauri bundles (AppImage, .deb, .rpm), Flatpak, sign, checksums, GitHub Release.
- Provide a `Makefile` with: `build test lab test-matrix package lint`.

## 12. BUILD ORDER (strict; commit after each milestone with tests passing)
1. M1: repo scaffold, manifest schema, scanner CLI (`scan --json`)
2. M2: backup with restic + excludes + events
3. M3: restore files and dotfiles with conflict policy and merge
4. M4: package mapper + installers + privilege handling
5. M5: rebuild runner (allowlist) + shell framework reinstall
6. M6: Tauri UI over the real CLI (all screens)
7. M7: packaging (AppImage, .deb, .rpm, Flatpak)
8. M8: Docker lab + GitHub Actions + release pipeline
9. M9: P1 polish (rclone remotes, GNOME/KDE settings via `dconf`, secret scan, git warnings, Ansible export, migration score)

After each milestone: run tests, update `CHANGELOG.md`, and print a 5-line status (done, tested, decisions, risks, next).

## 13. CODE QUALITY RULES
- Small packages with clear interfaces; no global state; context-aware cancellation; structured logging.
- Every exported function documented; errors wrapped with context; no panics in library code.
- Keep dependencies minimal and pinned. No unnecessary abstractions.
- Follow `gofmt`, `golangci-lint`, ESLint + Prettier. Conventional Commits.
- README must include: what it is, install instructions per distro, quick start with screenshots placeholders, CLI usage, FAQ, limitations, contributing, license.

## 14. DEFINITION OF DONE
All spec §18 acceptance criteria pass in CI; AppImage launches on clean Ubuntu 22.04, Fedora 40 and Arch VMs; docs are current; release artifacts are signed with checksums; no known data-loss bugs.

## 15. OUT OF SCOPE (do not build)
Windows/macOS support, disk imaging, bootloader/driver migration, a hosted cloud service, live database data backup.

## 16. FIRST ACTION
Start with Milestone M1 now. Create the repository scaffold, write `schemas/manifest.schema.json`, implement OS detection and the scanner for apt, dnf, pacman and flatpak first, add fixtures and unit tests, and show me `carryover scan --json` output. Then continue through the milestones in order.
