package verifier

import "github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"

// OAuthAdapter adapts *Verifier to the oauth.TokenVerifier interface, letting
// pkg/oauth handlers (e.g. /revoke, /verify/jwt) perform real hybrid
// signature verification without pkg/oauth importing pkg/verifier, which
// itself imports pkg/oauth.
type OAuthAdapter struct {
	*Verifier
}

// VerifyJWT implements oauth.TokenVerifier.
func (a *OAuthAdapter) VerifyJWT(tokenStr string) (*oauth.VerifiedTokenClaims, error) {
	claims, err := a.Verifier.VerifyJWT(tokenStr)
	if err != nil {
		return nil, err
	}
	return &oauth.VerifiedTokenClaims{
		Subject:   claims.Subject,
		Issuer:    claims.Issuer,
		TokenID:   claims.TokenID,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}
