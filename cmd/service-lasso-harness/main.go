package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/service-lasso/service-lasso-harness/internal/contract"
	"github.com/service-lasso/service-lasso-harness/internal/runflow"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage(os.Stderr)
		return 2
	}

	switch args[0] {
	case "version":
		fmt.Println(version)
		return 0
	case "validate-contract":
		return runValidateContract(args[1:])
	case "run":
		return runHarness(args[1:])
	case "-h", "--help", "help":
		printUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printUsage(os.Stderr)
		return 2
	}
}

func runValidateContract(args []string) int {
	fs := flag.NewFlagSet("validate-contract", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	contractPath := fs.String("contract", "", "Path to service-harness.json")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if strings.TrimSpace(*contractPath) == "" {
		fmt.Fprintln(os.Stderr, "validate-contract requires --contract")
		return 2
	}

	doc, err := contract.LoadFile(*contractPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "contract load failed: %v\n", err)
		return 1
	}

	if validationErrs := doc.Validate(); len(validationErrs) > 0 {
		for _, validationErr := range validationErrs {
			fmt.Fprintf(os.Stderr, "contract invalid: %s\n", validationErr)
		}
		return 1
	}

	resolvedArtifactPath := contract.ResolveArtifactPath(filepath.Clean(*contractPath), doc)
	fmt.Printf("contract valid: serviceId=%s artifact=%s health.type=%s\n", doc.ServiceID, resolvedArtifactPath, doc.Health.Type)
	return 0
}

func runHarness(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	contractPath := fs.String("contract", "", "Path to service-harness.json")
	outputDir := fs.String("output-dir", "", "Directory for result artifacts")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if strings.TrimSpace(*contractPath) == "" {
		fmt.Fprintln(os.Stderr, "run requires --contract")
		return 2
	}

	if strings.TrimSpace(*outputDir) == "" {
		fmt.Fprintln(os.Stderr, "run requires --output-dir")
		return 2
	}

	resultPath, summaryPath, err := runflow.Execute(*contractPath, *outputDir, version)
	if err != nil {
		var validationErr *runflow.ValidationFailure
		if errors.As(err, &validationErr) {
			fmt.Fprintf(os.Stderr, "run failed: %v\n", validationErr)
			fmt.Fprintf(os.Stderr, "result written to %s\nsummary written to %s\n", resultPath, summaryPath)
			return 1
		}

		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		if resultPath != "" || summaryPath != "" {
			fmt.Fprintf(os.Stderr, "result written to %s\nsummary written to %s\n", resultPath, summaryPath)
		}
		return 1
	}

	fmt.Printf("stub run complete: result=%s summary=%s\n", resultPath, summaryPath)
	return 0
}

func printUsage(stream *os.File) {
	fmt.Fprintln(stream, "service-lasso-harness")
	fmt.Fprintln(stream)
	fmt.Fprintln(stream, "Commands:")
	fmt.Fprintln(stream, "  version")
	fmt.Fprintln(stream, "  validate-contract --contract <path>")
	fmt.Fprintln(stream, "  run --contract <path> --output-dir <dir>")
}
