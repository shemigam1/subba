package nomba

const (
	RawEventPaymentSuccess  = "payment_success"
	RawEventPayoutSuccess   = "payout_success"
	RawEventPaymentFailed   = "payment_failed"
	RawEventPaymentReversal = "payment_reversal"
	RawEventPayoutFailed    = "payout_failed"
	RawEventPayoutRefund    = "payout_refund"
)

type TokenIssueRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenResponse struct {
	Data TokenData `json:"data"`
}

type TokenData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type TokenizedCardChargeOrderSplit struct {
	AccountID string `json:"accountId"`
	Value     string `json:"value"`
}
type TokenizedCardChargeOrderSplitRequest struct {
	SplitType string                          `json:"splitType"`
	SplitList []TokenizedCardChargeOrderSplit `json:"splitList"`
}
type TokenizedCardChargeOrder struct {
	OrderReference string                                `json:"orderReference"`
	CustomerID     string                                `json:"customerId"`
	CallbackURL    string                                `json:"callbackUrl"`
	CustomerEmail  string                                `json:"customerEmail"`
	Amount         string                                `json:"amount"` // Expects a string decimal representation (e.g. "10000.00")
	Currency       string                                `json:"currency"`
	AccountID      string                                `json:"accountId"`
	SplitRequest   *TokenizedCardChargeOrderSplitRequest `json:"splitRequest,omitempty"`
}
type TokenizedCardChargeRequest struct {
	Order    TokenizedCardChargeOrder `json:"order"`
	TokenKey string                   `json:"tokenKey"`
}

type ChargeResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Data        struct {
		Status  any    `json:"status"` // The docs show "true" as a string or true as bool
		Message string `json:"message"`
	} `json:"data"`
}

type CreateCheckoutOrderRequest struct {
	Order        CheckoutOrder `json:"order"`
	TokenizeCard bool          `json:"tokenizeCard,omitempty"`
}

type CheckoutOrder struct {
	OrderReference        string                                `json:"orderReference,omitempty"`
	CustomerID            string                                `json:"customerId,omitempty"`
	CallbackURL           string                                `json:"callbackUrl,omitempty"`
	CustomerEmail         string                                `json:"customerEmail,omitempty"`
	Amount                string                                `json:"amount"` // String decimal e.g. "10000.00"
	Currency              string                                `json:"currency"`
	AccountID             string                                `json:"accountId,omitempty"`
	AllowedPaymentMethods []string                              `json:"allowedPaymentMethods,omitempty"`
	SplitRequest          *TokenizedCardChargeOrderSplitRequest `json:"splitRequest,omitempty"`
}

type CreateCheckoutOrderResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Data        struct {
		CheckoutLink   string `json:"checkoutLink"`
		OrderReference string `json:"orderReference"`
	} `json:"data"`
}

type BankLookupRequest struct {
	BankCode      string `json:"bankCode"`
	AccountNumber string `json:"accountNumber"`
}

type BankLookupResponse struct {
	Data BankLookupData `json:"data"`
}

type BankLookupData struct {
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	BankCode      string `json:"bankCode"`
	BankName      string `json:"bankName"`
}

