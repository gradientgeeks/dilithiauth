/**
 * Utility functions for generating PKCE (Proof Key for Code Exchange) code_verifier and code_challenge.
 */

function base64UrlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  const base64 = typeof btoa !== 'undefined' ? btoa(binary) : Buffer.from(binary, 'binary').toString('base64');
  return base64
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

/**
 * Generate a random PKCE code verifier string (43 to 128 characters).
 */
export function generateCodeVerifier(length: number = 64): string {
  const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  const randomBytes = new Uint8Array(length);
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    crypto.getRandomValues(randomBytes);
  } else {
    for (let i = 0; i < length; i++) {
      randomBytes[i] = Math.floor(Math.random() * 256);
    }
  }
  let verifier = '';
  for (let i = 0; i < length; i++) {
    verifier += possible[randomBytes[i] % possible.length];
  }
  return verifier;
}

/**
 * Calculate the S256 code challenge for a given PKCE code verifier.
 */
export async function generateCodeChallenge(codeVerifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(codeVerifier);
  if (typeof crypto !== 'undefined' && crypto.subtle) {
    const hash = await crypto.subtle.digest('SHA-256', data);
    return base64UrlEncode(hash);
  } else if (typeof require !== 'undefined') {
    // Node.js fallback using crypto module
    const cryptoModule = require('crypto');
    const hash = cryptoModule.createHash('sha256').update(data).digest();
    return hash.toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  }
  throw new Error('Web Crypto API is not available in this environment.');
}

/**
 * Generate a random state parameter to prevent CSRF attacks.
 */
export function generateRandomState(length: number = 32): string {
  return generateCodeVerifier(length);
}
