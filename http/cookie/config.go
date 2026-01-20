package cookie

type OptFn func(sc *Cookie)

// WithEncoder sets the serializer used before encryption/MAC.
func WithEncoder(enc Encoder) OptFn {
	return func(sc *Cookie) {
		sc.encoder = enc
	}
}

// WithEncryptor overrides the encryptor (nil disables encryption).
func WithEncryptor(enc Encryptor) OptFn {
	return func(sc *Cookie) {
		sc.encryptor = enc
	}
}

// WithMAC sets the MAC hasher implementation.
func WithMAC(mac Hasher) OptFn {
	return func(sc *Cookie) {
		sc.mac = mac
	}
}

// WithURLEncoder sets the outer encoding (e.g., base64) for cookie values.
func WithURLEncoder(enc Encoder) OptFn {
	return func(sc *Cookie) {
		sc.urlEncoder = enc
	}
}

// WithMaxLength sets the maximum encoded length; 0 disables length checks.
func WithMaxLength(n int) OptFn {
	return func(sc *Cookie) {
		sc.maxLength = n
	}
}

// WithMaxAge sets the max age (seconds); 0 disables expiration checks.
func WithMaxAge(seconds int64) OptFn {
	return func(sc *Cookie) {
		sc.maxAge = seconds
	}
}

// WithMinAge sets the minimum age (seconds); 0 disables min-age checks.
func WithMinAge(seconds int64) OptFn {
	return func(sc *Cookie) {
		sc.minAge = seconds
	}
}
