/**
 * test_transfer.js
 * Verify Nomba BankLookup and Transfer endpoints
 */
const accountId = "f666ef9b-888e-4799-85ce-acb505b28023";
const subAccountId = "68b92816-31cc-4ccf-8151-0a110bc01723";
const clientId = "706df6c4-b8bb-4130-88c4-d21b052f8631";
const clientSecret = "k8UobYk3APgOoxUnNL7VpuxzwTsH4LsXtydfjcHs8RH0YISBB4OMqJsaafG+U8fWETu9YZ96bNXE+DelCDuMPw==";
const baseUrl = "https://sandbox.nomba.com";

async function run() {
  console.log("--- Fetching Token ---");
  const tokenRes = await fetch(`${baseUrl}/v1/auth/token/issue`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "accountId": accountId },
    body: JSON.stringify({ grant_type: "client_credentials", client_id: clientId, client_secret: clientSecret })
  });
  const tokenData = await tokenRes.json();
  const token = tokenData.data.access_token;

  console.log("\n--- Testing Bank Lookup ---");
  const lookupRes = await fetch(`${baseUrl}/v1/transfers/bank/lookup`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Authorization": `Bearer ${token}`, "accountId": accountId },
    body: JSON.stringify({ accountNumber: "0554772814", bankCode: "053" })
  });
  console.log("Lookup Status:", lookupRes.status);
  console.log("Lookup Body:", await lookupRes.text());

  console.log("\n--- Testing Sub-account Transfer ---");
  const transferRef = `txn-${Date.now()}`;
  const transferRes = await fetch(`${baseUrl}/v2/transfers/bank/${subAccountId}`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Authorization": `Bearer ${token}`, "accountId": accountId },
    body: JSON.stringify({
      amount: 100, // small amount
      accountNumber: "0554772814",
      accountName: "M.A Animashaun",
      bankCode: "053", // changed to match lookup
      merchantTxRef: transferRef,
      senderName: "Subba Testing",
      narration: "Hackathon Payout Test"
    })
  });
  
  const transferText = await transferRes.text();
  console.log("Transfer Status:", transferRes.status);
  console.log("Transfer Body:", transferText);
}

run().catch(err => console.error(err.message));
