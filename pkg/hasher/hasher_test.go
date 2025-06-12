package hasher

import (
	"strconv"
	"testing"
)

func TestMainHasherHash(t *testing.T) {
	h := New()
	pairs := []string{"admin", "alhlin", "jakoc", "pepoz"}
	for _, p := range pairs {
		t.Run("Hash["+p+"]", func(t *testing.T) {
			t.Parallel()
			hash, err := h.Hash(p)
			if err != nil {
				t.Errorf("Argon2id:Hash failed to generate hash for %v with %v", p, err)
			}
			result, err := h.Verify(hash, p)
			if err != nil && result != PasswordVerificationSuccess {
				t.Errorf("Argon2id:Hash verification generated for %v with %v", p, err)
			}
			if result != PasswordVerificationSuccess {
				t.Errorf("Argon2id:Hash failed to verify result: %d != %d", PasswordVerificationSuccess, result)
			}
		})
	}
}

func TestHasherVerify(t *testing.T) {
	h := New()
	pairs := []struct {
		password string
		hash     string
		result   PasswordVerificationResult
	}{
		{
			password: "admin",
			hash:     "AQAAAAIAAYagAAAAEP1uTAFupvq4+hfCQHpbWJSxwC2DeKAFSu0xrGQ9KGoXU6g0cQPjQzSH9m2SB+RWxg==",
			result:   PasswordVerificationNeedsRehash,
		},
		{
			password: "admin",
			hash:     "AgAAAAEAALgAAAAAAQAAABAYTUtPqa5tWauzS6vd7eJn1GRIVWGHgzgyepicyR+JNOabo7/EhDWwJFz1S8yP0b4=",
			result:   PasswordVerificationSuccess,
		},
		{
			password: "failed",
			hash:     "AQAAAAEAACcQAAAAEEOvBcn3YC/yfI8FXLvTBahIF0D3o6fZcBa1csQEUDWGqBOiuZdoW0gh9+wHP+iZKG==",
			result:   PasswordVerificationFailed,
		},
	}

	for _, p := range pairs {
		t.Run("Verify["+p.password+"]", func(t *testing.T) {
			t.Parallel()
			result, err := h.Verify(p.hash, p.password)

			if err != nil && p.result != PasswordVerificationFailed {
				t.Errorf("Argon2id:Verify failed for %v with %v", p, err)
			}

			if result != p.result {
				t.Errorf("Argon2Id:verify failed to verify result: %d != %d", p.result, result)
			}

		})
	}
}

// PBKDF2
func TestPbkdf2HasherHash(t *testing.T) {
	h := NewPbkdf2()
	pairs := []string{"admin", "alhlin", "jakoc", "pepoz"}
	for _, p := range pairs {
		t.Run("Hash["+p+"]", func(t *testing.T) {
			t.Parallel()
			hash, err := h.Hash(p)
			if err != nil {
				t.Errorf("Pbkdf2:Hash failed to generate hash for %v with %v", p, err)
			}

			result, err := h.Verify(hash, p)
			if err != nil && result != PasswordVerificationSuccess {
				t.Errorf("Pbkdf2:Hash verification generated for %v with %v", p, err)
			}

			if result != PasswordVerificationSuccess {
				t.Errorf("Pbkdf2:Hash failed to verify result: %d != %d, err: %v", PasswordVerificationSuccess, result, err)
			}
		})
	}
}

