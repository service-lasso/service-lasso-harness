# Service Lasso Service Harness

_Status: starter repo bootstrapped; forward-looking design notes remain where the implementation is still stubbed._

## Proposed repo target

Current proposed repo target for this project:
- `service-lasso-harness`

This is the working name for the shared harness project that validates whether service artifacts actually work inside Service Lasso.

## Preferred implementation and distribution direction

Current preferred direction:
- implement v1 in **Go**
- ship it as a simple **platform binary** via GitHub releases
- let service repos download/use a released harness binary rather than requiring a local Node toolchain just to execute validation

Why this is preferred:
- easier cross-repo consumption
- easier CI usage
- easier version pinning
- cleaner release/download story for a shared utility

## Purpose

This project exists to provide a **shared validation harness** for Service Lasso services.

Its job is to prove that a service:
- packages correctly
- installs correctly into an isolated Lasso root
- configures correctly where required
- starts correctly under Service Lasso semantics
- reaches the expected health/readiness state
- stops / uninstalls / resets correctly for its role

The important rule is:

**a service passing its own local smoke test is not enough.**
The harness exists to prove the service works **inside Lasso**.

## Current starter implementation

This repo is now a real local starter, not just a planning shell.

This repo now includes a first-pass Go starter with:
- `service-lasso-harness version`
- `service-lasso-harness validate-contract --contract <path>`
- `service-lasso-harness run --contract <path> --output-dir <dir>`
- `scripts/package.ps1` and `scripts/package.sh` to create the local example archive expected by `examples/service-template/service-harness.json`

## Quick local example

Build the bundled example artifact:

```powershell
pwsh -NoLogo -NoProfile -File .\scripts\package.ps1
```

Validate the bundled example contract:

```powershell
go run ./cmd/service-lasso-harness validate-contract --contract examples/service-template/service-harness.json
```

Run the current stub flow and write result artifacts:

```powershell
go run ./cmd/service-lasso-harness run --contract examples/service-template/service-harness.json --output-dir output/example-run
```

Current `run` behavior is intentionally stubbed:
- validates the contract
- applies default `health.type=process` when omitted
- resolves the artifact path relative to the contract file
- writes machine-readable `run-result.json` and `summary.json`
- records intended lifecycle stage status without executing the real engine yet

Current local example packaging behavior:
- packages `examples/service-template/source/` into `examples/service-template/dist/echo-service-win32.zip`
- gives the starter contract a real local dist artifact to point at during validation and stub runs
- provides a minimal example payload only, meant to exercise contract resolution and local packaging rather than full lifecycle execution

---

## Why this should be a shared harness

Without a shared harness, every service repo will invent its own:
- install scripts
- temp-root setup
- health polling
- log capture
- CI wiring
- cleanup logic

That leads to drift and weak validation.

A shared harness keeps the ecosystem consistent.

It should give every service repo one standard way to say:

> here is the service artifact, here are the expected checks, now prove it works in Lasso.

---

## Main goals

The harness should:
- run locally and in CI
- validate real release artifacts, not only source trees
- create an isolated temporary Lasso root for each run
- install services into that temp root
- optionally install required test dependencies
- run install -> config -> start -> health -> stop -> uninstall/reset flows
- capture logs, state, and result artifacts for debugging
- return machine-readable pass/fail output

## Proposed initial project structure

Current recommended first-pass layout for `service-lasso-harness`:

```text
service-lasso-harness/
  README.md
  CHANGELOG.md
  LICENSE
  go.mod
  docs/
    validation-contract.md
    pipeline-usage.md
    role-profiles.md
  examples/
    service-template/
      service-harness.json
  schemas/
    service-harness.schema.json
  scripts/
    run.ps1
    run.sh
  cmd/
    service-lasso-harness/
  internal/
  fixtures/
    sample-artifacts/
    sample-services/
  test/
    integration/
    matrix/
  output/
    .gitkeep
```

### Structure intent
- `docs/` -> explain the contract and CI/local usage
- `examples/` -> copyable starting examples for service authors
- `schemas/` -> machine-readable validation contract schema
- `scripts/` -> stable entrypoints for local/CI runs
- `src/` or `cmd/` -> actual harness implementation
- `fixtures/` -> controlled sample inputs for harness development
- `test/` -> harness self-tests and integration scenarios
- `output/` -> local artifact dump area during development (ignored in real repo where appropriate)

---

## Validation layers

The intended validation model has 3 layers.

### 1. Repo-self validation
Runs in the service repo itself.

Examples:
- manifest/schema validation
- packaging validation
- local smoke tests
- release artifact shape checks

This is fast and should run on every PR.

### 2. Lasso harness integration validation
Runs the service through a real Service Lasso harness.

Examples:
- install into clean temp services root
- config materialization
- start under Lasso-managed env/dependencies/ports
- health/readiness assertions
- state/log assertions
- stop/uninstall/reset behavior

This is the main proof that the service actually works in Lasso.

### 3. Optional matrix/system validation
Only used when needed.

Examples:
- Windows/Linux/macOS matrix
- runtime-provider combinations
- infrastructure services with heavier startup
- routing/network behavior
- dependency graph scenarios

Not every service needs the same depth.

---

## Suggested harness flow

For a normal integration run, the harness should execute roughly this sequence:

1. Create isolated temp root
2. Materialize test config
3. Install service artifact
4. Install any declared test dependencies
5. Run `install`
6. Run `config` if required
7. Run `start`
8. Wait for readiness / health
9. Validate expected outputs
10. Run `stop`
11. Run `uninstall` / `reset` checks where relevant
12. Collect artifacts and emit final result

---

## Service validation contract

Each service repo should provide a **small validation contract**.

The service repo should not have to implement the whole harness.
It should only declare what the shared harness needs to know.

