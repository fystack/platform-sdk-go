package fystack

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
)

func TestCanonicalJSONSortsNestedKeys(t *testing.T) {
	t.Parallel()
	value := map[string]any{"z": 1, "a": map[string]any{"y": 2, "b": 3}, "items": []any{map[string]any{"z": 4, "a": 5}}}
	canonical, err := CanonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":{"b":3,"y":2},"items":[{"a":5,"z":4}],"z":1}`
	if string(canonical) != want {
		t.Fatalf("canonical JSON = %s", canonical)
	}
}

func TestVerifyWebhookSignatures(t *testing.T) {
	t.Parallel()
	event := map[string]any{"webhook_id": "id", "payload": map[string]any{"amount": "1"}}
	hmacSignature, err := ComputeWebhookHMAC("secret", event)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := VerifyWebhookHMAC("secret", event, hmacSignature)
	if err != nil || !valid {
		t.Fatalf("HMAC validation = %v, %v", valid, err)
	}
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	canonical, _ := CanonicalJSON(event)
	signature := ed25519.Sign(privateKey, canonical)
	valid, err = VerifyWebhookEd25519(hex.EncodeToString(publicKey), event, hex.EncodeToString(signature))
	if err != nil || !valid {
		t.Fatalf("Ed25519 validation = %v, %v", valid, err)
	}
	valid, err = VerifyWebhookEd25519(hex.EncodeToString(publicKey), map[string]any{"tampered": true}, hex.EncodeToString(signature))
	if err != nil || valid {
		t.Fatalf("tampered validation = %v, %v", valid, err)
	}
}
