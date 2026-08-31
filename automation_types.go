package fystack

// CreateSweepTaskParams configures a workspace sweep task.
type CreateSweepTaskParams struct {
	Name                 string
	Strategy             SweepStrategy
	MinTriggerValueUSD   float64
	DestinationWalletID  string
	FrequencyInSeconds   int64
	WalletIDs            []string
	SweepType            SweepType
	DestinationType      DestinationType
	AssetIDs             []string
	ReserveType          ReserveType
	ReserveAmountUSD     *float64
	ReservePercentageUSD *float64
	Enabled              *bool
}

// SweepTask contains a workspace sweep task.
type SweepTask struct {
	ID                   string  `json:"id"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
	Name                 string  `json:"name"`
	WorkspaceID          string  `json:"workspace_id"`
	Strategy             string  `json:"strategy"`
	MinTriggerValueUSD   string  `json:"min_trigger_value_usd"`
	DestinationWalletID  string  `json:"destination_wallet_id"`
	FrequencyInSeconds   int64   `json:"frequency_in_seconds"`
	GasTankWalletID      string  `json:"gas_tank_wallet_id"`
	LastExecutedAt       *string `json:"last_executed_at"`
	IsExecuting          bool    `json:"is_executing"`
	Enabled              bool    `json:"enabled"`
	CreatedByUserID      string  `json:"created_by_user_id"`
	SweepType            string  `json:"sweep_type"`
	DestinationType      string  `json:"destination_type"`
	DeviationRate        string  `json:"deviation_rate"`
	ReserveType          string  `json:"reserve_type,omitempty"`
	ReserveAmountUSD     string  `json:"reserve_amount_usd,omitempty"`
	ReservePercentageUSD string  `json:"reserve_percentage_usd,omitempty"`
}
