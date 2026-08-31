package payment

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	fystack "github.com/fystack/platform-sdk-go"
)

func TestAuthentication(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("ACCESS-API-KEY") != "payment-key" {
			t.Errorf("API key = %q", request.Header.Get("ACCESS-API-KEY"))
		}
		if request.Header.Get("ACCESS-SIGN") != "" {
			t.Error("payment request should not contain an API signature")
		}
		_, _ = io.WriteString(response, `{"success":true,"data":{"id":"checkout-1","price":"10"}}`)
	}))
	defer server.Close()
	client, err := New("payment-key", fystack.ClientOptions{BaseURL: server.URL + "/api/v1", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	checkout, err := client.CreateCheckout(context.Background(), CreateCheckoutParams{
		Price: "10", Currency: "USD", DestinationWalletID: "11111111-1111-4111-8111-111111111111",
	})
	if err != nil {
		t.Fatal(err)
	}
	if checkout.ID != "checkout-1" {
		t.Fatalf("unexpected checkout: %#v", checkout)
	}
}
