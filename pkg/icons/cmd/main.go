package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/pudottapommin/golib"
)

//go:embed svg.gotmpl
var iconsTemplate string

const iconifyIconSetHugeicons = "https://raw.githubusercontent.com/iconify/icon-sets/refs/heads/master/json/hugeicons.json"

var buffers = golib.NewPool(func() *bytes.Buffer {
	return new(bytes.Buffer)
})

type IconifyHugeicons struct {
	Prefix string `json:"prefix"`
	Info   struct {
		Height int `json:"height"`
	} `json:"info"`
	LastModified uint `json:"lastModified"`
	Icons        map[string]struct {
		Body string `json:"body"`
	}
}

func main() {
	t := time.Now()
	req, err := http.NewRequest("GET", iconifyIconSetHugeicons, nil)
	checkErr(err)

	resp, err := http.DefaultClient.Do(req)
	checkErr(err)

	defer resp.Body.Close()
	var icons IconifyHugeicons
	if err := json.NewDecoder(resp.Body).Decode(&icons); err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.New("").Funcs(template.FuncMap{"pascal": PascalCase}).Parse(iconsTemplate)
	checkErr(err)
	f, err := os.Create("../hugeicons/" + icons.Prefix + ".go")
	checkErr(err)
	defer f.Close()
	if err := tmpl.Execute(f, icons); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated all icons in %dms", time.Since(t).Milliseconds())
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type parser uint8

const (
	_             parser = iota //   _$$_This is some text, OK?!
	idle                        // 1 ↑↑↑↑                  ↑   ↑
	firstAlphaNum               // 2     ↑    ↑  ↑    ↑     ↑
	alphaNum                    // 3      ↑↑↑  ↑  ↑↑↑  ↑↑↑   ↑
	delimiter                   // 4         ↑  ↑    ↑    ↑   ↑
)

func PascalCase(input string) string {
	b := buffers.Get()
	defer func() {
		b.Reset()
		buffers.Put(b)
	}()
	str := markLetterCaseChanges(input)
	state := idle
	for i := 0; i < len(str); {
		r, size := utf8.DecodeRuneInString(str[i:])
		i += size
		state = state.next(r)
		switch state {
		case firstAlphaNum:
			b.WriteRune(unicode.ToUpper(r))
		case alphaNum:
			b.WriteRune(unicode.ToLower(r))
		default:
		}
	}
	return b.String()
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func (s parser) next(r rune) parser {
	switch s {
	case idle:
		if isAlphaNum(r) {
			return firstAlphaNum
		}
	case firstAlphaNum:
		if isAlphaNum(r) {
			return alphaNum
		}
		return delimiter
	case alphaNum:
		if !isAlphaNum(r) {
			return delimiter
		}
	case delimiter:
		if isAlphaNum(r) {
			return firstAlphaNum
		}
		return idle
	}
	return s
}

// Mark letter case changes, i.e., "camelCaseTEXT" -> "camel_Case_TEXT".
func markLetterCaseChanges(input string) string {
	b := buffers.Get()
	defer func() {
		b.Reset()
		buffers.Put(b)
	}()

	wasLetter := false
	countConsecutiveUpperLetters := 0

	for i := 0; i < len(input); {
		r, size := utf8.DecodeRuneInString(input[i:])
		i += size

		if unicode.IsLetter(r) {
			if wasLetter && countConsecutiveUpperLetters > 1 && !unicode.IsUpper(r) {
				b.WriteRune('_')
			}
			if wasLetter && countConsecutiveUpperLetters == 0 && unicode.IsUpper(r) {
				b.WriteRune('_')
			}
		}

		wasLetter = unicode.IsLetter(r)
		if unicode.IsUpper(r) {
			countConsecutiveUpperLetters++
		} else {
			countConsecutiveUpperLetters = 0
		}
		b.WriteRune(r)
	}
	return b.String()
}
