import createClient from "openapi-fetch";
import type { paths } from "./types";

// When we are ready to connect to the real backend, this will be the backend URL.
// But since we are intercepting with MSW, we can just point it to a dummy or local URL.
export const apiClient = createClient<paths>({
  baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080/v1",
});
