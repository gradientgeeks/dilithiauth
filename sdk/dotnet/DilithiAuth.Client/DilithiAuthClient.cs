using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Web;

namespace DilithiAuth.Client;

public class DilithiAuthClient
{
    private readonly string _domain;
    private readonly string _clientId;
    private readonly string? _clientSecret;
    private readonly string? _redirectUri;
    private readonly HttpClient _httpClient;

    public DilithiAuthClient(DilithiAuthOptions options, HttpClient? httpClient = null)
    {
        ArgumentNullException.ThrowIfNull(options);
        if (string.IsNullOrWhiteSpace(options.Domain)) throw new ArgumentException("Domain is required", nameof(options));
        if (string.IsNullOrWhiteSpace(options.ClientId)) throw new ArgumentException("ClientId is required", nameof(options));

        _domain = options.Domain.TrimEnd('/');
        _clientId = options.ClientId;
        _clientSecret = options.ClientSecret;
        _redirectUri = options.RedirectUri;
        _httpClient = httpClient ?? new HttpClient();
    }

    public string GetAuthorizationUrl(string? redirectUri = null, string? state = null, string? codeChallenge = null, string scope = "openid profile email")
    {
        string targetRedirect = redirectUri ?? _redirectUri ?? throw new InvalidOperationException("RedirectUri must be specified.");

        var query = HttpUtility.ParseQueryString(string.Empty);
        query["client_id"] = _clientId;
        query["redirect_uri"] = targetRedirect;
        query["response_type"] = "code";
        query["scope"] = scope;

        if (!string.IsNullOrEmpty(state)) query["state"] = state;
        if (!string.IsNullOrEmpty(codeChallenge))
        {
            query["code_challenge"] = codeChallenge;
            query["code_challenge_method"] = "S256";
        }

        return $"{_domain}/authorize?{query}";
    }

    public async Task<TokenResponse> ExchangeCodeAsync(string code, string codeVerifier, string? redirectUri = null)
    {
        string targetRedirect = redirectUri ?? _redirectUri ?? throw new InvalidOperationException("RedirectUri must be specified.");
        string endpoint = $"{_domain}/api/v1/oauth/token";

        var formData = new Dictionary<string, string>
        {
            { "grant_type", "authorization_code" },
            { "client_id", _clientId },
            { "code", code },
            { "redirect_uri", targetRedirect },
            { "code_verifier", codeVerifier }
        };

        var request = new HttpRequestMessage(HttpMethod.Post, endpoint)
        {
            Content = new FormUrlEncodedContent(formData)
        };

        if (!string.IsNullOrEmpty(_clientSecret))
        {
            string auth = Convert.ToBase64String(Encoding.UTF8.GetBytes($"{_clientId}:{_clientSecret}"));
            request.Headers.Authorization = new AuthenticationHeaderValue("Basic", auth);
        }

        HttpResponseMessage response = await _httpClient.SendAsync(request);
        string json = await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
        {
            throw new HttpRequestException($"Token exchange failed ({response.StatusCode}): {json}");
        }

        return JsonSerializer.Deserialize<TokenResponse>(json) ?? throw new InvalidOperationException("Failed to deserialize TokenResponse");
    }

    public async Task<UserInfo> GetUserInfoAsync(string accessToken)
    {
        var request = new HttpRequestMessage(HttpMethod.Get, $"{_domain}/userinfo");
        request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", accessToken);

        HttpResponseMessage response = await _httpClient.SendAsync(request);
        string json = await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
        {
            throw new HttpRequestException($"UserInfo request failed ({response.StatusCode}): {json}");
        }

        return JsonSerializer.Deserialize<UserInfo>(json) ?? throw new InvalidOperationException("Failed to deserialize UserInfo");
    }

    public static Claims ParseClaims(string token)
    {
        string[] parts = token.Split('.');
        if (parts.Length != 3) throw new ArgumentException("Invalid JWT format");

        string payloadSegment = parts[1];
        int padding = 4 - (payloadSegment.Length % 4);
        if (padding < 4) payloadSegment += new string('=', padding);

        byte[] bytes = Convert.FromBase64String(payloadSegment.Replace('-', '+').Replace('_', '/'));
        string json = Encoding.UTF8.GetString(bytes);

        return JsonSerializer.Deserialize<Claims>(json) ?? throw new InvalidOperationException("Failed to deserialize Claims");
    }

    public Claims VerifyToken(string token)
    {
        var claims = ParseClaims(token);
        long now = DateTimeOffset.UtcNow.ToUnixTimeSeconds();
        if (claims.ExpiresAt > 0 && now > claims.ExpiresAt)
        {
            throw new InvalidOperationException("Token has expired");
        }
        return claims;
    }
}
