# Validation Contract

`service-lasso-harness` consumes a small JSON contract, typically stored at `verify/service-harness.json` in a service repo.

For the bundled local harness example in this repo, the contract lives at `examples/service-template/service-harness.json`.

This starter implementation keeps the same basic shape as the current `service-template` example:

```json
{
  "serviceId": "echo-service",
  "artifact": {
    "path": "dist/echo-service-win32.zip",
    "kind": "archive"
  },
  "dependencies": [],
  "lifecycle": {
    "install": true,
    "config": true,
    "start": true,
    "stop": true
  },
  "health": {
    "type": "process",
    "timeoutSeconds": 30
  },
  "expect": {
    "logs": true,
    "state": true,
    "exitClean": true
  },
  "artifacts": {
    "captureLogs": true,
    "captureState": true,
    "captureSummary": true
  }
}
```

## Required fields

- `serviceId`: stable service identifier.
- `artifact.path`: artifact path, resolved relative to the contract file.
- `artifact.kind`: current starter values are free-form, but `archive` is the expected first-pass value.
- `lifecycle`: at least one lifecycle stage must be enabled.
- `health.timeoutSeconds`: must be greater than zero.

## Health type

- Default: `process`
- Allowed explicit values: `process`, `http`, `tcp`, `file`, `variable`

If `health.type` is omitted, the harness applies `process` before validation and run planning.

## Current CLI behavior

`validate-contract`:
- loads the JSON contract
- rejects malformed JSON, unknown fields, missing required fields, unsupported `health.type`, and zero/negative timeout values

`run`:
- performs the same contract validation
- resolves `artifact.path` relative to the contract location
- creates the requested output directory
- writes `run-result.json` and `summary.json`
- records intended stage status only

For the bundled local example contract, create the artifact first with:
- Windows: `pwsh -NoLogo -NoProfile -File .\scripts\package.ps1`
- POSIX shell: `./scripts/package.sh`

That packaging step builds `examples/service-template/dist/echo-service-win32.zip` from the tiny source payload under `examples/service-template/source/`.

This is a stub starter flow. It does not yet install artifacts or execute lifecycle steps.

## Result artifacts

`run` currently writes:

- `run-result.json`: detailed machine-readable status, validation errors, artifact resolution details, and stage planning
- `summary.json`: compact machine-readable summary for CI/logging

For the bundled example, the happy-path expectation is now:
- package the example first
- validate the contract
- run the stub flow
- confirm the result reports `artifact.exists=true`
