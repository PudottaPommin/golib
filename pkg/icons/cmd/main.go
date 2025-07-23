package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
	"maps"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/sync/errgroup"
)

type (
	TemplateModel struct {
		*IconPackage
	}
)

var (
	//go:embed svg.gotmpl
	iconsTemplate string
	tmpl          = template.Must(template.New("").Parse(iconsTemplate))

	iconSets = map[string]string{
		"phosphor":     "ph.json",
		"hugeicons":    "hugeicons.json",
		"flagicons":    "flag.json",
		"circle-flags": "circle-flags.json",
	}

	goKeywords = []string{
		"break", "case", "chan", "const", "continue", "default", "defer",
		"else", "fallthrough", "for", "func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return", "select", "struct",
		"switch", "type", "var",
	}
)

func main() {
	g, ctx := errgroup.WithContext(context.Background())

	t := time.Now()
	for name, url := range iconSets {
		g.Go(func() error {
			return fetchIconset(ctx, name, url)
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated all icons in %dms", time.Since(t).Milliseconds())
}

func fetchIconset(ctx context.Context, name, url string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://raw.githubusercontent.com/iconify/icon-sets/refs/heads/master/json/%s", url), nil)
	if err != nil {
		return fmt.Errorf("[%s]: %w", name, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[%s]: %w", name, err)
	}
	defer resp.Body.Close()

	var collection iconCollectionInfo
	if err = json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return fmt.Errorf("[%s]: %w", name, err)
	}

	packageName := strings.ToLower(strcase.ToSnake(collection.Prefix))
	if slices.Contains(goKeywords, packageName) {
		packageName = packageName + "_icons"
	}

	iconPkg := IconPackage{
		Name:     newNameVariants(packageName),
		ViewBox:  make(map[string][2]int),
		Version:  collection.Info.Version,
		Width:    collection.Info.Height,
		Height:   collection.Info.Height,
		Variants: make(map[string]map[string]string),
	}

	if iconPkg.Width == 0 && iconPkg.Height > 0 {
		iconPkg.Width = iconPkg.Height
	} else if iconPkg.Height == 0 && iconPkg.Width > 0 {
		iconPkg.Height = iconPkg.Width
	} else if iconPkg.Width == 0 && iconPkg.Height == 0 {
		iconPkg.Width = 24
		iconPkg.Height = 24
	}

	iconPkg.Icons = make([]Icon, 0, len(collection.Icons))
	if collection.Suffixes == nil {
		iconPkg.Variants[""] = make(map[string]string)
		iconPkg.ViewBox[""] = [2]int{collection.Width, collection.Height}

		for k, v := range collection.Icons {
			iconPkg.Icons = append(iconPkg.Icons, Icon{
				Name:   newNameVariants(k),
				Body:   v.Body,
				Width:  v.Width,
				Height: v.Height,
			})
			iconPkg.Variants[""][k] = strcase.ToCamel(k)
		}
	} else {
		suffixes := slices.SortedFunc(maps.Keys(collection.Suffixes), func(l string, r string) int {
			if l == "" {
				return 1
			}
			if r == "" {
				return -1
			}
			return strings.Compare(l, r)
		})
		for _, k := range suffixes {
			iconPkg.Variants[k] = make(map[string]string)
		}

		for k, v := range collection.Icons {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var suffix string

			for _, ks := range suffixes {
				if strings.HasSuffix(k, ks) {
					suffix = ks
					break
				}
			}

			w, h := v.Width, v.Height
			if w == 0 {
				w = collection.Width
			}
			if h == 0 {
				h = collection.Height
			}

			iconPkg.Icons = append(iconPkg.Icons, Icon{
				Name:    newNameVariants(k),
				Variant: suffix,
				Body:    v.Body,
				Width:   w,
				Height:  h,
			})

			// if suffix == "" {
			// 	log.Printf("No suffix for %s", k)
			// 	continue
			// }

			withoutSuffix := strings.TrimSuffix(strings.TrimSuffix(k, suffix), "-")
			iconPkg.Variants[suffix][withoutSuffix] = strcase.ToCamel(k)
			if vb, ok := iconPkg.ViewBox[suffix]; !ok {
				w, h = v.Width, v.Height
				if w == 0 {
					w = collection.Width
				}
				if h == 0 {
					h = collection.Height
				}
				iconPkg.ViewBox[suffix] = [2]int{w, h}
			} else if vb[0] == 0 || vb[1] == 0 {
				vb[0] = max(vb[0], v.Width)
				vb[1] = max(vb[1], v.Height)
			}
		}
	}

	slices.SortFunc(iconPkg.Icons, func(a, b Icon) int {
		return strings.Compare(a.Name.Original, b.Name.Original)
	})

	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	tm := TemplateModel{&iconPkg}
	if err = tmpl.Execute(b, tm); err != nil {
		return fmt.Errorf("[%s]: %w", name, err)
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("[%s]: %w", name, err)
	}

	dirPath := path.Join("..", iconPkg.Name.Lower)
	if err = os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("[%s]: %w", name, err)
	}
	if err = os.WriteFile(path.Join(dirPath, iconPkg.Name.Lower+".gen.go"), formatted, 0644); err != nil {
		return fmt.Errorf("[%s]: %w", name, err)
	}
	return nil
}
