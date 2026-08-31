# API Key Request Signing

The SDK signs requests automatically. This document describes the wire format
for interoperability and troubleshooting.

## Headers

Signed requests contain:

| Header | Value |
| --- | --- |
| `ACCESS-API-KEY` | API key ID |
| `ACCESS-TIMESTAMP` | Unix timestamp in seconds |
| `ACCESS-SIGN` | Base64-encoded scheme-specific signature |

The timestamp must be within five minutes of server time.

## Canonical request

Both schemes sign the same byte string:

```text
method=<HTTP_METHOD>&path=<HTTP_PATH>&timestamp=<TIMESTAMP>&body=<RAW_BODY>
```

The path includes `/api/v1` and excludes the query string. The body is the
exact byte sequence sent on the wire. Requests without a body use an empty
string.

```text
method=POST&path=/api/v1/workspaces/abc-123/invite&timestamp=1735689600&body={"email":"a@b.com"}
```

## HMAC-SHA256

The API secret is used as its literal UTF-8 string. It is not hex-decoded.

1. Compute HMAC-SHA256 over the canonical request.
2. Encode the digest as lowercase hexadecimal.
3. Base64-encode the UTF-8 hexadecimal string.

```text
ACCESS-SIGN = base64(hex(HMAC-SHA256(api_secret, canonical_request)))
```

Configure this scheme with `Credentials.APISecret`.

## Ed25519

Sign the canonical request bytes directly with Ed25519, then base64-encode the
raw 64-byte signature:

```text
ACCESS-SIGN = base64(Ed25519-Sign(private_key, canonical_request))
```

Use `fystack.NewLocalPrivateKeySigner` for a PEM PKCS8 private key, or
`awskms.New` from `github.com/fystack/platform-sdk-go/signer/awskms` for an AWS
KMS `ECC_NIST_EDWARDS25519` `SIGN_VERIFY` key.
The KMS signer sends `MessageType=RAW` and
`SigningAlgorithm=ED25519_SHA_512`; the `PH` variant is not used.

`Credentials.Signer` takes precedence if both a signer and API secret are
provided.

## Key rotation

For an Ed25519 API key, register `new_public_key`, switch the client to the new
private key, and make a signed request. After the first successful request
with the new key, the server promotes it and retires the previous key.

## Additional server checks

The server may also enforce an IP allowlist, key expiration, disabled status,
and revocation status.
