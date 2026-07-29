# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`golib` (module `github.com/pudottapommin/golib`) is a general-purpose Go library of reusable
components for HTTP services: secure cookies, request binding, a set of `net/http`-native
middlewares, password hashing, ID generation, and small utility packages. It is a library, not
an application — there is no `main` package to run except the internal icon-generation tool.

## Commands

The `pkg/icons/generator` tool (a `package main`) uses the Go 1.25+ experimental
`encoding/json/v2` and `encoding/json/jsontext` packages, which require the `jsonv2` build
experiment. **`GOEXPERIMENT=jsonv2` is required for `go build ./...`, `go vet ./...`, and any
other whole-module command** — without it, building `./...` fails on that one package even
though the rest of the module is unaffected.

```bash
# Build everything (required for pkg/icons/generator to compile)
GOEXPERIMENT=jsonv2 go build ./...

# Vet
GOEXPERIMENT=jsonv2 go vet ./...

# Run all tests
GOEXPERIMENT=jsonv2 go test ./...

# Run tests for a single package
go test ./pkg/hasher/...

# Run a single test by name
go test ./http/binding/ -run TestBoolBinder -v

# Run benchmarks (e.g. hasher has dedicated benchmark files)
go test ./pkg/hasher/... -bench=. -run=^$

# Format
gofmt -l -w . && goimports -w .

# Lint (config in .golangci.yml, golangci-lint v2)
golangci-lint run
```

Test files live alongside the code under test using the *same* package name (e.g.
`package binding` in both `binder.go` and `binder_test.go`), not `_test`-suffixed packages. This
is legacy/majority style and does trip golangci-lint's `testpackage` linter, which is otherwise
enabled. `pkg/principal`, `pkg/principal/cookie`, and `http/middleware/principal` use the
`_test`-suffixed external package instead (e.g. `package principal_test`), with a single dot
import of the package under test (`. "github.com/pudottapommin/golib/pkg/principal"`) so call
sites read the same as same-package tests would. Prefer this `_test` + dot-import style for new
packages; existing same-package test files are not being migrated wholesale.

The `.goreleaser.yaml` config runs `go generate ./...` before release and skips actual binary
builds (`builds: - skip: true`) — this repo is only released as source/changelog artifacts.

## Architecture

The module is split into two top-level trees:

- `http/` — components tied to `net/http`.
- `pkg/` — dependency-free core logic and general-purpose utilities.

Root package `golib` (files directly in the repo root: `pool.go`, `date.go`) holds tiny
generic helpers (`Pool[T]` wrapping `sync.Pool`, date comparison helpers) that other packages
(e.g. `pkg/hasher`) depend on — keep these dependency-free since almost everything else imports
the root module path.

### Secure cookies (`http/cookie`)

`cookie.Cookie` (`http/cookie/cookie.go`) implements a Gorilla-securecookie-inspired encode/MAC
scheme: `Secure()` JSON-encodes a value, optionally encrypts it (`Encryptor` interface,
`http/cookie/encryptor.go`), then wraps it in a binary envelope
`[timestamp:8][value_len:4][value:N][mac:M]` (HMAC over `name + timestamp + value`) before
base64-encoding it. `Decrypt()` reverses this, verifying the MAC before decrypting/decoding, and
enforces `minAge`/`maxAge` windows. This one type underlies both the CSRF/session token stores
and the identity cookie store described below — read `cookie.go` before touching any of them.

### Identity/auth (`pkg/principal`)

`pkg/principal` is the identity/auth stack, generic over the identity's ID type `T comparable`
(a caller using `uuid.UUID` IDs and one using `string`/`int64` IDs both instantiate the same
types, e.g. `principal.Identity[uuid.UUID]`). `Identity[T]` (`ID() T`, `Username()`,
`SecurityStamp()`) is resolved by an `IdentityResolver[T]`, stored by an `IdentityStorer[T]`, and
revoked by an `IdentityRevoker` (revocation doesn't need the ID type, so it stays
non-generic) — `IdentityStore[T]` composes all three. `pkg/principal/cookie.Store[T]`
(constructed via `NewCookieStore[T](...)`) is the concrete `IdentityStore[T]`, built on
`http/cookie.Cookie` (symmetric encrypt+MAC, not ECDSA signing); it requires non-empty `hashKey`
and `blockKey` (identity is PII — a missing hash key lets anyone forge the MAC, a missing block
key leaves the payload readable) and defaults the cookie to an 8h max age (callers override via
`cookie.OptFn`, e.g. `WithMaxAge`, or the store's own `WithCookiePath`/`WithCookieDomain`/
`WithCookieSameSite`/`WithCookieHTTPOnly`/`WithCookieSecure`/`WithCookieFactory` setters).
`Resolve` classifies failures as `principal.ErrNoIdentity` (no cookie — anonymous, don't clear
it) or `principal.ErrInvalidIdentity` (present but unverifiable/expired — clear it).

