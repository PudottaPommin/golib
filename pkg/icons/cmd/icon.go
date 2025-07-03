package main

import (
	"strings"

	"github.com/iancoleman/strcase"
)

type (
	iconCollectionInfo struct {
		Prefix string `json:"prefix"`
		Info   struct {
			Name    string  `json:"name,omitempty"`
			Total   int     `json:"total,omitempty"`
			Version string  `json:"version,omitempty"`
			License License `json:"license,omitempty"`
			Height  int     `json:"height,omitempty"`
		} `json:"info"`
		LastModified int64 `json:"lastModified,omitempty"`
		Icons        map[string]struct {
			Body   string `json:"body"`
			Width  int    `json:"width,omitempty"`
			Height int    `json:"height,omitempty"`
		}
		Suffixes map[string]string `json:"suffixes,omitempty"`
		Width    int               `json:"width,omitempty"`
		Height   int               `json:"height,omitempty"`
	}
	License struct {
		Title string `json:"title,omitempty"`
		SPDX  string `json:"spdx,omitempty"`
		URL   string `json:"url,omitempty"`
	}

	NameVariants struct {
		Original string
		Pascal   string
		Lower    string
	}

	IconPackage struct {
		Name     NameVariants
		Version  string
		Width    int
		Height   int
		ViewBox  map[string][2]int
		Icons    []Icon
		Variants map[string]map[string]string
	}

	Icon struct {
		Name    NameVariants
		Body    string
		Width   int
		Height  int
		Variant string
	}
)

func newNameVariants(s string) NameVariants {
	return NameVariants{
		Original: s,
		Pascal:   strcase.ToCamel(s),
		Lower:    strings.ToLower(s),
	}
}
