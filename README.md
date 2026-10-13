# Fystack Platform SDK for Go

The Go SDK for Fystack wallet, payment, automation, and chain APIs. Wallet
addresses and withdrawals support EVM, Solana, TRON, and Bitcoin.

## Requirements

- Go 1.24 or newer, using the latest patch release

## Installation

```bash
go get github.com/fystack/platform-sdk-go
```

## Quick start

```go
package main

import (
	"context"
	"log"

	fystack "github.com/fystack/platform-sdk-go"
)

func main() {
	client, err := fystack.NewClient(fystack.ClientOptions{
		Credentials: fystack.Credentials{
			APIKey:    "YOUR_API_KEY",
			APISecret: "YOUR_API_SECRET",
		},
		WorkspaceID: "YOUR_WORKSPACE_ID",
		Environment: fystack.EnvironmentProduction,
	})
	if err != nil {
		log.Fatal(err)
	}

	wallet, err := client.CreateWallet(context.Background(), fystack.CreateWalletParams{
		Name:          "Treasury Wallet",
		WalletType:    fystack.WalletTypeHyper,
		WalletPurpose: fystack.WalletPurposeGeneral,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("wallet %s is %s", wallet.WalletID, wallet.Status)
}
```

Every network operation accepts `context.Context`. `CreateWallet` waits for an
asynchronous wallet to reach a terminal state; use `CreateWalletAsync` to
return its initial state immediately.

Set `CreateWalletParams.NetworkIDs` to provision only selected networks. This
applies equally to EVM, Solana, TRON, and Bitcoin network UUIDs.

## Authentication

HMAC-SHA256 is used when `APISecret` is set. For Ed25519, use a PKCS8 private
key:

```go
privateKeyPEM, err := os.ReadFile("private-key.pem")
if err != nil {
	log.Fatal(err)
}
signer, err := fystack.NewLocalPrivateKeySigner(privateKeyPEM)
if err != nil {
	log.Fatal(err)
}
client, err := fystack.NewClient(fystack.ClientOptions{
	Credentials: fystack.Credentials{APIKey: "YOUR_API_KEY", Signer: signer},
})
```

AWS KMS uses the default AWS configuration and credential provider chain:

```go
import "github.com/fystack/platform-sdk-go/signer/awskms"

signer, err := awskms.New(ctx, awskms.Options{
	KeyID: "alias/signer",
})
```

Pass a preconfigured `*kms.Client` through `awskms.Options.Client` for a custom
endpoint, role assumption, or client reuse. The KMS key must be an
`ECC_NIST_EDWARDS25519` `SIGN_VERIFY` key.

If neither `Signer` nor `APISecret` is configured, `AuthToken` is sent as the
`Authorization` header. See [API key signing](docs/api-key-signing.md) for the
wire protocol.

## Self-hosted deployments

```go
client, err := fystack.NewClient(fystack.ClientOptions{
	Credentials: credentials,
	Domain:      "api.your-company.com",
})
```

`Domain` takes precedence over `Environment` and resolves to
`https://<domain>/api/v1`. `BaseURL` is available for local testing and custom
HTTP gateways. Supply `HTTPClient` for custom timeouts, tracing, proxies, or
transport policy.

## Wallets and withdrawals

```go
address, err := client.GetDepositAddress(ctx, walletID, fystack.AddressTypeEVM)

withdrawal, err := client.RequestWithdrawal(ctx, walletID, fystack.RequestWithdrawalParams{
	Asset:            "ETH",
	Network:          "ETHEREUM_MAINNET",
	Amount:           "0.01",
	RecipientAddress: "0xRecipientAddress",
	Notes:            "Monthly payout",
})

latest, err := client.GetWithdrawalStatus(ctx, walletID, withdrawal.Withdrawal.ID)
```

Use `AssetID` instead of `Asset` and `Network` when the asset UUID is known.
Amounts remain strings so decimal values are never rounded.

### TRON and Bitcoin

TRON addresses can be selected directly by address type. For Bitcoin, prefer
the asset-specific method so Apex can select the correct mainnet or testnet
Bech32 address:

```go
tronAddress, err := client.GetDepositAddress(ctx, walletID, fystack.AddressTypeTron)
bitcoinAddress, err := client.GetDepositAddressByAssetID(ctx, walletID, bitcoinAssetID)
```

Withdrawals use the same API for every supported chain:

