package fystack

import "encoding/json"

// Credentials configures request authentication. Signer takes precedence over APISecret.
type Credentials struct {
	APIKey    string
	APISecret string
	Signer    RequestSigner
	AuthToken string
}

// WebhookEvent is the canonical webhook envelope.
type WebhookEvent struct {
	WebhookID  string          `json:"webhook_id"`
	ResourceID string          `json:"resource_id"`
	URL        string          `json:"url"`
	Payload    json.RawMessage `json:"payload"`
	Event      string          `json:"event"`
}

// WebhookPublicKeyResponse contains a base64-encoded Ed25519 public key.
type WebhookPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}
