// Package solana provides a Fystack-backed Solana transaction and message signer.
package solana

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/blocto/solana-go-sdk/types"
	fystack "github.com/fystack/platform-sdk-go"
	"github.com/fystack/platform-sdk-go/internal/poller"
	"github.com/mr-tron/base58"
)

const chainID = "1399811149"

// Option configures a Signer.
type Option func(*Signer)

// WithPollerOptions overrides asynchronous status polling defaults.
func WithPollerOptions(options fystack.PollerOptions) Option {
	return func(signer *Signer) {
		signer.pollerOptions = options
	}
}

// Signer signs Solana transactions and messages with a Fystack wallet.
type Signer struct {
	client        *fystack.Client
	walletID      string
	address       string
	pollerOptions poller.Options
}

// New creates a Solana signer backed by client.
func New(client *fystack.Client, options ...Option) (*Signer, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	signer := &Signer{client: client}
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

// Address returns and caches the selected wallet's Solana address.
func (s *Signer) Address(ctx context.Context) (string, error) {
	if s.address != "" {
		return s.address, nil
	}
	detail, err := s.client.GetWalletDetail(ctx, s.walletID, fystack.AddressTypeSolana)
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

// SignTransaction submits a base64-encoded serialized transaction and waits
// for its transaction signature.
func (s *Signer) SignTransaction(ctx context.Context, transaction string, options fystack.RequestOptions) (string, error) {
	address, err := s.Address(ctx)
	if err != nil {
		return "", err
	}
	body := struct {
		Data    string         `json:"data"`
		From    string         `json:"from"`
		Meta    map[string]any `json:"meta"`
		ChainID string         `json:"chainId"`
	}{transaction, address, map[string]any{"tx_method": "solana_signTransaction"}, chainID}
	response, err := s.client.SignTransaction(ctx, s.walletID, body, options)
	if err != nil {
		return "", err
	}
	return s.waitForTransaction(ctx, response.TransactionID)
}

// SignMessage signs a UTF-8 Solana message.
func (s *Signer) SignMessage(ctx context.Context, message string, options fystack.RequestOptions) (string, error) {
	return s.signMessage(ctx, message, options)
}

// SignMessageBytes signs bytes encoded as base64.
func (s *Signer) SignMessageBytes(ctx context.Context, message []byte, options fystack.RequestOptions) (string, error) {
	return s.signMessage(ctx, base64.StdEncoding.EncodeToString(message), options)
}

func (s *Signer) signMessage(ctx context.Context, message string, options fystack.RequestOptions) (string, error) {
	if _, err := s.Address(ctx); err != nil {
		return "", err
	}
	response, err := s.client.RequestSign(ctx, s.walletID, fystack.SignRequest{
		Method: "solana_signMessage", Message: message, ChainID: 0,
	}, options)
	if err != nil {
		return "", err
	}
	return s.waitForTransaction(ctx, response.TransactionID)
}

// SignAndSendTransaction signs and broadcasts a base64-encoded transaction.
func (s *Signer) SignAndSendTransaction(ctx context.Context, transaction string, options fystack.RequestOptions) (string, error) {
	address, err := s.Address(ctx)
	if err != nil {
		return "", err
	}
	body := struct {
		Transaction string `json:"transaction"`
		From        string `json:"from"`
		Method      string `json:"method"`
	}{transaction, address, "solana_signAndSendTransaction"}
	response, err := s.client.SignTransaction(ctx, s.walletID, body, options)
	if err != nil {
		return "", err
	}
	return s.waitForTransaction(ctx, response.TransactionID)
}

// SignAllTransactions signs and incorporates signatures into serialized transactions.
func (s *Signer) SignAllTransactions(ctx context.Context, transactions []string, options fystack.RequestOptions) ([]string, error) {
	result := make([]string, len(transactions))
	for index, transaction := range transactions {
		signature, err := s.SignTransaction(ctx, transaction, options)
		if err != nil {
			return nil, fmt.Errorf("sign transaction %d: %w", index, err)
		}
		result[index], err = IncorporateSignature(transaction, signature)
		if err != nil {
			return nil, fmt.Errorf("incorporate signature into transaction %d: %w", index, err)
		}
	}
	return result, nil
}

// IncorporateSignature adds a base58 signature to a base64 serialized legacy
// or versioned Solana transaction.
func IncorporateSignature(transactionBase64, signatureBase58 string) (string, error) {
	serialized, err := base64.StdEncoding.DecodeString(transactionBase64)
	if err != nil {
		return "", fmt.Errorf("decode transaction: %w", err)
	}
	transaction, err := types.TransactionDeserialize(serialized)
	if err != nil {
		return "", fmt.Errorf("deserialize transaction: %w", err)
	}
	signature, err := base58.Decode(signatureBase58)
	if err != nil {
		return "", fmt.Errorf("decode signature: %w", err)
	}
	if len(signature) != 64 {
		return "", fmt.Errorf("invalid Solana signature length %d", len(signature))
	}
	if err := transaction.AddSignature(signature); err != nil {
		return "", fmt.Errorf("add signature: %w", err)
	}
	output, err := transaction.Serialize()
	if err != nil {
		return "", fmt.Errorf("serialize transaction: %w", err)
	}
	return base64.StdEncoding.EncodeToString(output), nil
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
	return status.Hash, nil
}
