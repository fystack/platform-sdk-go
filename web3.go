package fystack

import (
	"context"
	"net/http"
	"net/url"
)

// GetWalletDetail returns a wallet address for the requested address family.
func (c *Client) GetWalletDetail(ctx context.Context, walletID string, addressType AddressType) (WalletDetail, error) {
	endpoint := "/web3/wallet-detail"
	if walletID != "" {
		endpoint += "/" + url.PathEscape(walletID)
	}
	query := make(url.Values)
	query.Set("address_type", string(addressType))
	endpoint += "?" + query.Encode()
	var result WalletDetail
	err := c.do(ctx, http.MethodGet, endpoint, nil, nil, &result)
	return result, err
}

// RequestSign submits an asynchronous message-signing request.
func (c *Client) RequestSign(ctx context.Context, walletID string, params SignRequest, options RequestOptions) (SignResponse, error) {
	var result SignResponse
	err := c.do(ctx, http.MethodPost, "/web3/"+url.PathEscape(walletID)+"/sign", params, requestHeaders(options), &result)
	return result, err
}

// GetSignStatus returns the latest message-signing status.
func (c *Client) GetSignStatus(ctx context.Context, walletID, transactionID string) (SignatureStatusResponse, error) {
	var result SignatureStatusResponse
	err := c.do(ctx, http.MethodGet, "/web3/"+url.PathEscape(walletID)+"/sign/"+url.PathEscape(transactionID), nil, nil, &result)
	return result, err
}

// SignTransaction submits transaction data for asynchronous signing and broadcast.
func (c *Client) SignTransaction(ctx context.Context, walletID string, body any, options RequestOptions) (SignResponse, error) {
	var result SignResponse
	err := c.do(ctx, http.MethodPost, "/web3/transaction/"+url.PathEscape(walletID)+"/signRaw", body, requestHeaders(options), &result)
	return result, err
}

// GetTransactionStatus returns the latest on-chain transaction status.
func (c *Client) GetTransactionStatus(ctx context.Context, walletID, transactionID string) (TransactionStatusResponse, error) {
	var result TransactionStatusResponse
	err := c.do(ctx, http.MethodGet, "/web3/transaction/"+url.PathEscape(walletID)+"/"+url.PathEscape(transactionID), nil, nil, &result)
	return result, err
}