type BankTransferRequest struct {
	AccountID     string `json:"accountId,omitempty"`
	Amount        int64  `json:"amount"`
	BankCode      string `json:"bankCode"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	SenderName    string `json:"senderName"`
	Narration     string `json:"narration"`
	MerchantTxRef string `json:"merchantTxRef"`
}

type TransferResponse struct {
	Data TransactionResult `json:"data"`
}

type TransactionResult struct {
	TransactionID string `json:"transactionId"`
	MerchantTxRef string `json:"merchantTxRef,omitempty"`
	Status        any    `json:"status"`
	ResponseCode  string `json:"responseCode,omitempty"`
	Message       string `json:"message,omitempty"`
}

type CreateVirtualAccountRequest struct {
	AccountRef     string  `json:"accountRef"`
	AccountName    string  `json:"accountName"`
	Currency       string  `json:"currency"`
	ExpiryDate     string  `json:"expiryDate,omitempty"`
	ExpectedAmount float64 `json:"expectedAmount,omitempty"`
}

type VirtualAccountResponse struct {
	Data VirtualAccount `json:"data"`
}

type VirtualAccount struct {
	AccountID      string `json:"accountId,omitempty"`
	AccountRef     string `json:"accountRef"`
	AccountName    string `json:"accountName"`
	AccountNumber  string `json:"bankAccountNumber"`
	BankName       string `json:"bankName"`
	BankCode       string `json:"bankCode,omitempty"`
	ExpiryDate     string `json:"expiryDate,omitempty"`
	AmountExpected int64  `json:"amount,omitempty"`
}

// type CreateSubAccountRequest struct {
// 	AccountName string `json:"accountName"`
// 	AccountRef  string `json:"accountRef"`
// }

// type SubAccountResponse struct {
// 	Data SubAccount `json:"data"`
// }

// type SubAccount struct {
// 	ID          string `json:"id"`
// 	AccountRef  string `json:"accountRef"`
// 	AccountName string `json:"accountName"`
// 	Status      string `json:"status,omitempty"`
// }

// type WebhookPayload struct {
// 	EventType string          `json:"event_type"`
// 	RequestID string          `json:"requestId"`
// 	Data      WebhookData     `json:"data"`
// 	Raw       json.RawMessage `json:"-"`
// }

// type WebhookData struct {
// 	Merchant    WebhookMerchant       `json:"merchant"`
// 	Terminal    json.RawMessage       `json:"terminal,omitempty"`
// 	Transaction WebhookTransaction    `json:"transaction"`
// 	Customer    WebhookCustomer       `json:"customer"`
// 	Extra       map[string]any        `json:"-"`
// }

// type WebhookMerchant struct {
// 	WalletID      string  `json:"walletId,omitempty"`
// 	WalletBalance float64 `json:"walletBalance,omitempty"`
// 	UserID        string  `json:"userId,omitempty"`
// }

// type WebhookTransaction struct {
// 	Fee                    float64 `json:"fee,omitempty"`
// 	SessionID              string  `json:"sessionId,omitempty"`
// 	Type                   string  `json:"type,omitempty"`
// 	TransactionID          string  `json:"transactionId,omitempty"`
// 	MerchantTxRef          string  `json:"merchantTxRef,omitempty"`
// 	ResponseCode           string  `json:"responseCode,omitempty"`
// 	ResponseCodeMessage    string  `json:"responseCodeMessage,omitempty"`
// 	OriginatingFrom        string  `json:"originatingFrom,omitempty"`
// 	TransactionAmount      float64 `json:"transactionAmount,omitempty"`
// 	Narration              string  `json:"narration,omitempty"`
// 	Time                   string  `json:"time,omitempty"`
// 	AliasAccountNumber     string  `json:"aliasAccountNumber,omitempty"`
// 	AliasAccountName       string  `json:"aliasAccountName,omitempty"`
// 	AliasAccountReference  string  `json:"aliasAccountReference,omitempty"`
// 	AliasAccountType       string  `json:"aliasAccountType,omitempty"`
// 	RRN                    string  `json:"rrn,omitempty"`
// 	CardIssuer             string  `json:"cardIssuer,omitempty"`
// 	CardBank               string  `json:"cardBank,omitempty"`
// 	TerminalSerialNumber   string  `json:"terminalSerialNumber,omitempty"`
// }

// type WebhookCustomer struct {
// 	BankCode      string `json:"bankCode,omitempty"`
// 	SenderName    string `json:"senderName,omitempty"`
// 	RecipientName string `json:"recipientName,omitempty"`
// 	BankName      string `json:"bankName,omitempty"`
// 	AccountNumber string `json:"accountNumber,omitempty"`
// 	ProductID     string `json:"productId,omitempty"`
// 	CardPAN       string `json:"cardPan,omitempty"`
// }

// type ErrorResponse struct {
// 	Message string          `json:"message,omitempty"`
// 	Code    string          `json:"code,omitempty"`
// 	Data    json.RawMessage `json:"data,omitempty"`
// }
