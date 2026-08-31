package fystack

import "fmt"

// SignRequest contains a remote message-signing request.
type SignRequest struct {
	Method    string `json:"method"`
	Message   string `json:"message"`
	ChainID   int64  `json:"chain_id"`
	TypedData string `json:"typed_data,omitempty"`
}

// SignResponse identifies an asynchronous signing request.
type SignResponse struct {
	TransactionID string `json:"transaction_id"`
}

// ApprovalInfo describes one transaction approval.
type ApprovalInfo struct {
	UserID string                    `json:"user_id"`
	Status TransactionApprovalStatus `json:"status"`
}

// SignatureStatusResponse contains the state of a message-signing request.
type SignatureStatusResponse struct {
	Status        TransactionStatus `json:"status"`
	Signature     string            `json:"signature,omitempty"`
	TransactionID string            `json:"transaction_id"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
	Approvals     []ApprovalInfo    `json:"approvals"`
}

// TransactionStatusResponse contains the state of an on-chain transaction.
type TransactionStatusResponse struct {
	TransactionID string            `json:"transaction_id"`
	Status        TransactionStatus `json:"status"`
	Method        string            `json:"method"`
	Hash          string            `json:"hash,omitempty"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
	Approvals     []ApprovalInfo    `json:"approvals"`
	FailedReason  string            `json:"failed_reason,omitempty"`
}

// WalletDetail contains a signing wallet's address details.
type WalletDetail struct {
	APIKey      string      `json:"ap_key"`
	WalletID    string      `json:"wallet_id"`
	Name        string      `json:"name"`
	AddressType AddressType `json:"address_type"`
	Address     string      `json:"address"`
}

// TransactionError reports a terminal asynchronous transaction failure.
type TransactionError struct {
	Message       string
	Code          string
	TransactionID string
}

func (e *TransactionError) Error() string {
	if e.TransactionID == "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s (transaction %s)", e.Code, e.Message, e.TransactionID)
}
