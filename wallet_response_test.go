package fystack

import (
	"encoding/json"
	"testing"
)

func TestDepositAddressResponseCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"apex", `{"asset_id":"asset","deposit_address":"deposit","qr_code":"qr"}`, true},
		{"alias", `{"asset_id":"asset","address":"deposit","qr_code":"qr"}`, true},
		{"matching", `{"asset_id":"asset","address":"deposit","deposit_address":"deposit","qr_code":"qr"}`, true},
		{"conflict", `{"asset_id":"asset","address":"wrong","deposit_address":"deposit"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result DepositAddressResponse
			err := json.Unmarshal([]byte(tc.body), &result)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
			if tc.valid && (result.Address != "deposit" || result.AssetID != "asset" || result.QRCode != "qr") {
				t.Fatalf("response=%+v", result)
			}
		})
	}
}
