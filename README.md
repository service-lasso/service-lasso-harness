# Service Lasso Harness

A Go validation runner for service artifacts. Service repositories own their validation contract; the harness owns shared execution. The current runner extracts an archive and executes its manifest command directly. It does not yet exercise the real Service Lasso runtime lifecycle or implement all declared health and evidence requirements.

For the reader journey, use Core's [Harness starter guide](https://github.com/service-lasso/service-lasso/blob/develop/docs/service-authoring/harness-starter.md) and [Validate and release](https://github.com/service-lasso/service-lasso/blob/develop/docs/service-authoring/05-validate-release.md). The current-source behavior and limitations below take precedence over the central guide's older planning-only description until Core #1418 reconciles it.

## Component contracts

- [Validation contract and current CLI behavior](docs/validation-contract.md)
- [Harness design requirements and open questions](docs/openspec-drafts/SPEC-SERVICE-LASSO-HARNESS.md)
- [Draft spec tracker](docs/openspec-drafts/OPENSPEC-TRACKER.md)
- [Example contract](examples/service-template/service-harness.json)

The draft spec retains the intended isolated Core lifecycle, dependency bootstrapping, role-based validation, evidence outputs and service-template relationship. These requirements are distinct from delivered runner behavior.

## Development and distribution

Use the Go version declared in `go.mod`. From this checkout:

```sh
go test ./...
go build ./cmd/service-lasso-harness
```

The CLI exposes `version`, `validate-contract --contract <path>` and `run --contract <path> --output-dir <directory>`.

The [release workflow](.github/workflows/release.yml) tests and builds Windows amd64, Linux amd64 and macOS amd64/arm64 binaries. It runs on matching version tags or manual dispatch; it is not a pull-request workflow. Consumer repositories should pin a released binary identity. A workflow definition does not establish that a particular release exists or that platform runtime acceptance passed.

Reader migration: Core #1418 / #1265, SPEC-002 AC-4AJ and AC-4AJ.3. The former generic `docs/usage-flow.md` journey is replaced by the central guides; schemas, CLI/build contracts and governing design requirements stay here.