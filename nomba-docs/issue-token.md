Nomba uses OAuth 2.0 client_credentials for server-to-server calls. You exchange your clientId and clientSecret for a short-lived access token (1 hour) and attach it as a Bearer token on every subsequent request, plus the accountId header.

POST
/auth/token/issue
Issue an access token for your account
POST
/auth/token/refresh
Refresh an access token before it expires
Issuing a token
Node.js
Python
Copy
import fetch from "node-fetch";

const res = await fetch("https://api.nomba.com/v1/auth/token/issue", {
method: "POST",
headers: {
"Content-Type": "application/json",
"accountId": process.env.NOMBA_ACCOUNT_ID!,
},
body: JSON.stringify({
grant_type: "client_credentials",
client_id: process.env.NOMBA_CLIENT_ID,
client_secret: process.env.NOMBA_CLIENT_SECRET,
}),
});

const { data } = await res.json();
console.log("access_token:", data.access_token);
Cache your tokens
Tokens are valid for 60 minutes. Cache in memory or Redis and refresh at the 55-minute mark — do not request a fresh token per call.

Required headers on every authenticated call
Header Value
Authorization Bearer <access_token>
accountId Your Nomba account ID
Content-Type application/json
Knowledge check
Which OAuth 2.0 grant type does a server-side Nomba integration use?

password

authorization_code

client_credentials

implicit
