package fystack

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// CreateSweepTask creates an automation sweep task in the configured workspace.
func (c *Client) CreateSweepTask(ctx context.Context, params CreateSweepTaskParams) (SweepTask, error) {
	if c.workspaceID == "" {
		return SweepTask{}, errors.New("workspace ID is required")
	}
	if err := validateUUID(params.DestinationWalletID, "destination wallet ID"); err != nil {
		return SweepTask{}, err
	}
	for _, walletID := range params.WalletIDs {
		if err := validateUUID(walletID, "wallet ID"); err != nil {
			return SweepTask{}, err
		}
	}
	for _, assetID := range params.AssetIDs {
		if err := validateUUID(assetID, "asset ID"); err != nil {
			return SweepTask{}, err
		}
	}
	body := struct {
		Name                 string          `json:"name"`
		Strategy             SweepStrategy   `json:"strategy"`
		MinTriggerValueUSD   float64         `json:"min_trigger_value_usd"`
		DestinationWalletID  string          `json:"destination_wallet_id"`
		FrequencyInSeconds   int64           `json:"frequency_in_seconds"`
		WalletIDs            []string        `json:"wallet_ids"`
		SweepType            SweepType       `json:"sweep_type,omitempty"`
		DestinationType      DestinationType `json:"destination_type,omitempty"`
		AssetIDs             []string        `json:"asset_ids,omitempty"`
		ReserveType          ReserveType     `json:"reserve_type,omitempty"`
		ReserveAmountUSD     *float64        `json:"reserve_amount_usd,omitempty"`
		ReservePercentageUSD *float64        `json:"reserve_percentage_usd,omitempty"`
		Enabled              *bool           `json:"enabled,omitempty"`
	}{params.Name, params.Strategy, params.MinTriggerValueUSD, params.DestinationWalletID, params.FrequencyInSeconds, params.WalletIDs, params.SweepType, params.DestinationType, params.AssetIDs, params.ReserveType, params.ReserveAmountUSD, params.ReservePercentageUSD, params.Enabled}
	var result SweepTask
	err := c.do(ctx, http.MethodPost, "/workspaces/"+url.PathEscape(c.workspaceID)+"/automation/sweep-task", body, nil, &result)
	return result, err
}