Initial likely shape:
- `scripts/verify.ps1`
- `scripts/verify.sh`
- `verify/service-harness.json`

Recommended first machine-readable validation contract path:
- `verify/service-harness.json`

Example responsibilities of the service-specific contract:
- identify the service id
- point to the built/released artifact
- declare required test dependencies
- declare required actions to run
- define health/readiness expectations
- define expected ports / URLs / exported env values
- define cleanup expectations

The shared harness should consume this contract and perform the actual orchestration.

---

## Role-based assertions

Different service roles need different harness checks.

### Runtime provider service
Examples:
- `@node`
- `@python`
- `@java`

Typical checks:
- runtime installs successfully
- expected binary is callable
- expected env is exported
- dependent sample service can run through it

### Infrastructure archive service
Examples:
- postgres
- traefik
- mongo

Typical checks:
- archive install works
- config materializes correctly
- port opens
- health succeeds
- stop/uninstall behavior is correct

### App service
Examples:
- node app
- python app
- API service

Typical checks:
- runtime dependency resolves
- app starts
- health endpoint or sample route responds
- expected logs/state are written

### Utility/bootstrap service
Examples:
- `@archive`
- `@localcert`

Typical checks:
- expected action completes
- expected outputs/artifacts are created
- exported paths/env are usable by dependent services

---

## Pipeline expectations

To work well in CI, the harness should be:
- headless
- deterministic
- temp-dir isolated
- port-collision safe
- timeout-bounded
- artifact-friendly for logs/state/debug dumps
- runnable against real release artifacts

Recommended pipeline shape:

### Fast PR pipeline
- lint / schema validation
- package build
- repo-self smoke checks
- one primary harness smoke run

### Release / heavier pipeline
- artifact build
- full harness integration run
- optional OS matrix
- dependency scenario checks where required
- publish only after harness passes

---

## Relationship to `service-template`

The `service-template` repo should include the first golden example of this harness pattern.

That template should prove:
- how a service declares its validation contract
- how local verification runs
- how CI runs the same validation flow
- what evidence a passing service produces

The harness project and the template repo should evolve together.

---

## Recommended first implementation slice

The best first slice for the harness is:

1. one sample service from `service-template`
2. one isolated temp-root harness run
3. one install -> start -> health -> stop proof
4. captured logs/state/artifacts
5. one CI workflow using the same path

That is enough to prove the model before broadening to heavier matrix scenarios.

---

## Early non-goals

This harness should **not** start as:
- a full UI end-to-end framework
- a huge generalized orchestration platform
- a replacement for all unit tests
- a substitute for service-local smoke tests

Its core job is narrower:

**prove that a service artifact works inside Service Lasso.**

---

## Open questions

- What exact machine-readable validation contract should the harness consume first?
- Should the first harness implementation start in `service-template` and then be extracted into `service-lasso-harness`, or should it begin in the dedicated repo immediately?
- Which artifacts should be mandatory from every run?
- How much dependency bootstrapping should the harness own vs the service contract declare?
- Which reset/uninstall checks should be universal versus role-specific?

---

## Bootstrap checklist for starting `service-lasso-harness`

Use this when the project is actually created.

### Phase 1 - create the repo and baseline files
- [ ] create repo `service-lasso/service-lasso-harness`
- [ ] use **Go** for v1 and commit to the single-binary runner path
- [ ] create baseline files:
  - [ ] `README.md`
  - [ ] `CHANGELOG.md`
  - [ ] `LICENSE`
  - [ ] `.gitignore`
  - [ ] CI workflow skeleton
- [ ] create starter folder structure:
  - [ ] `docs/`
  - [ ] `examples/`
  - [ ] `schemas/`
  - [ ] `scripts/`
  - [ ] `cmd/service-lasso-harness/`
  - [ ] `internal/`
  - [ ] `fixtures/`
  - [ ] `test/`

### Phase 2 - lock the first contract
- [ ] define the first machine-readable validation contract at `verify/service-harness.json`
- [ ] write `schemas/service-harness.schema.json`
- [ ] document the contract in `docs/validation-contract.md`
- [ ] define the first stable runner CLI/API surface
- [ ] define the first required output/artifact set (logs, state dump, result summary, timings, failure reason)

### Phase 3 - make the first runnable slice work
- [ ] add `scripts/run.ps1`
- [ ] add `scripts/run.sh`
- [ ] implement temp-root creation
- [ ] implement artifact install into isolated harness root
- [ ] implement `install -> start -> health -> stop` flow
- [ ] emit machine-readable result output
- [ ] archive logs/state/artifacts on both pass and fail

### Phase 4 - connect it to `service-template`
- [ ] add a sample service contract under `examples/service-template/`
- [ ] wire the harness against the `service-template` sample service
- [ ] prove one green local run
- [ ] prove one green CI run
- [ ] document the exact service-author usage flow in the README

### Phase 5 - harden for real ecosystem use
- [ ] add dependency-install support
- [ ] add `config` step support where required
- [ ] add uninstall/reset validation rules
- [ ] add role profiles (runtime provider / app / infrastructure / utility)
- [ ] add timeout handling and clearer failure diagnostics
- [ ] add optional OS matrix support
- [ ] add release packaging for downloadable GitHub release binaries
- [ ] document version pinning / binary acquisition for consumer repos

### First success bar
The first milestone should be considered successful when:
- one sample service artifact can be validated end-to-end under an isolated Lasso harness root
- the same flow works locally and in CI
- the run produces useful artifacts for debugging
- service authors can understand how to adopt the contract without reading donor-analysis docs first

## Working summary

The Service Lasso service harness should become the shared ecosystem proof layer that answers:

> does this service actually work in Lasso?

Not just:

> does this service look valid in its own repo?
