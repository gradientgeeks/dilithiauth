import { DilithiAuthClientOptions, RedirectLoginOptions, RedirectCallbackResult, UserClaims, JWKSResponse } from './types.js';
export declare class DilithiAuthClient {
    private domain;
    private clientId;
    private clientSecret?;
    private redirectUri?;
    private tokenStrategy;
    private storage;
    constructor(options: DilithiAuthClientOptions);
    private createMemoryStorage;
    /**
     * Initiates PKCE login redirect flow to DilithiAuth authorization server.
     */
    loginWithRedirect(options?: RedirectLoginOptions): Promise<void>;
    /**
     * Processes OAuth callback URL, exchanges code for Post-Quantum token, and caches user session.
     */
    handleRedirectCallback(urlStr?: string): Promise<RedirectCallbackResult>;
    /**
     * Gets cached access token if valid. Throws error if expired or missing.
     */
    getTokenSilently(): Promise<string>;
    /**
     * Returns current user claims if authenticated.
     */
    getUser(): UserClaims | null;
    /**
     * Check if user is currently authenticated.
     */
    isAuthenticated(): boolean;
    /**
     * Fetch public keys (JWKS) from DilithiAuth IdP discovery endpoint.
     */
    getJwks(): Promise<JWKSResponse>;
    /**
     * Log out user, clear cached tokens, and optionally redirect.
     */
    logout(options?: {
        returnTo?: string;
    }): void;
}
