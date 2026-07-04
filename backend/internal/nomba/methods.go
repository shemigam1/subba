package nomba

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Charge, Transfer, and CreateVirtualAccount implement the real Nomba API calls.
// Every method calls getToken() first so callers never manage auth themselves.

// Charge bills a saved tokenized card. merchantTxRef is used by Nomba as the
// idempotency key — send the same ref to safely retry without double-charging.
func (c *Client) Charge(ctx context.Context, req TokenizedCardChargeRequest) (*ChargeResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("nomba charge: get token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("nomba charge: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/tokenized-card/charge", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("nomba charge: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("accountId", c.accountID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("nomba charge: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nomba charge: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var out ChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("nomba charge: decode response: %w", err)
	}
	return &out, nil
}

// Transfer initiates a bank transfer from the merchant's Nomba balance.
// merchantTxRef serves as Nomba's idempotency key — use event.RequestID
// (prefixed to stay unique across payout types) so retries are safe.
func (c *Client) Transfer(ctx context.Context, req BankTransferRequest) (*TransferResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("nomba transfer: get token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("nomba transfer: marshal request: %w", err)
	}

	url := c.baseURL + "/transfers/bank"
	accountID := c.accountID
	if req.AccountID != "" {
		accountID = req.AccountID
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("nomba transfer: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("accountId", accountID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("nomba transfer: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nomba transfer: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var out TransferResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("nomba transfer: decode response: %w", err)
	}
	return &out, nil
}

// GetTransferStatus polls for the outcome of a previously initiated transfer.
// Use this when Transfer returns a non-final status so the payout handler can
// decide whether to retry or mark the record as failed.
func (c *Client) GetTransferStatus(ctx context.Context, merchantTxRef string) (*TransferResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("nomba get transfer status: get token: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v1/transfers/"+merchantTxRef, nil)
	if err != nil {
		return nil, fmt.Errorf("nomba get transfer status: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("accountId", c.accountID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("nomba get transfer status: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("nomba get transfer status: ref %q not found", merchantTxRef)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nomba get transfer status: status %d", resp.StatusCode)
	}

	var out TransferResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("nomba get transfer status: decode response: %w", err)
	}
	return &out, nil
}

// CreateVirtualAccount provisions a virtual bank account for a customer, scoped
// to subAccountID so that Nomba settles received funds directly into that
// sub-account's balance rather than the parent account.
// Endpoint: POST /v1/accounts/virtual/{subAccountId}
func (c *Client) CreateVirtualAccount(ctx context.Context, subAccountID string, req CreateVirtualAccountRequest) (*VirtualAccountResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("nomba create virtual account: get token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("nomba create virtual account: marshal request: %w", err)
	}

	url := c.baseURL + "/accounts/virtual"
	accountID := c.accountID
	if subAccountID != "" {
		accountID = subAccountID
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("nomba create virtual account: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("accountId", accountID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("nomba create virtual account: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nomba create virtual account: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var out VirtualAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("nomba create virtual account: decode response: %w", err)
	}
	return &out, nil
}

// BankLookup resolves an account number to an account name before a transfer.
func (c *Client) BankLookup(ctx context.Context, req BankLookupRequest) (*BankLookupResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("nomba bank lookup: get token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("nomba bank lookup: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/transfers/bank/lookup", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("nomba bank lookup: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("accountId", c.accountID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("nomba bank lookup: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nomba bank lookup: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var out BankLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("nomba bank lookup: decode response: %w", err)
	}
	return &out, nil
}
