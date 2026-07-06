/**
 * test_nomba.js
 * Quick script to verify Nomba Virtual Account API functionality.
 */
const crypto = require("crypto");

const accountId = "f666ef9b-888e-4799-85ce-acb505b28023";
const subAccountId = "68b92816-31cc-4ccf-8151-0a110bc01723";
const clientId = "706df6c4-b8bb-4130-88c4-d21b052f8631";
const clientSecret = "k8UobYk3APgOoxUnNL7VpuxzwTsH4LsXtydfjcHs8RH0YISBB4OMqJsaafG+U8fWETu9YZ96bNXE+DelCDuMPw==";
const baseUrl = "https://sandbox.nomba.com";

async function run() {
  console.log("--- Fetching Token ---");
  const tokenRes = await fetch(`${baseUrl}/v1/auth/token/issue`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "accountId": accountId
    },
    body: JSON.stringify({
      grant_type: "client_credentials",
      client_id: clientId,
      client_secret: clientSecret
    })
  });

  if (!tokenRes.ok) {
    throw new Error(`Token fetch failed: ${tokenRes.status} ${await tokenRes.text()}`);
  }

  const tokenData = await tokenRes.json();
  const token = tokenData.data.access_token;
  console.log("Token obtained successfully.");

  console.log("\n--- Creating Virtual Account ---");
  const accountRef = `test-ref-${Date.now()}`;
  
  // Notice we use the CORRECT endpoint: /v1/accounts/virtual/{subAccountId}
  const createRes = await fetch(`${baseUrl}/v1/accounts/virtual/${subAccountId}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${token}`,
      "accountId": accountId
    },
    body: JSON.stringify({
      accountRef: accountRef,
      accountName: "Agent Test Account",
      currency: "NGN",
      expectedAmount: 5000.00
    })
  });

  const responseText = await createRes.text();
  if (!createRes.ok && createRes.status !== 201) {
    throw new Error(`Virtual Account creation failed: ${createRes.status}\n${responseText}`);
  }

  const resData = JSON.parse(responseText);
  console.log("\nSUCCESS! Virtual Account Created.");
  console.log("Full Response:");
  console.log(JSON.stringify(resData, null, 2));

  // Verify the exact key that contains the NUBAN
  console.log("\n--- Verification ---");
  console.log("Is it 'accountNumber'? :", resData.data.accountNumber || "MISSING");
  console.log("Is it 'bankAccountNumber'? :", resData.data.bankAccountNumber || "MISSING");
}

run().catch(err => {
  console.error("FATAL ERROR:", err.message);
});
