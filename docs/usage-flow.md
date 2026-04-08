# Service Lasso Harness - Usage Flow

_Status: planning doc_

## Purpose

This doc explains how the Service Lasso harness is intended to be used in practice.

It now also has a small local starter path in this repo, so the contract and packaging flow can be exercised before the full harness engine exists.

The key idea is simple:

- a service repo should declare **how it expects to be validated**
- the shared harness should perform the **real isolated Lasso validation flow**
- the same flow should work **locally and in CI**

This is meant to prove that a service works **inside Service Lasso**, not just by itself.

---

## Core model

The usage model is split into two parts.

### Service repo responsibility
Each service repo provides:
- the service artifact/source
- the service manifest
- a small validation contract
- service-specific expected checks

### Harness repo responsibility
The harness repo provides:
- isolated temp-root setup
- Lasso-oriented install/config/start/stop orchestration
- dependency bootstrapping for validation
- health/readiness polling
- logs/state/artifact capture
- machine-readable pass/fail output
- a released downloadable runner binary for consumer repos/CI

This split prevents every service repo from inventing its own orchestration logic.

---

## Service repo side

A service repo is expected to contain something like:

```text
lasso-fastapi/
  service.json
  scripts/
    verify.ps1
    verify.sh
  verify/
    service-harness.json
```

### Intended meaning
- `scripts/verify.ps1` / `scripts/verify.sh`
  - stable local and CI entrypoints
  - should hand off to the shared harness rather than reimplement the whole validation engine
- `verify/service-harness.json`
  - machine-readable validation contract for the service

### What the service contract should describe
The contract should define things like:
- service id
- artifact location or artifact selection rule
- required dependency services for validation
- actions that must run (`install`, `config`, `start`, `stop`, etc.)
- health/readiness expectations
- expected ports / URLs / exported env values
- cleanup expectations
- role-specific checks

The service repo should declare expectations, not own the full harness behavior.

---

## Harness repo side

The harness repo is the executor.

Current planned location:
- `service-lasso-harness`

Current preferred implementation/distribution direction:
- implement the harness in **Go**
- release platform binaries via GitHub releases
- let service repos invoke the released binary rather than requiring a local Node toolchain

Current planned structure includes:

```text
service-lasso-harness/
  README.md
  docs/
  examples/
  schemas/
  scripts/
  src/ | cmd/
  fixtures/
  test/
```

### What the harness should do
Given a service validation contract, the harness should:

1. create an isolated temporary Lasso root
2. materialize any test config
3. install the service artifact
4. install any declared test dependencies
5. run `install`
6. run `config` if required
7. run `start`
8. wait for readiness / health
9. validate expected outputs
10. run `stop`
11. run uninstall/reset checks where relevant
12. collect logs/state/artifacts
13. emit a machine-readable result

This is the shared proof path for “works in Lasso”.

---

## Local developer flow

The local developer experience should be straightforward.

### Starter flow in this repo

Before the full harness engine is finished, this repo now supports a small local proof path:

1. build the bundled example artifact with `scripts/package.ps1` or `scripts/package.sh`
2. validate `examples/service-template/service-harness.json`
3. run the stub flow to emit `output/example-run/run-result.json` and `summary.json`
4. confirm the result sees the packaged artifact

This is not the final Lasso execution engine. It is the first coherent repo-local proof that packaging, contract resolution, and machine-readable result artifacts fit together.

### Consumer-repo example

Example:

```powershell
cd C:\projects\service-lasso\lasso-fastapi
.\scripts\verify.ps1
```

Conceptually that script should be able to call a released harness binary such as:

The bundled local example in this repo follows that same shape, but uses a tiny packaged sample payload so the starter flow can be exercised immediately.

```powershell
service-lasso-harness.exe run --contract .\verify\service-harness.json
```

Or on shell-based environments:

```bash
./scripts/verify.sh
```

