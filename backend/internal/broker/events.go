package broker

const (
    ExchangeName = "subba.events"
    DlxExchangeName = "subba.dlx"

    RoutingKeyPaymentSucceeded  = "payment.succeeded"
    RoutingKeyPaymentFailed     = "payment.failed"
    RoutingKeyPaymentReversal   = "payment.reversal"
    RoutingKeySubscriptionRenew = "subscription.renew"

    QueueInvoicing         = "subba.invoicing"
    QueuePayouts           = "subba.payouts"
    QueueSubscriptionState = "subba.subscription_state"

    QueueInvoicingRetry         = "subba.invoicing.retry"
    QueuePayoutsRetry           = "subba.payouts.retry"
    QueueSubscriptionStateRetry = "subba.subscription_state.retry"

    QueueInvoicingDead         = "subba.invoicing.dead"
    QueuePayoutsDead           = "subba.payouts.dead"
    QueueSubscriptionStateDead = "subba.subscription_state.dead"
)

// NombaWebhookPayload is the raw shape Nomba POSTs to your webhook endpoint.
// Do not publish this to the broker — translate it into Event first.
type NombaWebhookPayload struct {
    EventType string      `json:"event_type"`
    RequestID string      `json:"requestId"`
    Data      NombaData   `json:"data"`
}

type NombaData struct {
    Merchant    NombaMerchant    `json:"merchant"`
    Transaction NombaTransaction `json:"transaction"`
    Customer    NombaCustomer    `json:"customer"`
}

type NombaMerchant struct {
    WalletID      string  `json:"walletId"`
    WalletBalance float64 `json:"walletBalance"`
    UserID        string  `json:"userId"`
}

type NombaTransaction struct {
    TransactionID         string  `json:"transactionId"`
    Type                  string  `json:"type"`
    TransactionAmount     float64 `json:"transactionAmount"`
    Fee                   float64 `json:"fee"`
    Time                  string  `json:"time"`
    ResponseCode          string  `json:"responseCode"`
    Narration             string  `json:"narration"`
    AliasAccountNumber    string  `json:"aliasAccountNumber"`  // virtual account number — use this to look up tenant+customer
    AliasAccountReference string  `json:"aliasAccountReference"`
    SessionID             string  `json:"sessionId"`
    OriginatingFrom       string  `json:"originatingFrom"`
}

type NombaCustomer struct {
    BankCode      string `json:"bankCode"`
    SenderName    string `json:"senderName"`
    BankName      string `json:"bankName"`
    AccountNumber string `json:"accountNumber"`
}

// Event is your internal envelope published to the broker.
// All consumers read this shape — never the raw Nomba payload.
type Event struct {
    RequestID      string `json:"requestId"`       // from NombaWebhookPayload.RequestID
    TenantID       string `json:"tenantId"`        // looked up from DB via virtual account
    EventType      string `json:"eventType"`       // canonical routing key, e.g. "payment.succeeded"
    Amount         int64  `json:"amount"`          // in kobo — convert from Nomba's float; 0 on subscription.renew
    Currency       string `json:"currency"`        // always "NGN" for now
    CustomerID     string `json:"customerId"`      // looked up from DB via virtual account
    SubscriptionID string `json:"subscriptionId"`  // looked up from DB via customer
    NombaReference string `json:"nombaReference,omitempty"` // NombaTransaction.TransactionID; absent on subscription.renew
    PlanID         string `json:"planId,omitempty"`         // plan that triggered subscription.renew; absent on payment.succeeded
}