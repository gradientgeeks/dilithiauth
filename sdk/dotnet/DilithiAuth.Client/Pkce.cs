using System;
using System.Security.Cryptography;
using System.Text;

namespace DilithiAuth.Client;

public record PkcePair(string CodeVerifier, string CodeChallenge);

public static class Pkce
{
    private static string Base64UrlEncode(byte[] buffer)
    {
        return Convert.ToBase64String(buffer)
            .Replace("+", "-")
            .Replace("/", "_")
            .TrimEnd('=');
    }

    public static PkcePair GeneratePair()
    {
        byte[] bytes = new byte[32];
        using (var rng = RandomNumberGenerator.Create())
        {
            rng.GetBytes(bytes);
        }

        string verifier = Base64UrlEncode(bytes);
        string challenge = CalculateChallenge(verifier);

        return new PkcePair(verifier, challenge);
    }

    public static string CalculateChallenge(string codeVerifier)
    {
        byte[] bytes = Encoding.UTF8.GetBytes(codeVerifier);
        using var sha256 = SHA256.Create();
        byte[] hash = sha256.ComputeHash(bytes);
        return Base64UrlEncode(hash);
    }

    public static string GenerateState()
    {
        byte[] bytes = new byte[16];
        using (var rng = RandomNumberGenerator.Create())
        {
            rng.GetBytes(bytes);
        }
        return Base64UrlEncode(bytes);
    }
}
