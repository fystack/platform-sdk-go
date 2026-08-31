package fystack

import (
	"bytes"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// CanonicalJSON recursively orders object keys and returns compact JSON.
func CanonicalJSON(value any) ([]byte, error) {
	input, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode JSON: %w", err)
	}
	var normalized any
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()
	if err := decoder.Decode(&normalized); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}
	if _, ok := normalized.(map[string]any); !ok {
		return nil, errors.New("input must be a non-null object")
	}
	return json.Marshal(normalized)
}

// ComputeWebhookHMAC returns the lowercase hex HMAC-SHA256 webhook signature.
func ComputeWebhookHMAC(apiSecret string, event any) (string, error) {
	canonical, err := CanonicalJSON(event)
	if err != nil {
		return "", err
	}
	digest := hmac.New(sha256.New, []byte(apiSecret))
	_, _ = digest.Write(canonical)
	return hex.EncodeToString(digest.Sum(nil)), nil
}

// VerifyWebhookHMAC verifies a lowercase hex HMAC webhook signature in constant time.
func VerifyWebhookHMAC(apiSecret string, event any, signatureHex string) (bool, error) {
	expected, err := ComputeWebhookHMAC(apiSecret, event)
	if err != nil {
		return false, err
	}
	expectedBytes, _ := hex.DecodeString(expected)
	actualBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, nil
	}
	return hmac.Equal(expectedBytes, actualBytes), nil
}

// VerifyWebhookEd25519 verifies a hex-encoded Ed25519 webhook signature.
func VerifyWebhookEd25519(publicKeyHex string, event any, signatureHex string) (bool, error) {
	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return false, errors.New("public key must be a 32-byte hex-encoded Ed25519 key")
	}
	signature, err := hex.DecodeString(signatureHex)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return false, nil
	}
	canonical, err := CanonicalJSON(event)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(publicKey, canonical, signature), nil
}
