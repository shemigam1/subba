package nomba

import "context"

// Charge, Transfer, and CreateVirtualAccount are stubbed for now — signatures
// only, matching the real Nomba request/response shapes from types.go. Every
// method calls getToken() first so callers never manage auth themselves.
// Real Nomba API calls land in Phase 2.

func (c *Client) Charge(ctx context.Context, req TokenizedCardChargeRequest) (*ChargeResponse, error) {
	if _, err := c.getToken(ctx); err != nil {
		return nil, err
	}
	// TODO: real Nomba tokenized card charge call
	return &ChargeResponse{}, nil
}

func (c *Client) Transfer(ctx context.Context, req BankTransferRequest) (*TransferResponse, error) {
	if _, err := c.getToken(ctx); err != nil {
		return nil, err
	}
	// TODO: real Nomba bank transfer call
	return &TransferResponse{}, nil
}

func (c *Client) CreateVirtualAccount(ctx context.Context, req CreateVirtualAccountRequest) (*VirtualAccountResponse, error) {
	if _, err := c.getToken(ctx); err != nil {
		return nil, err
	}
	// TODO: real Nomba createVirtualAccount call
	return &VirtualAccountResponse{}, nil
}
