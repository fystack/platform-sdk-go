package fystack

// RequestWithdrawalParams configures a withdrawal. Set AssetID or both Asset and Network.
type RequestWithdrawalParams struct {
	AssetID          string
	Network          string
	Asset            string
	Amount           string
	RecipientAddress string
	Notes            string
	SkipBalanceCheck *bool
	CategoryIDs      []string
	FeePriority      *FeePriority
}

// WithdrawalApproval describes an approval attached to a withdrawal.
type WithdrawalApproval struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// WithdrawalTransaction describes the transaction executing a withdrawal.
type WithdrawalTransaction struct {
	ID        string `json:"id"`
	Hash      string `json:"hash,omitempty"`
	TxHash    string `json:"tx_hash,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// TransactionCategory describes a withdrawal category.
type TransactionCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Withdrawal contains the current withdrawal details.
type Withdrawal struct {
	ID                  string                 `json:"id"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
	Amount              string                 `json:"amount"`
	Status              WithdrawalStatus       `json:"status"`
	RecipientAddress    string                 `json:"recipient_address"`
	Notes               string                 `json:"notes,omitempty"`
	WithdrawalApprovals []WithdrawalApproval   `json:"withdrawal_approvals"`
	CreatorID           string                 `json:"creator_id"`
	AssetID             string                 `json:"asset_id"`
	Asset               *WalletAssetDetail     `json:"asset,omitempty"`
	WalletID            string                 `json:"wallet_id"`
	TransactionID       string                 `json:"transaction_id,omitempty"`
	Transaction         *WithdrawalTransaction `json:"transaction,omitempty"`
	AssetHoldID         string                 `json:"asset_hold_id"`
	ErrorReason         string                 `json:"error_reason,omitempty"`
	Categories          []TransactionCategory  `json:"categories,omitempty"`
}

// RequestWithdrawalResponse contains the submitted withdrawal and approval mode.
type RequestWithdrawalResponse struct {
	AutoApproved             bool       `json:"auto_approved"`
	ApprovalBypassedByPolicy bool       `json:"approval_bypassed_by_policy,omitempty"`
	Withdrawal               Withdrawal `json:"withdrawal"`
}
