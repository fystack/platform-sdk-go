package fystack

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	testWalletID     = "11111111-1111-4111-8111-111111111111"
	testWithdrawalID = "22222222-2222-4222-8222-222222222222"
)

func TestClientSignsExactRequestBodyAndPath(t *testing.T) {
	t.Parallel()
	var failure error
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			failure = err
		}
		if got, want := request.URL.Path, "/api/v1/wallets/creation-status/"+testWalletID; got != want {
			failure = errors.New("unexpected request path: " + got)
		}
		canonical := CanonicalString(http.MethodGet, request.URL.Path, "1735689600", body)
		expected := base64.StdEncoding.EncodeToString([]byte(ComputeHMAC("secret", canonical)))
		if request.Header.Get("ACCESS-SIGN") != expected {
			failure = errors.New("unexpected request signature")
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"success":true,"data":{"wallet_id":"`+testWalletID+`","status":"success"}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server, Credentials{APIKey: "key", APISecret: "secret"})
	client.now = func() time.Time { return time.Unix(1735689600, 0) }
	result, err := client.GetWalletCreationStatus(context.Background(), testWalletID)
	if err != nil {
		t.Fatal(err)
	}
	if failure != nil {
		t.Fatal(failure)
	}
	if result.Status != WalletCreationStatusSuccess {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

func TestClientSignsBodyBytesThatItSends(t *testing.T) {
	t.Parallel()
	var failure error
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		canonical := CanonicalString(http.MethodPost, request.URL.Path, request.Header.Get("ACCESS-TIMESTAMP"), body)
		expected := hmacAccessSignature("secret", canonical)
		if request.Header.Get("ACCESS-SIGN") != expected {
			failure = errors.New("signature does not cover transmitted body")
		}
		if string(body) != `{"name":"wallet","wallet_type":"standard"}` {
			failure = errors.New("unexpected body: " + string(body))
		}
		_, _ = io.WriteString(response, `{"success":true,"data":{"wallet_id":"`+testWalletID+`","status":"success"}}`)
	}))
	defer server.Close()
	client := newTestClient(t, server, Credentials{APIKey: "key", APISecret: "secret"})
	_, err := client.CreateWalletAsync(context.Background(), CreateWalletParams{Name: "wallet"})
	if err != nil {
		t.Fatal(err)
	}
	if failure != nil {
		t.Fatal(failure)
	}
}

func TestClientUsesAuthTokenWithoutSigning(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer token" {
			t.Errorf("Authorization = %q", got)
		}
		if request.Header.Get("ACCESS-SIGN") != "" {
			t.Error("request should not be signed")
		}
		_, _ = io.WriteString(response, `{"success":true,"data":{"id":"`+testWithdrawalID+`","status":"PENDING"}}`)
	}))
	defer server.Close()
	client := newTestClient(t, server, Credentials{AuthToken: "Bearer token"})
	result, err := client.GetWithdrawalStatus(context.Background(), testWalletID, testWithdrawalID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != WithdrawalStatusPending {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(response, `{"success":false,"message":"not allowed"}`)
	}))
	defer server.Close()
	client := newTestClient(t, server, Credentials{})
	_, err := client.GetWalletCreationStatus(context.Background(), testWalletID)
	var apiError *APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusForbidden || apiError.Message != "not allowed" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestCreateWalletPollsToCompletion(t *testing.T) {
	t.Parallel()
	var statusCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if request.Method == http.MethodPost {
			_, _ = io.WriteString(response, `{"success":true,"data":{"wallet_id":"`+testWalletID+`","status":"pending"}}`)
			return
		}
		status := "pending"
		if statusCalls.Add(1) == 2 {
			status = "success"
		}
		_, _ = io.WriteString(response, `{"success":true,"data":{"wallet_id":"`+testWalletID+`","status":"`+status+`"}}`)
	}))
	defer server.Close()
	client := newTestClient(t, server, Credentials{})
	client.pollerOptions = PollerOptions{MaxAttempts: 3, Interval: time.Nanosecond, MaxInterval: time.Nanosecond, BackoffFactor: 1, Timeout: time.Second}
	result, err := client.CreateWallet(context.Background(), CreateWalletParams{Name: "wallet"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != WalletCreationStatusSuccess || statusCalls.Load() != 2 {
		t.Fatalf("result = %#v, calls = %d", result, statusCalls.Load())
	}
}

func TestRequestWithdrawalWireFormat(t *testing.T) {
	t.Parallel()
	skip := false
	priority := FeePriorityFast
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["recipient_address"] != "TAddress" || body["skip_balance_check"] != false || body["fee_priority"] != "fast" {
			t.Errorf("unexpected body: %#v", body)
		}
		if request.Header.Get("Idempotency-Key") != "withdrawal-1" {
			t.Errorf("idempotency key = %q", request.Header.Get("Idempotency-Key"))
		}
		_, _ = io.WriteString(response, `{"success":true,"data":{"auto_approved":false,"withdrawal":{"id":"`+testWithdrawalID+`"}}}`)
	}))
	defer server.Close()
	client := newTestClient(t, server, Credentials{})
	_, err := client.RequestWithdrawalWithOptions(context.Background(), testWalletID, RequestWithdrawalParams{
		Asset: "TRX", Network: "TRON_MAINNET", Amount: "1", RecipientAddress: "TAddress", SkipBalanceCheck: &skip, FeePriority: &priority,
	}, RequestOptions{IdempotencyKey: "withdrawal-1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTronAndBitcoinDepositAddresses(t *testing.T) {
	t.Parallel()
	assetID := "33333333-3333-4333-8333-333333333333"
	requests := make(chan *http.Request, 3)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests <- request.Clone(request.Context())
		_, _ = io.WriteString(response, `{"success":true,"data":{"address":"deposit-address"}}`)
	}))
	defer server.Close()
	client := newTestClient(t, server, Credentials{})
	if _, err := client.GetDepositAddress(context.Background(), testWalletID, AddressTypeTron); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDepositAddress(context.Background(), testWalletID, AddressTypeBitcoin); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDepositAddressByAssetID(context.Background(), testWalletID, assetID); err != nil {
		t.Fatal(err)
	}
	tronRequest := <-requests
	bitcoinTypeRequest := <-requests
	bitcoinAssetRequest := <-requests
	if got := tronRequest.URL.Query().Get("address_type"); got != "tron" {
		t.Fatalf("TRON address type = %q", got)
	}
	if got := bitcoinTypeRequest.URL.Query().Get("address_type"); got != "btc" {
		t.Fatalf("Bitcoin address type = %q", got)
	}
	if got := bitcoinAssetRequest.URL.Query().Get("asset_id"); got != assetID {
		t.Fatalf("Bitcoin asset ID = %q", got)
	}
}

func TestValidationDoesNotSendRequest(t *testing.T) {
	t.Parallel()
	client, err := NewClient(ClientOptions{BaseURL: "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetDepositAddress(context.Background(), "bad", AddressTypeEVM)
	if err == nil || !strings.Contains(err.Error(), "UUID v4") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newTestClient(t *testing.T, server *httptest.Server, credentials Credentials) *Client {
	t.Helper()
	client, err := NewClient(ClientOptions{
		Credentials: credentials,
		BaseURL:     server.URL + "/api/v1",
		HTTPClient:  server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
