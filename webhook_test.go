package fystack

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
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

func TestCanonicalJSONMatchesApexJCS(t *testing.T) {
	t.Parallel()
	value := map[string]any{"url": "https://x.io/hook?a=1&b=<2>", "amount": json.Number("1.50"), "big": json.Number("1e2")}
	canonical, err := CanonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"amount":1.5,"big":100,"url":"https://x.io/hook?a=1&b=<2>"}`
	if string(canonical) != want {
		t.Fatalf("canonical JSON = %s, want %s", canonical, want)
	}
	if _, err := CanonicalJSON([]any{1}); err == nil {
		t.Fatal("expected error for non-object input")
	}
}

func TestVerifyWebhookEd25519ApexVector(t *testing.T) {
	t.Parallel()
	// Delivered by Apex and shared with the TypeScript SDK test suite.
	event := WebhookEvent{
		WebhookID:  "73893532-8e3d-4bc7-b3f9-148c9b82b811",
		ResourceID: "c6702df0-b8ef-4a38-882d-dabf0d56127d",
		URL:        "https://webhook.site/56e92433-823d-4b28-9cc9-9298076af671",
		Payload:    json.RawMessage(`{"amount":"1000","asset":null,"asset_hold":null,"asset_hold_id":null,"asset_id":"f1be472d-1338-43ff-b547-06faba7ecce7","block_number":8960538,"created_at":"2025-08-11T17:36:12+07:00","description":"","direction":"in","from_address":"0xe6EBF81E9C225BbCEa9b5399E0D0d0f29f30f119","id":"c6702df0-b8ef-4a38-882d-dabf0d56127d","method":"transfer","network":null,"network_id":"23d13c03-1f3c-4e2a-9aef-1771a244b7f6","omitempty":null,"price_native_token":"4260.480000022453","price_token":"0.9997516371944777","status":"confirmed","to_address":"0xA118fa3d039BA6ad3BB0bA8ac8F243F727B2F99F","tx_fee":"0.0000515434459912","tx_hash":"0x75789cbbe4d55328e6be7b7155b4028ea8afa384ffe13a1146d8bdb3eed6ba22","type":"token_transfer","updated_at":"2025-08-11T17:36:28.532806+07:00","user":null,"user_id":"075243c1-962e-4803-83da-d73c78b8b060","wallet":null,"wallet_id":"e0851607-417a-475a-b399-5f3aa262b81d","workspace":null,"workspace_id":"6bcfe6c7-281b-4773-a188-e78f5cad3336"}`),
		Event:      "deposit.confirmed",
	}
	valid, err := VerifyWebhookEd25519(
		"503d18d8375e60667a4cdb879e72c249b3dc054e7d6443c5ccd93aef7f6547fb",
		event,
		"84bfb2defeb09c443204f2ab4002d983fc1873907613c4d63bad5a95eadf91e7ac7a29d0746f8a12295b0c3e54ea919c17c959a094e3da11529a29687a258204",
	)
	if err != nil || !valid {
		t.Fatalf("Apex vector validation = %v, %v", valid, err)
	}
}
