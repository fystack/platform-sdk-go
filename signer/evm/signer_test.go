package evm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	fystack "github.com/fystack/platform-sdk-go"
)

const walletID = "11111111-1111-4111-8111-111111111111"

type fakeProvider struct{}

func (fakeProvider) ChainID(context.Context) (int64, error) { return 1, nil }
func (fakeProvider) PopulateTransaction(_ context.Context, transaction Transaction) (Transaction, error) {
	return transaction, nil
}

func TestSignMessage(t *testing.T) {
	t.Parallel()
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		switch {
		case request.URL.Path == "/api/v1/web3/wallet-detail/"+walletID:
			_, _ = io.WriteString(response, `{"success":true,"data":{"wallet_id":"`+walletID+`","address":"0xabc"}}`)
		case request.Method == http.MethodPost:
			_, _ = io.WriteString(response, `{"success":true,"data":{"transaction_id":"tx-1"}}`)
		default:
			_, _ = io.WriteString(response, `{"success":true,"data":{"transaction_id":"tx-1","status":"confirmed","signature":"0xsigned"}}`)
		}
	}))
	defer server.Close()
	client, err := fystack.NewClient(fystack.ClientOptions{BaseURL: server.URL + "/api/v1", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	signer, err := New(client, fakeProvider{}, WithPollerOptions(fystack.PollerOptions{
		MaxAttempts: 1, Interval: time.Nanosecond, MaxInterval: time.Nanosecond, BackoffFactor: 1, Timeout: time.Second,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := signer.SetWallet(walletID); err != nil {
		t.Fatal(err)
	}
	signature, err := signer.SignMessage(context.Background(), "hello", fystack.RequestOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if signature != "0xsigned" || requests != 3 {
		t.Fatalf("signature = %q, requests = %d", signature, requests)
	}
}
