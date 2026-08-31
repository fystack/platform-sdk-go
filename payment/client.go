// Package payment provides a client for Fystack checkout APIs.
package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	fystack "github.com/fystack/platform-sdk-go"
)

// Status is the state of a checkout payment.
type Status string

const (
	StatusPending    Status = "Pending"
	StatusConfirming Status = "Confirming"
	StatusSuccess    Status = "Success"
)

// AcceptedAsset identifies an accepted asset and network by code.
type AcceptedAsset struct {
	Asset   string `json:"asset"`
	Network string `json:"network"`
}

// CreateCheckoutParams configures a checkout.
type CreateCheckoutParams struct {
	Price                 string          `json:"price"`
	Currency              string          `json:"currency"`
	SupportedAssets       []string        `json:"supported_assets,omitempty"`
	AcceptedAssets        []AcceptedAsset `json:"accepted_assets,omitempty"`
	UnderpaidCoverPercent *float64        `json:"underpaid_cover_percent,omitempty"`
	Description           string          `json:"description,omitempty"`
	SuccessURL            string          `json:"success_url,omitempty"`
	CancelURL             string          `json:"cancel_url,omitempty"`
	ProductID             string          `json:"product_id,omitempty"`
	CustomerID            string          `json:"customer_id,omitempty"`
	OrderID               string          `json:"order_id,omitempty"`
	EnableLocalization    *bool           `json:"enable_localization,omitempty"`
	DestinationWalletID   string          `json:"destination_wallet_id"`
	FromAddress           string          `json:"from_address,omitempty"`
	ExpiryDurationSeconds int64           `json:"expiry_duration_seconds"`
}

// CreatePaymentParams configures a checkout payment.
type CreatePaymentParams struct {
	CustomerEmail string `json:"customer_email,omitempty"`
	PayAssetID    string `json:"pay_asset_id"`
}

// SupportedAsset describes an asset accepted by a checkout.
type SupportedAsset struct {
	ID            string                      `json:"id"`
	CreatedAt     string                      `json:"created_at"`
	UpdatedAt     string                      `json:"updated_at"`
	Name          string                      `json:"name"`
	Symbol        string                      `json:"symbol"`
	Decimals      int                         `json:"decimals"`
	LogoURL       string                      `json:"logo_url"`
	IsNative      bool                        `json:"is_native"`
	AddressType   string                      `json:"address_type"`
	IsWhitelisted bool                        `json:"is_whitelisted"`
	Address       string                      `json:"address,omitempty"`
	NetworkID     string                      `json:"network_id"`
	Network       *fystack.WalletAssetNetwork `json:"network,omitempty"`
}

// Checkout contains checkout configuration and status.
type Checkout struct {
	ID                    string           `json:"id"`
	CreatedAt             string           `json:"created_at"`
	UpdatedAt             string           `json:"updated_at"`
	Currency              string           `json:"currency"`
	ProductID             string           `json:"product_id"`
	OrderID               string           `json:"order_id"`
	WorkspaceID           string           `json:"workspace_id"`
	UnderpaidCoverPercent float64          `json:"underpaid_cover_percent"`
	Description           string           `json:"description"`
	Price                 string           `json:"price"`
	SupportedAssets       []SupportedAsset `json:"supported_assets"`
	CustomerID            string           `json:"customer_id"`
	ExpiryDurationSeconds int64            `json:"expiry_duration_seconds"`
	EnableLocalization    bool             `json:"enable_localization"`
	SuccessURL            string           `json:"success_url"`
	CancelURL             string           `json:"cancel_url"`
	Status                Status           `json:"status"`
	Outcome               json.RawMessage  `json:"outcome,omitempty"`
	FromAddress           string           `json:"from_address"`
}

// Payment contains checkout payment instructions and status data.
type Payment struct {
	ID                   string                      `json:"id"`
	CreatedAt            string                      `json:"created_at"`
	UpdatedAt            string                      `json:"updated_at"`
	CheckoutID           string                      `json:"checkout_id"`
	Checkout             *Checkout                   `json:"checkout,omitempty"`
	CustomerEmail        *string                     `json:"customer_email"`
	PayAssetID           string                      `json:"pay_asset_id"`
	PayAsset             *fystack.WalletAssetDetail  `json:"pay_asset,omitempty"`
	NetworkID            string                      `json:"network_id"`
	Network              *fystack.WalletAssetNetwork `json:"network,omitempty"`
	ExpiredAt            string                      `json:"expired_at"`
	PayerAddress         *string                     `json:"payer_address"`
	TxHash               *string                     `json:"tx_hash"`
	ConvertedCryptoPrice string                      `json:"converted_crypto_price"`
	DepositAddress       string                      `json:"deposit_address"`
	DepositQR            string                      `json:"deposit_qr"`
}

// Client is a Fystack checkout API client.
type Client struct {
	client *fystack.Client
	apiKey string
}

// New creates a checkout API client.
func New(apiKey string, options fystack.ClientOptions) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("API key is required")
	}
	options.Credentials = fystack.Credentials{}
	client, err := fystack.NewClient(options)
	if err != nil {
		return nil, err
	}
	return &Client{client: client, apiKey: apiKey}, nil
}

// CreateCheckout creates a checkout using the payment API key.
func (c *Client) CreateCheckout(ctx context.Context, params CreateCheckoutParams) (Checkout, error) {
	headers := make(http.Header)
	headers.Set("ACCESS-API-KEY", c.apiKey)
	var result Checkout
	err := c.client.Do(ctx, fystack.APIRequest{Method: http.MethodPost, Path: "/checkouts", Body: params, Header: headers}, &result)
	return result, err
}

// GetCheckout returns a public checkout.
func (c *Client) GetCheckout(ctx context.Context, checkoutID string) (Checkout, error) {
	var result Checkout
	err := c.client.Do(ctx, fystack.APIRequest{Method: http.MethodGet, Path: "/checkouts/" + url.PathEscape(checkoutID)}, &result)
	return result, err
}

// CreatePayment creates payment instructions for a checkout.
func (c *Client) CreatePayment(ctx context.Context, checkoutID string, params CreatePaymentParams) (Payment, error) {
	var result Payment
	err := c.client.Do(ctx, fystack.APIRequest{Method: http.MethodPost, Path: "/checkouts/" + url.PathEscape(checkoutID) + "/payment", Body: params}, &result)
	return result, err
}

// GetPayment returns a public checkout payment.
func (c *Client) GetPayment(ctx context.Context, paymentID string) (Payment, error) {
	var result Payment
	err := c.client.Do(ctx, fystack.APIRequest{Method: http.MethodGet, Path: "/checkouts/payment/" + url.PathEscape(paymentID)}, &result)
	return result, err
}
