"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.DilithiAuthClient = void 0;
const pkce_js_1 = require("./pkce.js");
const jwt_js_1 = require("./jwt.js");
const STORAGE_KEY_PREFIX = 'dilithi_auth_';
class DilithiAuthClient {
    constructor(options) {
        if (!options.domain)
            throw new Error('DilithiAuthClient: "domain" option is required');
        if (!options.clientId)
            throw new Error('DilithiAuthClient: "clientId" option is required');
        // Normalize domain (remove trailing slash)
        this.domain = options.domain.replace(/\/+$/, '');
        this.clientId = options.clientId;
        this.clientSecret = options.clientSecret;
        this.redirectUri = options.redirectUri;
        this.tokenStrategy = options.tokenStrategy || 'thin';
        const storageType = options.storageType || 'sessionStorage';
        this.storage = typeof window !== 'undefined' && window[storageType] ? window[storageType] : this.createMemoryStorage();
    }
    createMemoryStorage() {
        const store = {};
        return {
            getItem: (key) => store[key] || null,
            setItem: (key, val) => { store[key] = val; },
            removeItem: (key) => { delete store[key]; },
            clear: () => { Object.keys(store).forEach(k => delete store[k]); },
            length: 0,
            key: () => null
        };
    }
    /**
     * Initiates PKCE login redirect flow to DilithiAuth authorization server.
     */
    async loginWithRedirect(options = {}) {
        const redirectUri = options.redirectUri || this.redirectUri;
        if (!redirectUri) {
            throw new Error('loginWithRedirect: redirectUri is required in client options or method parameters');
        }
        const codeVerifier = (0, pkce_js_1.generateCodeVerifier)(64);
        const codeChallenge = await (0, pkce_js_1.generateCodeChallenge)(codeVerifier);
        const state = options.state || (0, pkce_js_1.generateRandomState)(32);
        // Save PKCE verifier & state in session storage
        this.storage.setItem(STORAGE_KEY_PREFIX + 'code_verifier', codeVerifier);
        this.storage.setItem(STORAGE_KEY_PREFIX + 'state', state);
        const authUrl = new URL(`${this.domain}/authorize`);
        authUrl.searchParams.set('client_id', this.clientId);
        authUrl.searchParams.set('redirect_uri', redirectUri);
        authUrl.searchParams.set('response_type', 'code');
        authUrl.searchParams.set('state', state);
        authUrl.searchParams.set('code_challenge', codeChallenge);
        authUrl.searchParams.set('code_challenge_method', 'S256');
        if (options.sub) {
            authUrl.searchParams.set('sub', options.sub);
        }
        if (options.scope) {
            authUrl.searchParams.set('scope', options.scope);
        }
        if (typeof window !== 'undefined') {
            window.location.href = authUrl.toString();
        }
    }
    /**
     * Processes OAuth callback URL, exchanges code for Post-Quantum token, and caches user session.
     */
    async handleRedirectCallback(urlStr) {
        const currentUrl = urlStr || (typeof window !== 'undefined' ? window.location.href : '');
        const url = new URL(currentUrl);
        const code = url.searchParams.get('code');
        const state = url.searchParams.get('state');
        const error = url.searchParams.get('error');
        if (error) {
            const errorDesc = url.searchParams.get('error_description') || error;
            throw new Error(`Authentication error: ${errorDesc}`);
        }
        if (!code) {
            throw new Error('handleRedirectCallback: No authorization code found in callback URL');
        }
        const savedState = this.storage.getItem(STORAGE_KEY_PREFIX + 'state');
        if (savedState && state !== savedState) {
            throw new Error('handleRedirectCallback: Invalid state parameter (potential CSRF attack)');
        }
        const codeVerifier = this.storage.getItem(STORAGE_KEY_PREFIX + 'code_verifier');
        if (!codeVerifier) {
            throw new Error('handleRedirectCallback: PKCE code_verifier missing from storage');
        }
        // Clean up temporary PKCE state
        this.storage.removeItem(STORAGE_KEY_PREFIX + 'state');
        this.storage.removeItem(STORAGE_KEY_PREFIX + 'code_verifier');
        // Exchange code for token
        const tokenEndpoint = `${this.domain}/api/v1/oauth/token`;
        const body = new URLSearchParams({
            grant_type: 'authorization_code',
            client_id: this.clientId,
            code: code,
            redirect_uri: this.redirectUri || url.origin + url.pathname,
            code_verifier: codeVerifier
        });
        const headers = {
            'Content-Type': 'application/x-www-form-urlencoded'
        };
        if (this.clientSecret) {
            headers['Authorization'] = 'Basic ' + (typeof btoa !== 'undefined' ? btoa(`${this.clientId}:${this.clientSecret}`) : Buffer.from(`${this.clientId}:${this.clientSecret}`).toString('base64'));
        }
        const response = await fetch(tokenEndpoint, {
            method: 'POST',
            headers,
            body
        });
        if (!response.ok) {
            const errText = await response.text();
            throw new Error(`Token exchange failed (${response.status}): ${errText}`);
        }
        const tokenData = await response.json();
        const accessToken = tokenData.access_token;
        const userClaims = (0, jwt_js_1.decodeJwt)(accessToken);
        // Cache session in storage
        this.storage.setItem(STORAGE_KEY_PREFIX + 'access_token', accessToken);
        this.storage.setItem(STORAGE_KEY_PREFIX + 'user', JSON.stringify(userClaims));
        return {
            accessToken,
            tokenType: tokenData.token_type || 'Bearer',
            expiresIn: tokenData.expires_in,
            pqcRef: tokenData.pqc_ref || userClaims.pqc_ref,
            user: userClaims
        };
    }
    /**
     * Gets cached access token if valid. Throws error if expired or missing.
     */
    async getTokenSilently() {
        const token = this.storage.getItem(STORAGE_KEY_PREFIX + 'access_token');
        if (!token || (0, jwt_js_1.isTokenExpired)(token)) {
            this.logout();
            throw new Error('No active or valid session available. User must log in again.');
        }
        return token;
    }
    /**
     * Returns current user claims if authenticated.
     */
    getUser() {
        const userJson = this.storage.getItem(STORAGE_KEY_PREFIX + 'user');
        if (!userJson)
            return null;
        try {
            return JSON.parse(userJson);
        }
        catch {
            return null;
        }
    }
    /**
     * Check if user is currently authenticated.
     */
    isAuthenticated() {
        const token = this.storage.getItem(STORAGE_KEY_PREFIX + 'access_token');
        return !!token && !(0, jwt_js_1.isTokenExpired)(token);
    }
    /**
     * Fetch public keys (JWKS) from DilithiAuth IdP discovery endpoint.
     */
    async getJwks() {
        const jwksUrl = `${this.domain}/.well-known/jwks.json`;
        const res = await fetch(jwksUrl);
        if (!res.ok)
            throw new Error('Failed to fetch JWKS key set');
        return res.json();
    }
    /**
     * Log out user, clear cached tokens, and optionally redirect.
     */
    logout(options) {
        this.storage.removeItem(STORAGE_KEY_PREFIX + 'access_token');
        this.storage.removeItem(STORAGE_KEY_PREFIX + 'user');
        this.storage.removeItem(STORAGE_KEY_PREFIX + 'state');
        this.storage.removeItem(STORAGE_KEY_PREFIX + 'code_verifier');
        if (options?.returnTo && typeof window !== 'undefined') {
            window.location.href = options.returnTo;
        }
    }
}
exports.DilithiAuthClient = DilithiAuthClient;
