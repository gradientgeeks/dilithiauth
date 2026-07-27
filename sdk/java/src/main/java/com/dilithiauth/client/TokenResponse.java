package com.dilithiauth.client;

public class TokenResponse {
    private String accessToken;
    private String idToken;
    private String tokenType;
    private int expiresIn;
    private String pqcRef;
    private String pqcAlg;
    private String refreshToken;

    public String getAccessToken() { return accessToken; }
    public void setAccessToken(String accessToken) { this.accessToken = accessToken; }

    public String getIdToken() { return idToken; }
    public void setIdToken(String idToken) { this.idToken = idToken; }

    public String getTokenType() { return tokenType; }
    public void setTokenType(String tokenType) { this.tokenType = tokenType; }

    public int getExpiresIn() { return expiresIn; }
    public void setExpiresIn(int expiresIn) { this.expiresIn = expiresIn; }

    public String getPqcRef() { return pqcRef; }
    public void setPqcRef(String pqcRef) { this.pqcRef = pqcRef; }

    public String getPqcAlg() { return pqcAlg; }
    public void setPqcAlg(String pqcAlg) { this.pqcAlg = pqcAlg; }

    public String getRefreshToken() { return refreshToken; }
    public void setRefreshToken(String refreshToken) { this.refreshToken = refreshToken; }
}
