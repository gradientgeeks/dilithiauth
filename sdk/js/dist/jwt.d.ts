import { UserClaims } from './types.js';
/**
 * Decode base64url string to UTF-8 text.
 */
export declare function urlBase64Decode(str: string): string;
/**
 * Parse JWT token string and extract claims payload.
 */
export declare function decodeJwt<T = UserClaims>(token: string): T;
/**
 * Extract JWT header.
 */
export declare function decodeJwtHeader(token: string): Record<string, any>;
/**
 * Check if a JWT token is expired based on its 'exp' claim.
 */
export declare function isTokenExpired(token: string, offsetSeconds?: number): boolean;
