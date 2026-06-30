Tokenized Cards & Recurring Payments
Knowledge checks · 0 / 1
0%
After a successful checkout, Nomba returns a card token representing the customer's card. You can charge that token later — for subscriptions, top-ups, or one-click re-orders — without the customer re-entering details. Tokens are scoped to your merchant and cannot be used elsewhere.

POST
/tokenized-card/charge
Charge a previously saved card token
GET
/tokenized-card/list
List saved tokens for a customer
DELETE
/tokenized-card/{tokenId}
Revoke a stored card token
Node.js
Python
Copy
await nomba.post("/tokenized-card/charge", {
amount: 500000, // ₦5,000.00
currency: "NGN",
cardId: "tok*5fa12b...",
customerId: "cus_8821",
merchantTxRef: "sub_2026_03*" + customerId,
});
Subscriptions are your job
Nomba does not run the schedule — you do. Store the token, run a cron, and charge on your billing cycle. Always send a unique merchantTxRef per attempt to make retries idempotent.
