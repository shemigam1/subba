Build-Week Checklist
Knowledge checks · 1 / 1
100%
Run through this list before your submission. If any item is unchecked you will lose points — these are the same checks senior engineers use during a production launch.

Security
clientSecret and webhookSecret loaded from environment variables — not in source
All webhook handlers verify the nomba-signature HMAC
Idempotency: every external write keyed on a unique merchantTxRef
Correctness
All amounts converted to kobo before sending
Recipient name verified via /transfers/bank/lookup before transfers
Webhook handler is idempotent against duplicate requestId values
Over- and under-payment branches handled for virtual accounts
Operations
Nightly reconciliation job comparing /transactions to your ledger
Structured logging on every Nomba call with merchantTxRef tagged
Health-check endpoint your judges can hit to see green status
