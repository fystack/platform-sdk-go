package fystack

// SweepTaskParams configures a sweep task attached during wallet creation.
type SweepTaskParams struct {
	MinTriggerValueUSD  float64         `json:"min_trigger_value_usd"`
	DestinationWalletID string          `json:"destination_wallet_id"`
	DestinationType     DestinationType `json:"destination_type"`
}

// CreateWalletParams configures a new wallet.
type CreateWalletParams struct {
	Name                  string
	WalletType            WalletType
	WalletPurpose         WalletPurpose
	SweepTaskParams       *SweepTaskParams
	SweepTaskID           string
	NetworkIDs            []string
	GasSponsorshipEnabled *bool
	ApprovalGroupID       string
}

// CreateWalletResponse contains a wallet provisioning state.
type CreateWalletResponse struct {
	WalletID string               `json:"wallet_id"`
	Status   WalletCreationStatus `json:"status"`
}

// WalletCreationStatusResponse contains the latest wallet provisioning state.
type WalletCreationStatusResponse = CreateWalletResponse

// WalletAssetNetwork describes the network associated with an asset.
type WalletAssetNetwork struct {
	ID                     string `json:"id"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
	Name                   string `json:"name"`
	Description            string `json:"description,omitempty"`
	IsEVM                  bool   `json:"is_evm"`
	ChainID                int64  `json:"chain_id"`
	NativeCurrency         string `json:"native_currency"`
	IsTestnet              bool   `json:"is_testnet,omitempty"`
	InternalCode           string `json:"internal_code"`
	ExplorerTx             string `json:"explorer_tx"`
	ExplorerAddress        string `json:"explorer_address"`
	ExplorerToken          string `json:"explorer_token"`
	ConfirmationBlocks     int64  `json:"confirmation_blocks"`
	BlockIntervalInSeconds int64  `json:"block_interval_in_seconds"`
	Disabled               bool   `json:"disabled"`
	LogoURL                string `json:"logo_url"`
}

// WalletAssetDetail describes an asset held by a wallet.
type WalletAssetDetail struct {
	ID            string              `json:"id"`
	CreatedAt     string              `json:"created_at"`
	UpdatedAt     string              `json:"updated_at"`
	Name          string              `json:"name"`
	Symbol        string              `json:"symbol"`
	Decimals      int                 `json:"decimals"`
	LogoURL       string              `json:"logo_url"`
	IsNative      bool                `json:"is_native"`
	AddressType   AddressType         `json:"address_type"`
	IsWhitelisted bool                `json:"is_whitelisted"`
	Address       string              `json:"address,omitempty"`
	NetworkID     string              `json:"network_id"`
	Network       *WalletAssetNetwork `json:"network,omitempty"`
}

// WalletAsset describes a wallet balance asset.
type WalletAsset struct {
	ID             string            `json:"id"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
	WalletID       string            `json:"wallet_id"`
	AssetID        string            `json:"asset_id"`
	DepositAddress string            `json:"deposit_address"`
	Hidden         bool              `json:"hidden"`
	Asset          WalletAssetDetail `json:"asset"`
}

// DepositAddressResponse contains a wallet deposit address and QR code.
type DepositAddressResponse struct {
	AssetID string `json:"asset_id,omitempty"`
	Address string `json:"address"`
	QRCode  string `json:"qr_code"`
}

// RescanTransactionParams identifies a transaction to rescan.
type RescanTransactionParams struct {
	TxHash    string
	NetworkID string
}

// TopAsset is a compact wallet asset summary.
type TopAsset struct {
	Symbol  string `json:"symbol"`
	LogoURL string `json:"logo_url"`
}

// Wallet describes a workspace wallet.
type Wallet struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Role          string     `json:"role"`
	WalletType    string     `json:"wallet_type"`
	ValueUSD      string     `json:"value_usd"`
	TopAssets     []TopAsset `json:"top_assets"`
	WalletPurpose string     `json:"wallet_purpose"`
}
