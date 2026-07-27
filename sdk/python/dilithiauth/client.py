import base64
import json
import time
import urllib.parse
import urllib.request
from typing import Dict, Any, Optional
from .pkce import generate_pkce_pair, generate_random_state

class DilithiAuthException(Exception):
    """Base exception for DilithiAuth SDK errors."""
    pass

DEFAULT_DOMAIN = "https://dilithiauthapis.gradientgeeks.tech"

class DilithiAuthClient:
    """
    Official Python Client SDK for DilithiAuth Post-Quantum Identity Provider.
    """

    def __init__(
        self,
        client_id: str,
        domain: Optional[str] = DEFAULT_DOMAIN,
        client_secret: Optional[str] = None,
        redirect_uri: Optional[str] = None
    ):
        if not client_id:
            raise ValueError("DilithiAuthClient: client_id parameter is required")

        target_domain = domain or DEFAULT_DOMAIN
        self.domain = target_domain.rstrip('/')
        self.client_id = client_id
        self.client_secret = client_secret
        self.redirect_uri = redirect_uri

    def get_authorization_url(
        self,
        redirect_uri: Optional[str] = None,
        state: Optional[str] = None,
        code_challenge: Optional[str] = None,
        scope: str = "openid profile email"
    ) -> str:
        """
        Constructs the OAuth 2.0 PKCE authorization redirect URL.
        """
        target_redirect_uri = redirect_uri or self.redirect_uri
        if not target_redirect_uri:
            raise ValueError("get_authorization_url: redirect_uri must be specified")

        params = {
            "client_id": self.client_id,
            "redirect_uri": target_redirect_uri,
            "response_type": "code",
            "scope": scope,
        }

        if state:
            params["state"] = state
        if code_challenge:
            params["code_challenge"] = code_challenge
            params["code_challenge_method"] = "S256"

        return f"{self.domain}/authorize?{urllib.parse.urlencode(params)}"

    def exchange_code(
        self,
        code: str,
        code_verifier: str,
        redirect_uri: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Exchanges authorization code for Post-Quantum access token and refresh token.
        """
        target_redirect_uri = redirect_uri or self.redirect_uri
        if not target_redirect_uri:
            raise ValueError("exchange_code: redirect_uri must be specified")

        endpoint = f"{self.domain}/api/v1/oauth/token"
        payload = {
            "grant_type": "authorization_code",
            "client_id": self.client_id,
            "code": code,
            "redirect_uri": target_redirect_uri,
            "code_verifier": code_verifier,
        }

        encoded_payload = urllib.parse.urlencode(payload).encode('utf-8')
        req = urllib.request.Request(endpoint, data=encoded_payload, method="POST")
        req.add_header("Content-Type", "application/x-www-form-urlencoded")

        if self.client_secret:
            auth_str = f"{self.client_id}:{self.client_secret}"
            b64_auth = base64.b64encode(auth_str.encode('utf-8')).decode('utf-8')
            req.add_header("Authorization", f"Basic {b64_auth}")

        try:
            with urllib.request.urlopen(req) as resp:
                data = resp.read().decode('utf-8')
                return json.loads(data)
        except urllib.error.HTTPError as e:
            error_body = e.read().decode('utf-8')
            raise DilithiAuthException(f"Token exchange failed ({e.code}): {error_body}")
        except Exception as e:
            raise DilithiAuthException(f"Network request error: {str(e)}")

    def get_user_info(self, access_token: str) -> Dict[str, Any]:
        """
        Fetches claims for the authenticated user from /userinfo endpoint.
        """
        endpoint = f"{self.domain}/userinfo"
        req = urllib.request.Request(endpoint, method="GET")
        req.add_header("Authorization", f"Bearer {access_token}")

        try:
            with urllib.request.urlopen(req) as resp:
                data = resp.read().decode('utf-8')
                return json.loads(data)
        except urllib.error.HTTPError as e:
            error_body = e.read().decode('utf-8')
            raise DilithiAuthException(f"Userinfo request failed ({e.code}): {error_body}")
        except Exception as e:
            raise DilithiAuthException(f"Network request error: {str(e)}")

    @staticmethod
    def parse_claims(token: str) -> Dict[str, Any]:
        """
        Decodes payload claims from JWT token string without signature verification.
        """
        parts = token.split('.')
        if len(parts) != 3:
            raise DilithiAuthException("Invalid JWT token format")

        payload_segment = parts[1]
        # Pad base64url string
        padding = 4 - (len(payload_segment) % 4)
        if padding != 4:
            payload_segment += '=' * padding

        decoded_bytes = base64.urlsafe_b64decode(payload_segment)
        return json.loads(decoded_bytes.decode('utf-8'))

    def verify_token(self, token: str) -> Dict[str, Any]:
        """
        Decodes JWT token claims and checks expiration.
        """
        claims = self.parse_claims(token)
        exp = claims.get('exp')
        if exp and time.time() > exp:
            raise DilithiAuthException("Token has expired")
        return claims
