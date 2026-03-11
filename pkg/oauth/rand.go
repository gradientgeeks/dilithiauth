package oauth

import "crypto/rand"

// cryptoRead is a package-level alias so that tests can stub it out.
var cryptoRead = func(b []byte) (int, error) {
	return rand.Read(b)
}
