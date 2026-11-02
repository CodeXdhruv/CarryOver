# Carryover

> Minimal, production-quality Linux OS migration application. Switch operating systems without losing anything.

---

## What is Carryover?

Carryover backs up personal files, dotfiles, shell configuration, user-installed applications, and developer project rebuild recipes. It excludes regenerable folders (like `node_modules`, `venv`, `.gradle`, and `target`), encrypts data using pinned `restic`, and allows restoring onto any Linux distribution with automatic cross-distro package mapping.

## Key Features

- **Cross-Distro Package Mapping:** Automatically maps package names across `apt`, `dnf`, `pacman`, `zypper`, `flatpak`, and `snap`.
- **Developer Rebuild Recipes:** Skips gigabytes of regenerable dependency directories and automatically generates rebuild commands (`npm ci`, `cargo fetch`, `poetry install`, etc.).
- **Safe Dotfile Merging:** Timestamped backup copies of existing files into `~/.carryover-old/<timestamp>/`, with merge markers for `.bashrc`, `.zshrc`, and `.profile`.
- **Engine-First Architecture:** Standalone Go binary CLI with JSON event streaming + lightweight Tauri UI.
- **Zero Telemetry & Offline Capable:** Encrypted locally, no cloud dependencies required.

## Installation

### AppImage (Universal)
```bash
chmod +x carryover-x86_64.AppImage
./carryover-x86_64.AppImage
```

### Ubuntu / Debian (.deb)
```bash
sudo dpkg -i carryover_amd64.deb
```

### Fedora / RHEL (.rpm)
```bash
sudo rpm -i carryover_x86_64.rpm
```

### Arch Linux (AUR)
```bash
yay -S carryover-bin
```

---

## Quick Start (CLI)

### 1. Scan your system
```bash
# Scan and view human-readable summary
carryover scan

# Scan and stream JSON events
carryover scan --json

# Save scan manifest to JSON or YAML
carryover scan --out manifest.yaml
```

### 2. Create an encrypted backup
```bash
carryover backup --dest /media/usb/backup --password-file /path/to/key
```

### 3. Restore on new OS
```bash
# Dry run / preview
carryover restore --dest /media/usb/backup --plan-only

# Execute restore
carryover restore --dest /media/usb/backup
```

---

## Architecture

```
carryover/
├── engine/             # Go single-binary engine (CLI + business logic)
│   ├── cmd/carryover/  # CLI entrypoint
│   └── internal/       # scanner, manifest, backup, restore, mapper, events
├── ui/                 # Tauri 2 + Svelte desktop UI
├── schemas/            # JSON Schema for manifest (v1)
├── packaging/          # AppImage, deb, rpm, flatpak recipes
├── docker/             # Distro test lab containers
└── tests/              # Unit tests, integration matrix, fixtures
```

---

## FAQ

#### Does Carryover backup `/etc` or the whole OS?
No. Carryover is an application and personal configuration migration tool, not a full disk imaging tool like Timeshift or Clonezilla.

#### How are sensitive items handled?
SSH keys, GPG keys, and token files are flagged as sensitive, encrypted by default, and restored with strict file permissions (`0700` for `~/.ssh`, `0600` for keys).

---

## Contributing

1. Clone repository
2. Run `make build` and `make test`
3. Submit a pull request following Conventional Commits

## License

MIT © 2026 Carryover Authors
