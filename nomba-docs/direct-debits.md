Direct Debits (Mandates)
Knowledge checks · 0 / 1
0%
A mandate is the customer's standing authorisation to debit their bank account on a recurring or on-demand basis. Use mandates for lending, BNPL, or any service that needs to pull funds without the customer initiating each charge. Mandates require explicit customer consent via OTP or in-app approval.

POST
/mandates/create
Create a mandate request — returns a consent URL
POST
/mandates/{mandateId}/debit
Debit a previously approved mandate
DELETE
/mandates/{mandateId}
Cancel an active mandate
Node.js
Python
Copy
const mandate = await nomba.post("/mandates/create", {
customerId: "cus_8821",
maxAmount: 5000000, // ₦50,000 ceiling per debit
frequency: "monthly",
startDate: "2026-04-01",
endDate: "2027-04-01",
});
// redirect customer to mandate.data.consentUrl
Respect the ceiling
Attempting to debit more than maxAmount will fail. If your billing exceeds the ceiling, create a new mandate — do not split debits to bypass it.
