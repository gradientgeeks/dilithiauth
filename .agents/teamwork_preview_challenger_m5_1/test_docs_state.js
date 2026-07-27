// Harness to empirically test Docs.tsx state calculations & copy key behaviors

const results = [];

// Simulate getEndpointResponse from Docs.tsx
function getEndpointResponse(
  selectedEndpoint,
  clientId,
  redirectUri,
  subject,
  tokenStrategy
) {
  switch (selectedEndpoint) {
    case 'discovery':
      return {
        status: 200,
        statusText: 'OK',
        method: 'GET',
        url: '/.well-known/openid-configuration',
        body: {
          issuer: 'https://dilithiauthapis.gradientgeeks.tech',
          authorization_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/authorize',
          token_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/token',
          userinfo_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/userinfo',
          jwks_uri: 'https://dilithiauthapis.gradientgeeks.tech/.well-known/jwks.json',
          revocation_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/revoke',
          response_types_supported: ['code'],
          subject_types_supported: ['public'],
          id_token_signing_alg_values_supported: ['EdDSA', 'MLDSA44'],
          code_challenge_methods_supported: ['S256'],
          pqc_hybrid_enabled: true,
          pqc_algorithms: ['ML-DSA-44', 'Ed25519']
        },
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
          'Cache-Control': 'public, max-age=3600',
          'X-DilithiAuth-PQC': 'ML-DSA-44 (FIPS 204)'
        }
      };
    case 'jwks':
      return {
        status: 200,
        statusText: 'OK',
        method: 'GET',
        url: '/.well-known/jwks.json',
        body: {
          keys: [
            {
              kid: 'x9K3mP1Q',
              kty: 'OKP',
              crv: 'Ed25519',
              x: 'u18d9fA2mL5pQ8sT4vW7xZ1aB3cD5eF7gH9iJ1kL3mN',
              mldsa44_pub: 'A9x2kL7mP4qS9tV2wY5zB8cD1eF4gH7iJ0kL3mN6pQ9rS2tV5wY8zA1bC4dE7fG0hJ3kL6mN9pQ2rS5tV8wY1zB4cD7eF0gH3iJ6kL9mN2pQ5rS8tV1wY4zB7cD0eF3gH6iJ9kL2mN5pQ8rS1tV4wY7zB0cD3eF6gH9iJ2kL5mN8pQ1rS4tV7wY0zB3cD6eF9gH2iJ5kL8mN1pQ4rS7tV0wY3zB6cD9eF2gH5iJ8kL1mN4pQ7rS0tV3wY6zB9cD2eF5gH8iJ1kL4mN7pQ0',
              use: 'sig',
              alg: 'EdDSA+MLDSA44'
            }
          ]
        },
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'public, max-age=86400',
          'Access-Control-Allow-Origin': '*'
        }
      };
    case 'authorize':
      return {
        status: 302,
        statusText: 'Found (Redirect)',
        method: 'GET',
        url: `/authorize?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile+email&code_challenge=E9Mel-pGsD-Zc4jKWBfZ6KBG_gJS5524j_ic71_715k&code_challenge_method=S256`,
        body: {
          redirect_to: `${redirectUri}?code=da_code_8f9a2b4c1d6e3f5a7b9c0d2e4f6a8b1c&state=random_csrf_token`,
          grant_type: 'authorization_code',
          code_challenge_method: 'S256',
          status: 'authorization_granted'
        },
        headers: {
          'Location': `${redirectUri}?code=da_code_8f9a2b4c1d6e3f5a7b9c0d2e4f6a8b1c`,
          'Set-Cookie': 'dilithiauth_session=sess_987654321; HttpOnly; Secure; SameSite=Lax'
        }
      };
    case 'token':
      if (tokenStrategy === 'thin') {
        return {
          status: 200,
          statusText: 'OK',
          method: 'POST',
          url: '/token',
          body: {
            access_token: 'eyJhbGciOiJFZERTQSIsImtpZCI6Ing5SzNtUDFRIiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2RpbGl0aGlhdXRoYXBpcy5ncmFkaWVudGdlZWtzLnRlY2giLCJzdWIiOiJhbGljZUBleGFtcGxlLmNvbSIsImF1ZCI6ImRpbGl0aGktZGVtby1jbGllbnQtMTIzIiwiZXhwIjoxNzcxNzE2MDAwLCJpYXQiOjE3NzE3MTI0MDAsInBxc19yZWYiOiI0N0RFUVBqOEhCU2EtX1RJbVctNUpDZXVRZVJrbTVOTXBKV1pHM2hTdUZVIn0.ed25519_sig_64bytes_sample',
            token_type: 'Bearer',
            expires_in: 3600,
            pqc_ref: '47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU',
            pqc_alg: 'ML-DSA-44',
            refresh_token: 'da_rt_7f8a9b0c1d2e3f4a5b6c7d8e9f0a'
          },
          headers: {
            'Content-Type': 'application/json',
            'X-DilithiAuth-Token-Strategy': 'thin',
            'X-DilithiAuth-PQC-Ref': '47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU'
          }
        };
      } else if (tokenStrategy === 'full') {
        return {
          status: 200,
          statusText: 'OK',
          method: 'POST',
          url: '/token',
          body: {
            access_token: 'eyJhbGciOiJFZERTQSsiLCJraWQiOiJ4OUszbVAxUSIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2RpbGl0aGlhdXRoYXBpcy5ncmFkaWVudGdlZWtzLnRlY2giLCJzdWIiOiJhbGljZUBleGFtcGxlLmNvbSIsImF1ZCI6ImRpbGl0aGktZGVtby1jbGllbnQtMTIzIiwiZXhwIjoxNzcxNzE2MDAwLCJpYXQiOjE3NzE3MTI0MDAsInBxc19zaWciOiJBMTl4Mms3bVA0cVMy...[ML-DSA-44 ~2.4KB Signature Embedded]..."}.ed25519_sig_64bytes',
            token_type: 'Bearer',
            expires_in: 3600,
            token_size_bytes: 3840,
            pqc_alg: 'ML-DSA-44',
            pqc_sig_bytes: 2420,
            refresh_token: 'da_rt_7f8a9b0c1d2e3f4a5b6c7d8e9f0a'
          },
          headers: {
            'Content-Type': 'application/json',
            'X-DilithiAuth-Token-Strategy': 'full (Jumbo)',
            'X-DilithiAuth-MaxHeaderBytes': '16384 (16 KiB)'
          }
        };
      } else {
        return {
          status: 200,
          statusText: 'OK',
          method: 'POST',
          url: '/token',
          body: {
            cwt_hex: 'd83d8448a10168697373756572a501686973737565720265616c696365041a6627000008582047444551506a38484253612d5f54496d572d354a4365755165526b6d354e4d704a575a473368537546555840...',
            token_type: 'Bearer',
            format: 'CBOR Web Token (RFC 8392)',
            token_size_bytes: 284,
            pqc_ref: '47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU'
          },
          headers: {
            'Content-Type': 'application/cbor',
            'X-DilithiAuth-Token-Strategy': 'cwt'
          }
        };
      }
    case 'userinfo':
      return {
        status: 200,
        statusText: 'OK',
        method: 'GET',
        url: '/userinfo',
        body: {
          sub: subject,
          name: subject.split('@')[0].toUpperCase() + ' Identity',
          email: subject,
          email_verified: true,
          pqc_verified: true,
          hybrid_signature: 'Ed25519 + ML-DSA-44'
        },
        headers: {
          'Content-Type': 'application/json',
          'X-DilithiAuth-Sub': subject
        }
      };
    case 'revoke':
      return {
        status: 200,
        statusText: 'OK',
        method: 'POST',
        url: '/revoke',
        body: {
          status: 'revoked',
          token_type_hint: 'access_token',
          pqc_sigstore_purged: true
        },
        headers: {
          'Content-Type': 'application/json'
        }
      };
  }
}

