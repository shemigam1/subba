# Deploying Subba to AWS

One EC2 instance runs the whole backend (API, worker, scheduler, Postgres,
Redis, RabbitMQ, Prometheus/Grafana) via `docker-compose.prod.yml`, with Caddy
terminating HTTPS. The Next.js frontend deploys to Vercel. Total cost:
~$15–30/month for the EC2 instance; everything else is free tier.

Everything referenced here already exists in the repo:

| Artifact | Purpose |
|---|---|
| `backend/Dockerfile` | builds `api`, `worker`, `scheduler` images (multi-stage, ~5–11 MB each) |
| `docker-compose.prod.yml` | the full production stack; only Caddy is exposed |
| `infra/caddy/Caddyfile` | HTTPS termination, auto Let's Encrypt |
| `infra/prometheus/prometheus.prod.yml` | scrape config for containerised services |
| `.env.production.example` | template for the server-side secrets file |

---

## 0. What you need before starting

- An **AWS account** (root or an IAM user able to create EC2 resources).
- A **domain** you control (e.g. from Namecheap/Cloudflare, ~$10/yr). You'll
  create two subdomains:
  - `api.<domain>` → the EC2 box (backend + webhooks)
  - `app.<domain>` → Vercel (frontend)

  > **Why a real domain, and why subdomains of the SAME domain:** Caddy needs a
  > domain for Let's Encrypt TLS, and the browser only sends our session
  > cookies (SameSite=Lax) when the frontend and API are on the *same site* —
  > `app.foo.com` → `api.foo.com` qualifies; `something.vercel.app` →
  > `api.foo.com` does **not** (logins would silently fail).
- The **live Nomba credentials** — and rotate the live private key before
  using it (it was shared in plaintext in chat).
- A **Resend API key**, with the sender domain's SPF/DKIM/DMARC verified.

## 1. Launch the EC2 instance (~10 min, AWS console)

1. EC2 → **Launch instance**.
2. Name `subba-prod`. AMI: **Ubuntu Server 24.04 LTS (64-bit x86)**.
3. Instance type: **t3.small** (2 vCPU / 2 GB — enough for this stack;
   `t3.micro` is free-tier but 1 GB is tight with RabbitMQ + Postgres).
4. Key pair: create one (`subba-prod-key`), download the `.pem`, then
   `chmod 400 ~/Downloads/subba-prod-key.pem`.
