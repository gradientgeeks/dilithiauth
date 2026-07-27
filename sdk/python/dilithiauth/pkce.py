import base64
import hashlib
import secrets
from typing import Tuple

def generate_code_verifier(length: int = 64) -> str:
    """
    Generates a cryptographically secure PKCE code verifier.
    """
    token = secrets.token_bytes(length)
    return base64.urlsafe_b64encode(token).decode('utf-8').rstrip('=')

def generate_code_challenge(code_verifier: str) -> str:
    """
    Calculates the S256 code challenge for a given PKCE code verifier.
    """
    digest = hashlib.sha256(code_verifier.encode('utf-8')).digest()
    return base64.urlsafe_b64encode(digest).decode('utf-8').rstrip('=')

def generate_pkce_pair(length: int = 64) -> Tuple[str, str]:
    """
    Generates a tuple of (code_verifier, code_challenge).
    """
    verifier = generate_code_verifier(length)
    challenge = generate_code_challenge(verifier)
    return verifier, challenge

def generate_random_state(length: int = 32) -> str:
    """
    Generates a random state parameter to prevent CSRF attacks.
    """
    return secrets.token_urlsafe(length)