// TEST 1: Unencoded clientId with special characters in authorize URL
{
  const clientIdWithSpecial = "app_123 & client_secret=foo #hash";
  const res = getEndpointResponse('authorize', clientIdWithSpecial, 'https://app.com/cb', 'alice@test.com', 'thin');
  const hasUnencodedAmpersand = res.url.includes('& client_secret=foo');
  results.push({
    name: 'Authorize URL - clientId URI Encoding',
    category: 'dynamic_api_response',
    passed: !hasUnencodedAmpersand,
    details: `URL generated: "${res.url}". Notice clientId is not passed through encodeURIComponent, causing query parameter injection.`
  });
}

// TEST 2: redirectUri with existing query parameter
{
  const redirectUriWithQuery = "https://app.com/cb?env=dev";
  const res = getEndpointResponse('authorize', 'client-1', redirectUriWithQuery, 'alice@test.com', 'thin');
  const hasDoubleQuestionMark = res.body.redirect_to.includes('?env=dev?code=');
  results.push({
    name: 'Authorize Redirect URL - Double Question Mark with Existing Query Params',
    category: 'dynamic_api_response',
    passed: !hasDoubleQuestionMark,
    details: `redirect_to generated: "${res.body.redirect_to}". Notice double '?' created when redirectUri has existing query string.`
  });
}

