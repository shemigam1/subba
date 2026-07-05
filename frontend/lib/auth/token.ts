// Bearer-token storage. The frontend (Vercel) and API (CloudFront) are cross-site,
// so cookies can't be used for auth — the browser would drop them. Instead we keep
// the session token in localStorage and send it as an Authorization header.
//
// Two independent sessions: `tenant` (dashboard) and `portal` (customer). All access
// is SSR-safe (localStorage only exists in the browser).

const TENANT_KEY = "subba.tok.tenant";
const PORTAL_KEY = "subba.tok.portal";

function read(key: string): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(key);
}

function write(key: string, token: string): void {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(key, token);
}

function remove(key: string): void {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(key);
}

export const getTenantToken = (): string | null => read(TENANT_KEY);
export const setTenantToken = (token: string): void => write(TENANT_KEY, token);
export const clearTenantToken = (): void => remove(TENANT_KEY);

export const getPortalToken = (): string | null => read(PORTAL_KEY);
export const setPortalToken = (token: string): void => write(PORTAL_KEY, token);
export const clearPortalToken = (): void => remove(PORTAL_KEY);
