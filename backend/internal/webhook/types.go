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
	Customer    RawCustomer    `json:"customer"`
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
	Narration     string  `json:"narration,omitempty"`
	SessionID     string  `json:"sessionId,omitempty"`
	OriginatingFrom string `json:"originatingFrom,omitempty"`

	// AliasAccountNumber is the virtual account number that received the
	// funds.
	AliasAccountNumber string `json:"aliasAccountNumber,omitempty"`

	// AliasAccountName is the display name associated with the virtual account.
	AliasAccountName string `json:"aliasAccountName,omitempty"`

	// AliasAccountReference is the custom tag we provided when creating the Virtual Account.
	// We format this as "{tenantID}:{customerID}" to instantly route payments.
	AliasAccountReference string `json:"aliasAccountReference,omitempty"`

	// AliasAccountType distinguishes VIRTUAL from other account types.
	AliasAccountType string `json:"aliasAccountType,omitempty"`
}

// RawCustomer represents the sender/payer details from a Nomba webhook.
type RawCustomer struct {
	BankCode      string `json:"bankCode,omitempty"`
	SenderName    string `json:"senderName,omitempty"`
	BankName      string `json:"bankName,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
}
