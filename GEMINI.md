# GEMINI.md

## Project Overview
**GoLIB** is a comprehensive Go library providing a set of high-quality, reusable components for web applications and general systems development. It focuses on security, performance, and idiomatic Go patterns.

### Main Technologies
- **Language:** Go (1.26+)
- **HTTP:** Standard library `net/http` with extensions for secure cookies, CSRF protection, and parameter binding.
- **Security:** ECDSA-based authentication, Argon2id/PBKDF2 hashing, and AES/ChaCha20-Poly1305 encryption.
- **Utilities:** UUID generation, property-based testing helpers, and type-safe collections (sets).

### Architecture
The project is structured into two main hierarchies:
- `http/`: Web-specific components including:
  - `cookie/`: Secure, length-prefixed binary cookies with encryption and HMAC.
  - `middleware/`: A rich set of middlewares (CSRF, Authentication, HSTS, Logger, etc.).
  - `binding/`: A request parameter binder inspired by the Gin framework.
- `pkg/`: Core logic and non-web utilities:
  - `auth/`: Identity management and signed token handling.
  - `hasher/`: Pluggable password hashing.
  - `id/`: Specialized ID generation (Short, Long).
  - `utils/`: Common cryptographic and file system helpers.

## Building and Running
The project uses `just` as a command runner. Key tasks include:

- **Check Code:** `just check` (runs `go vet`)
- **Format Code:** `just fmt` (runs `go fmt` and `goimports`)
- **Run Tests:** `just test` (runs all tests with multiple CPU configurations)
- **Benchmarking:** `just bench` (runs performance benchmarks)
- **Update Dependencies:** `just update` (updates modules and tidies `go.mod`)

## Development Conventions
- **Error Handling:** Errors are handled explicitly. Web-specific errors are often wrapped in custom types within the `binding` package.
- **Testing:** Comprehensive test coverage is expected. Use `TestMain` or package-level tests. Benchmarks are encouraged for performance-critical components (e.g., in `pkg/hasher`).
- **Security:** 
  - Never log sensitive information like keys or MACs (as enforced in `hasher.go`).
  - Use `subtle.ConstantTimeCompare` for all security-sensitive comparisons.
  - Prefer the binary length-prefixed format for serializing secure data.
- **Naming:** Follow standard Go naming conventions. Receiver names should be short and consistent (e.g., `sc` for `Cookie`, `mw` for `Middleware`).
