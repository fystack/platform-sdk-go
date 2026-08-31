package solana

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/mr-tron/base58"
)

func TestIncorporateSignature(t *testing.T) {
	t.Parallel()
	account, err := types.AccountFromSeed(make([]byte, ed25519.SeedSize))
	if err != nil {
		t.Fatal(err)
	}
	message := types.Message{
		Version:         types.MessageVersionLegacy,
		Header:          types.MessageHeader{NumRequireSignatures: 1},
		Accounts:        []common.PublicKey{account.PublicKey},
		RecentBlockHash: base58.Encode(make([]byte, 32)),
	}
	transaction, err := types.NewTransaction(types.NewTransactionParam{Message: message})
	if err != nil {
		t.Fatal(err)
	}
	serialized, err := transaction.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	messageBytes, err := message.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	signature := account.Sign(messageBytes)
	encoded, err := IncorporateSignature(base64.StdEncoding.EncodeToString(serialized), base58.Encode(signature))
	if err != nil {
		t.Fatal(err)
	}
	output, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := types.TransactionDeserialize(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(parsed.Signatures[0]) != string(signature) {
		t.Fatal("signature was not incorporated")
	}
}
