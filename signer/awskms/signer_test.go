package awskms

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

type fakeClient struct {
	input *kms.SignInput
}

func (f *fakeClient) Sign(_ context.Context, input *kms.SignInput, _ ...func(*kms.Options)) (*kms.SignOutput, error) {
	f.input = input
	return &kms.SignOutput{Signature: []byte("signature")}, nil
}

func TestSigner(t *testing.T) {
	t.Parallel()
	fake := &fakeClient{}
	signer, err := New(context.Background(), Options{KeyID: "alias/signer", Client: fake})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := signer.Sign(context.Background(), "canonical")
	if err != nil {
		t.Fatal(err)
	}
	if encoded != base64.StdEncoding.EncodeToString([]byte("signature")) {
		t.Fatalf("unexpected signature: %s", encoded)
	}
	if fake.input.MessageType != types.MessageTypeRaw || string(fake.input.SigningAlgorithm) != "ED25519_SHA_512" {
		t.Fatalf("unexpected KMS input: %#v", fake.input)
	}
}