func TestPbkdf2HasherVerify(t *testing.T) {
	h := NewPbkdf2()
	pairs := []struct {
		password string
		hash     string
		result   PasswordVerificationResult
	}{
		{
			password: "admin",
			hash:     "AQAAAAIAAYagAAAAEP1uTAFupvq4+hfCQHpbWJSxwC2DeKAFSu0xrGQ9KGoXU6g0cQPjQzSH9m2SB+RWxg==",
			result:   PasswordVerificationSuccess,
		},
		{
			password: "alhlin",
			hash:     "AQAAAAEAACcQAAAAEEZ3P+Fn4I8U2fXpmUBxjJS1Ls3ABDUPSjkRUcZMQ+TOUmhRmmHqP+nusLp8tpDNgQ==",
			result:   PasswordVerificationNeedsRehash,
		},
		{
			password: "jakoc",
			hash:     "AQAAAAEAACcQAAAAEEOvBcn3YC/yfI8FXLvTBahIF0D3o6fZcBa1csQEUDWGqBOiuZdoW0gh9+wHP+iZKg==",
			result:   PasswordVerificationNeedsRehash,
		},
		{
			password: "pepoz",
			hash:     "AQAAAAEAACcQAAAAED2K/sYYl6NKhE8u5U6aw9B80c06zIp7t5AroXLWPqbztFWL+bp/CEho+HKBbMltNA==",
			result:   PasswordVerificationNeedsRehash,
		},
		{
			password: "failed",
			hash:     "AQAAAAEAACcQAAAAEEOvBcn3YC/yfI8FXLvTBahIF0D3o6fZcBa1csQEUDWGqBOiuZdoW0gh9+wHP+iZKG==",
			result:   PasswordVerificationFailed,
		},
	}

	for _, p := range pairs {
		t.Run("Verify["+p.password+"]", func(t *testing.T) {
			t.Parallel()
			result, err := h.Verify(p.hash, p.password)
			if err != nil && p.result != PasswordVerificationFailed {
				t.Errorf("Argon2id:Verify failed for %v with %v", p, err)
			}

			if result != p.result {
				t.Errorf("Argon2Id:verify failed to verify result: %d != %d", p.result, result)
			}

		})
	}
}

// Argon2id
func TestArgon2HasherHash(t *testing.T) {
	h := NewArgon2id()
	pairs := []string{"admin", "alhlin", "jakoc", "pepoz"}
	for _, p := range pairs {
		t.Run("Hash["+p+"]", func(t *testing.T) {
			t.Parallel()
			hash, err := h.Hash(p)
			// t.Log(hash)
			if err != nil {
				t.Errorf("Argon2id:Hash failed to generate hash for %v with %v", p, err)
			}

			result, err := h.Verify(hash, p)

			if err != nil && result != PasswordVerificationSuccess {
				t.Errorf("Argon2id:Hash verification generated for %v with %v", p, err)
			}

			if result != PasswordVerificationSuccess {
				t.Errorf("Argon2id:Hash failed to verify result: %d != %d", PasswordVerificationSuccess, result)
			}

		})
	}
}

func TestArgon2idHasherVerify(t *testing.T) {
	h := NewArgon2id()
	pairs := []struct {
		password string
		hash     string
		result   PasswordVerificationResult
	}{
		{
			password: "admin",
			hash:     "AgAAAAEAALgAAAAAAQAAABAYTUtPqa5tWauzS6vd7eJn1GRIVWGHgzgyepicyR+JNOabo7/EhDWwJFz1S8yP0b4=",
			result:   PasswordVerificationSuccess,
		},
		{
			password: "alhlin",
			hash:     "AgAAAAEAALgAAAAAAQAAABDaCe8Crm4E91aCJLFyFK2tVjCRZJ5gDhDJv/ClhUL52mLfcoU5V5L8K/5FSIsiKL0=",
			result:   PasswordVerificationSuccess,
		},
		{
			password: "jakoc",
			hash:     "AgAAAAEAALgAAAAAAQAAABBI7pLz+T46EXIruukPmkD+pR3Sk9GuPvud5AGHgdpT7bLpB70TGVgNQcNgP1JeCfQ=",
			result:   PasswordVerificationSuccess,
		},
		{
			password: "pepoz",
			hash:     "AgAAAAEAALgAAAAAAQAAABAK6Z+a6VGSnJPBqc16dX7A7NU/io7MJ0P1PQWM+DugNPesaHGSedVY6NvPYrJpI4U=",
			result:   PasswordVerificationSuccess,
		},
		{
			password: "failed",
			hash:     "AgAAAAEAALgAAAAAAQAAABByeKLzlyjniMm4zPHv0Qg69OIFHA/2Aiygi2JezmmYxeMTpmG9d9z23KErrrZmUnq=",
			result:   PasswordVerificationFailed,
		},
	}

	for i, p := range pairs {
		t.Run("Verify["+strconv.Itoa(i+1)+"]", func(t *testing.T) {
			t.Parallel()
			result, err := h.Verify(p.hash, p.password)

			if err != nil && p.result != PasswordVerificationFailed {
				t.Errorf("Argon2id:Verify failed for %v with %v", p, err)
			}

			if result != p.result {
				t.Errorf("Argon2Id:verify failed to verify result: %d != %d", p.result, result)
			}

		})
	}
}
