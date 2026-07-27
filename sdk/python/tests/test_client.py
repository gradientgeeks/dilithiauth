import base64
import json
import time
import unittest
from dilithiauth.client import DilithiAuthClient
from dilithiauth.pkce import generate_pkce_pair, generate_code_challenge

class TestDilithiAuthSDK(unittest.TestCase):

    def test_pkce_generation(self):
        verifier, challenge = generate_pkce_pair()
        self.assertTrue(len(verifier) >= 43)
        self.assertTrue(len(challenge) > 0)
        self.assertEqual(generate_code_challenge(verifier), challenge)

    def test_authorization_url(self):
        client = DilithiAuthClient(
            domain="https://auth.example.com",
            client_id="test_client_id",
            redirect_uri="https://myapp.com/callback"
        )
        url = client.get_authorization_url(state="xyz123", code_challenge="challenge123")
        self.assertIn("client_id=test_client_id", url)
        self.assertIn("response_type=code", url)
        self.assertIn("code_challenge=challenge123", url)
        self.assertIn("code_challenge_method=S256", url)

    def test_parse_claims(self):
        header = base64.urlsafe_b64encode(json.dumps({"alg": "Ed25519+ML-DSA-44"}).encode()).decode().rstrip('=')
        payload = base64.urlsafe_b64encode(json.dumps({
            "sub": "user_123",
            "exp": time.time() + 3600,
            "pqc_ref": "sha256:abc123ref"
        }).encode()).decode().rstrip('=')

        raw_jwt = f"{header}.{payload}.dummy_signature"

        claims = DilithiAuthClient.parse_claims(raw_jwt)
        self.assertEqual(claims["sub"], "user_123")
        self.assertEqual(claims["pqc_ref"], "sha256:abc123ref")

if __name__ == '__main__':
    unittest.main()
