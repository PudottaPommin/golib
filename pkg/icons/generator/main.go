package main

import (
	"context"
	_ "embed"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"flag"
	"fmt"
	"log"
	"os"
)

type (
	IconsJsonConfig struct {
		Phosphor *PhosphorJsonConfig `json:"ph,omitempty"`
	}
	PhosphorJsonConfig struct {
		Styles []PhosphorStyle `json:"styles,omitempty"`
		Icons  map[string]any  `json:"icons,omitempty"`
	}
)

var (
	//go:embed icons/phosphor-icons-source.svg
	phosphorIconsSVG []byte
)

func main() {
	var (
		refetchIcons bool
		cfgPath      string
		dir          string
		generate     bool
	)
	flag.StringVar(&cfgPath, "config", "config.json", "Path to icons configuration file")
	flag.StringVar(&dir, "dir", "icons/", "Output path for final SVG")
	flag.BoolVar(&refetchIcons, "refetch-icons", false, "Fetches icons for embedding")
	flag.BoolVar(&generate, "generate", true, "Builds icons from cache")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if refetchIcons {
		phH := NewPhosphorHandler(PhosphorCurrentVersion)
		if err = phH.downloadPhosphorSVG(ctx, dir); err != nil {
			log.Fatalf("Failed to download Phosphor SVG: %v", err)
		}
		return
	}
	if generate {
		phH := NewPhosphorHandler(PhosphorCurrentVersion)
		if err = phH.buildOptimizedSVG(cfg.Phosphor, dir); err != nil {
			log.Fatalf("Failed to build optimized SVG: %v", err)
		}
		log.Println("Generated final SVG")
	}
}

func loadConfig(path string) (IconsJsonConfig, error) {
	if path == "" {
		return IconsJsonConfig{}, fmt.Errorf("configuration file path cannot be empty")
	}
	f, err := os.Open(path)
	if err != nil {
		return IconsJsonConfig{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg IconsJsonConfig
	dec := jsontext.NewDecoder(f)
	if err = json.UnmarshalDecode(dec, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