### Intended local result
A local run should produce:
- pass/fail result
- useful logs
- state dump / relevant runtime outputs
- artifact locations
- clear failure reason when broken

The goal is to let a service author answer:

> does my service work inside Service Lasso right now?

before they push or publish.

---

## CI / pipeline flow

The same validation contract should run in CI.

### PR flow
Typical fast path:
- build artifact
- run repo-self smoke checks
- run one harness smoke validation

### Release flow
Typical heavier path:
- build release artifact
- acquire the pinned released harness binary for the target platform
- run full harness integration validation
- optionally run OS matrix / dependency scenarios
- publish only after harness passes

Current preferred release-tag convention for this repo:
- `vYYYY.M.D+<shortsha>`
- example: `v2026.4.9+3a3ae1b`

### Important rule
CI should not invent a different validation model.

The same contract and same harness runner should be usable:
- locally
- in PR validation
- in release validation

That keeps the ecosystem from drifting into different “truths” for local vs pipeline behavior.

---

## `service-template` as the first golden example

The first intended consumer of the harness is:
- `service-template`

The template should demonstrate:
- how a service declares its validation contract
- how local verification is run
- how CI runs the same validation flow
- what a passing service’s evidence looks like

This is important because future service authors should be able to copy a working pattern instead of reconstructing it from notes.

---

## Example end-to-end scenario

Consider a future service repo such as:
- `lasso-fastapi`

It might declare:
- dependency on `@python`
- release artifact path
- required actions: `install`, `config`, `start`, `stop`
- health contract: `GET /health`
- expected bound service URL
- uninstall behavior expectations

### Local run
A developer runs:

```powershell
cd C:\projects\service-lasso\lasso-fastapi
.\scripts\verify.ps1
```

### What happens under the hood
The shared harness then:
- creates temp Lasso root
- installs `@python` if needed
- installs `lasso-fastapi`
- runs install/config/start
- waits for health
- validates logs/state/output expectations
- runs stop/cleanup checks
- writes result artifacts

### CI run
CI runs the same service contract through the same shared harness path.

This means the service is being tested under real Lasso semantics rather than by ad hoc direct execution.

---

## Role-based usage

Different service roles use the same harness, but expect different assertions.

### Runtime provider service
Examples:
- `@node`
- `@python`
- `@java`

Likely checks:
- runtime installs successfully
- expected binary is callable
- expected env is exported
- dependent sample service can run through it

### Infrastructure archive service
Examples:
- postgres
- traefik
- mongo

Likely checks:
- archive install works
- config materializes correctly
- service port opens
- health succeeds
- stop/uninstall behavior is correct

### App service
Examples:
- node app
- python app
- API service

Likely checks:
- runtime dependency resolves
- app starts successfully
- health route responds
- expected logs/state are written

### Utility/bootstrap service
Examples:
- `@archive`
- `@localcert`

Likely checks:
- expected action completes
- expected outputs/artifacts are created
- exported paths/env are usable by dependent services

One harness model can support all of these if the service contract is explicit enough.

---

## Practical split of responsibility

### Service repo owns
- service code/artifact
- service manifest
- tiny validation contract
- service-specific expectations

### Harness repo owns
- isolated runtime setup
- orchestration engine
- dependency installation for validation
- health/readiness polling
- common artifact capture
- common result format
- CI-friendly runner behavior

That split is one of the core design decisions for the harness model.

---

## Why this approach was chosen

Without this split, every service repo would end up inventing its own:
- temp-root setup
- install logic
- health polling
- cleanup
- CI wiring
- debug artifact format

Then “validation” would mean something different in every repo.

The harness exists to make that consistent.

---

## Working summary

The Service Lasso harness is intended to be used like this:

1. service repo declares a small validation contract
2. shared harness runs the real isolated Lasso validation flow
3. the same path works locally and in CI
4. `service-template` provides the first golden example

The question the harness is meant to answer is:

> does this service actually work inside Service Lasso?

Not merely:

> does this service run by itself?
