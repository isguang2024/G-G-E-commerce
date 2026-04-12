# Changelog

All notable changes to this repository are documented in this file.

## [6.0.0] - 2026-04-13

### Changed

- Truth-source documents updated to reflect current project state as the baseline for business development.
- Truth sources reduced to four: `AGENTS.md`, `PROJECT_FRAMEWORK.md`, `FRONTEND_GUIDELINE.md`, `backend/CLAUDE.md`.
- All documentation rewritten to remove V5 refactoring context — project now described as stable architecture, not in-progress migration.

### Removed

- Removed `docs/V5_REFACTOR_TASKS.md` (V5 task tracker, no longer needed).
- Removed `docs/OPTIMIZATION_SUMMARY.md` (V5 optimization report).
- Removed `docs/reports/` directory (34 phase audit records from V5 refactoring).

## [5.0.1] - 2026-04-13

### Added

- Added `PROJECT_STRUCTURE.md` as the structure baseline.

### Changed

- Completed stage 7 verification closure with build, import/dependency, and documentation-navigation checks.

### Notes

- Front-end build and type-check passed.
- Back-end `go build ./cmd/server` and `go test ./internal/api/handlers -count=1` passed.

## [5.0.0] - 2026-04-12

### Added

- Established collaboration truth-source baseline (`AGENTS.md`, `PROJECT_FRAMEWORK.md`, `FRONTEND_GUIDELINE.md`, `backend/CLAUDE.md`).
- Added docs hub and index structure (`docs/README.md`, `docs/INDEX.md`, `docs/guides/*`) for navigation.

### Changed

- Completed folder optimization and structure audit.
- Root `README.md` updated with quick-start instructions.

### Removed

- Removed outdated front-end cleanup notes.
