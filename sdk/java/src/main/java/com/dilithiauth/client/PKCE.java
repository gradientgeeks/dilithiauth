package com.dilithiauth.client;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.util.Base64;

public class PKCE {

    public static class Pair {
        private final String codeVerifier;
        private final String codeChallenge;

        public Pair(String codeVerifier, String codeChallenge) {
            this.codeVerifier = codeVerifier;
            this.codeChallenge = codeChallenge;
        }

        public String getCodeVerifier() { return codeVerifier; }
        public String getCodeChallenge() { return codeChallenge; }
    }

    public static Pair generatePair() {
        SecureRandom random = new SecureRandom();
        byte[] bytes = new byte[32];
        random.nextBytes(bytes);
        String verifier = Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
        String challenge = calculateChallenge(verifier);
        return new Pair(verifier, challenge);
    }

    public static String calculateChallenge(String codeVerifier) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(codeVerifier.getBytes(StandardCharsets.UTF_8));
            return Base64.getUrlEncoder().withoutPadding().encodeToString(hash);
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("SHA-256 algorithm not available", e);
        }
    }

    public static String generateState() {
        SecureRandom random = new SecureRandom();
        byte[] bytes = new byte[16];
        random.nextBytes(bytes);
        return Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
    }
}
