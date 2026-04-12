# Changelog

All notable changes to this repository are documented in this file.

## [5.0.0] - 2026-04-12

### Added

- Established V5 collaboration truth-source baseline (`AGENTS.md`, `PROJECT_FRAMEWORK.md`, `FRONTEND_GUIDELINE.md`, `backend/CLAUDE.md`, `docs/V5_REFACTOR_TASKS.md`).
- Added docs hub and index structure (`docs/README.md`, `docs/INDEX.md`, `docs/guides/*`) for current-phase navigation.
- Added stage audit reports under `docs/reports/` to support cleanup and structure verification.

### Changed

- Completed V5 folder optimization stage 1 and stage 2 baseline cleanup and structure audit.
- Root `README.md` now uses V5-first entry flow and includes quick-start instructions.
- Documentation cleanup process moved to task-tree node workflow with explicit inventory and deletion logs.

### Removed

- Removed outdated `CLEANUP-V1` front-end notes:
  - `docs/frontend-cleanup-p1ab-notes.md`
  - `docs/frontend-cleanup-p2a-notes.md`
  - `docs/frontend-cleanup-p2b-notes.md`
  - `docs/frontend-cleanup-p3b-notes.md`

## [5.0.1] - 2026-04-13

### Added

- Added stage-7 closure reports:
  - `docs/reports/node-7-1-build-validation.md`
  - `docs/reports/node-7-2-imports-deps-audit.md`
  - `docs/reports/node-7-3-doc-navigation-check.md`
  - `docs/reports/node-7-4-project-structure-report.md`
  - `docs/reports/node-7-5-change-log-and-commit-plan.md`
  - `docs/reports/node-7-6-optimization-summary.md`
- Added `PROJECT_STRUCTURE.md` as the post-optimization structure baseline.

### Changed

- Completed stage 7 verification closure for task `V5-FOLDER-OPTIMIZE` with build, import/dependency, and documentation-navigation checks.
- Updated `docs/V5_REFACTOR_TASKS.md` with stage-7 closure progress and remaining risk notes.

### Notes

- Front-end build and type-check passed.
- Back-end `go build ./cmd/server` and `go test ./internal/api/handlers -count=1` passed.
- Back-end full `go test ./...` still has existing failures in `navigation`/`permission`; `go mod verify` is affected by local module-cache contamination.
