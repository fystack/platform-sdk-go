package fystack

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// RequestSigner signs the canonical API request string and returns the
// base64-encoded raw signature used in ACCESS-SIGN.
type RequestSigner interface {
	Sign(ctx context.Context, canonical string) (string, error)
}

// LocalPrivateKeySigner signs requests with an in-process PKCS8 Ed25519 key.
type LocalPrivateKeySigner struct {
	privateKey ed25519.PrivateKey
}

// NewLocalPrivateKeySigner parses a PEM-encoded PKCS8 Ed25519 private key.
func NewLocalPrivateKeySigner(privateKeyPEM []byte) (*LocalPrivateKeySigner, error) {
	if len(strings.TrimSpace(string(privateKeyPEM))) == 0 {
		return nil, errors.New("private key PEM is required")
	}
	block, rest := pem.Decode(privateKeyPEM)
	if block == nil || len(strings.TrimSpace(string(rest))) != 0 {
		return nil, errors.New("invalid private key PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private key must be Ed25519")
	}
	return &LocalPrivateKeySigner{privateKey: privateKey}, nil
}

// Sign signs a canonical request string.
func (s *LocalPrivateKeySigner) Sign(_ context.Context, canonical string) (string, error) {
	signature := ed25519.Sign(s.privateKey, []byte(canonical))
	return base64.StdEncoding.EncodeToString(signature), nil
}

// CanonicalString builds the request signing input in wire-order.
func CanonicalString(method, path, timestamp string, body []byte) string {
	return "method=" + method + "&path=" + path + "&timestamp=" + timestamp + "&body=" + string(body)
}

// ComputeHMAC returns the lowercase hex HMAC-SHA256 digest used by Fystack.
func ComputeHMAC(apiSecret, canonical string) string {
	digest := hmac.New(sha256.New, []byte(apiSecret))
	_, _ = digest.Write([]byte(canonical))
	return hex.EncodeToString(digest.Sum(nil))
}

func hmacAccessSignature(apiSecret, canonical string) string {
	hexDigest := ComputeHMAC(apiSecret, canonical)
	return base64.StdEncoding.EncodeToString([]byte(hexDigest))
}
