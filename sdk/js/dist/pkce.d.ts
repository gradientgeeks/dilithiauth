/**
 * Utility functions for generating PKCE (Proof Key for Code Exchange) code_verifier and code_challenge.
 */
/**
 * Generate a random PKCE code verifier string (43 to 128 characters).
 */
export declare function generateCodeVerifier(length?: number): string;
/**
 * Calculate the S256 code challenge for a given PKCE code verifier.
 */
export declare function generateCodeChallenge(codeVerifier: string): Promise<string>;
/**
 * Generate a random state parameter to prevent CSRF attacks.
 */
export declare function generateRandomState(length?: number): string;
