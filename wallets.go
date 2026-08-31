package fystack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/fystack/platform-sdk-go/internal/poller"
)

type createWalletRequest struct {
	Name                  string           `json:"name"`
	WalletType            WalletType       `json:"wallet_type"`
	WalletPurpose         WalletPurpose    `json:"wallet_purpose,omitempty"`
	SweepTaskParams       *SweepTaskParams `json:"sweep_task_params,omitempty"`
	SweepTaskID           string           `json:"sweep_task_id,omitempty"`
	NetworkIDs            []string         `json:"network_ids,omitempty"`
	GasSponsorshipEnabled *bool            `json:"gas_sponsorship_enabled,omitempty"`
	ApprovalGroupID       string           `json:"approval_group_id,omitempty"`
}

// GetWallets returns all wallets in the configured workspace.
func (c *Client) GetWallets(ctx context.Context) ([]Wallet, error) {
	if c.workspaceID == "" {
		return nil, errors.New("workspace ID is required")
	}
	var wallets []Wallet
	err := c.do(ctx, http.MethodGet, "/workspaces/"+url.PathEscape(c.workspaceID)+"/wallets", nil, nil, &wallets)
	return wallets, err
}

// CreateWallet creates a wallet and waits for provisioning to finish.
func (c *Client) CreateWallet(ctx context.Context, params CreateWalletParams) (CreateWalletResponse, error) {
	result, err := c.CreateWalletAsync(ctx, params)
	if err != nil || result.Status != WalletCreationStatusPending {
		return result, err
	}
	status, err := poller.Poll(ctx, c.pollerOptions,
		func(ctx context.Context) (WalletCreationStatusResponse, error) {
			c.log("polling wallet creation status", "wallet_id", result.WalletID)
			return c.GetWalletCreationStatus(ctx, result.WalletID)
		},
		func(status WalletCreationStatusResponse) bool {
			return status.Status == WalletCreationStatusSuccess || status.Status == WalletCreationStatusError
		}, nil,
	)
	return status, err
}

// CreateWalletAsync creates a wallet and returns its initial provisioning state.
func (c *Client) CreateWalletAsync(ctx context.Context, params CreateWalletParams) (CreateWalletResponse, error) {
	if params.WalletType == "" {
		params.WalletType = WalletTypeHyper
	}
	if err := validateCreateWallet(params); err != nil {
		return CreateWalletResponse{}, err
	}
	body := createWalletRequest(params)
	var result CreateWalletResponse
	err := c.do(ctx, http.MethodPost, "/wallets", body, nil, &result)
	return result, err
}

// GetWalletCreationStatus returns the latest wallet provisioning state.
func (c *Client) GetWalletCreationStatus(ctx context.Context, walletID string) (WalletCreationStatusResponse, error) {
	var result WalletCreationStatusResponse
	err := c.do(ctx, http.MethodGet, "/wallets/creation-status/"+url.PathEscape(walletID), nil, nil, &result)
	return result, err
}

// GetWalletAssets returns the assets associated with a wallet.
func (c *Client) GetWalletAssets(ctx context.Context, walletID string) ([]WalletAsset, error) {
	if strings.TrimSpace(walletID) == "" {
		return nil, errors.New("invalid wallet ID provided")
	}
	var result []WalletAsset
	err := c.do(ctx, http.MethodGet, "/wallets/"+url.PathEscape(walletID)+"/assets", nil, nil, &result)
	return result, err
}

// GetDepositAddress returns a wallet deposit address for an address family.
func (c *Client) GetDepositAddress(ctx context.Context, walletID string, addressType AddressType) (DepositAddressResponse, error) {
	if err := validateUUID(walletID, "wallet ID"); err != nil {
		return DepositAddressResponse{}, err
	}
	if addressType != AddressTypeEVM && addressType != AddressTypeBitcoin && addressType != AddressTypeSolana && addressType != AddressTypeTron {
		return DepositAddressResponse{}, fmt.Errorf("invalid address type %q: must be evm, btc, sol, or tron", addressType)
	}
	query := make(url.Values)
	query.Set("address_type", string(addressType))
	endpoint := "/wallets/" + url.PathEscape(walletID) + "/deposit-address?" + query.Encode()
	var result DepositAddressResponse
	err := c.do(ctx, http.MethodGet, endpoint, nil, nil, &result)
	return result, err
}

// GetDepositAddressByAssetID returns the network-specific deposit address for
// an asset. For Bitcoin, Apex uses the asset network to select mainnet or
// testnet Bech32 encoding.
func (c *Client) GetDepositAddressByAssetID(ctx context.Context, walletID, assetID string) (DepositAddressResponse, error) {
	if err := validateUUID(walletID, "wallet ID"); err != nil {
		return DepositAddressResponse{}, err
	}
	if err := validateUUID(assetID, "asset ID"); err != nil {
		return DepositAddressResponse{}, err
	}
	query := make(url.Values)
	query.Set("asset_id", assetID)
	endpoint := "/wallets/" + url.PathEscape(walletID) + "/deposit-address?" + query.Encode()
	var result DepositAddressResponse
	err := c.do(ctx, http.MethodGet, endpoint, nil, nil, &result)
	return result, err
}

// RescanTransaction requests a network transaction rescan.
func (c *Client) RescanTransaction(ctx context.Context, params RescanTransactionParams) error {
	if err := validateUUID(params.NetworkID, "network ID"); err != nil {
		return err
	}
	if strings.TrimSpace(params.TxHash) == "" {
		return errors.New("invalid transaction hash provided")
	}
	body := struct {
		TxHash    string `json:"tx_hash"`
		NetworkID string `json:"network_id"`
	}{params.TxHash, params.NetworkID}
	return c.do(ctx, http.MethodPost, "/networks/rescan-transaction", body, nil, nil)
}

