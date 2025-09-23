# Repository Guidelines

## Project Structure & Module Organization
gmenu is a Go module with the entry point in `cmd/main.go`. Shared logic lives in `core/` for menu orchestration, with `internal/cli/` providing Cobra command wiring and flag parsing. UI components and theming sit in `render/`, while configuration schemas are in `internal/config/` and `model/`. Persistent state and cache helpers live under `store/`. Supporting assets and examples are stored in `samples/`, and automation scripts in `scripts/`. Build outputs land in `bin/`; remove them with `just clean`.

## Build, Test, and Development Commands
- `just setup` installs Go toolchain deps and tidies modules.
- `just build` compiles `bin/gmenu` for your host; use `just build-all` for Darwin cross-builds via `scripts/build.sh`.
- `just dev` streams a directory listing into the binary for quick manual testing.
- `just clean` removes build artifacts.

## Coding Style & Naming Conventions
Follow idiomatic Go: tabs for indentation, `PascalCase` for exported identifiers, `camelCase` for internals, and constants in `ALL_CAPS` only when already present in `constant/`. Run `just fmt` before committing; it applies `go fmt` and strips trailing spaces. Lint with `just lint` (golangci-lint) to enforce vet, staticcheck, and style suites. Keep files focused per package and prefer small, composable functions.

## Testing Guidelines
Use `just test` to execute `gotestsum --format=short -- -timeout=30s ./...`. Target specific areas with `go test ./core/...` or other package globs. Add table-driven tests for menu behaviours and cover configuration edge cases. GUI regressions should be verified with `just test-visual` when altering `render/`. Run `just check` before opening a PR to combine format, lint, and test passes.

## Commit & Pull Request Guidelines
Write commit subjects in the imperative mood (e.g., "Add fuzzy scorer fallback") and keep bodies focused on rationale or follow-up steps. Squash fixup noise locally. Pull requests should describe the user-facing effect, link related issues, list validation commands run, and include screenshots or recordings when `render/` changes affect the GUI. Flag migrations or config changes in the PR summary so downstream users can adjust early.

## Configuration & Runtime Tips
Use `gmenu.yaml.example` as the baseline for new settings. Local overrides belong in `~/.config/gmenu/config.yaml`; keep repository defaults minimal. Place repeatable input samples in `samples/` so automated or manual tests stay reproducible.
