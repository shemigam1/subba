Transactions & Reconciliation
Knowledge checks · 0 / 1
0%
Reconciliation is the daily discipline of matching what your app thinks happened against what Nomba records. Skipping reconciliation is the single most common reason fintech startups lose money silently. Pull the transactions endpoint nightly, diff against your local ledger, and alert on any drift.

GET
/transactions
List transactions with filters: dateFrom, dateTo, status, type
GET
/transactions/{merchantTxRef}
Look up a single transaction by your reference
Node.js
Python
Copy
const { data } = await nomba.get("/transactions", {
params: { dateFrom: "2026-03-01", dateTo: "2026-03-31", status: "success" },
});

for (const tx of data.transactions) {
const local = await db.payments.findOne({ ref: tx.merchantTxRef });
if (!local) await alertOps("Orphan transaction on Nomba", tx);
else if (local.amount !== tx.amount) await alertOps("Amount drift", { local, tx });
}
Reconcile by reference, not by ID
Your merchantTxRef is the source of truth. Use it to join Nomba's view with yours — Nomba's internal IDs may rotate during retries.
