Sub-accounts
Knowledge checks · 0 / 1
0%
Sub-accounts let you split a single Nomba merchant into many logical accounts — perfect for marketplaces, multi-tenant SaaS, or any product where funds must be tracked per seller, per branch, or per project. Each sub-account has its own balance and its own virtual accounts.

POST
/accounts/sub-accounts
Create a new sub-account
GET
/accounts/sub-accounts
List sub-accounts under your parent
GET
/accounts/sub-accounts/{id}/balance
Fetch the available balance of a sub-account
Node.js
Python
Copy
const sub = await nomba.post("/accounts/sub-accounts", {
accountName: "Seller — Adaeze Kitchen",
accountRef: "seller_adaeze_001",
});
Use stable refs
Pass your own accountRef so you can look up Nomba sub-accounts from your database without storing Nomba IDs as primary keys.
