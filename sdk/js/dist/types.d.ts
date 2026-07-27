export type TokenStrategy = 'thin' | 'full' | 'cwt';
export interface DilithiAuthClientOptions {
    /** The base issuer domain of your DilithiAuth IdP instance (defaults to "https://dilithiauthapis.gradientgeeks.tech") */
    domain?: string;
    /** The registered OAuth 2.0 Client ID for your application */
    clientId: string;
    /** Client secret for confidential clients (leave undefined for Public/PKCE clients) */
    clientSecret?: string;
    /** Default redirect URI configured in your DilithiAuth application */
    redirectUri?: string;
    /** Token strategy preference: 'thin' (Ed25519 + sidecar ref), 'jumbo' (embedded ML-DSA-44), or 'cwt' (CBOR) */
    tokenStrategy?: TokenStrategy;
    /** Storage medium for session data: 'sessionStorage' (default) or 'localStorage' */
    storageType?: 'sessionStorage' | 'localStorage';
}
export interface RedirectLoginOptions {
    /** Override default redirect URI for this authorization request */
    redirectUri?: string;
    /** Custom state parameter */
    state?: string;
    /** Custom subject claim requested */
    sub?: string;
    /** Additional scope requested */
    scope?: string;
}
export interface RedirectCallbackResult {
    appState?: Record<string, any>;
    accessToken: string;
    idToken?: string;
    tokenType: string;
    expiresIn?: number;
    pqcRef?: string;
    user: UserClaims;
}
export interface TokenResponse {
    access_token: string;
    id_token?: string;
    token_type: string;
    expires_in?: number;
    pqc_ref?: string;
    pqc_alg?: string;
}
export interface UserClaims {
    sub: string;
    iss?: string;
    aud?: string | string[];
    exp?: number;
    iat?: number;
    pqc_ref?: string;
    pqc_alg?: string;
    [key: string]: any;
}
export interface JWKSKey {
    kty: string;
    use?: string;
    alg?: string;
    crv?: string;
    kid?: string;
    x?: string;
    y?: string;
    mldsa44_pub?: string;
}
export interface JWKSResponse {
    keys: JWKSKey[];
}
