# PRD: Carryover (Linux OS Migration Tool)

**Version:** 1.0  **Status:** Draft  **Date:** 2026-11-02

## 1. Summary
Carryover is a minimal desktop application for Linux that lets a user back up everything that matters before switching operating systems, then restore it on a new Linux distro with almost no manual work. It restores personal files to the correct folders, reinstalls applications using the new distro's package manager, restores shell and dotfile configuration, and rebuilds developer projects (e.g. `npm install`) instead of backing up heavy, regenerable folders such as `node_modules`.

## 2. Problem
When people change distro (Ubuntu to Fedora, Mint to Arch, etc.) or reinstall:
- They manually copy files to a drive and back, and forget folders.
- They cannot remember which apps they had installed.
- Shell config (`.bashrc`, aliases, git config, SSH keys) is lost or painful to recreate.
- Developers waste hours and gigabytes copying `node_modules`, `build`, `.gradle`, `venv`, which could be rebuilt in minutes.
- Package names differ per distro, so a plain "list of apps" does not transfer.

## 3. Goals
1. One-click-style backup of files, configs, app inventory and project rebuild recipes.
2. Restore on any mainstream Linux distro with automatic package-name mapping.
3. Backups are small (regenerable folders excluded), encrypted, incremental and resumable.
4. Minimal, calm UI with a guided wizard. A non-technical user can finish it in under 5 minutes of interaction.
5. Installable on any Linux distro (AppImage, .deb, .rpm, Flatpak).
6. Always show a preview (dry run) before changing anything, and a clear report afterwards.

## 4. Non-goals (v1)
- Windows or macOS support.
- Full disk imaging or system cloning (this is not Clonezilla or Timeshift).
- Migrating `/etc` system-wide config, kernel modules, drivers or bootloader.
- Backing up databases' live data (only configs and dump hints).
- Cloud service of our own. Users bring their own storage.

## 5. Target users
| Persona | Needs |
|---|---|
| **Distro hopper** (non-technical) | Simple wizard, safe defaults, no terminal. |
| **Developer** | Rebuild recipes, dotfiles, SSH/git config, VS Code, global npm/pip/cargo packages, Docker context. |
| **Student / hobbyist** | Free, offline-capable, works from a USB drive. |

## 6. User stories
- As a user, I scan my system and see what will be backed up, so I stay in control.
- As a user, I untick folders or apps I do not want.
- As a developer, I want `node_modules` and similar folders skipped and rebuilt automatically after restore.
- As a user, I protect my backup with a password because it may contain SSH keys.
- As a user, on a new distro I point Carryover at my backup and see exactly what will happen before I confirm.
- As a user, I see which apps could not be found on my new distro and get alternatives or a manual-action list.
- As a user, I want a final report of what succeeded and what needs attention.
- As a user, I can resume an interrupted backup or restore.

## 7. Features and priority
### P0 (MVP)
- System scan (OS, package manager, apps, folders, dotfiles, projects)
- Selection screen with sensible defaults
- Backup via restic (encrypted, deduplicated, incremental) to local/USB path
- Default exclusion rules for regenerable folders
- Project rebuild recipe detection and replay
- App inventory: apt, dnf, pacman, zypper, flatpak, snap, pip (user), npm -g, cargo
- Cross-distro package name mapping (curated table plus fuzzy fallback)
- Dotfile and shell config backup/restore with safe merge and timestamped backup of existing files
- Restore preview (dry run), execution, final report
- Official packaging: AppImage and .deb

### P1
- Cloud/remote targets via rclone (S3, Google Drive, Backblaze, SFTP)
- Flatpak and .rpm packaging
- GNOME/KDE settings (`dconf dump`, KDE config), fonts, wallpapers
- Browser profiles (Firefox/Chrome bookmarks and extensions list)
- Pre-backup checks: uncommitted/unpushed git repos, secret scan for `.env`/tokens
- Migration score (e.g. "92% of apps found")
- Ansible playbook export

### P2
- Docker image/compose list, cron jobs, systemd user services, `/etc/hosts` entries
- Scheduled automatic backups
- Plugin system for custom scanners
- Localization

## 8. UX principles
1. **Minimal:** one primary action per screen, no clutter.
2. **Guided:** wizard with a visible step indicator (Scan, Choose, Protect, Run).
3. **Safe:** nothing destructive without preview and confirmation. Never overwrite without making a backup copy.
4. **Transparent:** live progress with human-readable log, expandable details.
5. **Forgiving:** every failure produces a plain-language explanation and a next step.
6. **Accessible:** keyboard navigable, readable contrast, system light/dark theme, scalable text.

## 9. Screens (summary)
1. Home (Create Backup / Restore / Settings)
2. Scan results and selection
3. Destination and password
4. Backup progress, then Done report
5. Restore: choose backup, unlock, preview, run, report
6. Settings (excludes, mappings, destination defaults, theme)

(See `spec.md` for full UI and technical details.)

## 10. Success metrics
- 90% of first-time users complete a backup without opening docs.
- Backup size at least 40% smaller than naive copy for typical developer home directories.
- At least 85% of apps auto-mapped on first restore across Ubuntu, Fedora and Arch.
- Zero data loss incidents: verified via checksum comparison in CI restore tests.
- Cold start of the app under 2 seconds. Scan under 30 seconds on a typical machine.

## 11. Constraints and assumptions
- Same username is not guaranteed. Paths are stored relative to `$HOME`.
- Package installation requires privilege elevation (`pkexec`/`sudo`), always requested explicitly.
- Network is needed on restore only to install packages and rebuild projects.
- Target architectures: x86_64 and aarch64.

## 12. Risks
| Risk | Mitigation |
|---|---|
| Package names differ across distros | Curated mapping table, fuzzy search, "unmapped" list, user-editable mappings |
| Secrets leaked in backup | Mandatory encryption, secret scanner, clear warnings |
| Restore overwrites user data | Dry run, existing files moved to `.carryover-old/<timestamp>/` |
| Rebuild commands fail (missing toolchain) | Order: install toolchains first, then rebuild; report failures without aborting |
| WebKitGTK differences across distros | Ship AppImage bundling runtime deps |
| Running arbitrary rebuild commands is risky | Only run from an allowlist of known commands per detected manifest file |

## 13. Release plan
Alpha (CLI engine) → Beta (GUI, AppImage/.deb) → 1.0 (Flatpak/.rpm, rclone, GNOME/KDE settings).
