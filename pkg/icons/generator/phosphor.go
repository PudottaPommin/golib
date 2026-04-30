package main

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"encoding/xml"
	"fmt"
	"iter"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type (
	phosphorIconMetadata struct {
		ID        int    `json:"id"`
		Code      *int   `json:"code,omitempty"`
		Name      string `json:"name,omitempty"`
		Published bool   `json:"published,omitempty"`
	}
	phosphorIcon struct {
		Metadata phosphorIconMetadata `json:"icon"`
		//Svgs     phosphorIconSvgs     `json:"svgs"`
		Svgs map[PhosphorStyle]string `json:"svgs"`
	}
	PhosphorStyle string
)

const (
	PhosphorVariationUnknown PhosphorStyle = "_all_"
	PhosphorVariationRegular PhosphorStyle = "regular"
	PhosphorVariationBold    PhosphorStyle = "bold"
	PhosphorVariationDuotone PhosphorStyle = "duotone"
	PhosphorVariationFill    PhosphorStyle = "fill"
	PhosphorVariationLight   PhosphorStyle = "light"
	PhosphorVariationThin    PhosphorStyle = "thin"
)

const (
	PhosphorCurrentVersion = "2.2"
	phosphorEndpoint       = "https://api.phosphoricons.com/v1"
	phosphorMaxChannels    = 256
)

type PhosphorHandler struct {
	Version string
	client  *http.Client
}

func NewPhosphorHandler(version string) *PhosphorHandler {
	return &PhosphorHandler{
		Version: version,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        phosphorMaxChannels,
				MaxIdleConnsPerHost: phosphorMaxChannels,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: 60 * time.Second,
		},
	}
}

func (ph *PhosphorHandler) buildOptimizedSVG(cfg *PhosphorJsonConfig, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(dir, "phosphor-icons.svg"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	filter := ph.reduceSymbols(cfg)
	if err = StreamFilterSVG(bytes.NewReader(phosphorIconsSVG), f, filter); err != nil {
		return fmt.Errorf("failed to generate final SVG: %w", err)
	}
	return nil
}

func (ph *PhosphorHandler) downloadPhosphorSVG(ctx context.Context, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(dir, "phosphor-icons-source.svg"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open SVG file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n"); err != nil {
		return err
	}
	if _, err := f.WriteString(`<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">` + "\n"); err != nil {
		return err
	}

	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")

	svg := NewSVG()
	// Write the root svg tag manually to avoid Go's xml encoder prepending _xmlns: prefixes
	if _, err := f.WriteString(fmt.Sprintf(`<svg xmlns="%s" version="%s" xmlns:xlink="%s" xmlns:xml="%s">`+"\n",
		svg.Xmlns, svg.Version, svg.XmlnsXlink, svg.XmlnsXml)); err != nil {
		return err
	}

	for symbol := range ph.downloadAllSymbols(ctx) {
		if err := enc.Encode(symbol); err != nil {
			return err
		}
	}

	if _, err := f.WriteString("\n</svg>\n"); err != nil {
		return err
	}

	return enc.Flush()
}

func (ph *PhosphorHandler) reduceSymbols(cfg *PhosphorJsonConfig) func(SVGSymbol) bool {
	hasAllowedStyle := func(style PhosphorStyle) bool {
		if style == PhosphorVariationRegular && len(cfg.Styles) == 0 {
			return true
		}
		for _, allowedStyle := range cfg.Styles {
			if style == allowedStyle {
				return true
			}
		}
		return false
	}

	sanitizeID := func(id string) (string, PhosphorStyle) {
		if id, ok := strings.CutSuffix(id, "-bold"); ok {
			return id, PhosphorVariationBold
		}
		if id, ok := strings.CutSuffix(id, "-duotone"); ok {
			return id, PhosphorVariationDuotone
		}
		if id, ok := strings.CutSuffix(id, "-fill"); ok {
			return id, PhosphorVariationFill
		}
		if id, ok := strings.CutSuffix(id, "-light"); ok {
			return id, PhosphorVariationLight
		}
		if id, ok := strings.CutSuffix(id, "-thin"); ok {
			return id, PhosphorVariationThin
		}
		return id, PhosphorVariationRegular
	}

	return func(symbol SVGSymbol) bool {
		id, style := sanitizeID(symbol.ID)
		v, ok := cfg.Icons[id]

		if !ok {
			log.Println("skipping icon with unknown ID:", id)
			return false
		}

		switch t := v.(type) {
		case bool:
			// if it's false, then icon is disabled
			if !t {
				log.Println("skipping icon with unknown ID:", id)
				return false
			}
			// if it's true, then we return based on allowed styles
			return hasAllowedStyle(style)
		case map[PhosphorStyle]bool:
			return t[style]
		default:
			return false
		}
	}
}

func (ph *PhosphorHandler) fetchMetadata(ctx context.Context) ([]*phosphorIconMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, phosphorEndpoint+"/icons?v=.."+ph.Version+"&status=Implemented", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := ph.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Count int                     `json:"count"`
		Icons []*phosphorIconMetadata `json:"icons"`
	}

	dec := jsontext.NewDecoder(resp.Body)
	if err = json.UnmarshalDecode(dec, &data); err != nil {
		return nil, err
	}
	return data.Icons, nil
}

func (ph *PhosphorHandler) downloadAllSymbols(ctx context.Context) iter.Seq[SVGSymbol] {
	return func(yield func(SVGSymbol) bool) {
		metadata, err := ph.fetchMetadata(ctx)
		if err != nil {
			log.Printf("failed to fetch Phosphor metadata: %v", err)
			return
		}

		type chanData struct {
			m *phosphorIconMetadata
		}
		mdChan := make(chan chanData, 256)
		go func() {
			defer close(mdChan)
			for _, md := range metadata {
				select {
				case <-ctx.Done():
					return
				case mdChan <- chanData{m: md}:
				}
			}
		}()

		results := make(chan SVGSymbol, phosphorMaxChannels)
		eg, ectx := errgroup.WithContext(ctx)
		eg.SetLimit(phosphorMaxChannels)

		go func() {
			for md := range mdChan {
				gmd := md
				eg.Go(func() error {
					icon, err := ph.fetchIcon(ectx, gmd.m.ID)
					if err != nil {
						return fmt.Errorf("failed to fetch phosphor icon for id %d: %w", gmd.m.ID, err)
					}

					for k, symbolSVG := range icon.Svgs {
						if symbolSVG == "" {
							continue
						}
						id := gmd.m.Name
						switch k {
						case PhosphorVariationRegular:
						default:
							id += "-" + string(k)
						}

						symbol, err := parseRawSVG(id, symbolSVG)
						if err != nil {
							log.Printf("failed to parse icon %s: %v", id, err)
							continue
						}

						select {
						case <-ectx.Done():
							return ectx.Err()
						case results <- symbol:
						}
					}
					return nil
				})
			}
			if err = eg.Wait(); err != nil {
				log.Printf("error during icon download: %v", err)
			}
			close(results)
		}()

		for symbol := range results {
			if !yield(symbol) {
				return
			}
		}
	}
}

func (ph *PhosphorHandler) fetchIcon(ctx context.Context, id int) (*phosphorIcon, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(phosphorEndpoint+"/icon/%d", id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := ph.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data phosphorIcon
	dec := jsontext.NewDecoder(resp.Body)
	if err = json.UnmarshalDecode(dec, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