```go
priority := fystack.FeePriorityFast
withdrawal, err := client.RequestWithdrawalWithOptions(
	ctx,
	walletID,
	fystack.RequestWithdrawalParams{
		Asset:            "BTC",
		Network:          "BTC_MAINNET",
		Amount:           "0.001",
		RecipientAddress: bitcoinAddress.Address,
		FeePriority:      &priority,
	},
	fystack.RequestOptions{IdempotencyKey: "6f1c2a7e-3b9d-4c1a-9e2f-8a4b5c6d7e8f"},
)
```

Use `TRX` with `TRON_MAINNET` or the corresponding asset UUID for TRON.

`RequestOptions.IdempotencyKey` is sent as `X-Idempotency-Key`. Apex requires a
UUID and replays the first successful response for that key for 24 hours.

## Sweep automation

```go
task, err := client.CreateSweepTask(ctx, fystack.CreateSweepTaskParams{
	Name:                "Treasury sweep",
	Strategy:            fystack.SweepStrategyPeriodic,
	MinTriggerValueUSD:  50,
	DestinationWalletID: destinationWalletID,
	DestinationType:     fystack.DestinationTypeInternalWallet,
	FrequencyInSeconds:  3600,
	WalletIDs:           []string{sourceWalletID},
})
```

## Payments

```go
import "github.com/fystack/platform-sdk-go/payment"

payments, err := payment.New("YOUR_API_KEY", fystack.ClientOptions{
	Environment: fystack.EnvironmentProduction,
})
if err != nil {
	log.Fatal(err)
}

checkout, err := payments.CreateCheckout(ctx, payment.CreateCheckoutParams{
	Price:                 "66",
	Currency:              "USD",
	AcceptedAssets:        []payment.AcceptedAsset{{Asset: "ETH", Network: "BASE_MAINNET"}},
	DestinationWalletID:   walletID,
	ExpiryDurationSeconds: 3600,
})
```

## EVM signing

The `evm` signer accepts a small `Provider` interface. This keeps the SDK
compatible with any Ethereum RPC library while supporting transaction
population and chain ID lookup.

```go
import "github.com/fystack/platform-sdk-go/signer/evm"

signer, err := evm.New(client, provider)
if err != nil {
	log.Fatal(err)
}
if err := signer.SetWallet(walletID); err != nil {
	log.Fatal(err)
}

hash, err := signer.SendTransaction(ctx, transaction,
	fystack.RequestOptions{IdempotencyKey: "0d9e8f7a-6b5c-4d3e-8f1a-2b3c4d5e6f70"})
```

`evm.Transaction` integer quantities are decimal strings, matching the API and
avoiding precision loss. Apex builds EIP-1559 transactions only: set
`MaxFeePerGas` and `MaxPriorityFeePerGas`; legacy gas prices are not supported. `SignMessage`, `SignMessageBytes`, and `SignTypedData`
wait for the remote signature and return it directly.

## Solana signing

```go
import "github.com/fystack/platform-sdk-go/signer/solana"

signer, err := solana.New(client)
if err != nil {
	log.Fatal(err)
}
if err := signer.SetWallet(walletID); err != nil {
	log.Fatal(err)
}

signature, err := signer.SignTransaction(ctx, base64Transaction, fystack.RequestOptions{})
```

`SignTransaction`, `SignAndSendTransaction`, `SignMessage`, and
`SignAllTransactions` use base64 transactions and base58 signatures.
`solana.IncorporateSignature` supports legacy and versioned transactions.

The current Apex `/web3/transaction/{walletId}/signRaw` implementation rejects
Solana raw transactions. The Solana package is retained for compatibility with
deployments that enable that legacy SDK contract. Use wallet withdrawals for
current Apex deployments.

## Package layout

```text
platform-sdk-go
├── payment/          checkout client and payment models
├── signer/
│   ├── awskms/       optional AWS KMS request signer
│   ├── evm/          EVM transaction and message signer
│   └── solana/       legacy-compatible Solana signer
└── internal/poller/ shared bounded polling implementation
```

The root `fystack` package contains the HTTP client, authentication, wallet,
withdrawal, automation, and webhook APIs. Optional integrations are isolated
so applications only compile the integrations they import.

## Webhook verification

```go
valid, err := fystack.VerifyWebhookEd25519(publicKeyHex, event, signatureHex)
valid, err := fystack.VerifyWebhookHMAC(apiSecret, event, signatureHex)
```

Webhook JSON is canonicalized with RFC 8785 (JCS), the same encoding Apex signs.
HMAC comparison uses constant-time comparison. `GetWebhookPublicKey` returns the
hex-encoded key that `VerifyWebhookEd25519` expects.

## Development

```bash
gofmt -w .
go vet ./...
go test -race ./...
```
