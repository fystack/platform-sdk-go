package fystack

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestCanonicalStringAndHMAC(t *testing.T) {
	t.Parallel()
	canonical := CanonicalString("GET", "/api/v1/web3/wallet-detail", "1706783229", nil)
	if canonical != "method=GET&path=/api/v1/web3/wallet-detail&timestamp=1706783229&body=" {
		t.Fatalf("unexpected canonical string: %q", canonical)
	}
	digest := ComputeHMAC("c3c6aa8121b84d3aae4a4558956ee171", canonical)
	if digest != "f8dea93ca5272d175244c0c2b853f1fd8d4d01f4ea99abf6d2e62dde853fd9a6" {
		t.Fatalf("unexpected digest: %s", digest)
	}
}

func TestLocalPrivateKeySigner(t *testing.T) {
	t.Parallel()
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := NewLocalPrivateKeySigner(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	if err != nil {
		t.Fatal(err)
	}
	canonical := "method=POST&path=/api/v1/wallets&timestamp=1&body={}"
	encoded, err := signer.Sign(context.Background(), canonical)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(publicKey, []byte(canonical), signature) {
		t.Fatal("signature did not verify")
	}
}

func TestNewLocalPrivateKeySignerRejectsWrongKeyType(t *testing.T) {
	t.Parallel()
	if _, err := NewLocalPrivateKeySigner([]byte("not pem")); err == nil {
		t.Fatal("expected invalid PEM error")
	}
}
