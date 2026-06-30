import createClient from "openapi-fetch";
import type { paths } from "./v1";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080/v1";

// We create a singleton client instance to be used across the app.
export const api = createClient<paths>({ baseUrl: API_BASE_URL });
