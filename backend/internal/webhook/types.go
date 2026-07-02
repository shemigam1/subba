package webhook

// RawPayload is the shape of an incoming Nomba webhook body. Nomba's actual
// payloads vary slightly by transaction type (card checkout vs bank
// transfer vs payout), so fields are optional/omitempty except the ones we
// always rely on for routing and verification.
type RawPayload struct {
	EventType string  `json:"event_type"`
	RequestID string  `json:"requestId"`
	Data      RawData `json:"data"`
}

type RawData struct {
	Merchant    RawMerchant    `json:"merchant"`
	Transaction RawTransaction `json:"transaction"`
}

type RawMerchant struct {
	WalletID      string  `json:"walletId"`
	WalletBalance float64 `json:"walletBalance,omitempty"`
	UserID        string  `json:"userId"`
}

type RawTransaction struct {
	TransactionID string  `json:"transactionId"`
	MerchantTxRef string  `json:"merchantTxRef,omitempty"`
	Type          string  `json:"type"`
	Amount        float64 `json:"transactionAmount,omitempty"`
	Fee           float64 `json:"fee,omitempty"`
	Time          string  `json:"time"`
	ResponseCode  string  `json:"responseCode,omitempty"`

	// AliasAccountNumber is the virtual account number that received the
	// funds — this is what we use to look up which tenant/customer this
	// event belongs to.
	AliasAccountNumber string `json:"aliasAccountNumber,omitempty"`
}
