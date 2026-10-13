// Package evm provides a Fystack-backed EVM transaction and message signer.
package evm

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	fystack "github.com/fystack/platform-sdk-go"
	"github.com/fystack/platform-sdk-go/internal/poller"
)

// Provider supplies chain data needed for EVM signing.
type Provider interface {
	ChainID(context.Context) (int64, error)
	PopulateTransaction(context.Context, Transaction) (Transaction, error)
}

// Transaction contains fields accepted by the remote signer. Integer
// quantities are decimal strings to avoid precision loss. Apex builds EIP-1559
// transactions only, so MaxFeePerGas and MaxPriorityFeePerGas are required and
// legacy gas prices are not supported.
type Transaction struct {
	MaxFeePerGas         *string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *string `json:"maxPriorityFeePerGas"`
	To                   *string `json:"to"`
	From                 string  `json:"from"`
	Nonce                uint64  `json:"nonce"`
	GasLimit             *string `json:"gasLimit"`
	Data                 string  `json:"data"`
	Value                string  `json:"value"`
	// ChainID is encoded as a quoted decimal string, which Apex requires.
	ChainID    int64 `json:"chainId,string"`
	AccessList any   `json:"accessList"`
}

// Option configures a Signer.
type Option func(*Signer)

// WithPollerOptions overrides asynchronous status polling defaults.
func WithPollerOptions(options fystack.PollerOptions) Option {
	return func(signer *Signer) {
		signer.pollerOptions = options
	}
}

// Signer signs EVM transactions and messages with a Fystack wallet.
type Signer struct {
	client        *fystack.Client
	provider      Provider
	walletID      string
	address       string
	pollerOptions poller.Options
}

// New creates an EVM signer backed by client.
func New(client *fystack.Client, provider Provider, options ...Option) (*Signer, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	signer := &Signer{client: client, provider: provider}
	for _, option := range options {
		option(signer)
	}
	return signer, nil
}

// SetWallet selects the wallet used by subsequent operations.
func (s *Signer) SetWallet(walletID string) error {
	if strings.TrimSpace(walletID) == "" {
		return errors.New("invalid wallet ID provided")
	}
	s.walletID = walletID
	s.address = ""
	return nil
}

// Address returns and caches the selected wallet's EVM address.
func (s *Signer) Address(ctx context.Context) (string, error) {
	if s.address != "" {
		return s.address, nil
	}
	detail, err := s.client.GetWalletDetail(ctx, s.walletID, fystack.AddressTypeEVM)
	if err != nil {
		return "", err
	}
	if detail.Address == "" {
		return "", errors.New("wallet address not found")
	}
	if s.walletID == "" {
		if detail.WalletID == "" {
			return "", errors.New("wallet ID not found")
		}
		s.walletID = detail.WalletID
	}
	s.address = detail.Address
	return s.address, nil
}

// SignTransaction submits an EVM transaction and waits for its on-chain hash.
func (s *Signer) SignTransaction(ctx context.Context, transaction Transaction, options fystack.RequestOptions) (string, error) {
	address, err := s.Address(ctx)
	if err != nil {
		return "", err
	}
	if transaction.From != "" && !strings.EqualFold(transaction.From, address) {
		return "", errors.New("transaction from address mismatch")
	}
	if transaction.ChainID <= 0 {
		return "", errors.New("transaction chain ID is required")
	}
	if transaction.MaxFeePerGas == nil || transaction.MaxPriorityFeePerGas == nil {
		return "", errors.New("maxFeePerGas and maxPriorityFeePerGas are required: legacy gas price transactions are not supported")
	}
	transaction.From = address
	response, err := s.client.SignTransaction(ctx, s.walletID, transaction, options)
	if err != nil {
		return "", err
	}
	return s.waitForTransaction(ctx, response.TransactionID)
}

// SendTransaction populates missing chain fields, submits a transaction, and
// waits for its on-chain hash.
func (s *Signer) SendTransaction(ctx context.Context, transaction Transaction, options fystack.RequestOptions) (string, error) {
	if s.provider == nil {
		return "", errors.New("provider is required for send transaction")
	}
	populated, err := s.provider.PopulateTransaction(ctx, transaction)
	if err != nil {
		return "", fmt.Errorf("populate transaction: %w", err)
	}
	return s.SignTransaction(ctx, populated, options)
}

