/**
 * test_requery.js
 */
const accountId = "f666ef9b-888e-4799-85ce-acb505b28023";
const clientId = "706df6c4-b8bb-4130-88c4-d21b052f8631";
const clientSecret = "k8UobYk3APgOoxUnNL7VpuxzwTsH4LsXtydfjcHs8RH0YISBB4OMqJsaafG+U8fWETu9YZ96bNXE+DelCDuMPw==";
const baseUrl = "https://sandbox.nomba.com";

const sessionId = "260325515668563586728";
const merchantTxRef = "txn-1783338942046";

async function run() {
  const tokenRes = await fetch(`${baseUrl}/v1/auth/token/issue`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "accountId": accountId },
    body: JSON.stringify({ grant_type: "client_credentials", client_id: clientId, client_secret: clientSecret })
  });
  const token = (await tokenRes.json()).data.access_token;

  console.log("--- Testing Requery by SessionId ---");
  const sRes = await fetch(`${baseUrl}/v1/transactions/requery/${sessionId}`, {
    method: "GET",
    headers: { "Authorization": `Bearer ${token}`, "accountId": accountId }
  });
  console.log("Status:", sRes.status);
  console.log("Body:", await sRes.text());

  console.log("\n--- Testing Requery by merchantTxRef (Our existing logic) ---");
  const mRes = await fetch(`${baseUrl}/v1/transfers/${merchantTxRef}`, {
    method: "GET",
    headers: { "Authorization": `Bearer ${token}`, "accountId": accountId }
  });
  console.log("Status:", mRes.status);
  console.log("Body:", await mRes.text());
}

run().catch(console.error);
