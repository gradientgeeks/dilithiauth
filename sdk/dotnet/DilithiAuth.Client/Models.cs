using System.Text.Json.Serialization;

namespace DilithiAuth.Client;

public class DilithiAuthOptions
{
    public required string Domain { get; set; }
    public required string ClientId { get; set; }
    public string? ClientSecret { get; set; }
    public string? RedirectUri { get; set; }
}

public class TokenResponse
{
    [JsonPropertyName("access_token")]
    public string AccessToken { get; set; } = string.Empty;

    [JsonPropertyName("id_token")]
    public string? IdToken { get; set; }

    [JsonPropertyName("token_type")]
    public string TokenType { get; set; } = "Bearer";

    [JsonPropertyName("expires_in")]
    public int ExpiresIn { get; set; }

    [JsonPropertyName("pqc_ref")]
    public string? PqcRef { get; set; }

    [JsonPropertyName("pqc_alg")]
    public string? PqcAlg { get; set; }

    [JsonPropertyName("refresh_token")]
    public string? RefreshToken { get; set; }
}

public class UserInfo
{
    [JsonPropertyName("sub")]
    public string Sub { get; set; } = string.Empty;

    [JsonPropertyName("name")]
    public string? Name { get; set; }

    [JsonPropertyName("email")]
    public string? Email { get; set; }
}

public class Claims
{
    [JsonPropertyName("iss")]
    public string? Issuer { get; set; }

    [JsonPropertyName("sub")]
    public string Subject { get; set; } = string.Empty;

    [JsonPropertyName("aud")]
    public string? Audience { get; set; }

    [JsonPropertyName("exp")]
    public long ExpiresAt { get; set; }

    [JsonPropertyName("pqc_ref")]
    public string? PqcRef { get; set; }
}