`principal.Service[T]` (`pkg/principal/service.go`) composes an `IdentityResolver[T]` with an
optional `Validator[T]` to implement `AuthenticationService[T].Authenticate`: resolve, then
validate if a validator is set, passing errors through unwrapped so callers can `errors.Is` them.
`SecurityStampValidator[T]` compares a looked-up stamp against `Identity[T].SecurityStamp()` with
`subtle.ConstantTimeCompare`, failing closed (`principal.ErrIdentityRevoked`) if either stamp is
empty since two empty slices compare equal — this is the seam for session revocation
(logout-everywhere, password change, etc.).

`http/middleware/principal` follows the `csrf`/`session` shape (`config.go` + `OptsFn` + `Next` +
`FromContext[T]`) with both `Authentication[T]` and `Authorization[T]` in one package so they
share a typed unexported context key. `Authentication.Revoker` (if set) is invoked on any
authentication failure other than `ErrNoIdentity`, clearing a corrupt/revoked cookie instead of
looping; `AllowAnonymous` lets `ErrNoIdentity` fall through to `next` with no identity in
context. `Authorization`'s default failure is **403**.

There is no separate legacy stack anymore — `pkg/auth` and `http/middleware/{authentication,
authorization}` (ECDSA-signed, non-generic) were removed once `pkg/principal` replaced them.

### Request binding (`http/binding`)

`FormBinder.Bind(r, dst)` reflection-binds form/multipart values onto a struct using `form:"..."`
tags (`name,required,trim`). Struct field metadata is parsed once and cached in a package-level
`sync.Map` keyed by `reflect.Type`; per-field `Binder` resolution is cached per `FormBinder`
instance. Each primitive type (`bool.go`, `int8.go` … `uint64.go`, `string.go`, `float32/64.go`,
`any.go`) implements the `Binder` interface (`Mappable`, `Bind`, `BindMany`); custom types can
satisfy `BindUnmarshaler` or be registered via `AddDefaultBinder`/`FormWithBinders`. Default
binder resolution order matters (first match wins) — new default binders are prepended.

### Extractors (`http/extractors`)

A shared `Extractor` abstraction (`FromHeader`, `FromCookie`, `FromForm`, `FromQuery`,
`FromParam`, `FromAuthHeader`, `FromCustom`) pulls a token/value out of a request from a named
source, with `Chain(...)` trying multiple extractors in order. This is reused by
`http/middleware/csrf` (header + form field) and elsewhere; add new sources here rather than
duplicating request-parsing logic in a middleware.

### Middlewares (`http/middleware/*`)

Each middleware is its own package with a `config.go` (options/defaults, usually a
`Next func(w, r) bool` hook to skip the middleware) and the handler logic. Notable
cross-cutting patterns:
- `csrf` and `session` both follow the same shape: look up a token via a pluggable `Store`,
  regenerate + persist via `mw.Store.Save` if missing/invalid, set `Vary: Cookie` +
  `Cache-Control: no-cache="Set-Cookie"`, then stash the token in the request context under a
  package-level `ContextKey`.
- `static` serves from an `fs.FS`-like abstraction and understands pre-compressed/hashed assets
  via the `ZstdBytesProvider`/`HashedFileProvider` interfaces — pair with `pkg/assetsfs`.
- `principal` is the identity/auth middleware — see above.

### Layered filesystem (`pkg/assetsfs`)

`LayeredFS` composes multiple named `Layer`s (`Local(name, base, sub...)` for an OS dir,
`Blobs(name, fs.FS)` for an embedded/virtual FS) and resolves reads by falling through layers in
order — first layer that has the file wins. `ListFiles`/`ListAllFiles` merge and dedupe names
across *all* layers using `pkg/set.Set`. Used together with `http/middleware/static`.

### Password hashing (`pkg/hasher`)

`hasher.New()` returns a `Hasher` that always hashes new passwords with Argon2id
(`argon2id.go`) but can still *verify* legacy PBKDF2 hashes (`pbkdf2.go`), signaling
`PasswordVerificationNeedsRehash` so callers know to re-hash on next successful login. The
hash format's first byte is an algorithm tag (`pbkdf2Algorithm` / `argon2idAlgorithm`) read
before dispatching to the right verifier. Subkey comparisons use fixed-time comparison
(`compareSubkeysInFixedTime`) — never swap this for a plain `bytes.Equal`.

### IDs (`pkg/id`, `pkg/id/short`, `pkg/id/long`)

`pkg/id.ID` is a fixed-size (`idSize = 21`) crypto-random alphanumeric identifier stored as a
`[21]byte` array (not a string) with custom binary/JSON marshaling. `short`/`long` are separate
sub-packages for other ID shapes/lengths — check which one a given call site expects before
assuming they're interchangeable.

### Other utility packages

- `pkg/set` — generic `Set[T comparable]` over `map[T]struct{}`, with an `iter.Seq[T]` (`Seq()`)
  for range-over-func usage.
- `pkg/utils` — path-joining and other filesystem/crypto helpers shared by `assetsfs` and the
  cookie/hasher code.
- `pkg/clsx` — CSS class-name concatenation helper (à la the JS `clsx` library).
- `pkg/probability` — probability/weighted-choice helpers.
- `pkg/icons/generator` — a standalone code-gen tool (`package main`) that parses the bundled
  Phosphor icons SVG (`icons/phosphor-icons-source.svg`, via `//go:embed`) and filters/streams
  SVG symbols; only this package needs `GOEXPERIMENT=jsonv2`.
