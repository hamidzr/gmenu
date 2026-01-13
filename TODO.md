# Tasks

## Go gmenu (Current Implementation) TODO

- [ ] if there is no empty config file and an id is given gmenu doesn't come up eg no config.yaml found
- [ ] test with races
- [x] remove os.exit; doesn't let defers run

## Done

- Fix auto-accept panic on single-match initial query (see `core/gmenu.go`, `core/gmenurun.go`)
