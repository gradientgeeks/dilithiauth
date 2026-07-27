package com.dilithiauth.client;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

public class DilithiAuthClient {

    private final String domain;
    private final String clientId;
    private final String clientSecret;
    private final String redirectUri;

    private DilithiAuthClient(Builder builder) {
        if (builder.domain == null || builder.domain.isEmpty()) {
            throw new IllegalArgumentException("domain is required");
        }
        if (builder.clientId == null || builder.clientId.isEmpty()) {
            throw new IllegalArgumentException("clientId is required");
        }
        this.domain = builder.domain.replaceAll("/+$", "");
        this.clientId = builder.clientId;
        this.clientSecret = builder.clientSecret;
        this.redirectUri = builder.redirectUri;
    }

    public static class Builder {
        private String domain;
        private String clientId;
        private String clientSecret;
        private String redirectUri;

        public Builder domain(String domain) { this.domain = domain; return this; }
        public Builder clientId(String clientId) { this.clientId = clientId; return this; }
        public Builder clientSecret(String clientSecret) { this.clientSecret = clientSecret; return this; }
        public Builder redirectUri(String redirectUri) { this.redirectUri = redirectUri; return this; }

        public DilithiAuthClient build() {
            return new DilithiAuthClient(this);
        }
    }

    public static Builder builder() {
        return new Builder();
    }

    public String getAuthorizationUrl(String redirectUri, String state, String codeChallenge) {
        return getAuthorizationUrl(redirectUri, state, codeChallenge, "openid profile email");
    }

    public String getAuthorizationUrl(String redirectUri, String state, String codeChallenge, String scope) {
        String targetRedirect = (redirectUri != null && !redirectUri.isEmpty()) ? redirectUri : this.redirectUri;
        if (targetRedirect == null || targetRedirect.isEmpty()) {
            throw new IllegalArgumentException("redirectUri is required");
        }

        StringBuilder sb = new StringBuilder();
        sb.append(domain).append("/authorize?");
        sb.append("client_id=").append(encode(clientId));
        sb.append("&redirect_uri=").append(encode(targetRedirect));
        sb.append("&response_type=code");
        sb.append("&scope=").append(encode(scope));

        if (state != null) sb.append("&state=").append(encode(state));
        if (codeChallenge != null) {
            sb.append("&code_challenge=").append(encode(codeChallenge));
            sb.append("&code_challenge_method=S256");
        }

        return sb.toString();
    }

    public TokenResponse exchangeCode(String code, String codeVerifier, String redirectUri) throws Exception {
        String targetRedirect = (redirectUri != null && !redirectUri.isEmpty()) ? redirectUri : this.redirectUri;
        String endpoint = domain + "/api/v1/oauth/token";

        StringBuilder postData = new StringBuilder();
        postData.append("grant_type=authorization_code");
        postData.append("&client_id=").append(encode(clientId));
        postData.append("&code=").append(encode(code));
        postData.append("&redirect_uri=").append(encode(targetRedirect));
        postData.append("&code_verifier=").append(encode(codeVerifier));

        byte[] postBytes = postData.toString().getBytes(StandardCharsets.UTF_8);
        URL url = new URL(endpoint);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/x-www-form-urlencoded");
        conn.setDoOutput(true);

        if (clientSecret != null && !clientSecret.isEmpty()) {
            String auth = clientId + ":" + clientSecret;
            String encodedAuth = Base64.getEncoder().encodeToString(auth.getBytes(StandardCharsets.UTF_8));
            conn.setRequestProperty("Authorization", "Basic " + encodedAuth);
        }

        try (OutputStream os = conn.getOutputStream()) {
            os.write(postBytes);
        }

        int status = conn.getResponseCode();
        BufferedReader reader;
        if (status == 200) {
            reader = new BufferedReader(new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8));
        } else {
            reader = new BufferedReader(new InputStreamReader(conn.getErrorStream(), StandardCharsets.UTF_8));
        }

        StringBuilder responseBuilder = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            responseBuilder.append(line);
        }
        reader.close();

        if (status != 200) {
            throw new RuntimeException("Token exchange failed (" + status + "): " + responseBuilder.toString());
        }

        String json = responseBuilder.toString();
        TokenResponse response = new TokenResponse();
        response.setAccessToken(extractJsonVal(json, "access_token"));
        response.setTokenType(extractJsonVal(json, "token_type"));
        response.setPqcRef(extractJsonVal(json, "pqc_ref"));
        response.setRefreshToken(extractJsonVal(json, "refresh_token"));
        return response;
    }

    public UserInfo getUserInfo(String accessToken) throws Exception {
        URL url = new URL(domain + "/userinfo");
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        conn.setRequestProperty("Authorization", "Bearer " + accessToken);

        int status = conn.getResponseCode();
        if (status != 200) {
            throw new RuntimeException("UserInfo request failed with status: " + status);
        }

        BufferedReader reader = new BufferedReader(new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8));
        StringBuilder responseBuilder = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            responseBuilder.append(line);
        }
        reader.close();

        String json = responseBuilder.toString();
        UserInfo userInfo = new UserInfo();
        userInfo.setSub(extractJsonVal(json, "sub"));
        userInfo.setName(extractJsonVal(json, "name"));
        userInfo.setEmail(extractJsonVal(json, "email"));
        return userInfo;
    }

    private static String encode(String val) {
        try {
            return URLEncoder.encode(val, StandardCharsets.UTF_8.name());
        } catch (Exception e) {
            return val;
        }
    }

    private static String extractJsonVal(String json, String key) {
        String pattern = "\"" + key + "\":\"";
        int start = json.indexOf(pattern);
        if (start == -1) return null;
        start += pattern.length();
        int end = json.indexOf("\"", start);
        if (end == -1) return null;
        return json.substring(start, end);
    }
}
