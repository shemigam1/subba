import { http, HttpResponse } from 'msw'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/v1'

let plans = [
  {
    id: "a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22",
    name: "Pro Plan",
    amount_minor: 500000,
    currency: "NGN",
    interval: "month",
    deleted_at: null,
    created_at: new Date().toISOString()
  }
];

let customers = [
  {
    id: "111e4567-e89b-12d3-a456-426614174000",
    name: "John Doe",
    email: "john@example.com",
    has_card_on_file: true,
    created_at: new Date().toISOString(),
  },
  {
    id: "222e4567-e89b-12d3-a456-426614174000",
    name: "Jane Smith",
    email: "jane@example.com",
    has_card_on_file: false,
    created_at: new Date().toISOString(),
  },
];

let analyticsOverview = {
  mrr: { amount_minor: 25000000, currency: "NGN" },
  active_subscriptions: 142,
  payments_today: 18,
  failed_payments: 2,
  dlq_depth: 0,
  revenue_series: [
    { date: "2026-06-25", amount_minor: 500000 },
    { date: "2026-06-26", amount_minor: 1500000 },
    { date: "2026-06-27", amount_minor: 800000 },
    { date: "2026-06-28", amount_minor: 2000000 },
    { date: "2026-06-29", amount_minor: 1200000 },
    { date: "2026-06-30", amount_minor: 3500000 },
  ],
};

export const handlers = [
  http.get(`${API_BASE}/me`, () => {
    return HttpResponse.json({
      id: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
      name: "Acme Corp",
      email: "hello@acme.example",
      created_at: new Date().toISOString()
    })
  }),

  // Analytics
  http.get(`${API_BASE}/analytics/overview`, () => {
    return HttpResponse.json(analyticsOverview);
  }),

  // Plans
  http.get(`${API_BASE}/plans`, () => {
    return HttpResponse.json(plans);
  }),

  http.post(`${API_BASE}/plans`, async ({ request }) => {
    const data = await request.json() as any;
    const newPlan = {
      id: crypto.randomUUID(),
      name: data.name,
      amount_minor: data.amount_minor,
      currency: data.currency || "NGN",
      interval: data.interval,
      created_at: new Date().toISOString(),
      deleted_at: null,
    };
    plans = [newPlan, ...plans];
    return HttpResponse.json(newPlan, { status: 201 });
  }),

  // Customers
  http.get(`${API_BASE}/customers`, () => {
    return HttpResponse.json({ data: customers, next_cursor: null });
  }),

  http.get(`${API_BASE}/customers/:id`, ({ params }) => {
    const customer = customers.find(c => c.id === params.id) || customers[0];
    return HttpResponse.json(customer);
  }),

  // Portal Mocks
  http.get(`${API_BASE}/portal/invoices`, () => {
    return HttpResponse.json([
      {
        id: crypto.randomUUID(),
        amount_minor: 1500000,
        currency: "NGN",
        status: "paid",
        issued_at: new Date().toISOString(),
      },
      {
        id: crypto.randomUUID(),
        amount_minor: 500000,
        currency: "NGN",
        status: "open",
        issued_at: new Date().toISOString(),
      },
    ]);
  }),

  http.get(`${API_BASE}/portal/subscription`, () => {
    return HttpResponse.json({
      id: crypto.randomUUID(),
      customer_id: "111e4567-e89b-12d3-a456-426614174000",
      plan: plans[0],
      status: "active",
      current_period_start: new Date().toISOString(),
      current_period_end: "2026-06-30T00:00:00.000Z",
      cancel_at_period_end: false,
      canceled_at: null,
      created_at: new Date().toISOString()
    });
  }),

  http.post(`${API_BASE}/portal/subscription/cancel`, () => {
    return HttpResponse.json({
      id: crypto.randomUUID(),
      status: "canceled",
      cancel_at_period_end: true,
      canceled_at: new Date().toISOString(),
    });
  }),
];
