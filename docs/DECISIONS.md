# Architecture Decision Records (ADR)

## ADR-001: Go Engine Single-Binary Architecture
- **Date:** 2026-11-02
- **Status:** Accepted
- **Context:** Carryover needs to run natively across all Linux distributions without heavy runtime dependencies.
- **Decision:** Implement all core logic (scanner, restic wrapper, package mapper, restore planner, rebuild runner) in a single static Go binary (`carryover`) exposing a clean CLI and newline-delimited JSON events.
- **Consequences:** UI layer (Tauri) remains very thin; engine is 100% testable and operable headless from terminal or CI.

## ADR-002: JSON Schema Manifest Validation
- **Date:** 2026-11-02
- **Status:** Accepted
- **Context:** Backup metadata and system scan state must be structured, reproducible, and verifiable.
- **Decision:** Embed `schemas/manifest.schema.json` inside the Go binary (`//go:embed schema.json`) and validate every generated and loaded manifest using `github.com/santhosh-tekuri/jsonschema/v5`.
- **Consequences:** Invalid manifests are rejected before any destructive or restore operations occur.

## ADR-003: Rebuildable Directory Exclusion
- **Date:** 2026-11-02
- **Status:** Accepted
- **Context:** Developers store gigabytes of regenerable dependency directories (`node_modules`, `venv`, `.gradle`, `target`).
- **Decision:** Exclude regenerable directories automatically during scan, calculate size savings, and generate strict rebuild recipes from allowlisted project markers.
- **Consequences:** Backups are 40-70% smaller and faster; rebuild commands are never arbitrary free text.
