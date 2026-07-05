import createClient from "openapi-fetch";
import type { paths } from "./v1";
import { getPortalToken, getTenantToken } from "@/lib/auth/token";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080/v1";

// Bearer-token auth (not cookies): the frontend and API are cross-site
// (Vercel <-> CloudFront), so a session cookie would be dropped by the browser.
// We attach an Authorization header instead. Portal routes use the customer
// (portal) session token; everything else uses the tenant (dashboard) token.
export const api = createClient<paths>({ baseUrl: API_BASE_URL });

api.use({
  onRequest({ request, schemaPath }) {
    const token = schemaPath.startsWith("/portal") ? getPortalToken() : getTenantToken();
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }
    return request;
  },
});