// RequestWithdrawal requests a withdrawal from a wallet.
func (c *Client) RequestWithdrawal(ctx context.Context, walletID string, params RequestWithdrawalParams) (RequestWithdrawalResponse, error) {
	return c.RequestWithdrawalWithOptions(ctx, walletID, params, RequestOptions{})
}

// RequestWithdrawalWithOptions requests a withdrawal with request-level options.
func (c *Client) RequestWithdrawalWithOptions(ctx context.Context, walletID string, params RequestWithdrawalParams, options RequestOptions) (RequestWithdrawalResponse, error) {
	if err := validateWithdrawal(walletID, params); err != nil {
		return RequestWithdrawalResponse{}, err
	}
	body := struct {
		AssetID          string       `json:"asset_id,omitempty"`
		Network          string       `json:"network,omitempty"`
		Asset            string       `json:"asset,omitempty"`
		Amount           string       `json:"amount"`
		RecipientAddress string       `json:"recipient_address"`
		Notes            string       `json:"notes,omitempty"`
		SkipBalanceCheck *bool        `json:"skip_balance_check,omitempty"`
		CategoryIDs      []string     `json:"category_ids,omitempty"`
		FeePriority      *FeePriority `json:"fee_priority,omitempty"`
	}{params.AssetID, params.Network, params.Asset, params.Amount, params.RecipientAddress, params.Notes, params.SkipBalanceCheck, params.CategoryIDs, params.FeePriority}
	var result RequestWithdrawalResponse
	err := c.do(ctx, http.MethodPost, "/wallets/"+url.PathEscape(walletID)+"/request-withdrawal", body, requestHeaders(options), &result)
	return result, err
}

// GetWithdrawalStatus returns the latest withdrawal details.
func (c *Client) GetWithdrawalStatus(ctx context.Context, walletID, withdrawalID string) (Withdrawal, error) {
	if err := validateUUID(walletID, "wallet ID"); err != nil {
		return Withdrawal{}, err
	}
	if err := validateUUID(withdrawalID, "withdrawal ID"); err != nil {
		return Withdrawal{}, err
	}
	var result Withdrawal
	err := c.do(ctx, http.MethodGet, "/wallets/"+url.PathEscape(walletID)+"/withdrawals/"+url.PathEscape(withdrawalID), nil, nil, &result)
	return result, err
}

func validateCreateWallet(params CreateWalletParams) error {
	if strings.TrimSpace(params.Name) == "" || len(params.Name) > 100 {
		return errors.New("wallet name is required and must not exceed 100 characters")
	}
	if params.WalletType != WalletTypeHyper && params.WalletType != WalletTypeMPC {
		return fmt.Errorf("invalid wallet type %q: must be standard or mpc", params.WalletType)
	}
	if params.WalletPurpose != "" && params.WalletPurpose != WalletPurposeGeneral && params.WalletPurpose != WalletPurposeGasTank && params.WalletPurpose != WalletPurposeUser && params.WalletPurpose != WalletPurposePayment && params.WalletPurpose != WalletPurposeOneTimeUse {
		return fmt.Errorf("invalid wallet purpose %q", params.WalletPurpose)
	}
	for _, networkID := range params.NetworkIDs {
		if err := validateUUID(networkID, "network ID"); err != nil {
			return err
		}
	}
	if params.SweepTaskID != "" {
		if err := validateUUID(params.SweepTaskID, "sweep task ID"); err != nil {
			return err
		}
	}
	if params.ApprovalGroupID != "" {
		if err := validateUUID(params.ApprovalGroupID, "approval group ID"); err != nil {
			return err
		}
	}
	return nil
}

func validateWithdrawal(walletID string, params RequestWithdrawalParams) error {
	if err := validateUUID(walletID, "wallet ID"); err != nil {
		return err
	}
	if params.AssetID != "" {
		if err := validateUUID(params.AssetID, "asset ID"); err != nil {
			return err
		}
	} else if params.Asset == "" || params.Network == "" {
		return errors.New("either asset ID or both asset and network must be provided")
	}
	if strings.TrimSpace(params.Amount) == "" {
		return errors.New("invalid amount provided")
	}
	if strings.TrimSpace(params.RecipientAddress) == "" {
		return errors.New("invalid recipient address provided")
	}
	if len(params.RecipientAddress) > 256 {
		return errors.New("recipient address exceeds maximum length of 256 characters")
	}
	if len(params.Notes) > 500 {
		return errors.New("notes exceed maximum length of 500 characters")
	}
	for _, categoryID := range params.CategoryIDs {
		if err := validateUUID(categoryID, "category ID"); err != nil {
			return err
		}
	}
	if params.FeePriority != nil && *params.FeePriority != FeePriorityLow && *params.FeePriority != FeePriorityStandard && *params.FeePriority != FeePriorityFast {
		return fmt.Errorf("invalid fee priority %q: must be low, standard, or fast", *params.FeePriority)
	}
	return nil
}

func validateUUID(value, name string) error {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || value[14] != '4' || !strings.ContainsRune("89abAB", rune(value[19])) {
		return fmt.Errorf("invalid %s format: must be a valid UUID v4", name)
	}
	for index, char := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !strings.ContainsRune("0123456789abcdefABCDEF", char) {
			return fmt.Errorf("invalid %s format: must be a valid UUID v4", name)
		}
	}
	return nil
}
