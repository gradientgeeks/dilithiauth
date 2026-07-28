package developer

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const sessionCookieName = "da_sess"
const sessionTTL = 24 * time.Hour

// SessionClaims holds developer identity and a CSRF token inside the session JWT.
type SessionClaims struct {
	jwt.RegisteredClaims
	DeveloperID string `json:"did"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	CSRFToken   string `json:"csrf"` // random 16-byte hex; must match hidden form field
}

// ErrNoSession is returned when no valid session cookie is present.
var ErrNoSession = errors.New("no valid session")

type sessionContextKey struct{}

var sessionKey = sessionContextKey{}

// IssueSessionCookie signs a JWT session cookie and sets it on the response.
// It returns the session's CSRF token so callers can also return it in the
// JSON response body — the frontend needs it immediately to make
// state-changing requests without waiting for a subsequent /me call.
func IssueSessionCookie(w http.ResponseWriter, secret []byte, dev *Developer) (string, error) {
	csrf := make([]byte, 16)
	if _, err := rand.Read(csrf); err != nil {
		return "", err
	}
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", err
	}
	csrfToken := hex.EncodeToString(csrf)
	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        hex.EncodeToString(jti),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sessionTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		DeveloperID: dev.ID,
		Email:       dev.Email,
		Name:        dev.Name,
		CSRFToken:   csrfToken,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return csrfToken, nil
}

// ParseSessionJWT parses and validates a session JWT string against secret,
// restricting accepted algorithms to HS256. This is the single place that
// performs signature validation for session tokens — used both for the
// primary da_sess cookie (via SessionFromRequest) and the BFF session cookie
// fallback in requireSession — so the two paths can't drift out of sync on
// which algorithms they accept.
func ParseSessionJWT(raw string, secret []byte) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &SessionClaims{}, func(t *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, ErrNoSession
	}
	claims, ok := token.Claims.(*SessionClaims)
	if !ok {
		return nil, ErrNoSession
	}
	return claims, nil
}

// SessionFromRequest validates and parses the session cookie from the request.
func SessionFromRequest(r *http.Request, secret []byte) (*SessionClaims, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, ErrNoSession
	}
	return ParseSessionJWT(cookie.Value, secret)
}

// ClearSessionCookie removes the session cookie by setting MaxAge=-1.
// Path must exactly match the one used in IssueSessionCookie ("/") — browsers
// key cookies by name AND path, so a mismatched Path here would silently
// clear a different cookie and leave the real session cookie intact.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// csrfHeaderName is the header mutating JSON requests must echo the
// session's CSRF token back on. The developer portal API is JSON-only (no
// HTML form posts), so the token travels via a header, not a form field.
const csrfHeaderName = "X-CSRF-Token"

// ValidateCSRF checks the X-CSRF-Token request header against the session's
// CSRFToken using constant-time comparison. A session with an empty
// CSRFToken (e.g. one issued through a path that doesn't set it) never
// validates, even against an empty header — otherwise two empty strings
// would trivially "match".
func ValidateCSRF(r *http.Request, claims *SessionClaims) bool {
	if claims.CSRFToken == "" {
		return false
	}
	headerToken := r.Header.Get(csrfHeaderName)
	return subtle.ConstantTimeCompare([]byte(headerToken), []byte(claims.CSRFToken)) == 1
}

// WithSessionClaims attaches SessionClaims to the request context.
func WithSessionClaims(r *http.Request, claims *SessionClaims) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), sessionKey, claims))
}

// SessionClaimsFromContext retrieves SessionClaims from the request context.
func SessionClaimsFromContext(ctx context.Context) *SessionClaims {
	v, _ := ctx.Value(sessionKey).(*SessionClaims)
	return v
}
