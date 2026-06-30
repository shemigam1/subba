Webhooks
Knowledge checks · 0 / 3
0%
Webhooks are how Nomba tells your server that something happened — a payment succeeded, a virtual account was funded, a transfer completed. Every webhook is signed with HMAC-SHA256 using your webhook secret. You must verify the signature before trusting the payload.

Verifying signatures
Node.js
Python
Copy
import crypto from "crypto";

app.post("/webhooks/nomba", express.raw({ type: "application/json" }), (req, res) => {
const signature = req.header("nomba-signature");
const expected = crypto
.createHmac("sha256", process.env.NOMBA_WEBHOOK_SECRET!)
.update(req.body)
.digest("hex");

if (signature !== expected) return res.status(401).send("bad signature");

const event = JSON.parse(req.body.toString());
// Idempotency: ignore if we have already processed event.requestId
res.sendStatus(200);
});
Webhooks may fire twice
Network retries can deliver the same event multiple times. Store event.requestId in a unique index and reject duplicates — never apply a balance change twice.

Common event types
Event type Fires when
payment_success A checkout or token charge completes
virtual_account.funded A NUBAN you issued receives a transfer
transfer.success An outbound transfer settles to the recipient
transfer.failed An outbound transfer is reversed
mandate.debit_success A direct debit attempt clears
