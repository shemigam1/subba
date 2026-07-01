import { http, HttpResponse } from 'msw'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/v1'

export const handlers = [
  http.get(`${API_BASE}/me`, () => {
    return HttpResponse.json({
      id: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
      name: "Acme Corp",
      email: "hello@acme.example",
      created_at: new Date().toISOString()
    })
  }),

  http.get(`${API_BASE}/plans`, () => {
    return HttpResponse.json([
      {
        id: "a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22",
        name: "Pro Plan",
        amount_minor: 500000,
        currency: "NGN",
        interval: "month",
        deleted_at: null,
        created_at: new Date().toISOString()
      }
    ])
  }),

  // Add more handlers as needed for development
]
