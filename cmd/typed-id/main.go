package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func run(args []string) error {
	fs := flag.NewFlagSet("typed-id", flag.ContinueOnError)

	var (
		typeFlag    = fs.String("type", "", "Comma-separated list of type names to generate code for (required)")
		presetsFlag = fs.String("presets", "all", "Comma-separated list of presets: text, json, binary, sql/db, all")
		outputFlag  = fs.String("output", "", "Output file name (default: <type_lower>.typed_id.go)")
		packageFlag = fs.String("package", "", "Override package name for generated code")
		dirFlag     = fs.String("dir", ".", "Directory of the target Go package")
	)

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of typed-id:\n")
		fmt.Fprintf(
			fs.Output(),
			"  typed-id -type=TypeName [-presets=all] [-output=filename] [-package=pkg] [-dir=.]\n\n",
		)
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*typeFlag) == "" {
		fs.Usage()
		return fmt.Errorf("missing required -type flag")
	}

	var rawTypeNames []string
	for t := range strings.SplitSeq(*typeFlag, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			rawTypeNames = append(rawTypeNames, t)
		}
	}

	if len(rawTypeNames) == 0 {
		return fmt.Errorf("no valid type names specified in -type")
	}

	var rawPresets []Preset
	for p := range strings.SplitSeq(*presetsFlag, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			rawPresets = append(rawPresets, Preset(p))
		}
	}

	presets, err := NormalizePresets(rawPresets)
	if err != nil {
		return err
	}

	pkgInfo, types, err := InspectTypes(*dirFlag, rawTypeNames)
	if err != nil {
		return fmt.Errorf("inspection failed: %w", err)
	}

	pkgName := *packageFlag
	if pkgName == "" {
		pkgName = pkgInfo.PackageName
	}

	code, err := Generate(GenerateOptions{
		PackageName: pkgName,
		Types:       types,
		Presets:     presets,
	})
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	var outputPath string
	if *outputFlag != "" && filepath.IsAbs(*outputFlag) {
		outputPath = *outputFlag
	} else if *outputFlag != "" {
		outputPath = filepath.Join(*dirFlag, *outputFlag)
	} else {
		var filename string
		if len(types) == 1 {
			filename = fmt.Sprintf("%s.typed_id.go", strings.ToLower(types[0].Name))
		} else {
			filename = "typed_id_gen.go"
		}
		outputPath = filepath.Join(*dirFlag, filename)
	}

	if err = os.Remove(outputPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove output file %q: %w", outputPath, err)
	}
	if err = os.WriteFile(outputPath, code, 0o644); err != nil {
		return fmt.Errorf("failed to write output file %q: %w", outputPath, err)
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "typed-id error: %v\n", err)
		os.Exit(1)
	}
}
