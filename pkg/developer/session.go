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
func IssueSessionCookie(w http.ResponseWriter, secret []byte, dev *Developer) error {
	csrf := make([]byte, 16)
	if _, err := rand.Read(csrf); err != nil {
		return err
	}
	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sessionTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		DeveloperID: dev.ID,
		Email:       dev.Email,
		Name:        dev.Name,
		CSRFToken:   hex.EncodeToString(csrf),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    signed,
		Path:     "/developer",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return nil
}

// SessionFromRequest validates and parses the session cookie from the request.
func SessionFromRequest(r *http.Request, secret []byte) (*SessionClaims, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, ErrNoSession
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &SessionClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
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

// ClearSessionCookie removes the session cookie by setting MaxAge=-1.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/developer",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})
}

// ValidateCSRF checks the form's csrf_token against the session's CSRFToken
// using constant-time comparison.
func ValidateCSRF(r *http.Request, claims *SessionClaims) bool {
	formToken := r.FormValue("csrf_token")
	return subtle.ConstantTimeCompare([]byte(formToken), []byte(claims.CSRFToken)) == 1
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
