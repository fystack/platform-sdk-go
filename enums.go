package fystack

// Environment identifies a Fystack API deployment.
type Environment string

const (
	EnvironmentLocal      Environment = "local"
	EnvironmentSandbox    Environment = "sandbox"
	EnvironmentProduction Environment = "production"
)

// WalletType identifies the wallet key-management model.
type WalletType string

const (
	WalletTypeHyper WalletType = "standard"
	WalletTypeMPC   WalletType = "mpc"
)

// WalletPurpose identifies the intended use of a wallet.
type WalletPurpose string

const (
	WalletPurposeGeneral    WalletPurpose = "general"
	WalletPurposeGasTank    WalletPurpose = "gas_tank"
	WalletPurposeDeployment WalletPurpose = "deployment"
	WalletPurposeCustody    WalletPurpose = "custody"
	WalletPurposeUser       WalletPurpose = "user"
	WalletPurposePayment    WalletPurpose = "payment"
	WalletPurposeOneTimeUse WalletPurpose = "one_time_use"
)

// WalletCreationStatus is the provisioning state of a wallet.
type WalletCreationStatus string

const (
	WalletCreationStatusPending WalletCreationStatus = "pending"
	WalletCreationStatusSuccess WalletCreationStatus = "success"
	WalletCreationStatusError   WalletCreationStatus = "error"
)

// AddressType identifies a blockchain address family.
type AddressType string

const (
	AddressTypeEVM     AddressType = "evm"
	AddressTypeBitcoin AddressType = "btc"
	AddressTypeSolana  AddressType = "sol"
	AddressTypeTron    AddressType = "tron"
)

// AddressStandard identifies a chain-specific address encoding.
type AddressStandard string

const (
	AddressStandardBitcoinP2PKH         AddressStandard = "btc_p2pkh"
	AddressStandardBitcoinP2SH          AddressStandard = "btc_p2sh"
	AddressStandardBitcoinBech32        AddressStandard = "btc_bech32"
	AddressStandardBitcoinP2PKHTestnet  AddressStandard = "btc_p2pkh_testnet"
	AddressStandardBitcoinP2SHTestnet   AddressStandard = "btc_p2sh_testnet"
	AddressStandardBitcoinBech32Testnet AddressStandard = "btc_bech32_testnet"
)

// BitcoinBech32Standard returns the recommended Bitcoin address standard for a network.
func BitcoinBech32Standard(testnet bool) AddressStandard {
	if testnet {
		return AddressStandardBitcoinBech32Testnet
	}
	return AddressStandardBitcoinBech32
}

// FeePriority controls the confirmation-time target for a withdrawal fee.
type FeePriority string

const (
	FeePriorityLow      FeePriority = "low"
	FeePriorityStandard FeePriority = "standard"
	FeePriorityFast     FeePriority = "fast"
)

// DestinationType identifies a sweep destination.
type DestinationType string

const (
	DestinationTypeInternalWallet DestinationType = "internal_wallet"
	DestinationTypeAddressBook    DestinationType = "address_book"
)

// TransactionStatus is the state of a signing or transaction request.
type TransactionStatus string

const (
	TransactionStatusPending         TransactionStatus = "pending"
	TransactionStatusBroadcasting    TransactionStatus = "broadcasting"
	TransactionStatusCompleted       TransactionStatus = "completed"
	TransactionStatusConfirmed       TransactionStatus = "confirmed"
	TransactionStatusFailed          TransactionStatus = "failed"
	TransactionStatusPendingApproval TransactionStatus = "pending_approval"
	TransactionStatusRejected        TransactionStatus = "rejected"
)

// TransactionApprovalStatus is the state of an approval request.
type TransactionApprovalStatus string

const (
	TransactionApprovalStatusPending  TransactionApprovalStatus = "pending"
	TransactionApprovalStatusApproved TransactionApprovalStatus = "approved"
	TransactionApprovalStatusRejected TransactionApprovalStatus = "rejected"
)

// WalletRole identifies a user's role on a wallet.
type WalletRole string

const (
	WalletRoleAdmin  WalletRole = "wallet_admin"
	WalletRoleSigner WalletRole = "wallet_signer"
	WalletRoleViewer WalletRole = "wallet_viewer"
)

// WithdrawalStatus is the state of a withdrawal.
type WithdrawalStatus string

const (
	WithdrawalStatusPendingApproval WithdrawalStatus = "PENDING_APPROVAL"
	WithdrawalStatusCancelled       WithdrawalStatus = "CANCELLED"
	WithdrawalStatusPending         WithdrawalStatus = "PENDING"
	WithdrawalStatusFailed          WithdrawalStatus = "FAILED"
	WithdrawalStatusExecuted        WithdrawalStatus = "EXECUTED"
	WithdrawalStatusSuccess         WithdrawalStatus = "SUCCESS"
	WithdrawalStatusFundingGas      WithdrawalStatus = "FUNDING_GAS"
	WithdrawalStatusDeniedByPolicy  WithdrawalStatus = "DENIED_BY_POLICY"
)

// SweepStrategy identifies how a sweep task is scheduled.
type SweepStrategy string

const SweepStrategyPeriodic SweepStrategy = "periodic"

// SweepType identifies the source workflow for a sweep.
type SweepType string

const (
	SweepTypeDefault     SweepType = "default"
	SweepTypePaymentLink SweepType = "payment_link"
	SweepTypeCheckout    SweepType = "checkout"
)

// ReserveType identifies how funds are reserved during a sweep.
type ReserveType string

const ReserveTypeFixedUSD ReserveType = "fixed_usd"
