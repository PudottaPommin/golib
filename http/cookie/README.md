# Cookie
Inspired by and influenced by [Gorilla SecureCookie](https://github.com/gorilla/securecookie).

Secure, tamper-evident cookie values with optional encryption and built-in timestamp checks.

## Features
- HMAC-SHA256 signing to detect tampering
- Optional AES-CTR encryption
- Pluggable encoders (Gob, JSON, Noop) and URL encoders (Base64 by default)
- Max length, max age, and min age checks
- Simple API for round-tripping values

## Install
```bash
go get github.com/pudottapommin/golib
```

## Usage
```go
package main

import (
    "net/http"
    "time"

    "github.com/pudottapommin/golib/http/cookie"
)

type Session struct {
    UserID string
    Role   string
}

func handler(w http.ResponseWriter, r *http.Request) {
    hashKey := cookie.GenerateRandomKey(32)
    blockKey := cookie.GenerateRandomKey(32) // 16, 24, or 32 bytes for AES

    sc, err := cookie.New(
        hashKey,
        blockKey,
        cookie.WithMaxAge(int64((24*time.Hour).Seconds())),
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    value, err := sc.Secure("session", Session{UserID: "123", Role: "admin"})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:  "session",
        Value: string(value),
        Path:  "/",
    })
}

func read(r *http.Request) (*Session, error) {
    c, err := r.Cookie("session")
    if err != nil {
        return nil, err
    }

    var s Session
    if err := sc.Decrypt("session", c.Value, &s); err != nil {
        return nil, err
    }
    return &s, nil
}
```

## How it works
1) Serialize the value using the encoder (Gob by default).
2) Encrypt (if enabled).
3) Prefix with a timestamp and name, then append an HMAC.
4) Base64-url encode the payload.

## Options
- `WithEncoder`: serializer used before encryption/MAC.
- `WithEncryptor`: override or disable encryption (nil disables).
- `WithMAC`: custom MAC hasher implementation.
- `WithURLEncoder`: outer encoding for cookie values.
- `WithMaxLength`: set max encoded length (0 disables).
- `WithMaxAge`: set max age in seconds (0 disables).
- `WithMinAge`: set min age in seconds (0 disables).

## Notes
- If `blockKey` is nil, values are signed but not encrypted.
- Use consistent keys across requests; rotating keys will invalidate existing cookies.
