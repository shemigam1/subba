# How to Test the Subba Subscriptions Engine

Subba is an enterprise-grade, event-driven subscription layer. We chose to focus on deep infrastructure reliability (fault-tolerance, RabbitMQ fanout, Bearer token auth) rather than superficial UI features. 

**Prerequisites:**
You can test the application using our live hosted URLs:
- **Dashboard URL:** https://subba-theta.vercel.app
- **API URL:** https://asbestos-dale-serve-checks.trycloudflare.com/v1

**Step-by-Step Walkthrough:**

1. **Merchant Onboarding & Access:**
   - Visit the Dashboard URL and click **Login**. 
   - Use our pre-configured demo account to bypass signup:
     - **Email:** `demo@subba.com`
     - **Password:** `SubbaDemo2026!`
   - (Our system uses secure Bearer token authentication to handle cross-site domain restrictions smoothly).
   - Once logged in, navigate to **Customers** and click on the test customer.

2. **The "Cardless" Moat (Virtual Accounts):**
   - Click on your newly created customer to view their profile.
   - Note that our backend synchronously provisioned a **Nomba Virtual Account** specifically for this customer upon creation. Any funds transferred to this bank account will automatically trigger a webhook, credit the customer's wallet, and pay off their pending subscriptions!

3. **The Customer Portal (Hosted Checkout):**
   - On the customer profile, click **Generate Portal Link**.
   - Open this link in a new Incognito Window. This simulates the magic-link experience for the customer.
   - On the pending invoice, click **Pay with card**.
   - **Notice:** We have orchestrated `/v1/checkout/order` to redirect the user directly to **Nomba's Hosted Checkout Page**. This was a deliberate architectural decision to offload 3D Secure and OTP complexity entirely to Nomba, ensuring maximum conversion rates and minimizing PCI scope.
   
4. **Idempotency & Webhooks:**
   - Once a payment is completed (or a transfer is made to the Virtual Account), Nomba fires a webhook to our Go backend.
   - We perform standard **HMAC-SHA256 signature verification** on the raw HTTP body to ensure absolute security.
   - The webhook is then published to a RabbitMQ Topic Exchange, fanning out to isolated consumers to update the subscription state, ensuring exactly-once processing even during high-volume bursts.
