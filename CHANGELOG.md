# Changelog

All notable changes to Carryover will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

