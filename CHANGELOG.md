# Changelog

All notable changes to Carryover will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
