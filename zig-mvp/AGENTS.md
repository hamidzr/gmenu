# Repository Guidelines

## Project Structure & Module Organization
- `src/main.zig` is the entry point; keep app logic close to the UI scaffolding until the module grows.
- `build.zig` and `build.zig.zon` define the Zig build configuration and dependencies.
- `justfile` provides shorthand commands for common workflows.
- Build artifacts land in `zig-out/` with cache data in `.zig-cache/`.
- `README.md` documents runtime expectations and platform requirements.

## Build, Test, and Development Commands
- `zig build`: compile the app with the settings in `build.zig`.
- `zig build run`: build and launch the macOS AppKit proof of concept.
- `just build`: alias for `zig build`.
- `just run`: alias for `zig build run`.
- `just clean`: remove `zig-out/`, `.zig-cache/`, and `bin/`.
- `just zip`: package the repo into `../zmenu.zip`.

## Coding Style & Naming Conventions
- Use Zig standard formatting (`zig fmt src/*.zig`) and 4-space indentation.
- Prefer `camelCase` for locals and functions, `PascalCase` for types, and `SCREAMING_SNAKE_CASE` for constants.
- Keep functions small and focused; avoid large monolithic `main` blocks as features expand.

## Testing Guidelines
- There is no dedicated test target yet.
- When adding tests, use Zig `test` blocks and run them with `zig test src/main.zig` or add a `test` step in `build.zig`.
- Prefer small, focused tests around text input handling and event flow.

## Commit & Pull Request Guidelines
- Commit history uses short, informal, lowercase summaries; follow that style and describe behavior changes plainly.
- PRs should include a brief summary, how to run/verify (`zig build run`), and any screenshots or recordings if the UI changes.

## Configuration & Runtime Tips
- This project targets macOS and AppKit; ensure Xcode Command Line Tools are installed.
- Use a Terminal to run `zig build run` if you need to see stdout output.
- The installed `zmenu` is symlinked to the build output, so rebuilding updates the installed binary.
