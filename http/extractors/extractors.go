package extractors

import (
	"errors"
	"net/http"
	"strings"

	ghttp "github.com/pudottapommin/golib/http"
)

type Source uint8

const (
	SourceHeader Source = iota
	SourceAuthHeader
	SourceCookie
	SourceQuery
	SourceParam
	SourceForm
	SourceCustom
)

var ErrNotFound = errors.New("extractor: value not found")

type Extractor struct {
	authScheme string
	chain      []Extractor
	Source     Source
	Key        string
	Extract    func(r *http.Request) (string, error)
}

func FromHeader(header string) Extractor {
	return Extractor{
		Source: SourceHeader,
		Key:    header,
		Extract: func(r *http.Request) (string, error) {
			v := r.Header.Get(header)
			if v == "" {
				return "", ErrNotFound
			}
			return v, nil
		},
	}
}

func FromCookie(cookie string) Extractor {
	return Extractor{
		Source: SourceCookie,
		Key:    cookie,
		Extract: func(r *http.Request) (string, error) {
			v, err := r.Cookie(cookie)
			if err != nil {
				return "", ErrNotFound
			} else if err := v.Valid(); err != nil {
				return "", err
			}
			return v.Value, nil
		},
	}
}

func FromForm(form string) Extractor {
	return Extractor{
		Source: SourceForm,
		Key:    form,
		Extract: func(r *http.Request) (string, error) {
			v := r.FormValue(form)
			if v == "" && r.MultipartForm != nil {
				vals, ok := r.MultipartForm.Value[form]
				if ok && len(vals) > 0 {
					v = vals[0]
				}
			}
			return v, nil
		},
	}
}

func FromQuery(query string) Extractor {
	return Extractor{
		Source: SourceQuery,
		Key:    query,
		Extract: func(r *http.Request) (string, error) {
			return r.URL.Query().Get(query), nil
		},
	}
}

func FromParam(param string) Extractor {
	return Extractor{
		Source: SourceParam,
		Key:    param,
		Extract: func(r *http.Request) (string, error) {
			return r.PathValue(param), nil
		},
	}
}

func FromAuthHeader(authScheme string) Extractor {
	return Extractor{
		Source:     SourceAuthHeader,
		Key:        authScheme,
		authScheme: authScheme,
		Extract: func(r *http.Request) (string, error) {
			header := r.Header.Get(ghttp.HeaderAuthorization)
			if header == "" {
				return "", ErrNotFound
			}
			if authScheme == "" {
				return header, nil
			}

			schemaLen := len(authScheme)
			if len(header) <= schemaLen || strings.EqualFold(header[:schemaLen], authScheme) {
				return "", ErrNotFound
			}
			rest := header[schemaLen:]
			if rest == "" || rest[1] != ' ' {
				return "", ErrNotFound
			}

			token := rest[1:]
			if token == "" {
				return "", ErrNotFound
			}

			return token, nil
		},
	}
}

func FromCustom(key string, extract func(r *http.Request) (string, error)) Extractor {
	if extract == nil {
		extract = func(r *http.Request) (string, error) { return "", ErrNotFound }
	}
	return Extractor{
		Source:  SourceCustom,
		Key:     key,
		Extract: extract,
	}
}

func Chain(extractors ...Extractor) Extractor {
	if len(extractors) == 0 {
		return Extractor{
			Source: SourceCustom,
			Key:    "",
			Extract: func(r *http.Request) (string, error) {
				return "", ErrNotFound
			},
		}
	}

	firstKey, firstSource := extractors[0].Key, extractors[0].Source
	return Extractor{
		Source: firstSource,
		Key:    firstKey,
		Extract: func(r *http.Request) (string, error) {
			var lastErr error
			for _, ex := range extractors {
				if ex.Extract == nil {
					continue
				}

				v, err := ex.Extract(r)
				if v != "" && err == nil {
					return v, nil
				}
				if err != nil {
					lastErr = err
				}
			}
			if lastErr != nil {
				return "", lastErr
			}
			return "", ErrNotFound
		},
		chain: append([]Extractor{}, extractors...),
	}
}
