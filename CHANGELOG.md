# Changelog

All notable changes to Carryover will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Milestone M6: Tauri Desktop UI & Design System
- Built Tauri 2 desktop application scaffold with Svelte, TypeScript, Vite, and Tailwind CSS.
- Implemented design system adhering to specification: minimal aesthetics, Inter typography, `#4F7CFF` accent token, 12px radius, auto light/dark theme, and 880×620 window constraints.
- Built `HomeScreen` featuring primary action cards, quick status, and automatic USB backup detection banner.
- Built `BackupScreen` 4-stage guided wizard: live scanning progress, multi-tab category selection (Files, Apps, Dotfiles, Projects) with skip savings indicator, password strength meter, and progress monitoring with verified completion card.
- Built `RestoreScreen` 4-stage guided wizard: public metadata inspection (`manifest.public.json`), repository password unlock, package mapping preview chips (✓ matched, ⚠ suggested, ✕ unmapped), migration score, conflict policy selection, elevation explanation modal, and phase checklist with live status ticks.
- Built `SettingsScreen` with theme toggle, default destination, exclusion editor, mapping overrides editor, and security toggle.
- Created `EngineBridge` streaming engine events over stdout to UI stores.

### Added - Milestone M5: Rebuild Runner (Allowlist) & Shell Framework Reinstall
- Created strict rebuild command allowlist (`engine/internal/rebuild/allowlist.go`) for Node, Python, Rust, Go, Flutter, Gradle, Maven, PHP, Ruby, .NET, and Docker Compose.
- Implemented command re-derivation from detected project types, rejecting arbitrary free-text manifest scripts unless custom commands are authorized.
- Implemented rebuild execution runner (`engine/internal/rebuild/runner.go`) with project working directories, timeouts, and live event progress streaming.
- Built shell framework automated installer (`engine/internal/rebuild/frameworks.go`) restoring `nvm`, `oh-my-zsh`, `oh-my-bash`, `pyenv`, `rbenv`, `asdf`, and `starship` before shell rc placement.
- Integrated project rebuilding and framework restoration directly into the restore pipeline.

### Added - Milestone M4: Package Mapper, Cross-Distro Installers & Privilege Escalation
- Created curated cross-distro package mapping catalog in `engine/internal/mapper/data/packages.yaml`.
- Implemented package resolution order: curated catalog -> naming transforms (`-dev`/`-devel`, `python3-`/`python-`) -> repo-same fallback with confidence scoring.
- Implemented user-customizable mapping overrides from `~/.config/carryover/mappings.yaml`.
- Created privilege elevation module (`engine/internal/privilege/privilege.go`) wrapping `pkexec` (and fallback `sudo`) with user command preview and event emissions.
- Created batched package installers (`engine/internal/installer/installer.go`) for `apt`, `dnf`, `pacman`, `zypper`, `flatpak`, `snap`, `npm`, `pip`, `cargo`, and `code`.
- Implemented CLI command `carryover map <pkg> --from <pm> --to <pm> [--json]`.

### Added - Milestone M3: Restore Engine, Conflict Management & Dotfile Merge
- Built dry-run restore planner (`engine/internal/restore/plan.go`) with conflict detection against target `$HOME`.
- Implemented `ConflictKeepBoth` policy: automatically moves existing target files to timestamped directory `~/.carryover-old/<timestamp>/<relpath>`.
- Implemented non-destructive dotfile merge mode for `.bashrc`, `.zshrc`, and `.profile` using marker block `# >>> carryover restored >>>`.
- Implemented file permission hardening during restore (`0700` on `~/.ssh`, `0600` on private keys).
- Implemented CLI `carryover restore` supporting `--snapshot`, `--plan-only`, `--only`, `--conflict`, and `--json`.
- Added end-to-end integration test (`tests/integration/backup_restore_test.go`) validating backup, restore, conflict movement, dotfile merge, and permissions.

### Added - Milestone M2: Restic Backup Core, Excludes & Events
- Bundled and integrated restic 0.17.0 wrapper (`engine/internal/backup/restic.go`).
- Implemented secure temporary password handling preventing command line and process table leakage.
- Added disk space check and low storage warning emission before backup.
- Built dynamic file-list and exclude-list builder with global and per-project rule synthesis.
- Implemented restic JSON progress parsing into throttled `progress` events (≤ 4/sec).
- Implemented unencrypted `manifest.public.json` export alongside backup repository.
- Built automatic repository verification (`restic check --read-data-subset=5%`) and manifest checksum verification.
- Implemented CLI commands: `carryover backup`, `carryover list`, and `carryover verify`.
- Added end-to-end integration test verifying backup creation, exclusions, public metadata, snapshot listing, and repository verification.

### Added - Milestone M1: Repository Scaffold & System Scanner
- Established repository scaffold adhering to specification.
- Defined JSON Schema v1 in `schemas/manifest.schema.json` for migration manifests.
- Implemented embedded JSON Schema compilation and validation in `engine/internal/manifest`.
- Created JSON event stream emitter in `engine/internal/events`.
- Implemented OS detection (`/etc/os-release`) supporting Ubuntu/Debian (`apt`), Fedora/RHEL (`dnf`), Arch/Manjaro (`pacman`), and openSUSE (`zypper`).
- Built user-installed application scanners for `apt`, `dnf`, `pacman`, `zypper`, `flatpak`, `snap`, `pip`, `npm`, `cargo`, and `code`.
- Built candidate folder discovery and exclusion-aware disk sizing.
- Built dotfile inventory allowlist with sensitivity detection (`.ssh`, `.gnupg`, `.npmrc`, `.aws`, `.kube`).
- Built shell parser and framework detector (`oh-my-zsh`, `starship`, `nvm`, `pyenv`, etc.).
- Built project detector with rebuild command derivation for Node, Rust, Go, Python, Gradle, Maven, PHP, Ruby, .NET, and Docker Compose.
- Implemented Cobra CLI entrypoint with `carryover scan [--json] [--out <path>]`.
- Added test fixtures and unit tests with 100% test pass rate.

