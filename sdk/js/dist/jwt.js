"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.urlBase64Decode = urlBase64Decode;
exports.decodeJwt = decodeJwt;
exports.decodeJwtHeader = decodeJwtHeader;
exports.isTokenExpired = isTokenExpired;
/**
 * Decode base64url string to UTF-8 text.
 */
function urlBase64Decode(str) {
    let output = str.replace(/-/g, '+').replace(/_/g, '/');
    switch (output.length % 4) {
        case 0:
            break;
        case 2:
            output += '==';
            break;
        case 3:
            output += '=';
            break;
        default:
            throw new Error('Illegal base64url string!');
    }
    if (typeof atob !== 'undefined') {
        return decodeURIComponent(Array.prototype.map
            .call(atob(output), (c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
            .join(''));
    }
    return Buffer.from(output, 'base64').toString('utf8');
}
/**
 * Parse JWT token string and extract claims payload.
 */
function decodeJwt(token) {
    const parts = token.split('.');
    if (parts.length !== 3) {
        throw new Error('Invalid JWT format');
    }
    const payloadJson = urlBase64Decode(parts[1]);
    return JSON.parse(payloadJson);
}
/**
 * Extract JWT header.
 */
function decodeJwtHeader(token) {
    const parts = token.split('.');
    if (parts.length !== 3) {
        throw new Error('Invalid JWT format');
    }
    const headerJson = urlBase64Decode(parts[0]);
    return JSON.parse(headerJson);
}
/**
 * Check if a JWT token is expired based on its 'exp' claim.
 */
function isTokenExpired(token, offsetSeconds = 0) {
    try {
        const claims = decodeJwt(token);
        if (!claims.exp)
            return false;
        const currentTime = Math.floor(Date.now() / 1000);
        return claims.exp < currentTime + offsetSeconds;
    }
    catch {
        return true;
    }
}
