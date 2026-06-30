Virtual Accounts
Knowledge checks · 0 / 1
0%
Virtual accounts are dedicated NUBAN accounts you can issue to any customer or invoice. When the customer transfers to that NUBAN from any Nigerian bank, you get a webhook with the amount, sender, and your reference. Perfect for invoicing, escrow, and bank-transfer checkout flows.

POST
/accounts/virtual
Create a permanent or one-time virtual account
GET
/accounts/virtual/{accountId}
Fetch virtual account details and balance
Node.js
Python
Copy
const va = await nomba.post("/accounts/virtual", {
accountRef: "inv_9921",
accountName: "Acme Ltd — INV 9921",
expiryDate: "2026-12-31",
amount: 1000000, // ₦10,000.00 — optional, locks expected amount
});
Handle over- and under-payment
Even when you set an expected amount, the bank rails will accept any value. Compare amountReceived to amountExpected in your webhook handler — refund overpayments and surface short-payments to the customer.