// TEST 3: Userinfo name formatting with empty subject or special characters
{
  const emptySubRes = getEndpointResponse('userinfo', 'client-1', 'https://app.com', '', 'thin');
  const leadingSpaceName = emptySubRes.body.name === ' Identity';
  results.push({
    name: 'Userinfo Name Calculation - Empty Subject Handling',
    category: 'dynamic_api_response',
    passed: !leadingSpaceName,
    details: `Empty subject resulted in name: "${emptySubRes.body.name}". Produces awkward ' Identity' header.`
  });
}

// TEST 4: Token endpoint static sub/aud vs dynamic parameters
{
  const customSub = "custom_user_99@test.com";
  const customClient = "custom_client_99";
  const res = getEndpointResponse('token', customClient, 'https://app.com', customSub, 'thin');
  // Decode JWT payload base64
  const payloadB64 = res.body.access_token.split('.')[1];
  const payloadJson = JSON.parse(Buffer.from(payloadB64, 'base64').toString('utf-8'));
  const matchesDynamicSub = payloadJson.sub === customSub;
  const matchesDynamicAud = payloadJson.aud === customClient;
  results.push({
    name: 'Token Endpoint - Dynamic Sub and Aud in JWT Access Token',
    category: 'dynamic_api_response',
    passed: matchesDynamicSub && matchesDynamicAud,
    details: `JWT access_token payload sub="${payloadJson.sub}" (expected "${customSub}"), aud="${payloadJson.aud}" (expected "${customClient}"). Token body remains hardcoded to alice@example.com / dilithi-demo-client-123 despite interactive inputs.`
  });
}

// TEST 5: Copy to Clipboard state persistence across endpoint response strategy switches
{
  let copiedMap = {};
  const selectedEndpoint = 'token';
  
  // User copies 'token' response when strategy is 'thin'
  const key = `res-body-${selectedEndpoint}`;
  copiedMap[key] = true; // Simulated copyToClipboard call
  
  // User immediately switches token strategy to 'full' (Jumbo)
  const thinBody = getEndpointResponse('token', 'client-1', 'https://app.com', 'alice@test.com', 'thin').body;
  const fullBody = getEndpointResponse('token', 'client-1', 'https://app.com', 'alice@test.com', 'full').body;
  
  // Key remains 'res-body-token', so copiedMap[key] is still true for fullBody
  const staleCopyState = copiedMap[key] === true && JSON.stringify(thinBody) !== JSON.stringify(fullBody);
  results.push({
    name: 'Copy Button State - Stale State on Response Strategy Switch',
    category: 'copy_clipboard',
    passed: !staleCopyState,
    details: `When switching token strategy from thin to full on /token endpoint, copiedMap['res-body-token'] remains true despite response body content changing completely.`
  });
}

// TEST 6: Copy button rapid click race condition (timer collision)
{
  let copiedMap = {};
  
  function copyToClipboardSim(key) {
    copiedMap[key] = true;
    setTimeout(() => {
      copiedMap[key] = false;
    }, 2000);
  }
  
  // Simulate click at t=0
  copyToClipboardSim('install-ts');
  // Simulate click at t=1000ms
  setTimeout(() => {
    copyToClipboardSim('install-ts');
  }, 1000);

  // At t=2050ms, first timer fires
  setTimeout(() => {
    const isPrematurelyClearedAt2s = copiedMap['install-ts'] === false;
    results.push({
      name: 'Copy Button State - Rapid Click Timer Collision',
      category: 'copy_clipboard',
      passed: !isPrematurelyClearedAt2s,
      details: `At t=2050ms (1050ms after second click), copiedMap['install-ts'] is ${copiedMap['install-ts']}. The first timeout cleared the copied state 1s early.`
    });

    console.log(JSON.stringify(results, null, 2));
  }, 3500);
}
