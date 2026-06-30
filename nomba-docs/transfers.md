Transfers
Knowledge checks · 0 / 1
0%
Transfers move money out of your Nomba balance to any Nigerian bank account. Use them for payouts, refunds beyond the original card window, and treasury operations. Every transfer needs a verified recipient and a unique merchantTxRef.

POST
/transfers/bank/lookup
Resolve an account number to a name before sending
POST
/transfers/bank
Initiate a bank transfer
GET
/transfers/{merchantTxRef}
Check transfer status
Node.js
Python
Copy
const lookup = await nomba.post("/transfers/bank/lookup", {
bankCode: "044", // Access Bank
accountNumber: "0123456789",
});

await nomba.post("/transfers/bank", {
amount: 1500000, // ₦15,000.00
bankCode: "044",
accountNumber: "0123456789",
accountName: lookup.data.accountName,
senderName: "Acme Ltd",
narration: "Payout — March 2026",
merchantTxRef: "payout\_" + crypto.randomUUID(),
});
Always lookup before transfer
Sending to a wrong NUBAN can be irreversible. Display the resolved accountName to the user for confirmation before initiating the transfer.