// SignMessage signs a UTF-8 message using eth_sign.
func (s *Signer) SignMessage(ctx context.Context, message string, options fystack.RequestOptions) (string, error) {
	return s.signMessage(ctx, message, "eth_sign", "", options)
}

// SignMessageBytes signs bytes encoded as lowercase hexadecimal using eth_sign.
func (s *Signer) SignMessageBytes(ctx context.Context, message []byte, options fystack.RequestOptions) (string, error) {
	return s.signMessage(ctx, hex.EncodeToString(message), "eth_sign", "", options)
}

// SignTypedData signs EIP-712 data using eth_signTypedData_v4.
func (s *Signer) SignTypedData(ctx context.Context, domain, types, message any, options fystack.RequestOptions) (string, error) {
	typedData, err := json.Marshal(struct {
		Domain  any `json:"domain"`
		Types   any `json:"types"`
		Message any `json:"message"`
	}{domain, types, message})
	if err != nil {
		return "", fmt.Errorf("encode typed data: %w", err)
	}
	return s.signMessage(ctx, "", "eth_signTypedData_v4", string(typedData), options)
}

func (s *Signer) signMessage(ctx context.Context, message, method, typedData string, options fystack.RequestOptions) (string, error) {
	if s.provider == nil {
		return "", errors.New("provider is required for signing operations")
	}
	if _, err := s.Address(ctx); err != nil {
		return "", err
	}
	chainID, err := s.provider.ChainID(ctx)
	if err != nil {
		return "", fmt.Errorf("get chain ID from provider: %w", err)
	}
	response, err := s.client.RequestSign(ctx, s.walletID, fystack.SignRequest{
		Method: method, Message: message, ChainID: chainID, TypedData: typedData,
	}, options)
	if err != nil {
		return "", err
	}
	status, err := poller.Poll(ctx, s.pollerOptions,
		func(ctx context.Context) (fystack.SignatureStatusResponse, error) {
			return s.client.GetSignStatus(ctx, s.walletID, response.TransactionID)
		},
		func(status fystack.SignatureStatusResponse) bool {
			return status.Status == fystack.TransactionStatusConfirmed && status.Signature != ""
		},
		func(status fystack.SignatureStatusResponse) error {
			if status.Status == fystack.TransactionStatusFailed || status.Status == fystack.TransactionStatusRejected {
				return &fystack.TransactionError{Message: "signature request failed", Code: "SIGNATURE_FAILED", TransactionID: status.TransactionID}
			}
			return nil
		},
	)
	if err != nil {
		return "", err
	}
	return status.Signature, nil
}

func (s *Signer) waitForTransaction(ctx context.Context, transactionID string) (string, error) {
	status, err := poller.Poll(ctx, s.pollerOptions,
		func(ctx context.Context) (fystack.TransactionStatusResponse, error) {
			return s.client.GetTransactionStatus(ctx, s.walletID, transactionID)
		},
		func(status fystack.TransactionStatusResponse) bool {
			return (status.Status == fystack.TransactionStatusConfirmed || status.Status == fystack.TransactionStatusCompleted) && status.Hash != ""
		},
		func(status fystack.TransactionStatusResponse) error {
			switch status.Status {
			case fystack.TransactionStatusFailed:
				message := status.FailedReason
				if message == "" {
					message = "transaction failed"
				}
				return &fystack.TransactionError{Message: message, Code: "TRANSACTION_FAILED", TransactionID: status.TransactionID}
			case fystack.TransactionStatusRejected:
				return &fystack.TransactionError{Message: "transaction rejected", Code: "TRANSACTION_REJECTED", TransactionID: status.TransactionID}
			default:
				return nil
			}
		},
	)
	if err != nil {
		return "", err
	}
	if status.Hash == "" {
		return "", &fystack.TransactionError{Message: "transaction hash not found in successful response", Code: "TRANSACTION_HASH_MISSING", TransactionID: status.TransactionID}
	}
	return status.Hash, nil
}
