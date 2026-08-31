package fystack

import (
	"context"
	"net/http"
	"net/url"
)

// GetWebhookPublicKey returns the webhook verification key for a workspace.
func (c *Client) GetWebhookPublicKey(ctx context.Context, workspaceID string) (WebhookPublicKeyResponse, error) {
	if err := validateUUID(workspaceID, "workspace ID"); err != nil {
		return WebhookPublicKeyResponse{}, err
	}
	var result WebhookPublicKeyResponse
	err := c.do(ctx, http.MethodGet, "/workspaces/"+url.PathEscape(workspaceID)+"/webhook-verification-key", nil, nil, &result)
	return result, err
}
