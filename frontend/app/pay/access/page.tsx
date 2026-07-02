import { AccessClient } from './access-client'

// Magic-link landing: /pay/access?token=... (from the email) exchanges straight
// into a session; /pay/access?t=<tenant_id> shows the email request form.
export default async function AccessPage({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}) {
  const params = await searchParams
  const token = typeof params.token === 'string' ? params.token : undefined
  const tenantId = typeof params.t === 'string' ? params.t : undefined
  return <AccessClient token={token} tenantId={tenantId} />
}
