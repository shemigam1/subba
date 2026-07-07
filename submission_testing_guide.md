# Subba Testing Guide

This guide is split into two parts. **Part 1** is for you to test everything locally right now to make sure it's perfect. **Part 2** is the exact text you will copy/paste into your final hackathon submission form for the judges.

---

## Part 1: Your Final Local Test

Before submitting, you should run through the app exactly as the judges will. Follow these steps on your machine:

1. **Start the Backend**:
   Make sure Docker Desktop is running, then in your terminal run:
   ```bash
   cd backend
   docker compose up -d
   ```
   *(This starts Postgres, Redis, RabbitMQ, and the Go API server on port 8080).*

2. **Start the Frontend**:
   Open a new terminal window:
   ```bash
   cd frontend
   npm run dev
   ```

3. **Test the Merchant Dashboard flow**:
   - Go to `http://localhost:3000/signup` and create a test tenant.
   - You should be logged in automatically (testing our new Bearer token auth!).
   - Navigate to **Settings** and ensure your Nomba credentials (from your `.env`) are loaded.
   - Navigate to **Plans** and create a new Plan (e.g., "Premium Plan" for 5000 NGN).
   - Navigate to **Customers** and create a new Customer.

4. **Test the Customer Portal flow (The Checkout Link!)**:
   - On that customer's detail page, click **Generate Portal Link**.
   - Copy the link and open it in an **Incognito Window** (this proves our token authentication works perfectly across sessions).
   - Look at the pending invoice and click **Pay with card**.
   - You should be redirected to the **Nomba Hosted Checkout Page** (proving the orchestration works without needing an OTP form!).
   - Complete a test payment. 

Once the above works, you are 100% ready to submit!

---

## Part 2: Reviewer Testing Instructions (Copy & Paste for Submission)

*Copy the text below into the testing instructions section of your hackathon submission form.*

### How to Test the Subba Subscriptions Engine

Subba is an enterprise-grade, event-driven subscription layer. We chose to focus on deep infrastructure reliability (fault-tolerance, RabbitMQ fanout, Bearer token auth) rather than superficial UI features. 

**Prerequisites:**
You can test the application using our live hosted URLs:
- **Dashboard URL:** https://subba-theta.vercel.app
- **API URL:** https://asbestos-dale-serve-checks.trycloudflare.com

**Step-by-Step Walkthrough:**

1. **Merchant Onboarding:**
   - Visit the Dashboard URL and click **Sign Up**. 
   - Create a test account. Our system uses secure Bearer token authentication to handle cross-site domain restrictions smoothly.
   - Go to **Plans** and create a recurring plan (e.g., "Monthly Premium").
   - Go to **Customers** and add a test customer with your email address.

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
