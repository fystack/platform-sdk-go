// Package awskms provides an AWS KMS-backed Fystack request signer.
package awskms

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	fystack "github.com/fystack/platform-sdk-go"
)

// Client is the subset of the AWS KMS client used by Signer.
type Client interface {
	Sign(context.Context, *kms.SignInput, ...func(*kms.Options)) (*kms.SignOutput, error)
}

// Options configures an AWS KMS-backed Ed25519 signer.
type Options struct {
	KeyID  string
	Client Client
}

// Signer signs requests without exporting private key material from KMS.
type Signer struct {
	keyID  string
	client Client
}

var _ fystack.RequestSigner = (*Signer)(nil)

// New creates a signer. When Client is nil, the AWS default configuration and
// credential provider chain are used.
func New(ctx context.Context, options Options) (*Signer, error) {
	if strings.TrimSpace(options.KeyID) == "" {
		return nil, errors.New("KMS key ID is required")
	}
	client := options.Client
	if client == nil {
		cfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("load AWS configuration: %w", err)
		}
		client = kms.NewFromConfig(cfg)
	}
	return &Signer{keyID: options.KeyID, client: client}, nil
}

// Sign signs a canonical request string with an ECC_NIST_EDWARDS25519 KMS key.
func (s *Signer) Sign(ctx context.Context, canonical string) (string, error) {
	output, err := s.client.Sign(ctx, &kms.SignInput{
		KeyId:            &s.keyID,
		Message:          []byte(canonical),
		MessageType:      types.MessageTypeRaw,
		SigningAlgorithm: types.SigningAlgorithmSpecEd25519Sha512,
	})
	if err != nil {
		return "", fmt.Errorf("sign with KMS key %q: %w", s.keyID, err)
	}
	if len(output.Signature) == 0 {
		return "", fmt.Errorf("KMS returned no signature for key %q", s.keyID)
	}
	return base64.StdEncoding.EncodeToString(output.Signature), nil
}