5. Network settings → security group `subba-prod-sg`, inbound rules:
   - SSH (22) — **My IP** only
   - HTTP (80) — Anywhere (Let's Encrypt + redirect)
   - HTTPS (443) — Anywhere

   Nothing else. Postgres/Redis/RabbitMQ stay inside Docker's network.
6. Storage: **30 GB gp3**.
7. Launch.

### Elastic IP (so the address never changes)

EC2 → Elastic IPs → **Allocate** → **Associate** with the instance.
Note the IP — call it `<ELASTIC_IP>`.

## 2. DNS (~2 min + propagation)

At your DNS provider add:

| Type | Name | Value |
|---|---|---|
| A | `api` | `<ELASTIC_IP>` |
| CNAME | `app` | `cname.vercel-dns.com` (Vercel shows the exact target in step 6) |

Verify before continuing (Caddy can't get certificates until this resolves):

```bash
dig +short api.<domain>    # must print <ELASTIC_IP>
```

## 3. Bootstrap the server (~5 min)

```bash
ssh -i ~/Downloads/subba-prod-key.pem ubuntu@<ELASTIC_IP>

# Docker (official convenience script) + compose plugin
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker ubuntu
exit   # re-login so the group applies
ssh -i ~/Downloads/subba-prod-key.pem ubuntu@<ELASTIC_IP>

git clone https://github.com/shemigam1/subba.git && cd subba
```

## 4. Secrets

```bash
cp .env.production.example .env.production
nano .env.production
```

Fill every blank (the template documents each one). Generate strong values:

```bash
openssl rand -hex 24      # POSTGRES_PASSWORD, SUBBA_APP_PASSWORD, RABBITMQ_PASSWORD
openssl rand -base64 32   # MASTER_ENCRYPTION_KEY
```

Set `API_DOMAIN=api.<domain>`, `PORTAL_BASE_URL=https://app.<domain>/pay`,
`DASHBOARD_BASE_URL=https://app.<domain>`, and the live Nomba + Resend keys.

## 5. First deploy

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
```

First run takes a few minutes (image builds + TLS issuance). Then:

```bash
# One-time: the migration creates subba_app with a default password — replace it
# with the value you put in .env.production:
docker compose --env-file .env.production -f docker-compose.prod.yml \
  exec postgres psql -U postgres -d subba \
  -c "ALTER ROLE subba_app PASSWORD '<your SUBBA_APP_PASSWORD>'"
docker compose --env-file .env.production -f docker-compose.prod.yml restart api worker scheduler
```

Smoke-test from your laptop:

```bash
curl https://api.<domain>/healthz   # {"status":"ok"} with a valid certificate
curl https://api.<domain>/readyz    # {"status":"ready"} — Postgres/Redis/RabbitMQ all up
```

If `readyz` fails, read logs: `docker compose --env-file .env.production -f docker-compose.prod.yml logs api --tail 50`.

## 6. Frontend on Vercel (~10 min)

1. [vercel.com](https://vercel.com) → **Add New Project** → import the GitHub
   repo. Set **Root Directory: `frontend`** (framework auto-detects Next.js).
2. Environment variable: `NEXT_PUBLIC_API_BASE_URL=https://api.<domain>/v1`.
3. Deploy, then Project → Settings → **Domains** → add `app.<domain>` and
   create the CNAME it shows (that's the row from step 2).
4. Visit `https://app.<domain>` — landing page; `/signup` — dashboard;
   `/pay/access?t=<tenant-id>` — portal.

## 7. Register the webhook with Nomba

In the Nomba dashboard set the webhook URL to:

```
https://api.<domain>/webhooks/nomba
```

and confirm the signing secret matches `NOMBA_WEBHOOK_SECRET` in
`.env.production`. Send a test event; watch it land:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml logs api -f | grep webhook
```

A forged call must bounce: `curl -X POST https://api.<domain>/webhooks/nomba -d '{}'` → 401.

## 8. Day-2 operations

```bash
alias prod='docker compose --env-file .env.production -f docker-compose.prod.yml'

prod ps                      # status of every service
prod logs api -f             # follow a service's logs (api|worker|scheduler|caddy…)
git pull && prod up -d --build   # redeploy after a merge
prod restart worker          # bounce one service
prod exec postgres pg_dump -U postgres subba > backup-$(date +%F).sql   # DB backup
```

- **Grafana** isn't public by default. Either SSH-tunnel
  (`ssh -i <pem> -L 3001:localhost:3000 ubuntu@<ELASTIC_IP>` after adding a
  `ports: ["127.0.0.1:3001:3000"]` mapping to the grafana service), or
  uncomment the Grafana block in `infra/caddy/Caddyfile` (keep the basic-auth
  line — the hackathon Grafana config has anonymous admin).
- **Data** lives in named Docker volumes (`pgdata` etc.) — `up`/`down` and
  redeploys don't touch it. `docker compose down -v` DOES delete it.
- If the instance restarts, everything comes back on its own
  (`restart: unless-stopped` + the Elastic IP keeps the address).

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| Browser warns about the certificate | DNS not propagated when Caddy started → `prod restart caddy` after `dig` shows the right IP |
| `readyz` 503 | a datastore is down — `prod ps`, then logs of the unhealthy service |
| Frontend loads but login does nothing | frontend not on `app.<domain>` (cookie SameSite — see step 0), or `DASHBOARD_BASE_URL`/`PORTAL_BASE_URL` wrong → CORS blocked (check browser console) |
| Webhooks 401 | signing secret mismatch with the Nomba dashboard |
| `migrate` service failed | `prod logs migrate` — migrations run as superuser before anything starts |
